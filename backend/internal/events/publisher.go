package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	DomainExchange = "eventory.domain.v1"
)

// Publisher defines an asynchronous event publisher to the messaging broker.
type Publisher interface {
	Publish(ctx context.Context, routingKey string, event any) error
	Close() error
}

// NoopPublisher is a stub publisher that discards published events.
type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, string, any) error { return nil }
func (NoopPublisher) Close() error                              { return nil }

// RabbitMQPublisher publishes domain events to RabbitMQ with publisher confirms
// and OpenTelemetry trace context propagation.
type RabbitMQPublisher struct {
	mu       sync.Mutex
	conn     *amqp.Connection
	ch       *amqp.Channel
	confirms <-chan amqp.Confirmation
	returns  <-chan amqp.Return
	exchange string
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq publisher: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open rabbitmq publisher channel: %w", err)
	}

	if err := ch.ExchangeDeclare(DomainExchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare domain exchange: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("enable publisher confirms: %w", err)
	}

	pub := &RabbitMQPublisher{
		conn:     conn,
		ch:       ch,
		confirms: ch.NotifyPublish(make(chan amqp.Confirmation, 1)),
		returns:  ch.NotifyReturn(make(chan amqp.Return, 1)),
		exchange: DomainExchange,
	}

	return pub, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, routingKey string, event any) (err error) {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event for publish: %w", err)
	}

	ctx, span := otel.Tracer("events.publisher").Start(ctx, "publish "+p.exchange,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination.name", p.exchange),
			attribute.String("messaging.rabbitmq.destination.routing_key", routingKey),
		))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	headers := amqp.Table{"x-attempt": int32(0)}
	otel.GetTextMapPropagator().Inject(ctx, tableCarrier(headers))

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ch == nil || p.ch.IsClosed() {
		return fmt.Errorf("rabbitmq publisher channel is closed")
	}

	if err := p.ch.PublishWithContext(ctx, p.exchange, routingKey, true, false, amqp.Publishing{
		Headers:      headers,
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	}); err != nil {
		return err
	}

	select {
	case returned := <-p.returns:
		return fmt.Errorf("rabbitmq returned unroutable message: %s", returned.ReplyText)
	case confirmation := <-p.confirms:
		if !confirmation.Ack {
			return fmt.Errorf("rabbitmq negatively acknowledged publish")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *RabbitMQPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var errs []error
	if p.ch != nil && !p.ch.IsClosed() {
		if err := p.ch.Close(); err != nil {
			errs = append(errs, err)
		}
		p.ch = nil
	}
	if p.conn != nil && !p.conn.IsClosed() {
		if err := p.conn.Close(); err != nil {
			errs = append(errs, err)
		}
		p.conn = nil
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

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
