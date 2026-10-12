package grpcx

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func TestGRPCLogLevel(t *testing.T) {
	threshold := 100 * time.Millisecond
	tests := []struct {
		code    codes.Code
		dur     time.Duration
		wantLvl slog.Level
	}{
		{codes.OK, time.Millisecond, slog.LevelDebug},
		{codes.OK, 200 * time.Millisecond, slog.LevelWarn},
		{codes.Canceled, time.Millisecond, slog.LevelDebug},
		{codes.NotFound, time.Millisecond, slog.LevelWarn},
		{codes.Unavailable, time.Millisecond, slog.LevelWarn},
		{codes.Unimplemented, time.Millisecond, slog.LevelError},
		{codes.Internal, time.Millisecond, slog.LevelError},
	}
	for _, tt := range tests {
		got := grpcLogLevel(tt.code, tt.dur, threshold)
		if got != tt.wantLvl {
			t.Errorf("grpcLogLevel(%v, %v) = %v, want %v", tt.code, tt.dur, got, tt.wantLvl)
		}
	}
}

func TestUnaryServerLoggingLevels(t *testing.T) {
	h := &recordHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(prev) })

	tests := []struct {
		name    string
		handler grpc.UnaryHandler
		slowMS  int
		wantLvl slog.Level
	}{
		{
			name: "ok debug",
			handler: func(ctx context.Context, req any) (any, error) {
				return "ok", nil
			},
			slowMS:  1000,
			wantLvl: slog.LevelDebug,
		},
		{
			name: "not found warn",
			handler: func(ctx context.Context, req any) (any, error) {
				return nil, status.Error(codes.NotFound, "missing")
			},
			slowMS:  1000,
			wantLvl: slog.LevelWarn,
		},
		{
			name: "internal error",
			handler: func(ctx context.Context, req any) (any, error) {
				return nil, status.Error(codes.Internal, "boom")
			},
			slowMS:  1000,
			wantLvl: slog.LevelError,
		},
		{
			name: "slow warn",
			handler: func(ctx context.Context, req any) (any, error) {
				time.Sleep(40 * time.Millisecond)
				return "ok", nil
			},
			slowMS:  10,
			wantLvl: slog.LevelWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.mu.Lock()
			h.records = nil
			h.mu.Unlock()

			interceptor := UnaryServerLogging(tt.slowMS)
			info := &grpc.UnaryServerInfo{FullMethod: "/account.v1.AccountService/BatchGetAccounts"}
			_, _ = interceptor(context.Background(), struct{}{}, info, tt.handler)

			rec, ok := h.last()
			if !ok {
				t.Fatal("expected a log record")
			}
			if rec.Message != "grpc_server" {
				t.Fatalf("message = %q, want grpc_server", rec.Message)
			}
			if rec.Level != tt.wantLvl {
				t.Fatalf("level = %v, want %v", rec.Level, tt.wantLvl)
			}
		})
	}
}

func TestUnaryClientLoggingForwardsAndLogs(t *testing.T) {
	h := &recordHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(prev) })

	interceptor := UnaryClientLogging(1000)
	err := interceptor(context.Background(), "/account.v1.AccountService/Get", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	rec, ok := h.last()
	if !ok || rec.Message != "grpc_client" {
		t.Fatalf("expected grpc_client log, got ok=%v msg=%q", ok, rec.Message)
	}
	if rec.Level != slog.LevelDebug {
		t.Fatalf("level = %v, want debug", rec.Level)
	}
}
