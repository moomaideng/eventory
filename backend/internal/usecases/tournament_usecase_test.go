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

type mockTournamentRepository struct {
	filters    repositories.TournamentFilters
	items      []repositories.TournamentSearchItem
	total      int64
	err        error
	details    *models.Tournament
	detailsErr error
}

func (m *mockTournamentRepository) GetPublishedByID(
	_ context.Context,
	_ uuid.UUID,
) (*models.Tournament, error) {
	return m.details, m.detailsErr
}

func (m *mockTournamentRepository) Search(
	_ context.Context,
	filters repositories.TournamentFilters,
) ([]repositories.TournamentSearchItem, int64, error) {
	m.filters = filters
	return m.items, m.total, m.err
}

func TestSearchTournaments_NormalizesFilters(t *testing.T) {
	repo := &mockTournamentRepository{
		items: []repositories.TournamentSearchItem{{
			Tournament: models.Tournament{ID: uuid.New()}, RegisteredCount: 3,
		}},
		total: 5,
	}
	useCase := usecases.NewTournamentUseCase(repo)
	maxFee := int64(500)

	result, err := useCase.Search(context.Background(), usecases.SearchTournamentsInput{
		Query: "  valorant  ", StartFrom: "2026-09-01", StartTo: "2026-09-30",
		MaxEntryFee: &maxFee, Status: "registration_open",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.filters.Query != "valorant" || repo.filters.Status != models.TournamentStatusRegistrationOpen {
		t.Fatalf("filters were not normalized: %+v", repo.filters)
	}
	if repo.filters.Page != 1 || repo.filters.PageSize != 12 || repo.filters.Sort != "start_asc" {
		t.Fatalf("defaults were not applied: %+v", repo.filters)
	}
	if repo.filters.StartFrom == nil || repo.filters.StartTo == nil {
		t.Fatal("expected parsed date boundaries")
	}
	if result.Total != 5 {
		t.Fatalf("expected total 5, got %d", result.Total)
	}
	if len(result.Items) != 1 || result.Items[0].RegisteredCount != 3 {
		t.Fatalf("expected calculated registered count, got %+v", result.Items)
	}
}

func TestSearchTournaments_RejectsInvalidRanges(t *testing.T) {
	tests := []struct {
		name  string
		input usecases.SearchTournamentsInput
	}{ // mock test
		{name: "reversed dates", input: usecases.SearchTournamentsInput{StartFrom: "2026-10-01", StartTo: "2026-09-01"}},
		{name: "invalid date", input: usecases.SearchTournamentsInput{StartFrom: "09/01/2026"}},
		{name: "negative fee", input: usecases.SearchTournamentsInput{MaxEntryFee: int64Pointer(-1)}},
		{name: "reversed fees", input: usecases.SearchTournamentsInput{MinEntryFee: int64Pointer(501), MaxEntryFee: int64Pointer(500)}},
		{name: "invalid status", input: usecases.SearchTournamentsInput{Status: "DRAFT"}},
		{name: "oversized page", input: usecases.SearchTournamentsInput{PageSize: 101}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{})
			_, err := useCase.Search(context.Background(), test.input)
			if !errors.Is(err, usecases.ErrInvalidTournamentFilters) {
				t.Fatalf("expected invalid filters error, got %v", err)
			}
		})
	}
}

func TestGetTournamentDetails_CalculatesFundingStats(t *testing.T) {
	tournamentID := uuid.New()
	repo := &mockTournamentRepository{details: &models.Tournament{
		ID: tournamentID,
		Teams: []models.TournamentTeam{
			{
				ID: uuid.New(), TournamentID: tournamentID, Name: "Alpha",
				Status: models.TournamentTeamStatusAccepted,
				Members: []models.TournamentTeamMember{
					{ID: uuid.New(), Role: models.TournamentTeamMemberRoleCaptain},
					{ID: uuid.New(), Role: models.TournamentTeamMemberRoleMember},
				},
			},
			{
				ID: uuid.New(), TournamentID: tournamentID, Name: "Bravo",
				Status: models.TournamentTeamStatusAccepted,
				Members: []models.TournamentTeamMember{
					{ID: uuid.New(), Role: models.TournamentTeamMemberRoleCaptain},
				},
			},
			{
				ID: uuid.New(), TournamentID: tournamentID, Name: "Still Forming",
				Status: models.TournamentTeamStatusForming,
			},
		},
		Funding: &models.TournamentFunding{
			GoalAmount: 10000, RaisedAmount: 6250, SupporterCount: 14,
		},
	}}

	result, err := usecases.NewTournamentUseCase(repo).GetDetails(context.Background(), tournamentID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Funding.Percentage != 62.5 || result.Funding.RemainingAmount != 3750 {
		t.Fatalf("unexpected funding stats: %+v", result.Funding)
	}
	if result.RegisteredCount != 2 {
		t.Fatalf("expected 2 registered teams, got %d", result.RegisteredCount)
	}
	if len(result.Tournament.Teams) != 2 {
		t.Fatalf("expected only accepted teams, got %d", len(result.Tournament.Teams))
	}
}

func TestGetTournamentDetails_MapsNotFound(t *testing.T) {
	repo := &mockTournamentRepository{detailsErr: repositories.ErrTournamentNotFound}

	_, err := usecases.NewTournamentUseCase(repo).GetDetails(context.Background(), uuid.New())
	if !errors.Is(err, usecases.ErrTournamentNotFound) {
		t.Fatalf("expected tournament not found, got %v", err)
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}
