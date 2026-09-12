package apptest

import (
	"net/http/httptest"
	"testing"

	"github.com/moomaideng/eventory/internal/server"
	appconfig "github.com/moomaideng/eventory/pkg/config"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const basePath = "/api/v1"

// App hosts the in-process HTTP server wired with real GORM repositories and handlers.
type App struct {
	Server *httptest.Server
	DB     *gorm.DB
}

// NewApp instantiates the real HTTP application stack against the provided database DSN.
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

	cfg := appconfig.Config{
		Environment: "development",
		SupabaseURL: jwksServer.URL(),
	}

	router := server.NewRouter(db, cfg)

	ts := httptest.NewServer(router)
	tb.Cleanup(ts.Close)

	return &App{
		Server: ts,
		DB:     db,
	}
}

// BaseURL returns the root versioned API URL (e.g. http://127.0.0.1:port/api/v1).
func (a *App) BaseURL() string {
	return a.Server.URL + basePath
}
