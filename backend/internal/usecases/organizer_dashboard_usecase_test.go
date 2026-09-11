package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
	"github.com/moomaideng/eventory/internal/usecases"
)

type dashboardRepositoryStub struct {
	tournament models.Tournament
	accountID  uuid.UUID
	err        error
}

func (r *dashboardRepositoryStub) ListOwned(_ context.Context, accountID uuid.UUID, _, _ int) ([]models.Tournament, int64, error) {
	r.accountID = accountID
	return []models.Tournament{r.tournament}, 1, r.err
}

func (r *dashboardRepositoryStub) GetOwned(_ context.Context, accountID, _ uuid.UUID) (*models.Tournament, error) {
	r.accountID = accountID
	return &r.tournament, r.err
}

func TestOrganizerDashboardCountsConfirmedParticipantsAndEntryStates(t *testing.T) {
	participant := uuid.New()
	otherParticipant := uuid.New()
	repo := &dashboardRepositoryStub{tournament: models.Tournament{
		Capacity: 1,
		Teams: []models.TournamentTeam{
			{Status: models.TournamentTeamStatusForming, Members: []models.TournamentTeamMember{{AccountID: uuid.New()}}},
			{Status: models.TournamentTeamStatusLocked, Members: []models.TournamentTeamMember{{AccountID: uuid.New()}}},
			{Status: models.TournamentTeamStatusRejected, Members: []models.TournamentTeamMember{{AccountID: uuid.New()}}},
			{Status: models.TournamentTeamStatusAccepted, Members: []models.TournamentTeamMember{{AccountID: participant}}},
			{Status: models.TournamentTeamStatusAccepted, Members: []models.TournamentTeamMember{{AccountID: participant}, {AccountID: otherParticipant}}},
		},
		Funding: &models.TournamentFunding{GoalAmount: 100, RaisedAmount: 150, SupporterCount: 3},
	}}
	accountID := uuid.New()
	result, err := usecases.NewOrganizerDashboardUseCase(repo).Get(context.Background(), accountID, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	want := usecases.OrganizerDashboardMetrics{TotalEntries: 5, FormingEntries: 1, LockedEntries: 1, AcceptedEntries: 2, RejectedEntries: 1, ConfirmedParticipants: 2, AvailableSpots: 0}
	if result.Metrics != want {
		t.Fatalf("metrics = %+v, want %+v", result.Metrics, want)
	}
	if result.Funding.Percentage != 150 || result.Funding.RemainingAmount != 0 || result.Funding.SupporterCount != 3 {
		t.Fatalf("unexpected funding: %+v", result.Funding)
	}
	if repo.accountID != accountID {
		t.Fatal("authenticated account was not passed to ownership query")
	}
}

func TestOrganizerDashboardEmptyAndZeroGoal(t *testing.T) {
	for _, funding := range []*models.TournamentFunding{nil, {RaisedAmount: 50}} {
		repo := &dashboardRepositoryStub{tournament: models.Tournament{Capacity: 16, Funding: funding}}
		result, err := usecases.NewOrganizerDashboardUseCase(repo).Get(context.Background(), uuid.New(), uuid.New())
		if err != nil {
			t.Fatal(err)
		}
		if result.Metrics.TotalEntries != 0 || result.Metrics.ConfirmedParticipants != 0 || result.Metrics.AvailableSpots != 16 || result.Funding.Percentage != 0 {
			t.Fatalf("unexpected empty dashboard: %+v", result)
		}
	}
}

func TestOrganizerDashboardNotOwnedOrMissing(t *testing.T) {
	repo := &dashboardRepositoryStub{err: repositories.ErrTournamentNotFound}
	_, err := usecases.NewOrganizerDashboardUseCase(repo).Get(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, usecases.ErrTournamentNotFound) {
		t.Fatalf("error = %v, want not found", err)
	}
}

func TestOrganizerDashboardListPaginationAndOwnership(t *testing.T) {
	repo := &dashboardRepositoryStub{}
	u := usecases.NewOrganizerDashboardUseCase(repo)
	accountID := uuid.New()
	result, err := u.List(context.Background(), accountID, 2, 12)
	if err != nil || result.Page != 2 || result.PageSize != 12 || result.Total != 1 || repo.accountID != accountID {
		t.Fatalf("unexpected list result: %+v, error: %v", result, err)
	}
	for _, pair := range [][2]int{{0, 12}, {1, 0}, {1, 101}, {1000001, 12}} {
		_, err := u.List(context.Background(), accountID, pair[0], pair[1])
		if !errors.Is(err, usecases.ErrInvalidDashboardPage) {
			t.Fatalf("pagination %v: expected validation error, got %v", pair, err)
		}
	}
}
