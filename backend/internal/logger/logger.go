package logger

import (
	"log/slog"
	"os"

	"github.com/moomaideng/eventory/internal/config"
)

// SetDefault configures the process-wide slog default logger.
// cfg should already be normalized via Logging.WithDefaults.
func SetDefault(cfg config.Logging, environment, service string) *slog.Logger {
	log := New(cfg, environment, service)
	slog.SetDefault(log)
	return log
}

// New builds a slog.Logger from a normalized Logging config.
func New(cfg config.Logging, environment, service string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.Level)}
	var handler slog.Handler
	if cfg.IsJSON() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

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
