package accounts

import (
	"context"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

var app *apptest.App

func TestAccountsFeature(t *testing.T) {
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
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run account feature tests")
	}
}
