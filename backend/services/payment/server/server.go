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
	"github.com/moomaideng/eventory/services/payment/config"
	rest "github.com/moomaideng/eventory/services/payment/handlers/rest"
	"github.com/moomaideng/eventory/services/payment/internal/repositories"
	"github.com/moomaideng/eventory/services/payment/internal/usecases"
	"gorm.io/gorm"
)

type App struct{ HTTP http.Handler }

func NewApp(db *gorm.DB, cfg config.Config) *App {
	router := chi.NewMux()
	router.Use(middleware.Logger, middleware.Recoverer)
	// Checkout and history must not be cached by shared proxies.
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	})
	if len(cfg.CORSOrigins) > 0 {
		router.Use(cors.Handler(cors.Options{AllowedOrigins: cfg.CORSOrigins, AllowedMethods: []string{"GET", "POST", "OPTIONS"}, AllowedHeaders: []string{"Authorization", "Content-Type"}, MaxAge: 300}))
	}
	humaConfig := huma.DefaultConfig("Eventory Payment API (mock)", "1.0.0")
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer":     {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
		"serviceKey": {Type: "apiKey", In: "header", Name: "X-Payment-Service-Key", Description: "Server-only secret; never send this from a browser"},
	}
	api := humachi.New(router, humaConfig)
	huma.Register(api, huma.Operation{OperationID: "payment-health", Method: http.MethodGet, Path: "/health"}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, huma.Error503ServiceUnavailable("Database unreachable")
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			return nil, huma.Error503ServiceUnavailable("Database unreachable")
		}
		return nil, nil
	})
	uc := usecases.NewPaymentUseCase(repositories.NewPaymentRepository(db))
	rest.RegisterInternalRoutes(api, uc, cfg.ServiceKey)
	group := huma.NewGroup(api, "/api/v1/payments")
	group.UseMiddleware(middlewares.NewAuthMiddleware(api, cfg.SupabaseURL, cfg.Environment).HumaMiddleware())
	rest.RegisterPaymentRoutes(group, uc)
	return &App{HTTP: router}
}
