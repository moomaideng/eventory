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
	"github.com/moomaideng/eventory/internal/logger"
	"github.com/moomaideng/eventory/internal/telemetry"
	"github.com/moomaideng/eventory/services/account/config"
	"github.com/moomaideng/eventory/services/account/server"
)

func main() {
	appConfig, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		panic(err)
	}
	logger.SetDefault(appConfig.Logging, appConfig.Environment, config.ServiceName)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTelemetry, err := telemetry.Setup(ctx, telemetry.Config{ServiceName: config.ServiceName, Environment: appConfig.Environment, Endpoint: appConfig.OTel.ExporterOTLPEndpoint})
	if err != nil {
		slog.Error("failed to set up telemetry; continuing without tracing", "error", err)
	}

	db, err := database.ConnectPostgres(appConfig.DBDSN)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}
	slog.Info("database connection established")

	app := server.NewApp(db, appConfig)

	lis, err := net.Listen("tcp", ":"+appConfig.GRPCPort)
	if err != nil {
		slog.Error("failed to listen on gRPC port", "port", appConfig.GRPCPort, "error", err)
		panic(err)
	}

	errorsChannel := make(chan error, 2)
	go func() {
		slog.Info("account gRPC listening", "port", appConfig.GRPCPort)
		if err := app.GRPC.Serve(lis); err != nil {
			errorsChannel <- err
		}
	}()

	httpServer := &http.Server{Addr: ":" + appConfig.HTTPPort, Handler: app.HTTP}
	go func() {
		slog.Info("account HTTP starting", "port", appConfig.HTTPPort, "env", appConfig.Environment)
		slog.Info("API docs available", "url", "http://localhost:"+appConfig.HTTPPort+"/docs")
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errorsChannel <- err
		}
	}()

	select {
	case err := <-errorsChannel:
		if err != nil {
			slog.Error("account service stopped", "error", err)
		}
		stop()
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	app.GRPC.GracefulStop()
	if err := shutdownTelemetry(shutdownCtx); err != nil {
		slog.Error("telemetry shutdown failed", "error", err)
	}
}
