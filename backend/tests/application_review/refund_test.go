package application_review

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	paymentmodels "github.com/moomaideng/eventory/services/payment/models"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

func TestPaidRegistrationRefundedAfterRejection(t *testing.T) {
	pg := apptest.StartPostgres(context.Background(), t)
	app = apptest.NewApp(t, pg.DSN)
	for _, mode := range []models.TournamentRegistrationMode{models.TournamentRegistrationModeTeam, models.TournamentRegistrationModeSolo} {
		t.Run(string(mode), func(t *testing.T) {
			s := &reviewScenarioContext{client: apptest.NewClient(app.TournamentBaseURL())}
			if err := s.createTournamentIn(mode, 8); err != nil {
				t.Fatal(err)
			}
			var team *models.TournamentTeam
			var err error
			if mode == models.TournamentRegistrationModeSolo {
				team, err = s.seedSoloTeam(models.TournamentTeamStatusLocked)
			} else {
				team, err = s.seedTeam(models.TournamentTeamStatusLocked)
			}
			if err != nil {
				t.Fatal(err)
			}
			s.team = team
			var member models.TournamentTeamMember
			if err := app.DB.Where("tournament_team_id = ?", team.ID).First(&member).Error; err != nil {
				t.Fatal(err)
			}
			paidAt := time.Now().UTC()
			invoice := paymentmodels.Invoice{
				ID: uuid.New(), PayerAccountID: member.AccountID, Source: "TEAM_REGISTRATION",
				Reference: team.ID.String() + ":" + uuid.NewString(), IdempotencyKey: uuid.NewString(),
				Description: "Registration fee", AmountMinor: 15000, Currency: "THB", Provider: "MOCK",
				ExpiresAt: paidAt.Add(time.Hour), PaidAt: &paidAt,
			}
			if err := app.DB.Create(&invoice).Error; err != nil {
				t.Fatal(err)
			}
			if err := app.DB.Model(team).Update("payment_id", invoice.ID).Error; err != nil {
				t.Fatal(err)
			}
			if err := s.review("REJECT", "Not eligible"); err != nil || s.resp.StatusCode != http.StatusOK {
				t.Fatalf("review failed: %v, response: %+v", err, s.resp)
			}
			if err := app.DB.First(&invoice, "id = ?", invoice.ID).Error; err != nil || invoice.Status(time.Now()) != paymentmodels.Refunded {
				t.Fatalf("payment was not refunded: status=%s err=%v", invoice.Status(time.Now()), err)
			}
			var stored models.TournamentTeam
			if err := app.DB.First(&stored, "id = ?", team.ID).Error; err != nil || stored.RefundedAt == nil {
				t.Fatalf("rejected entry was not marked refunded: %+v err=%v", stored, err)
			}
		})
	}
}
