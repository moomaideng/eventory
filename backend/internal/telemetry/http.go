package telemetry

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HTTPMiddleware wraps a handler with an OTel server span per request.
// Health probes are excluded, matching the access-log skip policy.
func HTTPMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, serviceName+".http", otelhttp.WithFilter(skipHealthProbes))
	}
}

func skipHealthProbes(r *http.Request) bool {
	return r.URL.Path != "/health" && r.URL.Path != "/ready"
}
