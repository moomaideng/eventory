package handlers_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/handlers"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
	"github.com/moomaideng/eventory/internal/usecases"
)

type dashboardHTTPRepo struct {
	owner      uuid.UUID
	tournament models.Tournament
}

func (r *dashboardHTTPRepo) ListOwned(_ context.Context, accountID uuid.UUID, page, pageSize int) ([]models.Tournament, int64, error) {
	if accountID != r.owner {
		return nil, 0, nil
	}
	return []models.Tournament{r.tournament}, 1, nil
}

func (r *dashboardHTTPRepo) GetOwned(_ context.Context, accountID, tournamentID uuid.UUID) (*models.Tournament, error) {
	if accountID != r.owner || tournamentID != r.tournament.ID {
		return nil, repositories.ErrTournamentNotFound
	}
	return &r.tournament, nil
}

func TestOrganizerDashboardHTTP(t *testing.T) {
	owner, other, tournamentID := uuid.New(), uuid.New(), uuid.New()
	member := models.Account{ID: uuid.New(), Handle: "player_one", DisplayName: "Player One", Email: "private@example.com"}
	repo := &dashboardHTTPRepo{owner: owner, tournament: models.Tournament{
		ID: tournamentID, Name: "Owner tournament", Capacity: 8,
		Teams: []models.TournamentTeam{{ID: uuid.New(), Name: "First team", Status: models.TournamentTeamStatusAccepted,
			Members: []models.TournamentTeamMember{{ID: uuid.New(), AccountID: member.ID, Account: member}},
		}},
	}}
	accounts := usecases.NewAccountUseCase(&mockAccountRepo{accounts: []*models.Account{{ID: owner}, {ID: other}}})
	_, api := humatest.New(t)
	handlers.RegisterOrganizerDashboardRoutes(huma.NewGroup(api, "/api/v1"), usecases.NewOrganizerDashboardUseCase(repo), accounts)
	ownerCtx, otherCtx := makeAuthContext(owner, "owner@example.com"), makeAuthContext(other, "other@example.com")
	path := "/api/v1/tournaments/" + tournamentID.String() + "/dashboard"
	for _, tc := range []struct {
		name   string
		ctx    context.Context
		path   string
		status int
	}{
		{"anonymous list", context.Background(), "/api/v1/organizer/tournaments", 401},
		{"anonymous detail", context.Background(), path, 401},
		{"owner list", ownerCtx, "/api/v1/organizer/tournaments", 200},
		{"other list", otherCtx, "/api/v1/organizer/tournaments", 200},
		{"other detail", otherCtx, path, 404},
		{"missing detail", ownerCtx, "/api/v1/tournaments/" + uuid.NewString() + "/dashboard", 404},
		{"invalid ID", ownerCtx, "/api/v1/tournaments/invalid/dashboard", 422},
		{"invalid page", ownerCtx, "/api/v1/organizer/tournaments?page=0", 422},
		{"oversized page", ownerCtx, "/api/v1/organizer/tournaments?pageSize=101", 422},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := api.GetCtx(tc.ctx, tc.path)
			if resp.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", resp.Code, tc.status, resp.Body.String())
			}
		})
	}
	resp := api.GetCtx(ownerCtx, path)
	if resp.Code != 200 {
		t.Fatalf("owner dashboard: %s", resp.Body.String())
	}
	var output handlers.OrganizerDashboardOutput
	if err := json.Unmarshal(resp.Body.Bytes(), &output.Body); err != nil {
		t.Fatal(err)
	}
	if output.Body.Summary.Metrics.AcceptedEntries != 1 || output.Body.Summary.Metrics.ConfirmedParticipants != 1 {
		t.Fatalf("unexpected metrics: %+v", output.Body.Summary.Metrics)
	}
	got := output.Body.Entries[0].Members[0]
	if got.Handle != member.Handle || got.DisplayName != member.DisplayName {
		t.Fatalf("unexpected member: %+v", got)
	}
	if strings.Contains(resp.Body.String(), member.Email) || strings.Contains(resp.Body.String(), "inviteCode") {
		t.Fatal("dashboard must not expose private account or invite fields")
	}
}
