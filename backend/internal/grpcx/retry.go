package grpcx

import (
	"context"
	"log/slog"

	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryDecision reports whether a failed unary RPC attempt should be retried.
// method is the full gRPC method name, so a future non-idempotent RPC can get
// its own policy (e.g. never retry) without changing UnaryClientRetry itself.
type RetryDecision func(method string, err error) bool

// DefaultRetryDecision retries only transient codes: Unavailable, DeadlineExceeded,
// ResourceExhausted, Aborted. It ignores method and is safe for read-only RPCs;
// a non-idempotent RPC should pass its own RetryDecision to UnaryClientRetry instead.
func DefaultRetryDecision(_ string, err error) bool {
	return isRetryableGRPCError(err)
}

// UnaryClientRetry retries transient unary RPC failures with exponential backoff
// and full jitter. opts configures backoff only (zero value uses retry.Do's
// defaults); opts.RetryIf is ignored in favor of decide. decide may be nil, which
// uses DefaultRetryDecision.
func UnaryClientRetry(opts retry.Options, decide RetryDecision) grpc.UnaryClientInterceptor {
	if decide == nil {
		decide = DefaultRetryDecision
	}

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
		call := opts
		call.RetryIf = func(err error) bool { return decide(method, err) }
		call.OnRetry = func(ctx context.Context, attempt int, err error) {
			attrs := []any{
				"attempt", attempt,
				"grpc.method", method,
				"grpc.code", status.Code(err).String(),
				"error", err.Error(),
			}
			if requestID := middlewares.RequestIDFromContext(ctx); requestID != "" {
				attrs = append(attrs, "request_id", requestID)
			}
			slog.DebugContext(ctx, "grpc_client_retry", attrs...)
		}

		return retry.Do(ctx, call, func(ctx context.Context) error {
			return invoker(ctx, method, req, reply, cc, callOpts...)
		})
	}
}

func isRetryableGRPCError(err error) bool {
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}
