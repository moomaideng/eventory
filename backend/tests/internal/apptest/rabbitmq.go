package apptest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartRabbitMQ launches a disposable broker and returns its AMQP URL.
func StartRabbitMQ(ctx context.Context, tb testing.TB) string {
	tb.Helper()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:4-alpine",
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor:   wait.ForLog("Server startup complete").WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		tb.Fatalf("apptest: start rabbitmq container: %v", err)
	}
	tb.Cleanup(func() { _ = container.Terminate(context.Background()) })
	endpoint, err := container.PortEndpoint(ctx, "5672/tcp", "")
	if err != nil {
		tb.Fatalf("apptest: read rabbitmq endpoint: %v", err)
	}
	// Disposable loopback container without TLS; AMQPS adds nothing here.
	return fmt.Sprintf("amqp://guest:guest@%s/", endpoint) // NOSONAR go:S5332
}
