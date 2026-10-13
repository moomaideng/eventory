package main

import (
	"log/slog"
	"net"
	"net/http"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/internal/logger"
	"github.com/moomaideng/eventory/services/account/config"
	"github.com/moomaideng/eventory/services/account/server"
)

func main() {
	appConfig, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		panic(err)
	}
	logger.SetDefault(appConfig.Logging, appConfig.Environment, "account")

	db, err := database.ConnectPostgres(appConfig.DBDSN)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}
	slog.Info("database connection established")

	app := server.NewApp(db, appConfig)

	go func() {
		lis, err := net.Listen("tcp", ":"+appConfig.GRPCPort)
		if err != nil {
			slog.Error("failed to listen on gRPC port", "port", appConfig.GRPCPort, "error", err)
			panic(err)
		}
		slog.Info("account gRPC listening", "port", appConfig.GRPCPort)
		if err := app.GRPC.Serve(lis); err != nil {
			slog.Error("account gRPC failed", "error", err)
			panic(err)
		}
	}()

	slog.Info("account HTTP starting", "port", appConfig.HTTPPort, "env", appConfig.Environment)
	slog.Info("API docs available", "url", "http://localhost:"+appConfig.HTTPPort+"/docs")

	if err := http.ListenAndServe(":"+appConfig.HTTPPort, app.HTTP); err != nil {
		slog.Error("HTTP server failed", "error", err)
		panic(err)
	}
}
