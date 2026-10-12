package grpcx

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryClientRetryNonRetryable(t *testing.T) {
	var calls atomic.Int32
	interceptor := UnaryClientRetry(retry.Options{
		MaxAttempts: 5,
		Sleep:       func(context.Context, time.Duration) error { return nil },
	}, nil)

	err := interceptor(context.Background(), "/svc/Method", nil, nil, nil,
		func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
			calls.Add(1)
			return status.Error(codes.NotFound, "missing")
		},
	)
	if status.Code(err) != codes.NotFound {
		t.Fatalf("code = %v, want NotFound", status.Code(err))
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestUnaryClientRetryDoesNotRetryOK(t *testing.T) {
	var calls atomic.Int32
	interceptor := UnaryClientRetry(retry.Options{
		MaxAttempts: 3,
		Sleep:       func(context.Context, time.Duration) error { return nil },
	}, nil)

	err := interceptor(context.Background(), "/svc/Method", nil, nil, nil,
		func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
			calls.Add(1)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}

func TestUnaryClientRetryLogsDebugOnRetry(t *testing.T) {
	h := &recordHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(prev) })

	var calls atomic.Int32
	interceptor := UnaryClientRetry(retry.Options{
		MaxAttempts: 3,
		Sleep:       func(context.Context, time.Duration) error { return nil },
	}, nil)

	ctx := middlewares.WithRequestID(context.Background(), "req-1")
	_ = interceptor(ctx, "/account.v1.AccountService/BatchGetAccounts", nil, nil, nil,
		func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
			if calls.Add(1) < 3 {
				return status.Error(codes.Unavailable, "down")
			}
			return nil
		},
	)

	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.records) != 2 {
		t.Fatalf("retry logs = %d, want 2", len(h.records))
	}
	for _, rec := range h.records {
		if rec.Message != "grpc_client_retry" {
			t.Fatalf("message = %q, want grpc_client_retry", rec.Message)
		}
		if rec.Level != slog.LevelDebug {
			t.Fatalf("level = %v, want debug", rec.Level)
		}
	}
}

// TestUnaryClientRetryCustomDecision proves the per-method retry hook: a RetryDecision
// can veto retries for a specific method even on an otherwise-transient code, which is
// how a future non-idempotent RPC opts out of retry without changing this interceptor.
func TestUnaryClientRetryCustomDecision(t *testing.T) {
	const nonIdempotentMethod = "/svc/NonIdempotentMutate"
	var calls atomic.Int32
	decide := func(method string, err error) bool {
		return method != nonIdempotentMethod && DefaultRetryDecision(method, err)
	}
	interceptor := UnaryClientRetry(retry.Options{
		MaxAttempts: 3,
		Sleep:       func(context.Context, time.Duration) error { return nil },
	}, decide)

	err := interceptor(context.Background(), nonIdempotentMethod, nil, nil, nil,
		func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
			calls.Add(1)
			return status.Error(codes.Unavailable, "down")
		},
	)
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("code = %v, want Unavailable", status.Code(err))
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1 (custom decision must block retry for this method)", calls.Load())
	}
}

func TestIsRetryableGRPCError(t *testing.T) {
	retryable := []codes.Code{
		codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted,
	}
	for _, code := range retryable {
		if !isRetryableGRPCError(status.Error(code, "x")) {
			t.Errorf("code %v should be retryable", code)
		}
	}
	notRetryable := []codes.Code{
		codes.OK, codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
		codes.PermissionDenied, codes.Unauthenticated, codes.FailedPrecondition,
		codes.Unimplemented, codes.Internal, codes.DataLoss, codes.Unknown,
	}
	for _, code := range notRetryable {
		err := error(nil)
		if code != codes.OK {
			err = status.Error(code, "x")
		}
		if isRetryableGRPCError(err) {
			t.Errorf("code %v should not be retryable", code)
		}
	}
	if isRetryableGRPCError(errors.New("plain")) {
		t.Error("plain error should not be retryable (Unknown)")
	}
}
