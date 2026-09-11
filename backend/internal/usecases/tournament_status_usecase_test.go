package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
	"github.com/moomaideng/eventory/internal/usecases"
)

// statusRepositoryStub stands in for the database. It runs the use case's
// mutate callback against a fixed current status so transition rules can be
// exercised without Postgres, and records what the use case tried to persist.
type statusRepositoryStub struct {
	current models.TournamentStatus
	err     error

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
	accountID, tournamentID uuid.UUID,
	reason string,
	mutate func(current *models.Tournament) (models.TournamentStatus, error),
) (*models.TournamentStatusChange, error) {
	r.called = true
	r.gotAccountID, r.gotTournament, r.gotReason = accountID, tournamentID, reason
	if r.err != nil {
		return nil, r.err
	}

	tournament := models.Tournament{ID: tournamentID, Status: r.current}
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

func (r *statusRepositoryStub) ListHistory(_ context.Context, _, _ uuid.UUID) ([]models.TournamentStatusChange, error) {
	r.historyCalled = true
	return r.historyEntries, r.err
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
		// Force-starting an under-funded campaign is the whole point of the
		// crowdfunding escape hatch; the reason field records the risk taken.
		{"force start under-funded campaign", models.TournamentStatusCrowdfunding, models.TournamentStatusRegistrationOpen},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &statusRepositoryStub{current: test.from}
			accountID, tournamentID := uuid.New(), uuid.New()

			result, err := usecases.NewTournamentStatusUseCase(repo).
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

func TestOverrideRejectsForbiddenTransitions(t *testing.T) {
	tests := []struct {
		name string
		from models.TournamentStatus
		to   models.TournamentStatus
	}{
		// COMPLETED is a read-only archive.
		{"reopen a completed tournament", models.TournamentStatusCompleted, models.TournamentStatusOngoing},
		{"cancel a completed tournament", models.TournamentStatusCompleted, models.TournamentStatusCancelled},
		// Cancelling triggers refunds that flipping a status cannot reverse.
		{"revive a cancelled tournament", models.TournamentStatusCancelled, models.TournamentStatusRegistrationOpen},
		// DRAFT and CROWDFUNDING are declared placeholders; no screen handles
		// them, so a tournament must never be moved into one.
		{"regress to draft", models.TournamentStatusRegistrationOpen, models.TournamentStatusDraft},
		{"regress to crowdfunding", models.TournamentStatusOngoing, models.TournamentStatusCrowdfunding},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := &statusRepositoryStub{current: test.from}

			_, err := usecases.NewTournamentStatusUseCase(repo).
				Override(context.Background(), uuid.New(), uuid.New(), string(test.to), validReason)
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

	_, err := usecases.NewTournamentStatusUseCase(repo).
		Override(context.Background(), uuid.New(), uuid.New(), "ONGOING", validReason)
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

			_, err := usecases.NewTournamentStatusUseCase(repo).
				Override(context.Background(), uuid.New(), uuid.New(), "ONGOING", test.reason)
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

	result, err := usecases.NewTournamentStatusUseCase(repo).
		Override(context.Background(), uuid.New(), uuid.New(), "ONGOING", "   "+validReason+"   ")
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
		if _, err := usecases.NewTournamentStatusUseCase(repo).
			Override(context.Background(), uuid.New(), uuid.New(), "  ongoing  ", validReason); err != nil {
			t.Fatalf("expected normalised status to be accepted, got %v", err)
		}
		if repo.persistedTo != models.TournamentStatusOngoing {
			t.Fatalf("expected ONGOING persisted, got %s", repo.persistedTo)
		}
	})

	t.Run("rejects an unknown status", func(t *testing.T) {
		repo := &statusRepositoryStub{current: models.TournamentStatusRegistrationOpen}
		_, err := usecases.NewTournamentStatusUseCase(repo).
			Override(context.Background(), uuid.New(), uuid.New(), "BANANA", validReason)
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

	_, err := usecases.NewTournamentStatusUseCase(repo).
		Override(context.Background(), uuid.New(), uuid.New(), "ONGOING", validReason)
	if !errors.Is(err, usecases.ErrTournamentNotFound) {
		t.Fatalf("expected tournament-not-found error, got %v", err)
	}
}

func TestHistoryTranslatesOwnershipFailureToNotFound(t *testing.T) {
	repo := &statusRepositoryStub{err: repositories.ErrTournamentNotFound}

	_, err := usecases.NewTournamentStatusUseCase(repo).
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
