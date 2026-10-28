package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
)

// Config names the service for the trace resource and sets where spans go.
// An empty Endpoint disables tracing.
type Config struct {
	ServiceName string
	Environment string
	Endpoint    string
}

// noopShutdown lets callers always defer shutdown, even when tracing is off.
func noopShutdown(context.Context) error { return nil }

// Setup enables tracing only when an endpoint is configured, so a local run
// without Alloy doesn't spam export errors. Failures are returned, not
// fatal: callers log them and keep serving.
func Setup(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	if cfg.Endpoint == "" {
		slog.Info("tracing disabled: OTEL_EXPORTER_OTLP_ENDPOINT not set")
		return noopShutdown, nil
	}

	// Since otel v1.45, WithEndpointURL no longer appends the signal path, so
	// the base URL from OTEL_EXPORTER_OTLP_ENDPOINT needs /v1/traces added.
	tracesURL, err := url.JoinPath(cfg.Endpoint, "/v1/traces")
	if err != nil {
		return noopShutdown, fmt.Errorf("build OTLP traces URL from %q: %w", cfg.Endpoint, err)
	}
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(tracesURL))
	if err != nil {
		return noopShutdown, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironmentNameKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return noopShutdown, fmt.Errorf("build OTel resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		// Record every root span, but honour the caller's decision: a request
		// Traefik sampled out, or work wrapped in Untraced, stays out.
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("tracing enabled", "service", cfg.ServiceName)
	return tp.Shutdown, nil
}
