package sponsorship

import (
	"context"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

func TestSponsorshipTiersFeature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	app := apptest.NewApp(t, pg.DSN)
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) { initializeScenario(sc, app) },
		Options:             &godog.Options{Format: "pretty", Paths: []string{"features"}, TestingT: t, Strict: true},
	}
	if suite.Run() != 0 {
		t.Fatal("sponsorship tier feature tests failed")
	}
}
