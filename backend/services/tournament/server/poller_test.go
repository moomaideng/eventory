package server

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// recordSpans installs a recording provider with the sampler production uses.
func recordSpans(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	previous := otel.GetTracerProvider()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(recorder),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = provider.Shutdown(context.Background())
	})
	return recorder
}

// testPoller's list step starts a span like the GORM plugin does per query,
// and its reconcile step starts one like the payment HTTP client does.
func testPoller(items []string, reconcileErr error) poller[string] {
	return poller[string]{
		spanName: "reconcile pending test",
		listPending: func(ctx context.Context) ([]string, error) {
			_, span := otel.Tracer("test").Start(ctx, "select pending")
			span.End()
			return items, nil
		},
		reconcile: func(ctx context.Context, _ *string) error {
			_, span := otel.Tracer("test").Start(ctx, "call payment")
			span.End()
			return reconcileErr
		},
		onListError:      func(context.Context, error) {},
		onReconcileError: func(context.Context, *string, error) {},
	}
}

func TestIdlePollEmitsNoSpans(t *testing.T) {
	recorder := recordSpans(t)

	testPoller(nil, nil).tick(context.Background())

	if got := recorder.Ended(); len(got) != 0 {
		t.Fatalf("idle tick emitted %d spans, first %q", len(got), got[0].Name())
	}
}

func TestPollWithWorkIsOneTraceUnderASingleRoot(t *testing.T) {
	recorder := recordSpans(t)

	testPoller([]string{"a", "b"}, errors.New("boom")).tick(context.Background())

	var root sdktrace.ReadOnlySpan
	calls := 0
	for _, s := range recorder.Ended() {
		switch s.Name() {
		case "reconcile pending test":
			root = s
		case "call payment":
			calls++
		default:
			t.Errorf("unexpected span %q: the listing query must not be traced", s.Name())
		}
	}
	if root == nil {
		t.Fatal("no work span recorded")
	}
	if root.Parent().IsValid() {
		t.Error("work span has a parent; want a root")
	}
	if calls != 2 {
		t.Fatalf("reconcile spans = %d, want 2", calls)
	}
	for _, s := range recorder.Ended() {
		if s != root && s.Parent().SpanID() != root.SpanContext().SpanID() {
			t.Errorf("%q is not a child of the work span", s.Name())
		}
	}
	if root.Status().Code.String() != "Error" {
		t.Errorf("work span status = %v, want Error after failed reconciles", root.Status().Code)
	}
}
