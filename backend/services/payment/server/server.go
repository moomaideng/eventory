package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/internal/grpcx"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/telemetry"
	"github.com/moomaideng/eventory/services/payment/config"
	protohandlers "github.com/moomaideng/eventory/services/payment/handlers/proto"
	rest "github.com/moomaideng/eventory/services/payment/handlers/rest"
	"github.com/moomaideng/eventory/services/payment/internal/repositories"
	"github.com/moomaideng/eventory/services/payment/internal/usecases"
	paymentv1 "github.com/moomaideng/eventory/services/payment/pkg/proto/payment/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type App struct {
	HTTP http.Handler
	GRPC *grpc.Server
}

func NewApp(db *gorm.DB, cfg config.Config) *App {
	router := chi.NewMux()
	router.Use(telemetry.HTTPMiddleware(config.ServiceName))
	router.Use(middlewares.RequestID)
	router.Use(middlewares.AccessLog(cfg.Logging.SlowRequestMS))
	router.Use(middlewares.Recover)
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
		"bearer": {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
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
	var eventPublisher events.Publisher = events.NoopPublisher{}
	if cfg.BrokerURL != "" {
		if pub, err := events.NewRabbitMQPublisher(cfg.BrokerURL); err == nil {
			eventPublisher = pub
		} else {
			slog.Warn("failed to initialize rabbitmq publisher, using noop", "error", err)
		}
	}
	uc := usecases.NewPaymentUseCase(repositories.NewPaymentRepository(db), eventPublisher)
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()), grpc.ChainUnaryInterceptor(grpcx.UnaryServerLogging(cfg.Logging.SlowRequestMS), protohandlers.ServiceAuth(cfg.ServiceKey)))
	paymentv1.RegisterPaymentServiceServer(grpcServer, protohandlers.NewPaymentServer(uc))
	group := huma.NewGroup(api, "/api/v1/payments")
	group.UseMiddleware(middlewares.NewAuthMiddleware(api, cfg.SupabaseURL, cfg.Environment).HumaMiddleware())
	rest.RegisterPaymentRoutes(group, uc)
	return &App{HTTP: router, GRPC: grpcServer}
}
