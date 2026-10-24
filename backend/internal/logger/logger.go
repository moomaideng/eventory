package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/moomaideng/eventory/internal/config"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// SetDefault configures the process-wide slog default logger.
// cfg should already be normalized via Logging.WithDefaults.
func SetDefault(cfg config.Logging, environment, service string) *slog.Logger {
	log := New(cfg, environment, service)
	slog.SetDefault(log)
	return log
}

// New builds a slog.Logger from a normalized Logging config. Every record
// gets trace_id/span_id attrs when the call's context carries a live OTel
// span, correlating Loki lines with Tempo traces.
func New(cfg config.Logging, environment, service string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.Level)}
	var handler slog.Handler
	if cfg.IsJSON() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	handler = traceHandler{handler}

	attrs := []any{}
	if service != "" {
		attrs = append(attrs, "service", service)
	}
	if environment != "" {
		attrs = append(attrs, "env", environment)
	}
	return slog.New(handler).With(attrs...)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// traceHandler adds trace_id/span_id attrs from the record's context when a
// live OTel span is present, so every log line is correlatable in Grafana
// without touching call sites.
type traceHandler struct {
	slog.Handler
}

func (h traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := oteltrace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceHandler{h.Handler.WithAttrs(attrs)}
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return traceHandler{h.Handler.WithGroup(name)}
}
