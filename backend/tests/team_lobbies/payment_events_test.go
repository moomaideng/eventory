package team_lobbies

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	handlers "github.com/moomaideng/eventory/services/tournament/handlers/rest"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

// The captain never reloads the page: Payment's result alone must reach the
// roster, through the outbox, RabbitMQ and Tournament's consumer.
func TestPaymentResultSettlesRegistrationWithoutRefresh(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	broker := apptest.StartRabbitMQ(ctx, t)
	app = apptest.NewAppWithBroker(t, pg.DSN, broker)
	payments := apptest.NewClient(app.PaymentServer.URL + "/api/v1/payments")

	pendingLobby := func(t *testing.T) (*teamLobbyScenarioContext, uuid.UUID) {
		t.Helper()
		s := &teamLobbyScenarioContext{client: apptest.NewClient(app.TournamentBaseURL())}
		if err := s.aCaptainHasCreatedAFormingTeamLobbyWithAValidInviteCode(); err != nil {
			t.Fatal(err)
		}
		if err := app.DB.Model(s.tournament).Updates(map[string]any{"entry_fee_minor": 25000, "min_team_size": 1}).Error; err != nil {
			t.Fatal(err)
		}
		resp, err := s.client.Do(http.MethodPost, "/lobbies/"+s.lobby.ID.String()+"/lock", nil, s.captain.AuthHeaders())
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("lock: %v %v", err, resp)
		}
		var lobby handlers.TeamLobbyResponse
		if err := resp.JSON(&lobby); err != nil || lobby.Status != "PAYMENT_PENDING" || lobby.PaymentID == nil {
			t.Fatalf("lock result %+v: %v", lobby, err)
		}
		return s, *lobby.PaymentID
	}
	pay := func(t *testing.T, s *teamLobbyScenarioContext, paymentID uuid.UUID, result string) {
		t.Helper()
		resp, err := payments.Do(http.MethodPost, "/"+paymentID.String()+"/mock-result", map[string]string{"result": result}, s.captain.AuthHeaders())
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("mock-result %s: %v %v", result, err, resp)
		}
	}
	waitForStatus := func(t *testing.T, s *teamLobbyScenarioContext, want models.TournamentTeamStatus) {
		t.Helper()
		var team models.TournamentTeam
		for deadline := time.Now().Add(20 * time.Second); time.Now().Before(deadline); time.Sleep(200 * time.Millisecond) {
			team = models.TournamentTeam{}
			if err := app.DB.First(&team, "id = ?", s.lobby.ID).Error; err != nil {
				t.Fatal(err)
			}
			if team.Status == want && (want != models.TournamentTeamStatusPaymentPending || team.PaymentAttemptID == nil) {
				return
			}
		}
		t.Fatalf("team stayed %s, want %s", team.Status, want)
	}

	t.Run("paid", func(t *testing.T) {
		s, paymentID := pendingLobby(t)
		pay(t, s, paymentID, "SUCCEEDED")
		waitForStatus(t, s, models.TournamentTeamStatusLocked)
	})
	t.Run("failed", func(t *testing.T) {
		s, paymentID := pendingLobby(t)
		pay(t, s, paymentID, "FAILED")
		waitForStatus(t, s, models.TournamentTeamStatusPaymentPending)
		resp, err := s.client.Do(http.MethodPost, "/lobbies/"+s.lobby.ID.String()+"/lock", nil, s.captain.AuthHeaders())
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("retry: %v %v", err, resp)
		}
		var retried handlers.TeamLobbyResponse
		if err := resp.JSON(&retried); err != nil || retried.PaymentID == nil || *retried.PaymentID == paymentID {
			t.Fatalf("retry result: %+v %v", retried, err)
		}
		pay(t, s, *retried.PaymentID, "SUCCEEDED")
		waitForStatus(t, s, models.TournamentTeamStatusLocked)
	})
}
