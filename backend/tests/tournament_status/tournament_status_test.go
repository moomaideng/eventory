package tournament_status

import (
	"context"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

var app *apptest.App

func TestTournamentStatusFeature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pg := apptest.StartPostgres(ctx, t)
	app = apptest.NewApp(t, pg.DSN)

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
			Strict:   true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("tournament status override feature tests failed")
	}
}
