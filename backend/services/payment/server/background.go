package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/internal/telemetry"
	"github.com/moomaideng/eventory/services/payment/internal/usecases"
)

const (
	// The relay turns a committed invoice change into an event, so it runs
	// often; an idle pass is one query on a partial index.
	outboxInterval = time.Second
	// Expiry is deadline-driven; nobody waits on the exact second.
	expiryInterval = 30 * time.Second
	batchSize      = 100
	tickTimeout    = 30 * time.Second
)

// runBackground publishes the outbox and announces expired invoices until ctx ends.
func runBackground(ctx context.Context, uc *usecases.PaymentUseCase, publisher events.Publisher) {
	go every(ctx, expiryInterval, func(ctx context.Context) {
		drain(ctx, "announce expired payments", func(query context.Context) (int, error) {
			return uc.AnnounceExpired(query, time.Now().UTC(), batchSize)
		})
	})
	every(ctx, outboxInterval, func(ctx context.Context) {
		drain(ctx, "publish payment outbox", func(query context.Context) (int, error) {
			// Each publish starts from ctx, so it is its own trace.
			return uc.PublishOutbox(query, batchSize, func(routingKey string, payload []byte) error {
				return publisher.Publish(ctx, routingKey, json.RawMessage(payload))
			})
		})
	})
}

// every runs tick now and then once per interval. The first run catches up on
// whatever piled up while the service was down.
func every(ctx context.Context, interval time.Duration, tick func(context.Context)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		tickCtx, cancel := context.WithTimeout(ctx, tickTimeout)
		tick(tickCtx)
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// drain repeats step while it handles full batches, so a backlog clears in one
// tick. step's context is untraced: an idle pass leaves no trace.
func drain(ctx context.Context, what string, step func(context.Context) (int, error)) {
	for ctx.Err() == nil {
		handled, err := step(telemetry.Untraced(ctx))
		if err != nil {
			slog.WarnContext(ctx, what, "handled", handled, "error", err)
			return
		}
		if handled < batchSize {
			return
		}
	}
}
