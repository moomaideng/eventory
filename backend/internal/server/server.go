package server

import (
	"context"
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
	"gorm.io/gorm"
)

// NewRouter builds and configures the complete HTTP router, Huma API, middlewares, and routes.
func NewRouter(db *gorm.DB, cfg appconfig.Config) http.Handler {
	router := chi.NewMux()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Configure CORS
	if len(cfg.CORSOrigins) > 0 {
		router.Use(cors.Handler(cors.Options{
			AllowedOrigins: cfg.CORSOrigins,
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
	}

	// Configure Huma API with Bearer JWT security scheme
	humaConfig := huma.DefaultConfig("Eventory API", "1.0.0")
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Supabase Auth JWT Token (or 'Bearer dev-token' for local offline development)",
		},
	}
	api := humachi.New(router, humaConfig)

	// Public Healthcheck Endpoint
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health Check",
		Description: "Returns a 204 No Content status if the server is running.",
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			return nil, huma.Error500InternalServerError("Database unreachable", err)
		}
		return nil, nil
	})

	// Repositories & Use Cases
	accountRepo := repositories.NewAccountRepository(db)
	accountUseCase := usecases.NewAccountUseCase(accountRepo)
	tournamentRepo := repositories.NewTournamentRepository(db)
	tournamentUseCase := usecases.NewTournamentUseCase(tournamentRepo)
	teamLobbyRepo := repositories.NewTeamLobbyRepository(db)
	teamLobbyUseCase := usecases.NewTeamLobbyUseCase(teamLobbyRepo)
	dashboardUseCase := usecases.NewOrganizerDashboardUseCase(repositories.NewOrganizerDashboardRepository(db))

	// Middlewares
	authMiddleware := middlewares.NewAuthMiddleware(api, cfg.SupabaseURL, cfg.Environment)
	organizerMiddleware := middlewares.NewOrganizerMiddleware(api, accountRepo)

	// Scoped Route Groups
	accountGroup := huma.NewGroup(api, "/api/v1/accounts")
	accountGroup.UseMiddleware(authMiddleware.HumaMiddleware())

	teamLobbyGroup := huma.NewGroup(api, "/api/v1")
	teamLobbyGroup.UseMiddleware(authMiddleware.HumaMiddleware())

	organizerGroup := huma.NewGroup(api, "/api/v1")
	organizerGroup.UseMiddleware(authMiddleware.HumaMiddleware())

	organizerTournamentGroup := huma.NewGroup(api, "/api/v1/tournaments")
	organizerTournamentGroup.UseMiddleware(authMiddleware.HumaMiddleware(), organizerMiddleware.HumaMiddleware())

	// Handlers
	handlers.RegisterAccountRoutes(accountGroup, accountUseCase)
	handlers.RegisterTeamLobbyRoutes(teamLobbyGroup, teamLobbyUseCase, accountUseCase)
	handlers.RegisterOrganizerDashboardRoutes(organizerGroup, dashboardUseCase, accountUseCase)
	handlers.RegisterTournamentRoutes(api, organizerTournamentGroup, tournamentUseCase)

	return router
}
