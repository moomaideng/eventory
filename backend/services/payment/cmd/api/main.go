package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/internal/logger"
	"github.com/moomaideng/eventory/services/payment/config"
	"github.com/moomaideng/eventory/services/payment/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		panic(err)
	}
	logger.SetDefault(cfg.Logging, cfg.Environment, "payment")

	db, err := database.ConnectPostgres(cfg.DBDSN)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}
	app := server.NewApp(db, cfg)
	httpServer := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: app.HTTP, ReadHeaderTimeout: 10 * time.Second}
	slog.Info("payment API listening", "addr", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		slog.Error("HTTP server failed", "error", err)
		panic(err)
	}
}
