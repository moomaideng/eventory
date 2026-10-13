package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/internal/logger"
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
	logger.SetDefault(appConfig.Logging, appConfig.Environment, "tournament")

	db, err := database.ConnectPostgres(appConfig.DBDSN)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		panic(err)
	}
	slog.Info("database connection established")

	var mongoDB *mongo.Database
	if appConfig.MongoURI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		mongoClient, err := database.ConnectMongo(ctx, appConfig.MongoURI)
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
	go app.RunSubmissions(context.Background())

	slog.Info("tournament HTTP starting", "port", appConfig.HTTPPort, "env", appConfig.Environment)
	slog.Info("API docs available", "url", "http://localhost:"+appConfig.HTTPPort+"/docs")

	if err := http.ListenAndServe(":"+appConfig.HTTPPort, app.HTTP); err != nil {
		slog.Error("HTTP server failed", "error", err)
		panic(err)
	}
}
