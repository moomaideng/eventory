package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/internal/logger"
	"github.com/moomaideng/eventory/internal/telemetry"
	"github.com/moomaideng/eventory/services/payment/config"
	"github.com/moomaideng/eventory/services/payment/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		panic(err)
	}
	logger.SetDefault(cfg.Logging, cfg.Environment, config.ServiceName)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTelemetry, err := telemetry.Setup(ctx, telemetry.Config{ServiceName: config.ServiceName, Environment: cfg.Environment, Endpoint: cfg.OTel.ExporterOTLPEndpoint})
	if err != nil {
		slog.Error("failed to set up telemetry; continuing without tracing", "error", err)
	}

	db, err := database.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}
	publisher, err := events.NewRabbitMQPublisher(cfg.BrokerURL)
	if err != nil {
		slog.Error("failed to create event publisher", "error", err)
		panic(err)
	}
	app := server.NewApp(db, cfg, publisher)
	go app.RunBackground(ctx)
	httpServer := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: app.HTTP, ReadHeaderTimeout: 10 * time.Second}

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		panic(err)
	}
	errorsChannel := make(chan error, 2)
	go func() {
		slog.Info("payment gRPC listening", "port", cfg.GRPCPort)
		if err := app.GRPC.Serve(lis); err != nil {
			errorsChannel <- err
		}
	}()
	go func() {
		slog.Info("payment API listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errorsChannel <- err
		}
	}()

	select {
	case err := <-errorsChannel:
		if err != nil {
			slog.Error("payment service stopped", "error", err)
		}
		stop()
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	stopped := make(chan struct{})
	go func() { app.GRPC.GracefulStop(); close(stopped) }()
	select {
	case <-stopped:
	case <-shutdownCtx.Done():
		app.GRPC.Stop()
	}
	if err := shutdownTelemetry(shutdownCtx); err != nil {
		slog.Error("telemetry shutdown failed", "error", err)
	}
}
