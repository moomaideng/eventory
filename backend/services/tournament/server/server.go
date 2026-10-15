package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/config"
	resthandlers "github.com/moomaideng/eventory/services/tournament/handlers/rest"
	"github.com/moomaideng/eventory/services/tournament/internal/adapters/accountgrpc"
	"github.com/moomaideng/eventory/services/tournament/internal/adapters/paymenthttp"
	tournamentmiddlewares "github.com/moomaideng/eventory/services/tournament/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"

	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type App struct {
	HTTP           http.Handler
	RunSubmissions func(context.Context)
}

func NewApp(db *gorm.DB, mongoDB *mongo.Database, cfg config.Config) *App {
	accountSvc, err := accountgrpc.NewAccountClient(cfg.AccountGRPCAddr, cfg.Logging.SlowRequestMS)
	if err != nil {
		slog.Error("failed to create account gRPC client", "error", err)
		panic(err)
	}
	return newApp(db, mongoDB, cfg, accountSvc)
}

// NewAppWithAccountConn wires the HTTP API using an existing Account gRPC connection (tests).
func NewAppWithAccountConn(db *gorm.DB, mongoDB *mongo.Database, cfg config.Config, conn grpc.ClientConnInterface) *App {
	return newApp(db, mongoDB, cfg, accountgrpc.NewAccountClientFromConn(conn))
}

func newApp(db *gorm.DB, mongoDB *mongo.Database, cfg config.Config, accountSvc ports.AccountService) *App {
	payments := paymenthttp.New(cfg.PaymentURL, cfg.PaymentServiceKey)
	submission := usecases.NewTeamSubmissionUseCase(repositories.NewTeamSubmissionRepository(db), payments)
	return &App{HTTP: newHTTPRouter(db, mongoDB, cfg, accountSvc, submission), RunSubmissions: func(ctx context.Context) { submission.Run(ctx) }}
}

func newHTTPRouter(db *gorm.DB, mongoDB *mongo.Database, cfg config.Config, accountSvc ports.AccountService, submission *usecases.TeamSubmissionUseCase) http.Handler {
	router := chi.NewMux()
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

	humaConfig := huma.DefaultConfig("Eventory Tournament API", "1.0.0")
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
		if err != nil || sqlDB.PingContext(ctx) != nil {
			return nil, huma.Error500InternalServerError("Postgres database unreachable", err)
		}
		if mongoDB != nil {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			if err := mongoDB.Client().Ping(pingCtx, nil); err != nil {
				return nil, huma.Error500InternalServerError("MongoDB unreachable", err)
			}
		}
		return nil, nil
	})

	tournamentRepo := repositories.NewTournamentRepository(db)
	teamLobbyRepo := repositories.NewTeamLobbyRepository(db)
	organizerDashboardRepo := repositories.NewOrganizerDashboardRepository(db)
	tournamentStatusRepo := repositories.NewTournamentStatusRepository(db)
	tournamentStaffRepo := repositories.NewTournamentStaffRepository(db)

	tournamentUseCase := usecases.NewTournamentUseCase(tournamentRepo, accountSvc)
	teamLobbyUseCase := usecases.NewTeamLobbyUseCase(teamLobbyRepo, accountSvc).WithSubmission(submission)
	dashboardUseCase := usecases.NewOrganizerDashboardUseCase(organizerDashboardRepo, accountSvc)
	statusUseCase := usecases.NewTournamentStatusUseCase(tournamentStatusRepo, accountSvc)
	staffUseCase := usecases.NewTournamentStaffUseCase(tournamentStaffRepo, accountSvc)

	authMiddleware := middlewares.NewAuthMiddleware(api, cfg.SupabaseURL, cfg.Environment)
	organizerMiddleware := tournamentmiddlewares.NewOrganizerMiddleware(api, accountSvc)

	auth := authMiddleware.HumaMiddleware()
	organizer := organizerMiddleware.HumaMiddleware()

	resthandlers.RegisterTournamentRoutes(api, auth, organizer, tournamentUseCase)
	resthandlers.RegisterTeamLobbyRoutes(api, auth, teamLobbyUseCase)
	resthandlers.RegisterTournamentStaffRoutes(api, auth, staffUseCase)
	resthandlers.RegisterOrganizerDashboardRoutes(api, auth, dashboardUseCase)
	resthandlers.RegisterTournamentStatusRoutes(api, auth, organizer, statusUseCase)

	if mongoDB != nil {
		formRepo := repositories.NewRegistrationFormRepository(mongoDB)
		if err := formRepo.EnsureIndexes(context.Background()); err != nil {
			slog.Error("failed to ensure registration form indexes", "error", err)
		}
		fieldRegistry := forms.NewDefaultRegistry()
		formUseCase := usecases.NewRegistrationFormUseCase(tournamentRepo, formRepo, fieldRegistry)
		resthandlers.RegisterRegistrationFormRoutes(api, auth, organizer, formUseCase)
	}

	return router
}
