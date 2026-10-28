package grpcx

import (
	"fmt"

	"github.com/moomaideng/eventory/internal/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewClient creates a gRPC client connection with default unary interceptors:
// logging (outer) then retry (inner), so one access log reflects the final outcome
// while retries are logged at Debug inside the retry interceptor.
//
// Defaults to insecure credentials. Additional opts are appended and may override.
func NewClient(addr string, slowRequestMS int, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			UnaryClientLogging(slowRequestMS),
			UnaryClientRetry(retry.Options{}, nil),
		),
	}
	dialOpts = append(dialOpts, opts...)

	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", addr, err)
	}
	return conn, nil
}
