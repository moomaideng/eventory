package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/account/config"
	"github.com/moomaideng/eventory/services/account/models"
	"github.com/moomaideng/eventory/services/account/server"
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
	if err := db.AutoMigrate(models.All()...); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Println("Database connection established successfully.")

	app := server.NewApp(db, appConfig)

	go func() {
		lis, err := net.Listen("tcp", ":"+appConfig.GRPCPort)
		if err != nil {
			log.Fatalf("failed to listen on gRPC port %s: %v", appConfig.GRPCPort, err)
		}
		log.Printf("Account gRPC listening on :%s", appConfig.GRPCPort)
		if err := app.GRPC.Serve(lis); err != nil {
			log.Fatalf("Account gRPC failed: %v", err)
		}
	}()

	fmt.Printf("Account service starting on HTTP port %s (env: %s)...\n", appConfig.HTTPPort, appConfig.Environment)
	fmt.Printf("API Documentation available at http://localhost:%s/docs\n", appConfig.HTTPPort)

	if err := http.ListenAndServe(":"+appConfig.HTTPPort, app.HTTP); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
