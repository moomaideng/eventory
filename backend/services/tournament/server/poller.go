package server

import (
	"context"
	"time"

	"github.com/moomaideng/eventory/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// poller feeds every pending item to reconcile on a fixed interval.
type poller[T any] struct {
	interval    time.Duration
	spanName    string
	listPending func(context.Context) ([]T, error)
	reconcile   func(context.Context, *T) error
	// A failure is reported and the item retried on the next tick; the loop
	// only ends with ctx.
	onListError      func(context.Context, error)
	onReconcileError func(context.Context, *T, error)
}

func (p poller[T]) run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.tick(ctx)
		}
	}
}

func (p poller[T]) tick(ctx context.Context) {
	// Untraced: with no parent span, the listing query would become its own
	// one-span root trace every tick, even when nothing is pending.
	items, err := p.listPending(telemetry.Untraced(ctx))
	if err != nil {
		p.onListError(ctx, err)
		return
	}
	if len(items) == 0 {
		return
	}
	// Started only once there is work, from the original ctx, so everything
	// reconcile does (payment call, writes) shares this span as its root.
	ctx, span := otel.Tracer("tournament.reconciler").Start(ctx, p.spanName,
		trace.WithAttributes(attribute.Int("reconcile.pending", len(items))))
	defer span.End()
	for i := range items {
		if ctx.Err() != nil {
			return
		}
		if err := p.reconcile(ctx, &items[i]); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "reconcile failed")
			p.onReconcileError(ctx, &items[i], err)
		}
	}
}
