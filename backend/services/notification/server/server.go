package server

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/telemetry"
	"github.com/moomaideng/eventory/services/notification/config"
	"github.com/moomaideng/eventory/services/notification/handlers/rest"
	"github.com/moomaideng/eventory/services/notification/internal/usecases"
	"gorm.io/gorm"
)

type BrokerHealth interface{ Ready() bool }

type App struct{ HTTP http.Handler }

func NewApp(db *gorm.DB, broker BrokerHealth, uc *usecases.InAppNotificationUseCase, cfg config.Config) *App {
	router := chi.NewMux()
	router.Use(telemetry.HTTPMiddleware(config.ServiceName))
	router.Use(middlewares.RequestID)
	router.Use(middlewares.AccessLog(cfg.Logging.SlowRequestMS))
	router.Use(middlewares.Recover)

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

	// Plain health & readiness checks for Docker / Compose / K8s probes
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		if !databaseReady(r.Context(), db) || (broker != nil && !broker.Ready()) {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	humaConfig := huma.DefaultConfig("Eventory Notification API", "1.0.0")
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
	}

	api := humachi.New(router, humaConfig)

	// User-facing routes
	group := huma.NewGroup(api, "/api/v1/notifications")
	group.UseMiddleware(middlewares.NewAuthMiddleware(api, cfg.SupabaseURL, cfg.Environment).HumaMiddleware())
	rest.RegisterNotificationRoutes(group, uc)

	return &App{HTTP: router}
}

func databaseReady(ctx context.Context, db *gorm.DB) bool {
	sqlDB, err := db.DB()
	return err == nil && sqlDB.PingContext(ctx) == nil
}
