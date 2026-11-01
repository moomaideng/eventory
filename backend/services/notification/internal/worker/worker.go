package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/notification/internal/adapters/rabbitmq"
	"github.com/moomaideng/eventory/services/notification/internal/usecases"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Worker struct {
	broker         *rabbitmq.Broker
	useCase        *usecases.NotificationUseCase
	maxRetries     int
	reconnectDelay time.Duration
}

func New(broker *rabbitmq.Broker, useCase *usecases.NotificationUseCase, maxRetries int, reconnectDelay time.Duration) *Worker {
	return &Worker{broker: broker, useCase: useCase, maxRetries: maxRetries, reconnectDelay: reconnectDelay}
}

func (worker *Worker) Run(ctx context.Context) error {
	for {
		deliveries, err := worker.broker.Deliveries()
		if err != nil {
			if err := worker.reconnect(ctx, err); err != nil {
				return err
			}
			if ctx.Err() != nil {
				return nil
			}
			continue
		}

		connected := true
		for connected {
			select {
			case <-ctx.Done():
				return nil
			case delivery, ok := <-deliveries:
				if !ok {
					connected = false
					continue
				}
				msgCtx, span := otel.Tracer("notification.worker").Start(
					rabbitmq.ExtractTrace(ctx, delivery.Headers),
					"process registration.submitted",
					trace.WithSpanKind(trace.SpanKindConsumer),
				)
				worker.handle(msgCtx, delivery.Body, rabbitmq.Attempt(delivery.Headers), delivery.Ack, delivery.Nack)
				span.End()
			}
		}
		if err := worker.reconnect(ctx, fmt.Errorf("rabbitmq delivery channel closed")); err != nil {
			return err
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

func (worker *Worker) reconnect(ctx context.Context, cause error) error {
	slog.Warn("rabbitmq consumer disconnected", "error", cause)
	for {
		if err := worker.broker.Reconnect(); err == nil {
			slog.Info("rabbitmq consumer reconnected")
			return nil
		} else {
			slog.Warn("rabbitmq reconnect failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(worker.reconnectDelay):
		}
	}
}

func (worker *Worker) handle(
	ctx context.Context,
	body []byte,
	attempt int,
	ack func(bool) error,
	nack func(bool, bool) error,
) {
	type eventEnvelope struct {
		ID   uuid.UUID `json:"id"`
		Type string    `json:"type"`
	}

	var env eventEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		worker.deadLetter(ctx, body, attempt, "invalid JSON", ack, nack)
		return
	}

	finalAttempt := attempt >= worker.maxRetries
	var outcome usecases.Outcome
	var eventID = env.ID

	switch env.Type {
	case events.RegistrationSubmittedType:
		var event events.RegistrationSubmitted
		if err := json.Unmarshal(body, &event); err != nil {
			worker.deadLetter(ctx, body, attempt, "invalid RegistrationSubmitted payload", ack, nack)
			return
		}
		eventID = event.ID
		outcome, _ = worker.useCase.ProcessRegistrationSubmitted(ctx, event, finalAttempt)
	case events.RegistrationApprovedType:
		var event events.RegistrationReviewed
		if err := json.Unmarshal(body, &event); err != nil {
			worker.deadLetter(ctx, body, attempt, "invalid RegistrationReviewed payload", ack, nack)
			return
		}
		eventID = event.ID
		outcome, _ = worker.useCase.ProcessRegistrationApproved(ctx, event, finalAttempt)
	case events.RegistrationRejectedType:
		var event events.RegistrationReviewed
		if err := json.Unmarshal(body, &event); err != nil {
			worker.deadLetter(ctx, body, attempt, "invalid RegistrationReviewed payload", ack, nack)
			return
		}
		eventID = event.ID
		outcome, _ = worker.useCase.ProcessRegistrationRejected(ctx, event, finalAttempt)
	case events.PaymentConfirmedType:
		var event events.PaymentConfirmed
		if err := json.Unmarshal(body, &event); err != nil {
			worker.deadLetter(ctx, body, attempt, "invalid PaymentConfirmed payload", ack, nack)
			return
		}
		eventID = event.ID
		outcome, _ = worker.useCase.ProcessPaymentConfirmed(ctx, event, finalAttempt)
	default:
		worker.deadLetter(ctx, body, attempt, fmt.Sprintf("unsupported event type %q", env.Type), ack, nack)
		return
	}

	switch outcome {
	case usecases.OutcomeSent:
		slog.InfoContext(ctx, "notification sent", "event_id", eventID, "attempt", attempt)
		_ = ack(false)
	case usecases.OutcomeAlreadyHandled:
		slog.DebugContext(ctx, "notification duplicate skipped", "event_id", eventID)
		_ = ack(false)
	case usecases.OutcomeRetry:
		if attempt >= worker.maxRetries {
			reason := "retry budget exhausted"
			worker.deadLetter(ctx, body, attempt, reason, ack, nack)
			return
		}
		nextAttempt := attempt + 1
		if publishErr := worker.broker.PublishRetry(ctx, body, nextAttempt); publishErr != nil {
			slog.ErrorContext(ctx, "notification retry publish failed", "event_id", eventID, "error", publishErr)
			_ = nack(false, true)
			return
		}
		slog.InfoContext(ctx, "notification retry scheduled", "event_id", eventID, "attempt", nextAttempt)
		_ = ack(false)
	case usecases.OutcomePermanentFailure:
		reason := "permanent processing failure"
		worker.deadLetter(ctx, body, attempt, reason, ack, nack)
	}
}

func (worker *Worker) deadLetter(
	ctx context.Context,
	body []byte,
	attempt int,
	reason string,
	ack func(bool) error,
	nack func(bool, bool) error,
) {
	if len(reason) > 240 {
		reason = reason[:240]
	}
	if err := worker.broker.PublishDead(ctx, body, attempt, reason); err != nil {
		slog.ErrorContext(ctx, "notification dead-letter publish failed", "error", err, "attempt", attempt, "reason", reason)
		_ = nack(false, true)
		return
	}
	slog.WarnContext(ctx, "notification moved to dead-letter queue", "reason", reason, "attempt", attempt)
	_ = ack(false)
}
