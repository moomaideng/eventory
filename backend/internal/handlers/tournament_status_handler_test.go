package handlers_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/handlers"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
	"github.com/moomaideng/eventory/internal/usecases"
)

// statusHTTPRepo enforces ownership the way the real repository does: a
// tournament belonging to someone else is reported as not found, never as a
// distinct access-denied error.
type statusHTTPRepo struct {
	owner        uuid.UUID
	tournamentID uuid.UUID
	current      models.TournamentStatus
	actor        models.Account

	writes int
}

func (r *statusHTTPRepo) ApplyOverride(
	_ context.Context,
	accountID, tournamentID uuid.UUID,
	reason string,
	mutate func(current *models.Tournament) (models.TournamentStatus, error),
) (*models.TournamentStatusChange, error) {
	if accountID != r.owner || tournamentID != r.tournamentID {
		return nil, repositories.ErrTournamentNotFound
	}

	tournament := models.Tournament{ID: tournamentID, Status: r.current}
	to, err := mutate(&tournament)
	if err != nil {
		return nil, err
	}

	r.writes++
	return &models.TournamentStatusChange{
		ID: uuid.New(), TournamentID: tournamentID, FromStatus: r.current, ToStatus: to,
		Reason: reason, ActorAccountID: accountID, Actor: r.actor,
		CreatedAt: time.Now(),
	}, nil
}

func (r *statusHTTPRepo) ListHistory(_ context.Context, accountID, tournamentID uuid.UUID) ([]models.TournamentStatusChange, error) {
	if accountID != r.owner || tournamentID != r.tournamentID {
		return nil, repositories.ErrTournamentNotFound
	}
	return []models.TournamentStatusChange{{
		ID: uuid.New(), TournamentID: tournamentID,
		FromStatus: models.TournamentStatusRegistrationOpen, ToStatus: r.current,
		Reason:         "Venue double-booked, shifting the schedule",
		ActorAccountID: accountID, Actor: r.actor,
	}}, nil
}

func newStatusAPI(t *testing.T, repo *statusHTTPRepo, accountIDs ...uuid.UUID) humatest.TestAPI {
	t.Helper()
	accounts := make([]*models.Account, 0, len(accountIDs))
	for _, id := range accountIDs {
		accounts = append(accounts, &models.Account{ID: id})
	}
	_, api := humatest.New(t)
	handlers.RegisterTournamentStatusRoutes(
		huma.NewGroup(api, "/api/v1"),
		usecases.NewTournamentStatusUseCase(repo),
		usecases.NewAccountUseCase(&mockAccountRepo{accounts: accounts}),
	)
	return api
}

func overrideBody(status, reason string) map[string]any {
	return map[string]any{"status": status, "reason": reason}
}

// TestOverrideTournamentStatusAuthorization covers US3-3's authorization and
// permission boundaries at the HTTP layer.
func TestOverrideTournamentStatusAuthorization(t *testing.T) {
	owner, other, tournamentID := uuid.New(), uuid.New(), uuid.New()
	repo := &statusHTTPRepo{
		owner: owner, tournamentID: tournamentID,
		current: models.TournamentStatusRegistrationOpen,
		actor:   models.Account{ID: owner, Handle: "owner", DisplayName: "Owner", Email: "private@example.com"},
	}
	api := newStatusAPI(t, repo, owner, other)

	ownerCtx := makeAuthContext(owner, "owner@example.com")
	otherCtx := makeAuthContext(other, "other@example.com")
	path := "/api/v1/tournaments/" + tournamentID.String() + "/status"
	reason := "Venue double-booked, shifting the schedule"

	for _, tc := range []struct {
		name   string
		ctx    context.Context
		path   string
		body   map[string]any
		status int
	}{
		{"anonymous is rejected", context.Background(), path, overrideBody("ONGOING", reason), 401},
		// A non-owner must not be able to tell the tournament apart from one
		// that does not exist, so this is 404 rather than 403.
		{"non-owner sees not found", otherCtx, path, overrideBody("ONGOING", reason), 404},
		{"unknown tournament", ownerCtx, "/api/v1/tournaments/" + uuid.NewString() + "/status", overrideBody("ONGOING", reason), 404},
		{"malformed tournament id", ownerCtx, "/api/v1/tournaments/not-a-uuid/status", overrideBody("ONGOING", reason), 422},
		{"reason too short", ownerCtx, path, overrideBody("ONGOING", "short"), 422},
		{"reason missing", ownerCtx, path, map[string]any{"status": "ONGOING"}, 422},
		{"status missing", ownerCtx, path, map[string]any{"reason": reason}, 422},
		{"status outside the schema", ownerCtx, path, overrideBody("BANANA", reason), 422},
		{"no-op change", ownerCtx, path, overrideBody("REGISTRATION_OPEN", reason), 422},
		{"placeholder status as target", ownerCtx, path, overrideBody("DRAFT", reason), 422},
		{"owner performs a legal override", ownerCtx, path, overrideBody("ONGOING", reason), 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := api.PatchCtx(tc.ctx, tc.path, tc.body)
			if resp.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", resp.Code, tc.status, resp.Body.String())
			}
		})
	}

	// Only the single legal override should have reached the database.
	if repo.writes != 1 {
		t.Fatalf("expected exactly one persisted override, got %d", repo.writes)
	}
}

func TestOverrideTournamentStatusRejectsTerminalTransitions(t *testing.T) {
	owner, tournamentID := uuid.New(), uuid.New()
	repo := &statusHTTPRepo{
		owner: owner, tournamentID: tournamentID,
		current: models.TournamentStatusCompleted,
	}
	api := newStatusAPI(t, repo, owner)

	path := "/api/v1/tournaments/" + tournamentID.String() + "/status"
	body := overrideBody("ONGOING", "Trying to reopen a finished tournament")

	resp := api.PatchCtx(makeAuthContext(owner, "owner@example.com"), path, body)
	if resp.Code != 422 {
		t.Fatalf("status %d, want 422: %s", resp.Code, resp.Body.String())
	}
	if repo.writes != 0 {
		t.Fatal("a completed tournament must never be written to")
	}
}

func TestOverrideTournamentStatusResponseRecordsTheAuditFields(t *testing.T) {
	owner, tournamentID := uuid.New(), uuid.New()
	repo := &statusHTTPRepo{
		owner: owner, tournamentID: tournamentID,
		current: models.TournamentStatusRegistrationOpen,
		actor:   models.Account{ID: owner, Handle: "owner", DisplayName: "Owner", Email: "private@example.com"},
	}
	api := newStatusAPI(t, repo, owner)

	reason := "Venue double-booked, shifting the schedule"
	resp := api.PatchCtx(
		makeAuthContext(owner, "owner@example.com"),
		"/api/v1/tournaments/"+tournamentID.String()+"/status",
		overrideBody("ONGOING", reason),
	)
	if resp.Code != 200 {
		t.Fatalf("override failed: %s", resp.Body.String())
	}

	var output handlers.OverrideTournamentStatusOutput
	if err := json.Unmarshal(resp.Body.Bytes(), &output.Body); err != nil {
		t.Fatal(err)
	}
	change := output.Body.Change
	if change.FromStatus != "REGISTRATION_OPEN" || change.ToStatus != "ONGOING" {
		t.Fatalf("expected the previous status to be recorded, got %s -> %s", change.FromStatus, change.ToStatus)
	}
	if change.Reason != reason {
		t.Fatalf("expected the reason to be recorded, got %q", change.Reason)
	}
	if change.ActorID != owner {
		t.Fatalf("expected the acting account to be recorded, got %s", change.ActorID)
	}
	if change.CreatedAt.IsZero() {
		t.Fatal("expected the override to be timestamped")
	}
	if output.Body.Status != "ONGOING" {
		t.Fatalf("expected the new status to be returned, got %q", output.Body.Status)
	}
	if strings.Contains(resp.Body.String(), "private@example.com") {
		t.Fatal("the audit trail must not expose the actor's email")
	}
}

func TestTournamentStatusHistoryAuthorization(t *testing.T) {
	owner, other, tournamentID := uuid.New(), uuid.New(), uuid.New()
	repo := &statusHTTPRepo{
		owner: owner, tournamentID: tournamentID,
		current: models.TournamentStatusOngoing,
		actor:   models.Account{ID: owner, Handle: "owner", DisplayName: "Owner", Email: "private@example.com"},
	}
	api := newStatusAPI(t, repo, owner, other)

	path := "/api/v1/tournaments/" + tournamentID.String() + "/status/history"
	for _, tc := range []struct {
		name   string
		ctx    context.Context
		status int
	}{
		{"anonymous is rejected", context.Background(), 401},
		{"non-owner sees not found", makeAuthContext(other, "other@example.com"), 404},
		{"owner reads the trail", makeAuthContext(owner, "owner@example.com"), 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := api.GetCtx(tc.ctx, path)
			if resp.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", resp.Code, tc.status, resp.Body.String())
			}
		})
	}

	resp := api.GetCtx(makeAuthContext(owner, "owner@example.com"), path)
	if strings.Contains(resp.Body.String(), "private@example.com") {
		t.Fatal("the audit trail must not expose the actor's email")
	}
}
