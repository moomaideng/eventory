package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/tournament/config"
	"github.com/moomaideng/eventory/services/tournament/server"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	appConfig, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := database.ConnectPostgres(appConfig.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("Database connection established successfully.")

	var mongoDB *mongo.Database
	if appConfig.MongoURI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		mongoClient, err := database.ConnectMongo(ctx, appConfig.MongoURI)
		if err != nil {
			log.Fatalf("failed to connect to mongodb: %v", err)
		}
		defer func() {
			_ = mongoClient.Disconnect(context.Background())
		}()
		log.Println("MongoDB connection established successfully.")
		dbName := appConfig.MongoDBName
		if dbName == "" {
			dbName = "tournament_db"
		}
		mongoDB = mongoClient.Database(dbName)
	}

	app := server.NewApp(db, mongoDB, appConfig)

	fmt.Printf("Tournament service starting on HTTP port %s (env: %s)...\n", appConfig.HTTPPort, appConfig.Environment)
	fmt.Printf("API Documentation available at http://localhost:%s/docs\n", appConfig.HTTPPort)

	if err := http.ListenAndServe(":"+appConfig.HTTPPort, app.HTTP); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
