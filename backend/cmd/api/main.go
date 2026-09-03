package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	// 3. Initialize Router & Standard Chi Middlewares
	router := chi.NewMux()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Configure CORS using standard go-chi/cors
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: appConfig.CORSOrigins,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 4. Create Huma API instance with Bearer JWT security scheme definition
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

	// 5. Register Public Healthcheck Endpoint (Outside Auth Group)
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
		return nil, nil
	})

	// 6. Create Repositories & Use Cases
	accountRepo := repositories.NewAccountRepository(db)
	accountUseCase := usecases.NewAccountUseCase(accountRepo)
	tournamentRepo := repositories.NewTournamentRepository(db)
	tournamentUseCase := usecases.NewTournamentUseCase(tournamentRepo)

	// 7. Initialize Auth Middleware & Scoped Route Groups
	authMiddleware := middlewares.NewAuthMiddleware(api, appConfig.SupabaseURL, appConfig.Environment)

	// Create /api/v1/accounts Group with Auth Middleware
	accountGroup := huma.NewGroup(api, "/api/v1/accounts")
	accountGroup.UseMiddleware(authMiddleware.HumaMiddleware())

	// Register Account Handlers onto the scoped group
	handlers.RegisterAccountRoutes(accountGroup, accountUseCase)

	// Tournament discovery is public; joining and management will use authenticated routes.
	handlers.RegisterTournamentRoutes(api, tournamentUseCase)

	// 8. Start Server
	fmt.Printf("Server starting on port %s (env: %s)...\n", port, appConfig.Environment)
	fmt.Printf("API Documentation available at http://localhost:%s/docs\n", port)

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
