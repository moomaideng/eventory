package rabbitmq

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceContextSurvivesAMQPHeaders(t *testing.T) {
	previous := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(previous) })

	provider := sdktrace.NewTracerProvider()
	ctx, span := provider.Tracer("test").Start(context.Background(), "publish")
	defer span.End()

	headers := amqp.Table{"x-attempt": int32(3)}
	otel.GetTextMapPropagator().Inject(ctx, tableCarrier(headers))

	if Attempt(headers) != 3 {
		t.Errorf("injecting trace headers clobbered x-attempt: got %d", Attempt(headers))
	}

	got := trace.SpanContextFromContext(ExtractTrace(context.Background(), headers))
	if got.TraceID() != span.SpanContext().TraceID() || got.SpanID() != span.SpanContext().SpanID() {
		t.Errorf("extracted %s/%s, want %s/%s", got.TraceID(), got.SpanID(), span.SpanContext().TraceID(), span.SpanContext().SpanID())
	}
	if !got.IsRemote() {
		t.Error("extracted span context must be marked remote")
	}
}

func TestExtractTraceIgnoresMissingOrForeignHeaders(t *testing.T) {
	previous := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() { otel.SetTextMapPropagator(previous) })

	// Messages published before tracing existed have no traceparent; a
	// non-string value must not panic either.
	for name, headers := range map[string]amqp.Table{
		"nil":        nil,
		"empty":      {},
		"non-string": {"traceparent": int32(1)},
	} {
		if sc := trace.SpanContextFromContext(ExtractTrace(context.Background(), headers)); sc.IsValid() {
			t.Errorf("%s headers produced a valid span context %v", name, sc)
		}
	}
}
