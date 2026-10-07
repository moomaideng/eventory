package apptest

import (
	"net"
	"net/http/httptest"
	"testing"

	accountconfig "github.com/moomaideng/eventory/services/account/config"
	accountserver "github.com/moomaideng/eventory/services/account/server"
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
	AccountServer    *httptest.Server
	TournamentServer *httptest.Server
	DB               *gorm.DB
}

// NewApp wires both microservices with tournament calling account over in-process gRPC.
func NewApp(tb testing.TB, dsn string) *App {
	tb.Helper()

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

	tournamentApp := tournamentserver.NewAppWithAccountConn(db, tournamentconfig.Config{
		Environment: "development",
		SupabaseURL: jwksServer.URL(),
	}, conn)
	tournamentTS := httptest.NewServer(tournamentApp.HTTP)
	tb.Cleanup(tournamentTS.Close)

	return &App{
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
