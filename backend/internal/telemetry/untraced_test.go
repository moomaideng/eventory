package telemetry

import (
	"context"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// newRecordingTracer builds a tracer with the same sampler Setup installs.
func newRecordingTracer(t *testing.T) (*tracetest.SpanRecorder, func(context.Context, string) (context.Context, func())) {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(recorder),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	tracer := provider.Tracer("test")
	return recorder, func(ctx context.Context, name string) (context.Context, func()) {
		ctx, span := tracer.Start(ctx, name)
		return ctx, func() { span.End() }
	}
}

func TestUntracedDropsSpansStartedBeneathIt(t *testing.T) {
	recorder, start := newRecordingTracer(t)

	_, end := start(Untraced(context.Background()), "select tournament_teams")
	end()

	if got := recorder.Ended(); len(got) != 0 {
		t.Fatalf("spans recorded under Untraced = %d, want 0", len(got))
	}
}

func TestSpansWithoutAParentAreStillRecorded(t *testing.T) {
	recorder, start := newRecordingTracer(t)

	_, end := start(context.Background(), "root")
	end()

	if got := recorder.Ended(); len(got) != 1 {
		t.Fatalf("root spans recorded = %d, want 1", len(got))
	}
}

// The reconcile pattern: an untraced poll, then real work under its own span.
func TestSpanStartedFromOriginalContextIgnoresEarlierUntracedPoll(t *testing.T) {
	recorder, start := newRecordingTracer(t)
	ctx := context.Background()

	_, endPoll := start(Untraced(ctx), "poll")
	endPoll()
	workCtx, endWork := start(ctx, "reconcile")
	_, endChild := start(workCtx, "update")
	endChild()
	endWork()

	spans := recorder.Ended()
	if len(spans) != 2 {
		t.Fatalf("spans = %d, want 2 (reconcile + update)", len(spans))
	}
	for _, s := range spans {
		if s.Name() == "poll" {
			t.Fatalf("untraced poll span was recorded")
		}
	}
	if spans[0].Parent().SpanID() != spans[1].SpanContext().SpanID() {
		t.Fatalf("update is not a child of reconcile")
	}
}
