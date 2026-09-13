package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/moomaideng/eventory/internal/server"
	appconfig "github.com/moomaideng/eventory/pkg/config"
	"github.com/moomaideng/eventory/pkg/database"
)

func main() {
	// 1. Load Configuration
	appConfig, err := appconfig.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	port := appConfig.Port

	// 2. Initialize Database
	db, err := database.ConnectPostgres(appConfig.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	log.Println("Database connection established successfully.")

	// 3. Build HTTP Router via shared server package
	router := server.NewRouter(db, appConfig)

	// 4. Start Server
	fmt.Printf("Server starting on port %s (env: %s)...\n", port, appConfig.Environment)
	fmt.Printf("API Documentation available at http://localhost:%s/docs\n", port)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
