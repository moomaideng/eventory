package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/moomaideng/eventory/internal/config"
)

// AccessLog logs one structured line per HTTP request after it completes.
// Level: Error for 5xx, Warn for 4xx or slow requests, Debug otherwise.
// Must wrap Recover so panics are recorded as status 500.
func AccessLog(slowRequestMS int) func(http.Handler) http.Handler {
	if slowRequestMS <= 0 {
		slowRequestMS = config.DefaultSlowRequestMS
	}
	threshold := time.Duration(slowRequestMS) * time.Millisecond

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			// Skip successful healthcheck
			if r.URL.Path == "/health" && rec.status < 400 {
				return
			}

			duration := time.Since(start)
			level := httpLogLevel(rec.status, duration, threshold)
			slog.Log(r.Context(), level, "request",
				"system", "http",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", duration.Milliseconds(),
				"request_id", RequestIDFromContext(r.Context()),
			)
		})
	}
}

// Recover converts panics into HTTP 500 and logs them. Must sit inside
// AccessLog so the access line sees status 500.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec)
			}

			slog.ErrorContext(r.Context(), "panic recovered",
				"error", rec,
				"path", r.URL.Path,
				"request_id", RequestIDFromContext(r.Context()),
				"stack", string(debug.Stack()),
			)
			w.WriteHeader(http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}

func httpLogLevel(status int, duration, threshold time.Duration) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	case duration >= threshold:
		return slog.LevelWarn
	default:
		return slog.LevelDebug
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Unwrap exposes the underlying ResponseWriter for interfaces like Flusher.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
