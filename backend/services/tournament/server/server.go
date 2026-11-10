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
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/telemetry"
	"github.com/moomaideng/eventory/services/tournament/config"
	"github.com/moomaideng/eventory/services/tournament/handlers/messaging"
	resthandlers "github.com/moomaideng/eventory/services/tournament/handlers/rest"
	"github.com/moomaideng/eventory/services/tournament/internal/adapters/accountgrpc"
	"github.com/moomaideng/eventory/services/tournament/internal/adapters/paymentgrpc"
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
	Close         func() error
	HTTP          http.Handler
	RunBackground func(context.Context)
}

func NewApp(db *gorm.DB, mongoDB *mongo.Database, cfg config.Config) *App {
	accountSvc, err := accountgrpc.NewAccountClient(cfg.AccountGRPCAddr, cfg.Logging.SlowRequestMS)
	if err != nil {
		slog.Error("failed to create account gRPC client", "error", err)
		panic(err)
	}
	publisher, err := events.NewRabbitMQPublisher(cfg.BrokerURL)
	if err != nil {
		slog.Error("failed to create event publisher", "error", err)
		panic(err)
	}
	return newApp(db, mongoDB, cfg, accountSvc, publisher)
}

func NewAppForTest(db *gorm.DB, mongoDB *mongo.Database, cfg config.Config, conn grpc.ClientConnInterface, publisher events.Publisher) *App {
	return newApp(db, mongoDB, cfg, accountgrpc.NewAccountClientFromConn(conn), publisher)
}

func newApp(db *gorm.DB, mongoDB *mongo.Database, cfg config.Config, accountSvc ports.AccountService, eventPublisher events.Publisher) *App {
	payments, err := paymentgrpc.New(cfg.PaymentGRPCAddr, cfg.PaymentServiceKey, cfg.Logging.SlowRequestMS)
	if err != nil {
		panic(err)
	}
	tournamentRepo := repositories.NewTournamentRepository(db)
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
	sponsorships := usecases.NewSponsorshipUseCase(repositories.NewSponsorshipRepository(db), tournamentRepo, accountSvc, payments)

	teamReviews := usecases.NewTeamReviewUseCase(repositories.NewTeamReviewRepository(db), accountSvc, payments, eventPublisher)
	announcements := usecases.NewAnnouncementUseCase(repositories.NewAnnouncementRepository(db), accountSvc, eventPublisher)
	return &App{
		Close: payments.Close,
		HTTP:  newHTTPRouter(db, mongoDB, cfg, accountSvc, tournamentRepo, formRepo, fieldRegistry, teamLobbyUseCase, sponsorships, teamReviews, announcements),
		RunBackground: func(ctx context.Context) {
			go messaging.PaymentEventConsumer(cfg.BrokerURL).Run(ctx, messaging.PaymentEventHandler(teamLobbyUseCase, sponsorships))
			refundPoller(teamReviews).run(ctx)
		},
	}
}

// Rejected-registration refunds are retried by their own loop: a failed refund
// call to Payment is Tournament's to repeat, and no payment event announces it.
const refundRetryInterval = 10 * time.Second

func refundPoller(u *usecases.TeamReviewUseCase) poller[models.TournamentTeam] {
	return poller[models.TournamentTeam]{
		interval:    refundRetryInterval,
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

func newHTTPRouter(
	db *gorm.DB,
	mongoDB *mongo.Database,
	cfg config.Config,
	accountSvc ports.AccountService,
	tournamentRepo repositories.TournamentRepository,
	formRepo repositories.RegistrationFormRepository,
	fieldRegistry *forms.FieldRegistry,
	teamLobbyUseCase *usecases.TeamLobbyUseCase,
	sponsorships *usecases.SponsorshipUseCase,
	teamReviewUseCase *usecases.TeamReviewUseCase,
	announcementUseCase *usecases.AnnouncementUseCase,
) http.Handler {
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
	resthandlers.RegisterAnnouncementRoutes(api, auth, announcementUseCase)

	if formRepo != nil && fieldRegistry != nil {
		formUseCase := usecases.NewRegistrationFormUseCase(tournamentRepo, formRepo, fieldRegistry)
		resthandlers.RegisterRegistrationFormRoutes(api, auth, organizer, formUseCase)
	}

	return router
}
