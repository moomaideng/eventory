package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type recordHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *recordHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *recordHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r.Clone())
	return nil
}
func (h *recordHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *recordHandler) WithGroup(string) slog.Handler      { return h }

func (h *recordHandler) last() (slog.Record, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.records) == 0 {
		return slog.Record{}, false
	}
	return h.records[len(h.records)-1], true
}

func TestAccessLogLevels(t *testing.T) {
	h := &recordHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(prev) })

	tests := []struct {
		name     string
		status   int
		delay    time.Duration
		wantLvl  slog.Level
		slowMS   int
	}{
		{name: "ok debug", status: 200, wantLvl: slog.LevelDebug, slowMS: 1000},
		{name: "client warn", status: 404, wantLvl: slog.LevelWarn, slowMS: 1000},
		{name: "server error", status: 500, wantLvl: slog.LevelError, slowMS: 1000},
		{name: "slow warn", status: 200, delay: 50 * time.Millisecond, wantLvl: slog.LevelWarn, slowMS: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.mu.Lock()
			h.records = nil
			h.mu.Unlock()

			handler := RequestID(AccessLog(tt.slowMS)(Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.delay > 0 {
					time.Sleep(tt.delay)
				}
				w.WriteHeader(tt.status)
			}))))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			rec, ok := h.last()
			if !ok {
				t.Fatal("expected a log record")
			}
			if rec.Level != tt.wantLvl {
				t.Fatalf("level = %v, want %v", rec.Level, tt.wantLvl)
			}
			if rr.Header().Get(RequestIDHeader) == "" {
				t.Fatal("expected X-Request-Id response header")
			}
		})
	}
}

func TestRecoverLogsPanicAsError(t *testing.T) {
	h := &recordHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(prev) })

	handler := RequestID(AccessLog(1000)(Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}

	foundAccess := false
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, rec := range h.records {
		if rec.Message == "request" && rec.Level == slog.LevelError {
			foundAccess = true
		}
	}
	if !foundAccess {
		t.Fatal("expected access log at Error after panic")
	}
}
