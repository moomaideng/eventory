package server

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/account/config"
	protohandlers "github.com/moomaideng/eventory/services/account/handlers/proto"
	resthandlers "github.com/moomaideng/eventory/services/account/handlers/rest"
	"github.com/moomaideng/eventory/services/account/internal/repositories"
	"github.com/moomaideng/eventory/services/account/internal/usecases"
	accountv1 "github.com/moomaideng/eventory/services/account/pkg/proto/account/v1"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type App struct {
	HTTP http.Handler
	GRPC *grpc.Server
}

func NewApp(db *gorm.DB, cfg config.Config) *App {
	accountRepo := repositories.NewAccountRepository(db)
	accountUseCase := usecases.NewAccountUseCase(accountRepo)

	grpcServer := grpc.NewServer()
	accountv1.RegisterAccountServiceServer(grpcServer, protohandlers.NewAccountServer(accountUseCase))

	return &App{
		HTTP: newHTTPRouter(db, cfg, accountUseCase),
		GRPC: grpcServer,
	}
}

func newHTTPRouter(db *gorm.DB, cfg config.Config, accountUseCase *usecases.AccountUseCase) http.Handler {
	router := chi.NewMux()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

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

	humaConfig := huma.DefaultConfig("Eventory Account API", "1.0.0")
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "Supabase Auth JWT Token (or 'Bearer dev-token' for local offline development)",
		},
	}
	api := humachi.New(router, humaConfig)

	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health Check",
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			return nil, huma.Error500InternalServerError("Database unreachable", err)
		}
		return nil, nil
	})

	authMiddleware := middlewares.NewAuthMiddleware(api, cfg.SupabaseURL, cfg.Environment)
	accountGroup := huma.NewGroup(api, "/api/v1/accounts")
	accountGroup.UseMiddleware(authMiddleware.HumaMiddleware())
	resthandlers.RegisterAccountRoutes(accountGroup, accountUseCase)

	return router
}
