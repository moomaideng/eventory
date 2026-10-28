package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/internal/logger"
	"github.com/moomaideng/eventory/internal/telemetry"
	"github.com/moomaideng/eventory/services/tournament/config"
	"github.com/moomaideng/eventory/services/tournament/server"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

	var mongoDB *mongo.Database
	if appConfig.MongoURI != "" {
		connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		mongoClient, err := database.ConnectMongo(connectCtx, appConfig.MongoURI)
		if err != nil {
			slog.Error("failed to connect to mongodb", "error", err)
			panic(err)
		}
		defer func() {
			_ = mongoClient.Disconnect(context.Background())
		}()
		slog.Info("mongodb connection established")
		dbName := appConfig.MongoDBName
		if dbName == "" {
			dbName = "tournament_db"
		}
		mongoDB = mongoClient.Database(dbName)
	}

	app := server.NewApp(db, mongoDB, appConfig)
	go app.RunSubmissions(ctx)

	httpServer := &http.Server{Addr: ":" + appConfig.HTTPPort, Handler: app.HTTP}

	errorsChannel := make(chan error, 1)
	go func() {
		slog.Info("tournament HTTP starting", "port", appConfig.HTTPPort, "env", appConfig.Environment)
		slog.Info("API docs available", "url", "http://localhost:"+appConfig.HTTPPort+"/docs")
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errorsChannel <- err
		}
	}()

	select {
	case err := <-errorsChannel:
		if err != nil {
			slog.Error("tournament service stopped", "error", err)
		}
		stop()
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	if err := shutdownTelemetry(shutdownCtx); err != nil {
		slog.Error("telemetry shutdown failed", "error", err)
	}
}
