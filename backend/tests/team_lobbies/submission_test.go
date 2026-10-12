package team_lobbies

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	paymentmodels "github.com/moomaideng/eventory/services/payment/models"
	handlers "github.com/moomaideng/eventory/services/tournament/handlers/rest"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

func TestRosterSubmission(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	app = apptest.NewApp(t, pg.DSN)
	payments := apptest.NewClient(app.PaymentServer.URL + "/api/v1/payments")
	request := func(t *testing.T, client *apptest.Client, method, path string, body any, user *apptest.TestUser, code int) *apptest.Response {
		t.Helper()
		resp, err := client.Do(method, path, body, user.AuthHeaders())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != code {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, resp.StatusCode, code, resp.Body)
		}
		return resp
	}
	fixture := func(t *testing.T, fee int64) *teamLobbyScenarioContext {
		t.Helper()
		s := &teamLobbyScenarioContext{client: apptest.NewClient(app.TournamentBaseURL())}
		if err := s.aCaptainHasCreatedAFormingTeamLobbyWithAValidInviteCode(); err != nil {
			t.Fatal(err)
		}
		if err := app.DB.Model(s.tournament).Updates(map[string]any{"entry_fee": fee, "min_team_size": 1}).Error; err != nil {
			t.Fatal(err)
		}
		return s
	}
	lock := func(t *testing.T, s *teamLobbyScenarioContext) handlers.TeamLobbyResponse {
		t.Helper()
		resp := request(t, s.client, http.MethodPost, "/lobbies/"+s.lobby.ID.String()+"/lock", nil, s.captain, 200)
		var lobby handlers.TeamLobbyResponse
		if err := resp.JSON(&lobby); err != nil {
			t.Fatal(err)
		}
		return lobby
	}
	refresh := func(t *testing.T, s *teamLobbyScenarioContext, status string) {
		t.Helper()
		resp := request(t, s.client, http.MethodPost, "/lobbies/"+s.lobby.ID.String()+"/submission/refresh", nil, s.captain, 200)
		var lobby handlers.TeamLobbyResponse
		if err := resp.JSON(&lobby); err != nil {
			t.Fatal(err)
		}
		if lobby.Status != status {
			t.Fatalf("got %s, want %s", lobby.Status, status)
		}
	}
	review := func(t *testing.T, s *teamLobbyScenarioContext, count int) {
		t.Helper()
		owner := *s.captain
		owner.ID = s.tournament.Staff[0].AccountID
		owner.Token = apptest.GenerateTestJWT(owner.ID, owner.Email)
		resp := request(t, s.client, http.MethodGet, "/tournaments/"+s.tournament.ID.String()+"/dashboard", nil, &owner, 200)
		var dashboard struct {
			Entries []handlers.OrganizerDashboardEntry `json:"entries"`
		}
		if err := resp.JSON(&dashboard); err != nil {
			t.Fatal(err)
		}
		if len(dashboard.Entries) != count {
			t.Fatalf("review contains %d entries, want %d", len(dashboard.Entries), count)
		}
	}
	t.Run("free submission is idempotent and visible for review", func(t *testing.T) {
		s := fixture(t, 0)
		review(t, s, 0)
		for range 2 {
			if got := lock(t, s); got.Status != "LOCKED" || got.PaymentID != nil {
				t.Fatalf("unexpected free result: %+v", got)
			}
		}
		review(t, s, 1)
	})
	t.Run("paid submission freezes roster and reuses invoice", func(t *testing.T) {
		s := fixture(t, 250)
		var wg sync.WaitGroup
		results := make(chan *apptest.Response, 4)
		for range 4 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				r, e := s.client.Do(http.MethodPost, "/lobbies/"+s.lobby.ID.String()+"/lock", nil, s.captain.AuthHeaders())
				if e != nil {
					t.Error(e)
					return
				}
				results <- r
			}()
		}
		wg.Wait()
		close(results)
		var id uuid.UUID
		for r := range results {
			var lobby handlers.TeamLobbyResponse
			if r.StatusCode != 200 {
				t.Fatalf("%d: %s", r.StatusCode, r.Body)
			}
			if err := r.JSON(&lobby); err != nil {
				t.Fatal(err)
			}
			if lobby.Status != "PAYMENT_PENDING" || lobby.PaymentID == nil {
				t.Fatalf("unexpected pending result: %+v", lobby)
			}
			if id != uuid.Nil && id != *lobby.PaymentID {
				t.Fatal("duplicate invoice")
			}
			id = *lobby.PaymentID
		}
		var invoice paymentmodels.Invoice
		if err := app.DB.First(&invoice, "id = ?", id).Error; err != nil {
			t.Fatal(err)
		}
		if invoice.AmountMinor != 25000 || invoice.PayerAccountID != s.captain.ID {
			t.Fatalf("wrong invoice: %+v", invoice)
		}
		review(t, s, 0)
		request(t, s.client, http.MethodDelete, "/lobbies/"+s.lobby.ID.String(), nil, s.captain, 409)
		request(t, s.client, http.MethodPost, "/lobbies/"+s.lobby.ID.String()+"/invite/regenerate", nil, s.captain, 409)
		for range 2 {
			request(t, payments, http.MethodPost, "/"+id.String()+"/mock-result", map[string]string{"result": "SUCCEEDED"}, s.captain, 200)
		}
		for range 2 {
			refresh(t, s, "LOCKED")
		}
		review(t, s, 1)
		var receipts int64
		app.DB.Model(&paymentmodels.Receipt{}).Where("invoice_id = ?", id).Count(&receipts)
		if receipts != 1 {
			t.Fatalf("receipts = %d", receipts)
		}
	})
	for _, result := range []string{"FAILED", "CANCELLED"} {
		t.Run(result+" unlocks and permits a new attempt", func(t *testing.T) {
			s := fixture(t, 100)
			first := lock(t, s)
			request(t, payments, http.MethodPost, "/"+first.PaymentID.String()+"/mock-result", map[string]string{"result": result}, s.captain, 200)
			refresh(t, s, "FORMING")
			review(t, s, 0)
			request(t, payments, http.MethodPost, "/"+first.PaymentID.String()+"/mock-result", map[string]string{"result": "SUCCEEDED"}, s.captain, 409)
			next := lock(t, s)
			if next.PaymentID == nil || *next.PaymentID == *first.PaymentID {
				t.Fatal("retry reused closed invoice")
			}
		})
	}
	t.Run("expired attempt unlocks and cannot be paid", func(t *testing.T) {
		s := fixture(t, 100)
		first := lock(t, s)
		expired := time.Now().UTC().Add(-time.Minute).Truncate(time.Microsecond)
		if err := app.DB.Model(&paymentmodels.Invoice{}).Where("id = ?", first.PaymentID).Update("expires_at", expired).Error; err != nil {
			t.Fatal(err)
		}
		if err := app.DB.Model(&models.TournamentTeam{}).Where("id = ?", s.lobby.ID).Update("payment_expires_at", expired).Error; err != nil {
			t.Fatal(err)
		}
		refresh(t, s, "FORMING")
		review(t, s, 0)
		request(t, payments, http.MethodPost, "/"+first.PaymentID.String()+"/mock-result", map[string]string{"result": "SUCCEEDED"}, s.captain, 409)
	})
	t.Run("validation and captain authorization", func(t *testing.T) {
		s := fixture(t, 0)
		if err := s.anotherActiveCompetitor(); err != nil {
			t.Fatal(err)
		}
		request(t, s.client, http.MethodPost, "/lobbies/"+s.lobby.ID.String()+"/lock", nil, s.otherCompetitor, 403)
		for _, updates := range []map[string]any{{"min_team_size": 2}, {"min_team_size": 1, "registration_deadline": time.Now().Add(-time.Minute)}} {
			if err := app.DB.Model(s.tournament).Updates(updates).Error; err != nil {
				t.Fatal(err)
			}
			request(t, s.client, http.MethodPost, "/lobbies/"+s.lobby.ID.String()+"/lock", nil, s.captain, 409)
			refresh(t, s, "FORMING")
			review(t, s, 0)
		}
	})
}
