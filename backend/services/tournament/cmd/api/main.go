package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/moomaideng/eventory/internal/database"
	"github.com/moomaideng/eventory/services/tournament/config"
	"github.com/moomaideng/eventory/services/tournament/server"
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

	app := server.NewApp(db, appConfig)

	fmt.Printf("Tournament service starting on HTTP port %s (env: %s)...\n", appConfig.HTTPPort, appConfig.Environment)
	fmt.Printf("API Documentation available at http://localhost:%s/docs\n", appConfig.HTTPPort)

	if err := http.ListenAndServe(":"+appConfig.HTTPPort, app.HTTP); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
