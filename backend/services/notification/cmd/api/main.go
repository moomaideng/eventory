package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	shareddatabase "github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/internal/logger"
	"github.com/moomaideng/eventory/services/notification/config"
	"github.com/moomaideng/eventory/services/notification/internal/adapters/rabbitmq"
	smtpadapter "github.com/moomaideng/eventory/services/notification/internal/adapters/smtp"
	"github.com/moomaideng/eventory/services/notification/internal/repositories"
	"github.com/moomaideng/eventory/services/notification/internal/usecases"
	"github.com/moomaideng/eventory/services/notification/internal/worker"
	"github.com/moomaideng/eventory/services/notification/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		panic(err)
	}
	logger.SetDefault(cfg.Logging, cfg.Environment, "notification")

	db, err := shareddatabase.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		slog.Error("failed to connect to notification database", "error", err)
		panic(err)
	}
	broker, err := rabbitmq.Connect(cfg.BrokerURL, cfg.RetryDelays, cfg.ConsumerPrefetch)
	if err != nil {
		slog.Error("failed to connect to RabbitMQ", "error", err)
		panic(err)
	}
	defer broker.Close()

	mailer := smtpadapter.NewSender(smtpadapter.Config{
		Host: cfg.SMTPHost, Port: cfg.SMTPPort, Username: cfg.SMTPUsername,
		Password: cfg.SMTPPassword, FromAddress: cfg.SMTPFromAddress,
		FromName: cfg.SMTPFromName, StartTLS: cfg.SMTPStartTLS, Timeout: cfg.SMTPTimeout,
	})
	deliveryRepo := repositories.NewDeliveryRepository(db)
	useCase := usecases.NewNotificationUseCase(deliveryRepo, mailer, cfg.PublicWebURL, cfg.ProcessingLease)
	consumer := worker.New(broker, useCase, len(cfg.RetryDelays), cfg.BrokerReconnectDelay)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           server.New(db, broker),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errorsChannel := make(chan error, 2)
	go func() {
		slog.Info("notification health server listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errorsChannel <- err
		}
	}()
	go func() { errorsChannel <- consumer.Run(ctx) }()

	select {
	case err := <-errorsChannel:
		if err != nil {
			slog.Error("notification service stopped", "error", err)
		}
		stop()
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
