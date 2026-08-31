package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/moomaideng/eventory/internal/handlers"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/repositories"
	"github.com/moomaideng/eventory/internal/usecases"
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

	// 3. Initialize Router & API Framework
	router := chi.NewMux()

	// Create the Huma API instance with Bearer JWT security scheme definition
	config := huma.DefaultConfig("Eventory API", "1.0.0")
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Supabase Auth JWT Token (or 'Bearer dev-token' for local offline development)",
		},
	}
	api := humachi.New(router, config)

	// 4. Attach Global Auth Middleware with Official Supabase JWKS Verification & Environment Guard
	authMiddleware := middlewares.NewAuthMiddleware(api, appConfig.SupabaseURL, appConfig.Environment)
	api.UseMiddleware(authMiddleware.HumaMiddleware())

	// 5. Register Healthcheck Endpoint
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health Check",
		Description: "Returns a 204 No Content status if the server is running.",
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		sqlDB, _ := db.DB()
		if err := sqlDB.Ping(); err != nil {
			return nil, huma.Error500InternalServerError("Database unreachable", err)
		}
		return nil, nil // Nil response translates to HTTP 204
	})

	// 6. Wire Handlers & Use Cases
	accountRepo := repositories.NewAccountRepository(db)
	accountUseCase := usecases.NewAccountUseCase(accountRepo)
	handlers.RegisterAccountRoutes(api, accountUseCase)

	// 7. Start Server
	fmt.Printf("Server starting on port %s (env: %s)...\n", port, appConfig.Environment)
	fmt.Printf("API Documentation available at http://localhost:%s/docs\n", port)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
