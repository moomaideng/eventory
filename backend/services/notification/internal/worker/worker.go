package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/notification/internal/adapters/rabbitmq"
	"github.com/moomaideng/eventory/services/notification/internal/usecases"
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
				worker.handle(ctx, delivery.Body, rabbitmq.Attempt(delivery.Headers), delivery.Ack, delivery.Nack)
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
	log.Printf("rabbitmq consumer disconnected: %v", cause)
	for {
		if err := worker.broker.Reconnect(); err == nil {
			log.Printf("rabbitmq consumer reconnected")
			return nil
		} else {
			log.Printf("rabbitmq reconnect failed: %v", err)
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
	var event events.RegistrationSubmitted
	if err := json.Unmarshal(body, &event); err != nil {
		worker.deadLetter(ctx, body, attempt, "invalid JSON", ack, nack)
		return
	}
	outcome, _ := worker.useCase.Process(ctx, event, attempt >= worker.maxRetries)
	switch outcome {
	case usecases.OutcomeSent:
		log.Printf("notification sent event_id=%s attempt=%d", event.ID, attempt)
		_ = ack(false)
	case usecases.OutcomeAlreadyHandled:
		log.Printf("notification duplicate skipped event_id=%s", event.ID)
		_ = ack(false)
	case usecases.OutcomeRetry:
		if attempt >= worker.maxRetries {
			reason := "retry budget exhausted"
			worker.deadLetter(ctx, body, attempt, reason, ack, nack)
			return
		}
		nextAttempt := attempt + 1
		if publishErr := worker.broker.PublishRetry(ctx, body, nextAttempt); publishErr != nil {
			log.Printf("notification retry publish failed event_id=%s: %v", event.ID, publishErr)
			_ = nack(false, true)
			return
		}
		log.Printf("notification retry scheduled event_id=%s attempt=%d", event.ID, nextAttempt)
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
		log.Printf("notification dead-letter publish failed: %v", err)
		_ = nack(false, true)
		return
	}
	log.Printf("notification moved to dead-letter queue reason=%q attempt=%d", reason, attempt)
	_ = ack(false)
}
