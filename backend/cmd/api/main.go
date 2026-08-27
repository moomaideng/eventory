package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	appconfig "github.com/moomaideng/eventory/pkg/config"
)

func main() {
	// 1. Load Configuration
	appConfig, err := appconfig.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	port := appConfig.Port

	// 2. Initialize Database (To be implemented in pkg/database)
	// db := database.ConnectPostgres(viper.GetString("DB_DSN"))

	// Execute GORM AutoMigrate for your models here
	// db.AutoMigrate(&models.Account{}, &models.Profile{})

	// 3. Initialize Router & API Framework
	router := chi.NewMux()

	// Create the Huma API instance
	config := huma.DefaultConfig("Project API", "1.0.0")
	api := humachi.New(router, config)

	// 4. Register Healthcheck Endpoint
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health Check",
		Description: "Returns a 204 No Content status if the server is running.",
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		return nil, nil // Nil response translates to HTTP 204
	})

	// 5. Wire Handlers & Use Cases (To be implemented)
	// accountRepo := repositories.NewAccountRepository(db)
	// accountUseCase := usecases.NewAccountUseCase(accountRepo)
	// handlers.RegisterAccountRoutes(api, accountUseCase)

	// 6. Start Server
	fmt.Printf("Server starting on port %s...\n", port)
	fmt.Printf("API Documentation available at http://localhost:%s/docs\n", port)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
