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

type statusRepositoryStub struct {
	mockTournamentRepository
	current    models.TournamentStatus
	tournament *models.Tournament
	err        error

	called        bool
	gotAccountID  uuid.UUID
	gotTournament uuid.UUID
	persistedTo   models.TournamentStatus
}

func (r *statusRepositoryStub) GetByID(_ context.Context, id uuid.UUID) (*models.Tournament, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.tournament != nil {
		return r.tournament, nil
	}
	return &models.Tournament{
		ID:     id,
		Status: r.current,
		Staff:  []models.TournamentStaff{{AccountID: r.gotAccountID, Role: models.TournamentStaffRoleOwner}},
	}, nil
}

func (r *statusRepositoryStub) UpdateStatus(_ context.Context, id uuid.UUID, newStatus models.TournamentStatus) error {
	r.called = true
	r.gotTournament = id
	r.persistedTo = newStatus
	if r.err != nil {
		return r.err
	}
	r.current = newStatus
	return nil
}

func TestOverrideAllowsPermittedTransitions(t *testing.T) {
	tests := []struct {
		name string
		from models.TournamentStatus
		to   models.TournamentStatus
	}{
		{"draft to crowdfunding", models.TournamentStatusDraft, models.TournamentStatusCrowdfunding},
		{"draft to registration", models.TournamentStatusDraft, models.TournamentStatusRegistration},
		{"crowdfunding to registration", models.TournamentStatusCrowdfunding, models.TournamentStatusRegistration},
		{"registration to ongoing", models.TournamentStatusRegistration, models.TournamentStatusOngoing},
		{"ongoing to completed", models.TournamentStatusOngoing, models.TournamentStatusCompleted},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accountID, tournamentID := uuid.New(), uuid.New()
			repo := &statusRepositoryStub{current: test.from, gotAccountID: accountID}
			if test.from == models.TournamentStatusDraft && test.to == models.TournamentStatusCrowdfunding {
				repo.tournament = draftCrowdfundingReadyTournament(accountID)
				repo.activeTiersCount = 1
			}
			if test.from == models.TournamentStatusDraft && test.to == models.TournamentStatusRegistration {
				repo.tournament = draftRegistrationReadyTournament(accountID)
			}
			if test.from == models.TournamentStatusCrowdfunding {
				repo.tournament = crowdfundingReadyTournament(accountID)
			}
			if test.from == models.TournamentStatusRegistration && test.to == models.TournamentStatusOngoing {
				repo.tournament = &models.Tournament{
					ID:                   tournamentID,
					Status:               models.TournamentStatusRegistration,
					RegistrationDeadline: time.Now().Add(-time.Hour),
					Staff:                []models.TournamentStaff{{AccountID: accountID, Role: models.TournamentStaffRoleOwner}},
				}
			}

			result, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
				Override(context.Background(), accountID, tournamentID, string(test.to), "")
			if err != nil {
				t.Fatalf("expected transition to be allowed, got %v", err)
			}
			if result.FromStatus != test.from || result.ToStatus != test.to {
				t.Fatalf("expected %s -> %s, got %s -> %s", test.from, test.to, result.FromStatus, result.ToStatus)
			}
			if repo.persistedTo != test.to {
				t.Fatalf("expected %s persisted, got %s", test.to, repo.persistedTo)
			}
		})
	}
}

func draftCrowdfundingReadyTournament(ownerID uuid.UUID) *models.Tournament {
	now := time.Now()
	fundingDeadline := now.Add(24 * time.Hour)
	return &models.Tournament{
		Status: models.TournamentStatusDraft,
		Name:   "Community Cup", Game: "Valorant", Location: "Main Hall",
		RegistrationDeadline: now.Add(48 * time.Hour),
		StartsAt:             now.Add(72 * time.Hour), EndsAt: now.Add(96 * time.Hour),
		Funding: &models.TournamentFunding{GoalAmount: 1000, Deadline: &fundingDeadline},
		Staff:   []models.TournamentStaff{{AccountID: ownerID, Role: models.TournamentStaffRoleOwner}},
	}
}

func draftRegistrationReadyTournament(ownerID uuid.UUID) *models.Tournament {
	now := time.Now()
	return &models.Tournament{
		Status: models.TournamentStatusDraft,
		Name:   "Community Cup", Game: "Valorant", Location: "Main Hall",
		RegistrationDeadline: now.Add(48 * time.Hour),
		StartsAt:             now.Add(72 * time.Hour), EndsAt: now.Add(96 * time.Hour),
		RegistrationMode:     models.TournamentRegistrationModeTeam,
		MinTeamSize:          3, MaxTeamSize: 5, Capacity: 16,
		Staff: []models.TournamentStaff{{AccountID: ownerID, Role: models.TournamentStaffRoleOwner}},
	}
}

func crowdfundingReadyTournament(ownerID uuid.UUID) *models.Tournament {
	now := time.Now()
	fundingDeadline := now.Add(24 * time.Hour)
	return &models.Tournament{
		Status: models.TournamentStatusCrowdfunding,
		Name:   "Community Cup", Game: "Valorant", Location: "Main Hall",
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
		{name: "open when funded", to: models.TournamentStatusRegistration},
		{name: "open when overfunded", to: models.TournamentStatusRegistration,
			modify: func(t *models.Tournament) { t.Funding.RaisedAmount = 1200 }},
		{name: "underfunded normal launch", to: models.TournamentStatusRegistration,
			modify: func(t *models.Tournament) { t.Funding.RaisedAmount = 999 }, want: usecases.ErrCrowdfundingGoalNotReached},
		{name: "force launch underfunded", to: models.TournamentStatusRegistration, force: true,
			modify: func(t *models.Tournament) { t.Funding.RaisedAmount = 200 }},
		{name: "missing venue", to: models.TournamentStatusRegistration,
			modify: func(t *models.Tournament) { t.Location = " " }, want: usecases.ErrCrowdfundingPlanIncomplete, problem: "location"},
		{name: "registration deadline passed", to: models.TournamentStatusRegistration, force: true,
			modify: func(t *models.Tournament) { t.RegistrationDeadline = time.Now().Add(-time.Hour) }, want: usecases.ErrCrowdfundingPlanIncomplete, problem: "future registration deadline"},
		{name: "invalid roster rules", to: models.TournamentStatusRegistration, force: true,
			modify: func(t *models.Tournament) { t.MinTeamSize = 6 }, want: usecases.ErrCrowdfundingPlanIncomplete},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ownerID, tournamentID := uuid.New(), uuid.New()
			current := crowdfundingReadyTournament(ownerID)
			if test.modify != nil {
				test.modify(current)
			}
			repo := &statusRepositoryStub{current: current.Status, tournament: current, gotAccountID: ownerID}
			result, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
				OverrideWithForce(context.Background(), ownerID, tournamentID, string(test.to), "", test.force)
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
			if result.ToStatus != test.to || repo.persistedTo != test.to {
				t.Fatalf("unexpected change: result=%+v persisted=%s", result, repo.persistedTo)
			}
		})
	}
}

func TestDraftToCrowdfundingValidation(t *testing.T) {
	now := time.Now()
	ownerID, tournamentID := uuid.New(), uuid.New()

	tests := []struct {
		name       string
		modify     func(*models.Tournament)
		tiersCount int64
		want       error
	}{
		{name: "missing funding", modify: func(t *models.Tournament) { t.Funding = nil }, tiersCount: 1, want: usecases.ErrCrowdfundingSetupIncomplete},
		{name: "zero goal", modify: func(t *models.Tournament) { t.Funding.GoalAmount = 0 }, tiersCount: 1, want: usecases.ErrCrowdfundingSetupIncomplete},
		{name: "past deadline", modify: func(t *models.Tournament) { past := now.Add(-time.Hour); t.Funding.Deadline = &past }, tiersCount: 1, want: usecases.ErrCrowdfundingSetupIncomplete},
		{name: "no active tiers", modify: nil, tiersCount: 0, want: usecases.ErrCrowdfundingSetupIncomplete},
		{name: "funding deadline after registration deadline", modify: func(t *models.Tournament) {
			after := t.RegistrationDeadline.Add(time.Hour)
			t.Funding.Deadline = &after
		}, tiersCount: 1, want: usecases.ErrCrowdfundingSetupIncomplete},
		{name: "valid transition", modify: nil, tiersCount: 2, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := draftCrowdfundingReadyTournament(ownerID)
			if tt.modify != nil {
				tt.modify(current)
			}
			repo := &statusRepositoryStub{current: models.TournamentStatusDraft, tournament: current, gotAccountID: ownerID}
			repo.activeTiersCount = tt.tiersCount
			useCase := usecases.NewTournamentUseCase(repo, newStubAccountService())

			res, err := useCase.Override(context.Background(), ownerID, tournamentID, "CROWDFUNDING", "")
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected error %v, got %v", tt.want, err)
			}
			if tt.want == nil && (res == nil || res.ToStatus != models.TournamentStatusCrowdfunding) {
				t.Fatalf("expected transition to CROWDFUNDING, got %+v", res)
			}
		})
	}
}

func TestDraftToRegistrationValidation(t *testing.T) {
	ownerID, tournamentID := uuid.New(), uuid.New()

	tests := []struct {
		name   string
		modify func(*models.Tournament)
		want   error
	}{
		{name: "missing timeline", modify: func(t *models.Tournament) { t.StartsAt = time.Time{}; t.EndsAt = time.Time{} }, want: usecases.ErrRegistrationGateIncomplete},
		{name: "past registration deadline", modify: func(t *models.Tournament) { t.RegistrationDeadline = time.Now().Add(-time.Hour) }, want: usecases.ErrRegistrationGateIncomplete},
		{name: "zero capacity", modify: func(t *models.Tournament) { t.Capacity = 0 }, want: usecases.ErrRegistrationGateIncomplete},
		{name: "negative fee", modify: func(t *models.Tournament) { t.EntryFee = -5 }, want: usecases.ErrRegistrationGateIncomplete},
		{name: "invalid team size", modify: func(t *models.Tournament) { t.MinTeamSize = 5; t.MaxTeamSize = 2 }, want: usecases.ErrRegistrationGateIncomplete},
		{name: "valid direct transition", modify: nil, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := draftRegistrationReadyTournament(ownerID)
			if tt.modify != nil {
				tt.modify(current)
			}
			repo := &statusRepositoryStub{current: models.TournamentStatusDraft, tournament: current, gotAccountID: ownerID}
			useCase := usecases.NewTournamentUseCase(repo, newStubAccountService())

			res, err := useCase.Override(context.Background(), ownerID, tournamentID, "REGISTRATION", "")
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected error %v, got %v", tt.want, err)
			}
			if tt.want == nil && (res == nil || res.ToStatus != models.TournamentStatusRegistration) {
				t.Fatalf("expected transition to REGISTRATION, got %+v", res)
			}
		})
	}
}

func TestCrowdfundingLaunchRequiresOwner(t *testing.T) {
	ownerID, staffID, tournamentID := uuid.New(), uuid.New(), uuid.New()
	current := crowdfundingReadyTournament(ownerID)
	current.Staff = append(current.Staff, models.TournamentStaff{AccountID: staffID, Role: models.TournamentStaffRoleAdmin})
	current.Funding.RaisedAmount = 200
	repo := &statusRepositoryStub{current: models.TournamentStatusCrowdfunding, tournament: current, gotAccountID: ownerID}
	useCase := usecases.NewTournamentUseCase(repo, newStubAccountService())

	if _, err := useCase.OverrideWithForce(context.Background(), staffID, tournamentID, "REGISTRATION", "", true); !errors.Is(err, usecases.ErrNotTournamentOwner) {
		t.Fatalf("expected owner check, got %v", err)
	}
	if repo.persistedTo != "" {
		t.Fatal("staff member must not launch")
	}
}

func TestOverrideRejectsForbiddenTransitions(t *testing.T) {
	tests := []struct {
		name string
		from models.TournamentStatus
		to   models.TournamentStatus
	}{
		// COMPLETED is a read-only terminal status.
		{"reopen a completed tournament", models.TournamentStatusCompleted, models.TournamentStatusOngoing},
		// A running tournament cannot move back to its draft or funding stage.
		{"regress registration to draft", models.TournamentStatusRegistration, models.TournamentStatusDraft},
		{"regress ongoing to crowdfunding", models.TournamentStatusOngoing, models.TournamentStatusCrowdfunding},
		{"regress ongoing to registration", models.TournamentStatusOngoing, models.TournamentStatusRegistration},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accountID, tournamentID := uuid.New(), uuid.New()
			repo := &statusRepositoryStub{current: test.from, gotAccountID: accountID}
			if test.from == models.TournamentStatusDraft && test.to == models.TournamentStatusCrowdfunding {
				repo.tournament = draftCrowdfundingReadyTournament(accountID)
				repo.activeTiersCount = 1
			}
			if test.from == models.TournamentStatusDraft && test.to == models.TournamentStatusRegistration {
				repo.tournament = draftRegistrationReadyTournament(accountID)
			}
			if test.from == models.TournamentStatusCrowdfunding {
				repo.tournament = crowdfundingReadyTournament(accountID)
			}
			if test.from == models.TournamentStatusRegistration && test.to == models.TournamentStatusOngoing {
				repo.tournament = &models.Tournament{
					ID:                   tournamentID,
					Status:               models.TournamentStatusRegistration,
					RegistrationDeadline: time.Now().Add(-time.Hour),
					Staff:                []models.TournamentStaff{{AccountID: accountID, Role: models.TournamentStaffRoleOwner}},
				}
			}

			_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
				Override(context.Background(), accountID, tournamentID, string(test.to), "")
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
	accountID, tournamentID := uuid.New(), uuid.New()
	repo := &statusRepositoryStub{current: models.TournamentStatusOngoing, gotAccountID: accountID}

	_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
		Override(context.Background(), accountID, tournamentID, "ONGOING", "")
	if !errors.Is(err, usecases.ErrStatusUnchanged) {
		t.Fatalf("expected unchanged-status error, got %v", err)
	}
	if repo.persistedTo != "" {
		t.Fatal("expected no change persisted for a no-op override")
	}
}

func TestOverrideNormalisesAndValidatesStatusInput(t *testing.T) {
	t.Run("accepts lowercase and padding", func(t *testing.T) {
		accountID, tournamentID := uuid.New(), uuid.New()
		repo := &statusRepositoryStub{current: models.TournamentStatusRegistration, gotAccountID: accountID}
		if _, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
			Override(context.Background(), accountID, tournamentID, "  ongoing  ", ""); err != nil {
			t.Fatalf("expected normalised status to be accepted, got %v", err)
		}
		if repo.persistedTo != models.TournamentStatusOngoing {
			t.Fatalf("expected ONGOING persisted, got %s", repo.persistedTo)
		}
	})

	t.Run("rejects an unknown status", func(t *testing.T) {
		accountID, tournamentID := uuid.New(), uuid.New()
		repo := &statusRepositoryStub{current: models.TournamentStatusRegistration, gotAccountID: accountID}
		_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
			Override(context.Background(), accountID, tournamentID, "BANANA", "")
		if !errors.Is(err, usecases.ErrUnknownStatus) {
			t.Fatalf("expected unknown-status error, got %v", err)
		}
		if repo.called {
			t.Fatal("expected the repository never to be reached for an unknown status")
		}
	})
}

func TestOverrideTranslatesOwnershipFailureToNotFound(t *testing.T) {
	repo := &statusRepositoryStub{
		current: models.TournamentStatusRegistration,
		err:     repositories.ErrTournamentNotFound,
	}

	accountID, tournamentID := uuid.New(), uuid.New()
	_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).
		Override(context.Background(), accountID, tournamentID, "ONGOING", "")
	if !errors.Is(err, usecases.ErrTournamentNotFound) {
		t.Fatalf("expected tournament-not-found error, got %v", err)
	}
}

func TestAllowedTransitionsFromIsTerminalForFinishedTournaments(t *testing.T) {
	for _, status := range []models.TournamentStatus{
		models.TournamentStatusCompleted,
	} {
		if got := usecases.AllowedTransitionsFrom(status); len(got) != 0 {
			t.Fatalf("expected %s to be terminal, got %v", status, got)
		}
	}
}

func TestAllowedTransitionsFromReturnsACopy(t *testing.T) {
	first := usecases.AllowedTransitionsFrom(models.TournamentStatusRegistration)
	if len(first) == 0 {
		t.Fatal("expected REGISTRATION to have transitions")
	}
	first[0] = "MUTATED"

	if second := usecases.AllowedTransitionsFrom(models.TournamentStatusRegistration); second[0] == "MUTATED" {
		t.Fatal("expected AllowedTransitionsFrom to return a defensive copy")
	}
}

func TestRegistrationToOngoingValidation(t *testing.T) {
	ownerID, tournamentID := uuid.New(), uuid.New()

	tests := []struct {
		name         string
		openDeadline bool
		bracketErr   error
		want         error
	}{
		{name: "registration still open", openDeadline: true, want: usecases.ErrRegistrationStillOpen},
		{name: "unresolved applications", bracketErr: usecases.ErrUnresolvedApplicationsExist, want: usecases.ErrUnresolvedApplicationsExist},
		{name: "insufficient accepted teams", bracketErr: usecases.ErrInsufficientAcceptedTeams, want: usecases.ErrInsufficientAcceptedTeams},
		{name: "missing bracket", bracketErr: usecases.ErrBracketMissing, want: usecases.ErrBracketMissing},
		{name: "stage 1 capacity mismatch", bracketErr: usecases.ErrBracketStage1Mismatch, want: usecases.ErrBracketStage1Mismatch},
		{name: "bracket not seeded", bracketErr: usecases.ErrBracketNotSeeded, want: usecases.ErrBracketNotSeeded},
		{name: "valid transition to ongoing", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deadline := time.Now().Add(-time.Hour)
			if tt.openDeadline {
				deadline = time.Now().Add(time.Hour)
			}
			tourn := &models.Tournament{
				ID:                   tournamentID,
				Status:               models.TournamentStatusRegistration,
				RegistrationDeadline: deadline,
				Staff:                []models.TournamentStaff{{AccountID: ownerID, Role: models.TournamentStaffRoleOwner}},
			}
			repo := &statusRepositoryStub{current: models.TournamentStatusRegistration, tournament: tourn, gotAccountID: ownerID}
			repo.bracketReadinessErr = tt.bracketErr
			useCase := usecases.NewTournamentUseCase(repo, newStubAccountService())

			res, err := useCase.Override(context.Background(), ownerID, tournamentID, "ONGOING", "")
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected error %v, got %v", tt.want, err)
			}
			if tt.want == nil && (res == nil || res.ToStatus != models.TournamentStatusOngoing) {
				t.Fatalf("expected transition to ONGOING, got %+v", res)
			}
		})
	}
}

func TestOngoingToCompletedValidation(t *testing.T) {
	ownerID, tournamentID := uuid.New(), uuid.New()

	tests := []struct {
		name          string
		completionErr error
		want          error
	}{
		{name: "unfinished matches exist", completionErr: usecases.ErrUnfinishedMatchesExist, want: usecases.ErrUnfinishedMatchesExist},
		{name: "no single championship winner", completionErr: usecases.ErrNoChampionshipWinner, want: usecases.ErrNoChampionshipWinner},
		{name: "valid completion", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tourn := &models.Tournament{
				ID:     tournamentID,
				Status: models.TournamentStatusOngoing,
				Staff:  []models.TournamentStaff{{AccountID: ownerID, Role: models.TournamentStaffRoleOwner}},
			}
			repo := &statusRepositoryStub{current: models.TournamentStatusOngoing, tournament: tourn, gotAccountID: ownerID}
			repo.completionReadinessErr = tt.completionErr
			useCase := usecases.NewTournamentUseCase(repo, newStubAccountService())

			res, err := useCase.Override(context.Background(), ownerID, tournamentID, "COMPLETED", "")
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected error %v, got %v", tt.want, err)
			}
			if tt.want == nil && (res == nil || res.ToStatus != models.TournamentStatusCompleted) {
				t.Fatalf("expected transition to COMPLETED, got %+v", res)
			}
		})
	}
}
