package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ErrPermanent marks a failure that retrying cannot fix, such as an event that
// contradicts the consumer's own record. The event goes straight to the dead
// queue.
var ErrPermanent = errors.New("permanent event failure")

const (
	attemptHeader = "x-attempt"
	// routingKeyHeader keeps the key an event was published with. A retried
	// event returns to the queue under the queue's name, so the delivery's own
	// routing key no longer says what the event is.
	routingKeyHeader = "x-routing-key"
)

// Handler processes one event. It must be idempotent: delivery is at least once.
// A returned error retries the event after the next RetryDelays entry; once
// those run out, or on ErrPermanent, it moves to the dead queue.
type Handler func(ctx context.Context, routingKey string, body []byte) error

// Consumer reads domain events from a durable queue bound to DomainExchange.
// Instances sharing a Queue compete for deliveries, so each event is handled
// by one of them.
//
// Besides Queue it owns "<Queue>.retry.<n>" queues, whose message TTL is
// RetryDelays[n-1] and which dead-letter back into Queue, and "<Queue>.dead",
// where events that exhausted their retries wait for an operator.
type Consumer struct {
	URL            string
	Queue          string
	RoutingKeys    []string
	RetryDelays    []time.Duration
	Prefetch       int
	ReconnectDelay time.Duration
}

func (c Consumer) retryQueue(n int) string { return fmt.Sprintf("%s.retry.%d", c.Queue, n) }
func (c Consumer) deadQueue() string       { return c.Queue + ".dead" }

// Run consumes until ctx ends. A broker outage delays delivery: Run keeps
// reconnecting rather than returning.
func (c Consumer) Run(ctx context.Context, handle Handler) {
	for {
		err := c.consume(ctx, handle)
		if ctx.Err() != nil {
			return
		}
		slog.WarnContext(ctx, "rabbitmq consumer disconnected", "queue", c.Queue, "error", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(c.ReconnectDelay):
		}
	}
}

func (c Consumer) consume(ctx context.Context, handle Handler) error {
	conn, err := amqp.Dial(c.URL)
	if err != nil {
		return fmt.Errorf("connect rabbitmq consumer: %w", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq consumer channel: %w", err)
	}
	// Moving a failed event happens on its own confirmed channel: the original
	// is only acked once the copy is safely stored.
	moveCh, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq retry channel: %w", err)
	}
	if err := moveCh.Confirm(false); err != nil {
		return fmt.Errorf("enable retry confirms: %w", err)
	}
	if err := c.declare(ch); err != nil {
		return err
	}
	deliveries, err := ch.ConsumeWithContext(ctx, c.Queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue %s: %w", c.Queue, err)
	}
	slog.InfoContext(ctx, "rabbitmq consumer connected", "queue", c.Queue)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("rabbitmq delivery channel closed")
			}
			if err := c.deliver(ctx, moveCh, delivery, handle); err != nil {
				return err
			}
		}
	}
}

func (c Consumer) declare(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(DomainExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare domain exchange: %w", err)
	}
	if _, err := ch.QueueDeclare(c.Queue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue %s: %w", c.Queue, err)
	}
	for _, key := range c.RoutingKeys {
		if err := ch.QueueBind(c.Queue, key, DomainExchange, false, nil); err != nil {
			return fmt.Errorf("bind queue %s to %s: %w", c.Queue, key, err)
		}
	}
	for index, delay := range c.RetryDelays {
		// The default exchange routes by queue name, so an expired message
		// returns to Queue without a dedicated exchange.
		args := amqp.Table{
			"x-message-ttl":             delay.Milliseconds(),
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": c.Queue,
		}
		if _, err := ch.QueueDeclare(c.retryQueue(index+1), true, false, false, false, args); err != nil {
			return fmt.Errorf("declare retry queue: %w", err)
		}
	}
	if _, err := ch.QueueDeclare(c.deadQueue(), true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dead queue: %w", err)
	}
	if err := ch.Qos(c.Prefetch, 0, false); err != nil {
		return fmt.Errorf("set consumer prefetch: %w", err)
	}
	return nil
}

// deliver handles one delivery. A non-nil error means the broker connection is
// unusable and the caller must reconnect.
func (c Consumer) deliver(ctx context.Context, moveCh *amqp.Channel, delivery amqp.Delivery, handle Handler) error {
	routingKey := routingKeyOf(delivery)
	ctx, span := otel.Tracer("events.consumer").Start(
		otel.GetTextMapPropagator().Extract(ctx, tableCarrier(delivery.Headers)),
		"process "+routingKey,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination.name", c.Queue),
			attribute.String("messaging.rabbitmq.destination.routing_key", routingKey),
		))
	defer span.End()

	handleErr := handle(ctx, routingKey, delivery.Body)
	if handleErr == nil {
		if err := delivery.Ack(false); err != nil {
			slog.WarnContext(ctx, "ack event", "routing_key", routingKey, "error", err)
		}
		return nil
	}
	span.RecordError(handleErr)
	span.SetStatus(codes.Error, handleErr.Error())

	attempt := attemptOf(delivery.Headers)
	target, reason := c.deadQueue(), handleErr.Error()
	retry := !errors.Is(handleErr, ErrPermanent) && attempt < len(c.RetryDelays)
	if retry {
		target = c.retryQueue(attempt + 1)
	}
	slog.WarnContext(ctx, "handle event", "queue", c.Queue, "routing_key", routingKey, "attempt", attempt, "moved_to", target, "error", handleErr)

	headers := amqp.Table{}
	maps.Copy(headers, delivery.Headers)
	headers[attemptHeader] = int32(attempt + 1)
	headers[routingKeyHeader] = routingKey
	if !retry {
		headers["x-failure-reason"] = reason
	}
	if err := c.move(ctx, moveCh, target, delivery, headers); err != nil {
		_ = delivery.Nack(false, true)
		return fmt.Errorf("move event to %s: %w", target, err)
	}
	if err := delivery.Ack(false); err != nil {
		slog.WarnContext(ctx, "ack moved event", "routing_key", routingKey, "error", err)
	}
	return nil
}

func (c Consumer) move(ctx context.Context, ch *amqp.Channel, queue string, delivery amqp.Delivery, headers amqp.Table) error {
	confirmation, err := ch.PublishWithDeferredConfirmWithContext(ctx, "", queue, false, false, amqp.Publishing{
		Headers:      headers,
		ContentType:  delivery.ContentType,
		DeliveryMode: amqp.Persistent,
		Timestamp:    delivery.Timestamp,
		Body:         delivery.Body,
	})
	if err != nil {
		return err
	}
	acked, err := confirmation.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !acked {
		return errors.New("broker negatively acknowledged the move")
	}
	return nil
}

// routingKeyOf returns the key the event was originally published with.
func routingKeyOf(delivery amqp.Delivery) string {
	if key, ok := delivery.Headers[routingKeyHeader].(string); ok && key != "" {
		return key
	}
	return delivery.RoutingKey
}

func attemptOf(headers amqp.Table) int {
	switch value := headers[attemptHeader].(type) {
	case int32:
		return int(value)
	case int64:
		return int(value)
	case int:
		return value
	}
	return 0
}
