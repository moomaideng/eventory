package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// untracedParent is a made-up parent span context with the sampled flag
// cleared. Nothing is recorded or exported for it; it only tells the sampler
// "the parent was not sampled". The IDs just need to be non-zero so the
// context counts as valid; without a valid parent, spans become roots.
var untracedParent = trace.NewSpanContext(trace.SpanContextConfig{
	TraceID:    trace.TraceID{1},
	SpanID:     trace.SpanID{1},
	TraceFlags: 0, // unsampled
	Remote:     false,
})

// Untraced returns a context under which no spans are recorded. Use it for
// background work that is usually idle (poll loops), so an empty poll does
// not emit a one-span root trace every tick. Start a real span from the
// original context once there is work to trace.
func Untraced(ctx context.Context) context.Context {
	return trace.ContextWithSpanContext(ctx, untracedParent)
}
