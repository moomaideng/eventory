package apptest

import (
	"context"
	"net"
	"net/http/httptest"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/moomaideng/eventory/internal/events"
	accountconfig "github.com/moomaideng/eventory/services/account/config"
	accountserver "github.com/moomaideng/eventory/services/account/server"
	paymentconfig "github.com/moomaideng/eventory/services/payment/config"
	paymentdb "github.com/moomaideng/eventory/services/payment/database"
	paymentserver "github.com/moomaideng/eventory/services/payment/server"
	tournamentconfig "github.com/moomaideng/eventory/services/tournament/config"
	tournamentserver "github.com/moomaideng/eventory/services/tournament/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const basePath = "/api/v1"

// App hosts in-process account and tournament HTTP servers against one Postgres DSN.
type App struct {
	PaymentServer    *httptest.Server
	AccountServer    *httptest.Server
	TournamentServer *httptest.Server
	DB               *gorm.DB
}

// NewApp wires both microservices with tournament calling account over in-process gRPC.
// Payment events are discarded, so no broker is needed.
func NewApp(tb testing.TB, dsn string, formDB ...*mongo.Database) *App {
	tb.Helper()
	return newApp(tb, dsn, "", formDB...)
}

// NewAppWithBroker is NewApp with Payment and Tournament connected to the
// broker at brokerURL, and their background work running: payment events reach
// Tournament as they would in production.
func NewAppWithBroker(tb testing.TB, dsn, brokerURL string, formDB ...*mongo.Database) *App {
	tb.Helper()
	return newApp(tb, dsn, brokerURL, formDB...)
}

func newApp(tb testing.TB, dsn, brokerURL string, formDB ...*mongo.Database) *App {
	tb.Helper()

	var paymentEvents, tournamentEvents events.Publisher = events.NoopPublisher{}, events.NoopPublisher{}
	if brokerURL != "" {
		for _, target := range []*events.Publisher{&paymentEvents, &tournamentEvents} {
			publisher, err := events.NewRabbitMQPublisher(brokerURL)
			if err != nil {
				tb.Fatalf("apptest: connect event publisher: %v", err)
			}
			tb.Cleanup(func() { _ = publisher.Close() })
			*target = publisher
		}
	}

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		tb.Fatalf("apptest: connect to postgres: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		tb.Cleanup(func() { _ = sqlDB.Close() })
	}

	jwksServer := StartMockJWKSServer()

	accountApp := accountserver.NewApp(db, accountconfig.Config{
		Environment: "development",
		SupabaseURL: jwksServer.URL(),
	})

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		tb.Fatalf("apptest: listen gRPC: %v", err)
	}
	go func() { _ = accountApp.GRPC.Serve(lis) }()
	tb.Cleanup(accountApp.GRPC.Stop)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		tb.Fatalf("apptest: dial account gRPC: %v", err)
	}
	tb.Cleanup(func() { _ = conn.Close() })

	accountTS := httptest.NewServer(accountApp.HTTP)
	tb.Cleanup(accountTS.Close)

	if err := paymentdb.AutoMigrate(db); err != nil {
		tb.Fatal(err)
	}
	paymentApp := paymentserver.NewApp(db, paymentconfig.Config{Environment: "development", SupabaseURL: jwksServer.URL(), ServiceKey: "test-team-payment-service-key-12345678"}, paymentEvents)
	paymentLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		tb.Fatal(err)
	}
	go func() { _ = paymentApp.GRPC.Serve(paymentLis) }()
	tb.Cleanup(paymentApp.GRPC.Stop)
	paymentTS := httptest.NewServer(paymentApp.HTTP)
	tb.Cleanup(paymentTS.Close)
	var mongoDB *mongo.Database
	if len(formDB) > 0 {
		mongoDB = formDB[0]
	}
	tournamentApp := tournamentserver.NewAppForTest(db, mongoDB, tournamentconfig.Config{
		Environment:     "development",
		SupabaseURL:     jwksServer.URL(),
		PaymentGRPCAddr: paymentLis.Addr().String(), PaymentServiceKey: "test-team-payment-service-key-12345678",
		BrokerURL: brokerURL,
	}, conn, tournamentEvents)
	if brokerURL != "" {
		background, stop := context.WithCancel(context.Background())
		tb.Cleanup(stop)
		go paymentApp.RunBackground(background)
		go tournamentApp.RunBackground(background)
	}
	tb.Cleanup(func() { _ = tournamentApp.Close() })
	tournamentTS := httptest.NewServer(tournamentApp.HTTP)
	tb.Cleanup(tournamentTS.Close)

	return &App{
		PaymentServer:    paymentTS,
		AccountServer:    accountTS,
		TournamentServer: tournamentTS,
		DB:               db,
	}
}

// TournamentBaseURL returns the tournament service versioned API root.
func (a *App) TournamentBaseURL() string {
	return a.TournamentServer.URL + basePath
}

// AccountBaseURL returns the account service versioned API root.
func (a *App) AccountBaseURL() string {
	return a.AccountServer.URL + basePath
}
