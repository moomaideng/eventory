package rabbitmq

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	DomainExchange  = "eventory.domain.v1"
	RetryExchange   = "eventory.notification.retry.v1"
	DeadExchange    = "eventory.notification.dead.v1"
	RegistrationKey = "registration.submitted"
	MainQueue       = "notification.email.registration-submitted.v1"
	DeadQueue       = "notification.email.dlq.v1"
)

type Broker struct {
	mu          sync.RWMutex
	publishMu   sync.Mutex
	url         string
	prefetch    int
	conn        *amqp.Connection
	consumeCh   *amqp.Channel
	publishCh   *amqp.Channel
	confirms    <-chan amqp.Confirmation
	returns     <-chan amqp.Return
	retryDelays []time.Duration
}

func Connect(url string, retryDelays []time.Duration, prefetch int) (*Broker, error) {
	broker := &Broker{url: url, prefetch: prefetch, retryDelays: retryDelays}
	if err := broker.Reconnect(); err != nil {
		return nil, err
	}
	return broker, nil
}

func (broker *Broker) Reconnect() error {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	broker.closeLocked()

	url := broker.url
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	consumeCh, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	publishCh, err := conn.Channel()
	if err != nil {
		_ = consumeCh.Close()
		_ = conn.Close()
		return err
	}
	broker.conn = conn
	broker.consumeCh = consumeCh
	broker.publishCh = publishCh
	if err := broker.declareTopology(broker.prefetch); err != nil {
		broker.closeLocked()
		return err
	}
	return nil
}

func (broker *Broker) declareTopology(prefetch int) error {
	for _, exchange := range []struct{ name, kind string }{{DomainExchange, "topic"}, {RetryExchange, "direct"}, {DeadExchange, "direct"}} {
		if err := broker.consumeCh.ExchangeDeclare(exchange.name, exchange.kind, true, false, false, false, nil); err != nil {
			return err
		}
	}
	if _, err := broker.consumeCh.QueueDeclare(MainQueue, true, false, false, false, nil); err != nil {
		return err
	}
	if err := broker.consumeCh.QueueBind(MainQueue, RegistrationKey, DomainExchange, false, nil); err != nil {
		return err
	}
	for index, delay := range broker.retryDelays {
		queueName := retryQueueName(index + 1)
		args := amqp.Table{
			"x-message-ttl":             delay.Milliseconds(),
			"x-dead-letter-exchange":    DomainExchange,
			"x-dead-letter-routing-key": RegistrationKey,
		}
		if _, err := broker.consumeCh.QueueDeclare(queueName, true, false, false, false, args); err != nil {
			return err
		}
		if err := broker.consumeCh.QueueBind(queueName, retryRoutingKey(index+1), RetryExchange, false, nil); err != nil {
			return err
		}
	}
	if _, err := broker.consumeCh.QueueDeclare(DeadQueue, true, false, false, false, nil); err != nil {
		return err
	}
	if err := broker.consumeCh.QueueBind(DeadQueue, RegistrationKey, DeadExchange, false, nil); err != nil {
		return err
	}
	if err := broker.consumeCh.Qos(prefetch, 0, false); err != nil {
		return err
	}
	if err := broker.publishCh.Confirm(false); err != nil {
		return err
	}
	broker.confirms = broker.publishCh.NotifyPublish(make(chan amqp.Confirmation, 1))
	broker.returns = broker.publishCh.NotifyReturn(make(chan amqp.Return, 1))
	return nil
}

func (broker *Broker) Deliveries() (<-chan amqp.Delivery, error) {
	broker.mu.RLock()
	defer broker.mu.RUnlock()
	if broker.consumeCh == nil || broker.consumeCh.IsClosed() {
		return nil, fmt.Errorf("rabbitmq consumer channel is closed")
	}
	return broker.consumeCh.Consume(MainQueue, "notification-service", false, false, false, false, nil)
}

func (broker *Broker) PublishRegistration(ctx context.Context, body []byte) error {
	return broker.publish(ctx, DomainExchange, RegistrationKey, body, amqp.Table{"x-attempt": int32(0)})
}

func (broker *Broker) PublishRetry(ctx context.Context, body []byte, attempt int) error {
	return broker.publish(ctx, RetryExchange, retryRoutingKey(attempt), body, amqp.Table{"x-attempt": int32(attempt)})
}

func (broker *Broker) PublishDead(ctx context.Context, body []byte, attempt int, reason string) error {
	return broker.publish(ctx, DeadExchange, RegistrationKey, body, amqp.Table{"x-attempt": int32(attempt), "x-failure-reason": reason})
}

func (broker *Broker) publish(ctx context.Context, exchange, key string, body []byte, headers amqp.Table) (err error) {
	ctx, span := otel.Tracer("notification.rabbitmq").Start(ctx, "publish "+exchange,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination.name", exchange),
			attribute.String("messaging.rabbitmq.destination.routing_key", key),
		))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	// A confirm belongs to the next publish on this channel. Serialize publishes
	// so a future worker pool cannot consume another goroutine's confirmation.
	broker.publishMu.Lock()
	defer broker.publishMu.Unlock()
	broker.mu.RLock()
	defer broker.mu.RUnlock()
	if broker.publishCh == nil || broker.publishCh.IsClosed() {
		return fmt.Errorf("rabbitmq publisher channel is closed")
	}

	// Inject after starting the span so the consumer's span is a child of this publish.
	otel.GetTextMapPropagator().Inject(ctx, tableCarrier(headers))

	if err := broker.publishCh.PublishWithContext(ctx, exchange, key, true, false, amqp.Publishing{
		Headers: headers, ContentType: "application/json", DeliveryMode: amqp.Persistent,
		Timestamp: time.Now().UTC(), Body: body,
	}); err != nil {
		return err
	}
	select {
	case returned := <-broker.returns:
		return fmt.Errorf("rabbitmq returned unroutable message: %s", returned.ReplyText)
	case confirmation := <-broker.confirms:
		if !confirmation.Ack {
			return fmt.Errorf("rabbitmq negatively acknowledged publish")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Attempt(headers amqp.Table) int {
	value, ok := headers["x-attempt"]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case string:
		attempt, _ := strconv.Atoi(typed)
		return attempt
	default:
		return 0
	}
}

func (broker *Broker) Ready() bool {
	broker.mu.RLock()
	defer broker.mu.RUnlock()
	return broker.conn != nil && !broker.conn.IsClosed() && broker.consumeCh != nil && !broker.consumeCh.IsClosed()
}

func (broker *Broker) Close() error {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	return broker.closeLocked()
}

func (broker *Broker) closeLocked() error {
	if broker.consumeCh != nil {
		_ = broker.consumeCh.Close()
		broker.consumeCh = nil
	}
	if broker.publishCh != nil {
		_ = broker.publishCh.Close()
		broker.publishCh = nil
	}
	if broker.conn != nil && !broker.conn.IsClosed() {
		err := broker.conn.Close()
		broker.conn = nil
		return err
	}
	broker.conn = nil
	return nil
}

func retryQueueName(attempt int) string {
	return fmt.Sprintf("notification.email.retry.%d.v1", attempt)
}
func retryRoutingKey(attempt int) string { return fmt.Sprintf("retry.%d", attempt) }

// tableCarrier lets the OTel propagator read and write trace headers on AMQP
// message headers, so consumers continue the publisher's trace.
type tableCarrier amqp.Table

func (c tableCarrier) Get(key string) string {
	value, _ := c[key].(string)
	return value
}

func (c tableCarrier) Set(key, value string) { c[key] = value }

func (c tableCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for key := range c {
		keys = append(keys, key)
	}
	return keys
}

// ExtractTrace returns ctx carrying the trace context stored in headers.
func ExtractTrace(ctx context.Context, headers amqp.Table) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, tableCarrier(headers))
}
