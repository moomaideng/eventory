package grpcx

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/config"
	"github.com/moomaideng/eventory/internal/middlewares"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const requestIDMetadataKey = "x-request-id"

// UnaryServerLogging logs unary RPCs with the same level policy as HTTP access logs.
func UnaryServerLogging(slowRequestMS int) grpc.UnaryServerInterceptor {
	threshold := slowThreshold(slowRequestMS)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		requestID := incomingRequestID(ctx)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx = middlewares.WithRequestID(ctx, requestID)
		_ = grpc.SetHeader(ctx, metadata.Pairs(requestIDMetadataKey, requestID))

		start := time.Now()
		resp, err := handler(ctx, req)
		logGRPC(ctx, "grpc_server", info.FullMethod, err, time.Since(start), threshold, requestID)
		return resp, err
	}
}

// UnaryClientLogging logs outbound unary RPCs and forwards x-request-id metadata.
func UnaryClientLogging(slowRequestMS int) grpc.UnaryClientInterceptor {
	threshold := slowThreshold(slowRequestMS)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		requestID := middlewares.RequestIDFromContext(ctx)
		if requestID == "" {
			requestID = incomingRequestID(ctx)
		}
		if requestID == "" {
			requestID = uuid.NewString()
		}
		ctx = middlewares.WithRequestID(ctx, requestID)
		ctx = metadata.AppendToOutgoingContext(ctx, requestIDMetadataKey, requestID)

		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		logGRPC(ctx, "grpc_client", method, err, time.Since(start), threshold, requestID)
		return err
	}
}

func slowThreshold(slowRequestMS int) time.Duration {
	if slowRequestMS <= 0 {
		slowRequestMS = config.DefaultSlowRequestMS
	}
	return time.Duration(slowRequestMS) * time.Millisecond
}

func incomingRequestID(ctx context.Context) string {
	if id := middlewares.RequestIDFromContext(ctx); id != "" {
		return id
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(requestIDMetadataKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func logGRPC(ctx context.Context, msg, fullMethod string, err error, duration, threshold time.Duration, requestID string) {
	code := codes.OK
	if err != nil {
		code = status.Code(err)
	}
	service, method := splitFullMethod(fullMethod)
	level := grpcLogLevel(code, duration, threshold)
	attrs := []any{
		"system", "grpc",
		"grpc.service", service,
		"grpc.method", method,
		"grpc.code", code.String(),
		"duration_ms", duration.Milliseconds(),
		"request_id", requestID,
	}
	if err != nil && code != codes.OK {
		attrs = append(attrs, "error", err.Error())
	}
	slog.Log(ctx, level, msg, attrs...)
}

func grpcLogLevel(code codes.Code, duration, threshold time.Duration) slog.Level {
	switch code {
	case codes.OK:
		if duration >= threshold {
			return slog.LevelWarn
		}
		return slog.LevelDebug
	case codes.Canceled:
		return slog.LevelDebug
	case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.Unauthenticated,
		codes.DeadlineExceeded, codes.PermissionDenied, codes.ResourceExhausted,
		codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unavailable:
		return slog.LevelWarn
	case codes.Unknown, codes.Unimplemented, codes.Internal, codes.DataLoss:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

func splitFullMethod(fullMethod string) (service, method string) {
	// fullMethod is "/package.Service/Method"
	if len(fullMethod) == 0 {
		return "", ""
	}
	s := fullMethod
	if s[0] == '/' {
		s = s[1:]
	}
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}
