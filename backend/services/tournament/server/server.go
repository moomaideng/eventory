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
	"github.com/moomaideng/eventory/internal/telemetry"
	"github.com/moomaideng/eventory/services/tournament/config"
	resthandlers "github.com/moomaideng/eventory/services/tournament/handlers/rest"
	"github.com/moomaideng/eventory/services/tournament/internal/adapters/accountgrpc"
	"github.com/moomaideng/eventory/services/tournament/internal/adapters/paymenthttp"
	tournamentmiddlewares "github.com/moomaideng/eventory/services/tournament/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"

	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
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
	teamLobbyRepo := repositories.NewTeamLobbyRepository(db)
	var formRepo repositories.RegistrationFormRepository
	var fieldRegistry *forms.FieldRegistry
	if mongoDB != nil {
		formRepo = repositories.NewRegistrationFormRepository(mongoDB)
		if err := formRepo.EnsureIndexes(context.Background()); err != nil {
			slog.Error("failed to ensure registration form indexes", "error", err)
		}
		fieldRegistry = forms.NewDefaultRegistry()
	}
	teamLobbyUseCase := usecases.NewTeamLobbyUseCase(teamLobbyRepo, accountSvc, payments, formRepo, fieldRegistry)
	sponsorships := usecases.NewSponsorshipUseCase(repositories.NewSponsorshipRepository(db), repositories.NewTournamentRepository(db), accountSvc, payments)
	teamReviews := usecases.NewTeamReviewUseCase(repositories.NewTeamReviewRepository(db), accountSvc, payments)
	return &App{
		HTTP: newHTTPRouter(db, mongoDB, cfg, accountSvc, teamLobbyUseCase, sponsorships, teamReviews),
		RunSubmissions: func(ctx context.Context) {
			go sponsorshipPoller(sponsorships).run(ctx)
			go refundPoller(teamReviews).run(ctx)
			teamSubmissionPoller(teamLobbyUseCase).run(ctx)
		},
	}
}

const reconcileInterval = 10 * time.Second

func teamSubmissionPoller(u *usecases.TeamLobbyUseCase) poller[models.TournamentTeam] {
	return poller[models.TournamentTeam]{
		interval:    reconcileInterval,
		spanName:    "reconcile pending team submissions",
		listPending: u.Pending,
		reconcile:   u.Reconcile,
		onListError: func(ctx context.Context, err error) {
			slog.ErrorContext(ctx, "list pending team submissions", "error", err)
		},
		onReconcileError: func(ctx context.Context, team *models.TournamentTeam, err error) {
			slog.ErrorContext(ctx, "reconcile team submission", "team_id", team.ID, "error", err)
		},
	}
}

func sponsorshipPoller(u *usecases.SponsorshipUseCase) poller[models.SponsorshipPledge] {
	return poller[models.SponsorshipPledge]{
		interval:    reconcileInterval,
		spanName:    "reconcile pending sponsorships",
		listPending: u.Pending,
		reconcile:   u.Reconcile,
		onListError: func(ctx context.Context, err error) {
			slog.ErrorContext(ctx, "list pending sponsorships", "error", err)
		},
		onReconcileError: func(ctx context.Context, pledge *models.SponsorshipPledge, err error) {
			slog.WarnContext(ctx, "reconcile sponsorship", "pledge", pledge.ID, "error", err)
		},
	}
}

func refundPoller(u *usecases.TeamReviewUseCase) poller[models.TournamentTeam] {
	return poller[models.TournamentTeam]{
		interval:    reconcileInterval,
		spanName:    "retry rejected registration refunds",
		listPending: u.PendingRefunds,
		reconcile:   u.RetryRefund,
		onListError: func(ctx context.Context, err error) {
			slog.ErrorContext(ctx, "list rejected registration refunds", "error", err)
		},
		onReconcileError: func(ctx context.Context, team *models.TournamentTeam, err error) {
			slog.WarnContext(ctx, "retry rejected registration refund", "entry_id", team.ID, "error", err)
		},
	}
}

func newHTTPRouter(db *gorm.DB, mongoDB *mongo.Database, cfg config.Config, accountSvc ports.AccountService, teamLobbyUseCase *usecases.TeamLobbyUseCase, sponsorships *usecases.SponsorshipUseCase, teamReviewUseCase *usecases.TeamReviewUseCase) http.Handler {
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
	tournamentStaffRepo := repositories.NewTournamentStaffRepository(db)
	sponsorshipTierRepo := repositories.NewSponsorshipTierRepository(db)
	bracketRepo := repositories.NewBracketRepository(db)

	tournamentUseCase := usecases.NewTournamentUseCase(tournamentRepo, accountSvc)
	staffUseCase := usecases.NewTournamentStaffUseCase(tournamentStaffRepo, accountSvc)
	sponsorshipTierUseCase := usecases.NewSponsorshipTierUseCase(sponsorshipTierRepo, tournamentRepo)
	bracketUseCase := usecases.NewBracketUseCase(bracketRepo)

	authMiddleware := middlewares.NewAuthMiddleware(api, cfg.SupabaseURL, cfg.Environment)
	organizerMiddleware := tournamentmiddlewares.NewOrganizerMiddleware(api, accountSvc)

	auth := authMiddleware.HumaMiddleware()
	organizer := organizerMiddleware.HumaMiddleware()

	resthandlers.RegisterTournamentRoutes(api, auth, organizer, tournamentUseCase)
	resthandlers.RegisterTeamLobbyRoutes(api, auth, teamLobbyUseCase)
	resthandlers.RegisterTournamentStaffRoutes(api, auth, staffUseCase)
	resthandlers.RegisterOrganizerDashboardRoutes(api, auth, tournamentUseCase)
	resthandlers.RegisterTournamentStatusRoutes(api, auth, organizer, tournamentUseCase)
	resthandlers.RegisterSponsorshipTierRoutes(api, auth, organizer, sponsorshipTierUseCase)
	resthandlers.RegisterTeamReviewRoutes(api, auth, teamReviewUseCase)
	resthandlers.RegisterBracketRoutes(api, auth, bracketUseCase)
	resthandlers.RegisterSponsorshipRoutes(api, auth, sponsorships)

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
