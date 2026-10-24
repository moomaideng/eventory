package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
)

// OTEL_EXPORTER_OTLP_ENDPOINT is a base URL (http://alloy:4318); spans must be
// posted to its /v1/traces path, which Alloy serves.
func TestSetupPostsSpansToTracesPath(t *testing.T) {
	// Setup replaces the process-wide provider and propagator; put them back.
	previousProvider, previousPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})

	paths := make(chan string, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths <- r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	shutdown, err := Setup(context.Background(), Config{ServiceName: "test", Environment: "test", Endpoint: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	_, span := otel.Tracer("test").Start(context.Background(), "op")
	span.End()
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-paths:
		if got != "/v1/traces" {
			t.Fatalf("spans posted to %q, want /v1/traces", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("exporter sent nothing")
	}
}
