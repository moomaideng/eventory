package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

// statusRepositoryStub stands in for the database. It runs the use case's
// mutate callback against a fixed current status so transition rules can be
// exercised without Postgres, and records what the use case tried to persist.
type statusRepositoryStub struct {
	mockTournamentRepository
	current    models.TournamentStatus
	tournament *models.Tournament
	err        error

	called         bool
	gotAccountID   uuid.UUID
	gotTournament  uuid.UUID
	gotReason      string
	persistedTo    models.TournamentStatus
	historyCalled  bool
	historyEntries []models.TournamentStatusChange
}

func (r *statusRepositoryStub) ApplyOverride(
	_ context.Context,
	accountID uuid.UUID,
	_ string,
	tournamentID uuid.UUID,
	reason string,
	mutate func(current *models.Tournament) (models.TournamentStatus, error),
) (*models.TournamentStatusChange, error) {
	r.called = true
	r.gotAccountID, r.gotTournament, r.gotReason = accountID, tournamentID, reason
	if r.err != nil {
		return nil, r.err
	}

	tournament := models.Tournament{ID: tournamentID, Status: r.current}
	if r.tournament != nil {
		tournament = *r.tournament
	}
	to, err := mutate(&tournament)
	if err != nil {
		return nil, err
	}
	r.persistedTo = to

	return &models.TournamentStatusChange{
		TournamentID: tournamentID, FromStatus: r.current, ToStatus: to,
		Reason: reason, ActorAccountID: accountID,
	}, nil
}

func (r *statusRepositoryStub) ListHistory(_ context.Context, _, _ uuid.UUID) (*repositories.TournamentStatusHistory, error) {
	r.historyCalled = true
	if r.err != nil {
		return nil, r.err
	}
	return &repositories.TournamentStatusHistory{CurrentStatus: r.current, Changes: r.historyEntries}, nil
}

const validReason = "Venue double-booked, shifting the schedule"

func TestOverrideAllowsPermittedTransitions(t *testing.T) {
	tests := []struct {
		name string
		from models.TournamentStatus
		to   models.TournamentStatus
	}{
		{"close registration", models.TournamentStatusRegistrationOpen, models.TournamentStatusRegistrationClosed},
		{"skip ahead to ongoing", models.TournamentStatusRegistrationOpen, models.TournamentStatusOngoing},
		{"reopen after a delay", models.TournamentStatusRegistrationClosed, models.TournamentStatusRegistrationOpen},
		{"pull back a premature start", models.TournamentStatusOngoing, models.TournamentStatusRegistrationClosed},
		{"finish", models.TournamentStatusOngoing, models.TournamentStatusCompleted},
		{"cancel while live", models.TournamentStatusOngoing, models.TournamentStatusCancelled},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &statusRepositoryStub{current: test.from}
			accountID, tournamentID := uuid.New(), uuid.New()

			result, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
				Override(context.Background(), accountID, tournamentID, string(test.to), validReason)
			if err != nil {
				t.Fatalf("expected transition to be allowed, got %v", err)
			}
			if result.FromStatus != test.from || result.ToStatus != test.to {
				t.Fatalf("expected %s -> %s, got %s -> %s", test.from, test.to, result.FromStatus, result.ToStatus)
			}
			if repo.persistedTo != test.to {
				t.Fatalf("expected %s persisted, got %s", test.to, repo.persistedTo)
			}
			if repo.gotAccountID != accountID || repo.gotTournament != tournamentID {
				t.Fatal("expected the acting account and tournament to reach the repository")
			}
		})
	}
}

func crowdfundingReadyTournament(ownerID uuid.UUID) *models.Tournament {
	now := time.Now()
	fundingDeadline := now.Add(24 * time.Hour)
	return &models.Tournament{
		Status: models.TournamentStatusCrowdfunding, Published: true,
		Name: "Community Cup", Game: "Valorant", Location: "Main Hall",
		RegistrationDeadline: now.Add(48 * time.Hour),
		StartsAt:             now.Add(72 * time.Hour), EndsAt: now.Add(96 * time.Hour),
		RegistrationMode: models.TournamentRegistrationModeTeam,
		MinTeamSize:      3, MaxTeamSize: 5, Capacity: 16,
		Funding: &models.TournamentFunding{GoalAmount: 1000, RaisedAmount: 1000, Deadline: &fundingDeadline},
		Staff:   []models.TournamentStaff{{AccountID: ownerID, Role: models.TournamentStaffRoleOwner}},
	}
}

func TestCrowdfundingLifecycle(t *testing.T) {
	tests := []struct {
		name    string
		to      models.TournamentStatus
		force   bool
		modify  func(*models.Tournament)
		want    error
		problem string
	}{
		{name: "open when funded", to: models.TournamentStatusRegistrationOpen},
		{name: "open when overfunded", to: models.TournamentStatusRegistrationOpen,
			modify: func(t *models.Tournament) { t.Funding.RaisedAmount = 1200 }},
		{name: "underfunded normal launch", to: models.TournamentStatusRegistrationOpen,
			modify: func(t *models.Tournament) { t.Funding.RaisedAmount = 999 }, want: usecases.ErrCrowdfundingGoalNotReached},
		{name: "force launch underfunded", to: models.TournamentStatusRegistrationOpen, force: true,
			modify: func(t *models.Tournament) { t.Funding.RaisedAmount = 200 }},
		{name: "missing venue", to: models.TournamentStatusRegistrationOpen,
			modify: func(t *models.Tournament) { t.Location = " " }, want: usecases.ErrCrowdfundingPlanIncomplete, problem: "location"},
		{name: "registration deadline passed", to: models.TournamentStatusRegistrationOpen, force: true,
			modify: func(t *models.Tournament) { t.RegistrationDeadline = time.Now().Add(-time.Hour) }, want: usecases.ErrCrowdfundingPlanIncomplete, problem: "future registration deadline"},
		{name: "invalid roster rules", to: models.TournamentStatusRegistrationOpen, force: true,
			modify: func(t *models.Tournament) { t.MinTeamSize = 6 }, want: usecases.ErrCrowdfundingPlanIncomplete},
		{name: "cancel regardless of funding", to: models.TournamentStatusCancelled,
			modify: func(t *models.Tournament) { t.Funding.RaisedAmount = 0 }},
		{name: "force cannot cancel", to: models.TournamentStatusCancelled, force: true, want: usecases.ErrInvalidStatusTransition},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ownerID, tournamentID := uuid.New(), uuid.New()
			current := crowdfundingReadyTournament(ownerID)
			if test.modify != nil {
				test.modify(current)
			}
			repo := &statusRepositoryStub{current: current.Status, tournament: current}
			result, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
				OverrideWithForce(context.Background(), ownerID, tournamentID, string(test.to), validReason, test.force)
			if !errors.Is(err, test.want) {
				t.Fatalf("wanted %v, got %v", test.want, err)
			}
			if test.want != nil {
				if test.problem != "" && !strings.Contains(err.Error(), test.problem) {
					t.Fatalf("expected useful explanation containing %q, got %v", test.problem, err)
				}
				if repo.persistedTo != "" {
					t.Fatalf("unexpected status change to %s", repo.persistedTo)
				}
				return
			}
			if result.ToStatus != test.to || repo.persistedTo != test.to || repo.gotReason != validReason {
				t.Fatalf("unexpected change: result=%+v persisted=%s reason=%s", result, repo.persistedTo, repo.gotReason)
			}
		})
	}
}

func TestCrowdfundingLaunchRequiresOwnerAndReason(t *testing.T) {
	ownerID, staffID, tournamentID := uuid.New(), uuid.New(), uuid.New()
	repo := &statusRepositoryStub{current: models.TournamentStatusCrowdfunding, tournament: crowdfundingReadyTournament(ownerID)}
	repo.tournament.Staff = append(repo.tournament.Staff, models.TournamentStaff{AccountID: staffID, Role: models.TournamentStaffRoleAdmin})
	repo.tournament.Funding.RaisedAmount = 200
	useCase := usecases.NewTournamentUseCase(repo, newStubAccountService())
	if _, err := useCase.OverrideWithForce(context.Background(), staffID, tournamentID, "REGISTRATION_OPEN", validReason, true); !errors.Is(err, usecases.ErrNotTournamentOwner) {
		t.Fatalf("expected owner check, got %v", err)
	}
	if repo.persistedTo != "" {
		t.Fatal("staff member must not launch")
	}
	if _, err := useCase.OverrideWithForce(context.Background(), staffID, tournamentID, "CANCELLED", validReason, false); !errors.Is(err, usecases.ErrNotTournamentOwner) {
		t.Fatalf("expected owner check for cancellation, got %v", err)
	}
	if _, err := useCase.OverrideWithForce(context.Background(), ownerID, tournamentID, "REGISTRATION_OPEN", "", true); !errors.Is(err, usecases.ErrOverrideReasonRequired) {
		t.Fatalf("expected force-start reason check, got %v", err)
	}
	if _, err := useCase.OverrideWithForce(context.Background(), ownerID, tournamentID, "CANCELLED", "", false); !errors.Is(err, usecases.ErrOverrideReasonRequired) {
		t.Fatalf("expected reason check, got %v", err)
	}
}

func TestOverrideRejectsForbiddenTransitions(t *testing.T) {
	tests := []struct {
		name string
		from models.TournamentStatus
		to   models.TournamentStatus
	}{
		// COMPLETED is a read-only archive.
		{"reopen a completed tournament", models.TournamentStatusCompleted, models.TournamentStatusOngoing},
		{"cancel a completed tournament", models.TournamentStatusCompleted, models.TournamentStatusCancelled},
		// A cancelled tournament is terminal and cannot be revived.
		{"revive a cancelled tournament", models.TournamentStatusCancelled, models.TournamentStatusRegistrationOpen},
		// A running tournament cannot move back to its draft or funding stage.
		{"regress to draft", models.TournamentStatusRegistrationOpen, models.TournamentStatusDraft},
		{"regress to crowdfunding", models.TournamentStatusOngoing, models.TournamentStatusCrowdfunding},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &statusRepositoryStub{current: test.from}

			accountID, tournamentID := uuid.New(), uuid.New()
			_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
				Override(context.Background(), accountID, tournamentID, string(test.to), validReason)
			if !errors.Is(err, usecases.ErrInvalidStatusTransition) {
				t.Fatalf("expected invalid transition error, got %v", err)
			}
			if repo.persistedTo != "" {
				t.Fatalf("expected nothing persisted, got %s", repo.persistedTo)
			}
		})
	}
}

func TestOverrideRejectsNoOpChange(t *testing.T) {
	repo := &statusRepositoryStub{current: models.TournamentStatusOngoing}

	accountID, tournamentID := uuid.New(), uuid.New()
	_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
		Override(context.Background(), accountID, tournamentID, "ONGOING", validReason)
	if !errors.Is(err, usecases.ErrStatusUnchanged) {
		t.Fatalf("expected unchanged-status error, got %v", err)
	}
	if repo.persistedTo != "" {
		t.Fatal("expected no history row for a no-op override")
	}
}

func TestOverrideRequiresAReason(t *testing.T) {
	tests := []struct {
		name   string
		reason string
	}{
		{"empty", ""},
		{"whitespace only", "        \t  "},
		{"too short", "too short"},
		{"padded but still too short", "   hi   "},
		{"too long", strings.Repeat("a", 501)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &statusRepositoryStub{current: models.TournamentStatusRegistrationOpen}

			accountID, tournamentID := uuid.New(), uuid.New()
			_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
				Override(context.Background(), accountID, tournamentID, "ONGOING", test.reason)
			if !errors.Is(err, usecases.ErrOverrideReasonRequired) {
				t.Fatalf("expected reason-required error, got %v", err)
			}
			if repo.called {
				t.Fatal("expected the repository never to be reached for an invalid reason")
			}
		})
	}
}

func TestOverrideStoresTrimmedReason(t *testing.T) {
	repo := &statusRepositoryStub{current: models.TournamentStatusRegistrationOpen}

	accountID, tournamentID := uuid.New(), uuid.New()
	result, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
		Override(context.Background(), accountID, tournamentID, "ONGOING", "   "+validReason+"   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Reason != validReason || repo.gotReason != validReason {
		t.Fatalf("expected reason trimmed before storage, got %q", repo.gotReason)
	}
}

func TestOverrideNormalisesAndValidatesStatusInput(t *testing.T) {
	t.Run("accepts lowercase and padding", func(t *testing.T) {
		repo := &statusRepositoryStub{current: models.TournamentStatusRegistrationOpen}
		accountID, tournamentID := uuid.New(), uuid.New()
		if _, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
			Override(context.Background(), accountID, tournamentID, "  ongoing  ", validReason); err != nil {
			t.Fatalf("expected normalised status to be accepted, got %v", err)
		}
		if repo.persistedTo != models.TournamentStatusOngoing {
			t.Fatalf("expected ONGOING persisted, got %s", repo.persistedTo)
		}
	})

	t.Run("rejects an unknown status", func(t *testing.T) {
		repo := &statusRepositoryStub{current: models.TournamentStatusRegistrationOpen}
		accountID, tournamentID := uuid.New(), uuid.New()
		_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
			Override(context.Background(), accountID, tournamentID, "BANANA", validReason)
		if !errors.Is(err, usecases.ErrUnknownStatus) {
			t.Fatalf("expected unknown-status error, got %v", err)
		}
		if repo.called {
			t.Fatal("expected the repository never to be reached for an unknown status")
		}
	})
}

// A tournament the account does not own must be indistinguishable from one that
// does not exist, so the use case translates the repository's not-found error
// rather than surfacing an access-denied error the handler would map to 403.
func TestOverrideTranslatesOwnershipFailureToNotFound(t *testing.T) {
	repo := &statusRepositoryStub{
		current: models.TournamentStatusRegistrationOpen,
		err:     repositories.ErrTournamentNotFound,
	}

	accountID, tournamentID := uuid.New(), uuid.New()
	_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
		Override(context.Background(), accountID, tournamentID, "ONGOING", validReason)
	if !errors.Is(err, usecases.ErrTournamentNotFound) {
		t.Fatalf("expected tournament-not-found error, got %v", err)
	}
}

func TestHistoryTranslatesOwnershipFailureToNotFound(t *testing.T) {
	repo := &statusRepositoryStub{err: repositories.ErrTournamentNotFound}

	_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
		History(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, usecases.ErrTournamentNotFound) {
		t.Fatalf("expected tournament-not-found error, got %v", err)
	}
	if !repo.historyCalled {
		t.Fatal("expected history to be requested from the repository")
	}
}

func TestAllowedTransitionsFromIsTerminalForFinishedTournaments(t *testing.T) {
	for _, status := range []models.TournamentStatus{
		models.TournamentStatusCompleted,
		models.TournamentStatusCancelled,
	} {
		if got := usecases.AllowedTransitionsFrom(status); len(got) != 0 {
			t.Fatalf("expected %s to be terminal, got %v", status, got)
		}
	}
}

// The handler hands this list to the modal, so callers must not be able to
// mutate the use case's own transition table through it.
func TestAllowedTransitionsFromReturnsACopy(t *testing.T) {
	first := usecases.AllowedTransitionsFrom(models.TournamentStatusRegistrationOpen)
	if len(first) == 0 {
		t.Fatal("expected REGISTRATION_OPEN to have transitions")
	}
	first[0] = "MUTATED"

	if second := usecases.AllowedTransitionsFrom(models.TournamentStatusRegistrationOpen); second[0] == "MUTATED" {
		t.Fatal("expected AllowedTransitionsFrom to return a defensive copy")
	}
}
