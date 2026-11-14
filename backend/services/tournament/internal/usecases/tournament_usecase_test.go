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

type mockTournamentRepository struct {
	filters           repositories.TournamentFilters
	items             []repositories.TournamentSearchItem
	total             int64
	err               error
	details           *models.Tournament
	detailsErr        error
	createdTournament *models.Tournament
	createErr         error
	updatedTournament *models.Tournament
	updateErr         error
	activeTiersCount  int64
	tiersCountErr     error
}

func (m *mockTournamentRepository) GetPublishedByID(_ context.Context, _ uuid.UUID) (*models.Tournament, error) {
	return m.details, m.detailsErr
}

func (m *mockTournamentRepository) GetByID(_ context.Context, _ uuid.UUID) (*models.Tournament, error) {
	return m.details, m.detailsErr
}

func (m *mockTournamentRepository) Create(_ context.Context, tournament *models.Tournament) error {
	m.createdTournament = tournament
	return m.createErr
}

func (m *mockTournamentRepository) Update(_ context.Context, tournament *models.Tournament) error {
	m.updatedTournament = tournament
	return m.updateErr
}

func (m *mockTournamentRepository) Search(_ context.Context, filters repositories.TournamentFilters) ([]repositories.TournamentSearchItem, int64, error) {
	m.filters = filters
	return m.items, m.total, m.err
}

func (m *mockTournamentRepository) ListOwned(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Tournament, int64, error) {
	return nil, 0, nil
}

func (m *mockTournamentRepository) GetOwned(_ context.Context, _, _ uuid.UUID) (*models.Tournament, error) {
	return nil, nil
}

func (m *mockTournamentRepository) UpdateStatus(_ context.Context, _ uuid.UUID, _, _ models.TournamentStatus) error {
	return nil
}

func (m *mockTournamentRepository) CountActiveTiers(_ context.Context, _ uuid.UUID) (int64, error) {
	if m.tiersCountErr != nil {
		return 0, m.tiersCountErr
	}
	return m.activeTiersCount, nil
}

func (m *mockTournamentRepository) GetTeamCounts(_ context.Context, _ uuid.UUID) (int64, int64, error) {
	return 2, 0, nil
}

func (m *mockTournamentRepository) GetBracketMatches(_ context.Context, _ uuid.UUID) ([]models.Match, error) {
	rank1, rank2 := 1, 2
	return []models.Match{
		{
			ID:          uuid.New(),
			StageLevel:  1,
			TeamCount:   2,
			IsCompleted: true,
			Entries: []models.MatchEntry{
				{FinalRank: &rank1},
				{FinalRank: &rank2},
			},
		},
	}, nil
}

func ptr[T any](v T) *T { return &v }

func baseTournament(ownerID uuid.UUID) *models.Tournament {
	now := time.Now()
	return &models.Tournament{
		ID: uuid.New(),
		Staff: []models.TournamentStaff{
			{
				AccountID: ownerID,
				Role:      models.TournamentStaffRoleOwner,
			},
		},
		Name:                 "Original Name",
		Description:          "Original Description",
		Game:                 "Original Game",
		Location:             "Original Location",
		Status:               models.TournamentStatusDraft,
		RegistrationMode:     models.TournamentRegistrationModeTeam,
		MinTeamSize:          3,
		MaxTeamSize:          5,
		Capacity:             16,
		RegistrationDeadline: now.Add(24 * time.Hour),
		StartsAt:             now.Add(48 * time.Hour),
		EndsAt:               now.Add(72 * time.Hour),
		EntryFeeMinor:        10000,
		Currency:             "THB",
	}
}

func validCreateInput(now time.Time) usecases.CreateTournamentInput {
	start := now.Add(48 * time.Hour)
	end := now.Add(72 * time.Hour)
	deadline := now.Add(24 * time.Hour)
	mode := "TEAM"
	minSize := 3
	maxSize := 5
	cap := 16
	fee := int64(10000)
	return usecases.CreateTournamentInput{
		Name:                 "Valid Cup",
		Description:          "Tournament rules",
		Detail:               "## Full Rules\n\n- Fair play only",
		Game:                 "Chess",
		Location:             "Online",
		StartsAt:             &start,
		EndsAt:               &end,
		RegistrationDeadline: &deadline,
		RegistrationMode:     &mode,
		MinTeamSize:          &minSize,
		MaxTeamSize:          &maxSize,
		Capacity:             &cap,
		EntryFeeMinor:        &fee,
	}
}

func TestSearchTournaments_NormalizesFilters(t *testing.T) {
	repo := &mockTournamentRepository{
		items: []repositories.TournamentSearchItem{{
			Tournament: models.Tournament{ID: uuid.New()}, RegisteredCount: 3,
		}},
		total: 5,
	}
	useCase := usecases.NewTournamentUseCase(repo, newStubAccountService())

	result, err := useCase.Search(context.Background(), usecases.SearchTournamentsInput{
		Query: "  valorant  ", StartFrom: "2026-09-01", StartTo: "2026-09-30",
		MaxEntryFeeMinor: ptr(int64(50000)), Status: "registration",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.filters.Query != "valorant" || repo.filters.Status != models.TournamentStatusRegistration {
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
	}{
		{name: "reversed dates", input: usecases.SearchTournamentsInput{StartFrom: "2026-10-01", StartTo: "2026-09-01"}},
		{name: "invalid date", input: usecases.SearchTournamentsInput{StartFrom: "09/01/2026"}},
		{name: "negative fee", input: usecases.SearchTournamentsInput{MaxEntryFeeMinor: ptr(int64(-1))}},
		{name: "reversed fees", input: usecases.SearchTournamentsInput{MinEntryFeeMinor: ptr(int64(50100)), MaxEntryFeeMinor: ptr(int64(50000))}},
		{name: "invalid status", input: usecases.SearchTournamentsInput{Status: "NOT_A_STATUS"}},
		{name: "oversized page", input: usecases.SearchTournamentsInput{PageSize: 101}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{}, newStubAccountService())
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
				ID: "TM000001", TournamentID: tournamentID, Name: "Alpha",
				Status: models.TournamentTeamStatusAccepted,
				Members: []models.TournamentTeamMember{
					{ID: uuid.New(), Role: models.TournamentTeamMemberRoleCaptain},
					{ID: uuid.New(), Role: models.TournamentTeamMemberRoleMember},
				},
			},
			{
				ID: "TM000002", TournamentID: tournamentID, Name: "Bravo",
				Status: models.TournamentTeamStatusAccepted,
				Members: []models.TournamentTeamMember{
					{ID: uuid.New(), Role: models.TournamentTeamMemberRoleCaptain},
				},
			},
			{
				ID: "TM000003", TournamentID: tournamentID, Name: "Still Forming",
				Status: models.TournamentTeamStatusForming,
			},
		},
		Crowdfunding: models.CrowdfundingInfo{
			GoalAmountMinor: 1000000,
			Pledges:         []models.SponsorshipPledge{{Status: models.PledgeConfirmed, AmountMinor: 625000}},
		},
	}}

	result, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).GetDetails(context.Background(), tournamentID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Funding.Percentage != 62.5 || result.Funding.RemainingAmountMinor != 375000 {
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

	_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).GetDetails(context.Background(), uuid.New())
	if !errors.Is(err, usecases.ErrTournamentNotFound) {
		t.Fatalf("expected tournament not found, got %v", err)
	}
}

func TestCreateTournament_Success(t *testing.T) {
	organizerID := uuid.New()
	now := time.Now()

	t.Run("solo tournament creates in DRAFT status", func(t *testing.T) {
		repo := &mockTournamentRepository{}
		useCase := usecases.NewTournamentUseCase(repo, newStubAccountService())
		in := validCreateInput(now)
		in.RegistrationMode = ptr("SOLO")
		in.MinTeamSize, in.MaxTeamSize = ptr(5), ptr(5)

		got, err := useCase.CreateTournament(context.Background(), organizerID, in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Staff) != 1 || got.Staff[0].AccountID != organizerID || got.Staff[0].Role != models.TournamentStaffRoleOwner || got.Currency != "THB" ||
			got.Status != models.TournamentStatusDraft ||
			got.MinTeamSize != 1 || got.MaxTeamSize != 1 ||
			got.Crowdfunding.GoalAmountMinor != 0 || got.Crowdfunding.Deadline != nil {
			t.Fatalf("unexpected tournament state: %+v", got)
		}
	})

	t.Run("crowdfunding in DRAFT with configured goal and zero raised", func(t *testing.T) {
		repo := &mockTournamentRepository{}
		input := validCreateInput(now)
		goal := int64(2500000)
		deadline := now.Add(12 * time.Hour)
		input.FundingGoalMinor, input.FundingDeadline = &goal, &deadline
		got, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).CreateTournament(context.Background(), organizerID, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Status != models.TournamentStatusDraft ||
			got.Crowdfunding.GoalAmountMinor != goal || got.Crowdfunding.Deadline == nil || !got.Crowdfunding.Deadline.Equal(deadline) {
			t.Fatalf("unexpected crowdfunding state: tournament=%+v", got)
		}
	})

	t.Run("team tournament preserves team sizes", func(t *testing.T) {
		repo := &mockTournamentRepository{}
		got, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).CreateTournament(context.Background(), organizerID, validCreateInput(now))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.RegistrationMode != models.TournamentRegistrationModeTeam || got.MinTeamSize != 3 || got.MaxTeamSize != 5 {
			t.Fatalf("unexpected team tournament state: %+v", got)
		}
	})

	t.Run("minimal tournament creation in DRAFT with optional fields omitted", func(t *testing.T) {
		repo := &mockTournamentRepository{}
		minimal := usecases.CreateTournamentInput{
			Name:        "Minimal Cup",
			Description: "Short description",
			Detail:      "Detail here",
			Game:        "Dota 2",
			Location:    "Bangkok",
		}
		got, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).CreateTournament(context.Background(), organizerID, minimal)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Status != models.TournamentStatusDraft || !got.StartsAt.IsZero() || !got.EndsAt.IsZero() || !got.RegistrationDeadline.IsZero() || got.Capacity != 0 {
			t.Fatalf("expected minimal draft tournament, got: %+v", got)
		}
	})
}

func TestCreateTournament_ValidationErrors(t *testing.T) {
	organizerID := uuid.New()
	now := time.Now()

	tests := []struct {
		name    string
		modify  func(*usecases.CreateTournamentInput)
		wantErr error
	}{
		{name: "empty name", modify: func(in *usecases.CreateTournamentInput) { in.Name = "   " }, wantErr: usecases.ErrInvalidTournamentName},
		{name: "empty game", modify: func(in *usecases.CreateTournamentInput) { in.Game = "" }, wantErr: usecases.ErrInvalidTournamentGame},
		{name: "empty location", modify: func(in *usecases.CreateTournamentInput) { in.Location = "  " }, wantErr: usecases.ErrInvalidTournamentLocation},
		{name: "deadline after start", modify: func(in *usecases.CreateTournamentInput) { in.RegistrationDeadline = ptr(in.StartsAt.Add(time.Hour)) }, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "deadline equals start", modify: func(in *usecases.CreateTournamentInput) { in.RegistrationDeadline = in.StartsAt }, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "start after end", modify: func(in *usecases.CreateTournamentInput) { in.StartsAt = ptr(in.EndsAt.Add(time.Hour)) }, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "start equals end", modify: func(in *usecases.CreateTournamentInput) { in.StartsAt = in.EndsAt }, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "negative fee", modify: func(in *usecases.CreateTournamentInput) { in.EntryFeeMinor = ptr(int64(-10)) }, wantErr: usecases.ErrInvalidTournamentFee},
		{name: "zero capacity", modify: func(in *usecases.CreateTournamentInput) { in.Capacity = ptr(0) }, wantErr: usecases.ErrInvalidTournamentCapacity},
		{name: "invalid mode", modify: func(in *usecases.CreateTournamentInput) { in.RegistrationMode = ptr("SQUAD") }, wantErr: usecases.ErrInvalidRegistrationMode},
		{name: "team mode min > max", modify: func(in *usecases.CreateTournamentInput) { in.MinTeamSize, in.MaxTeamSize = ptr(5), ptr(3) }, wantErr: usecases.ErrInvalidTeamSize},
		{name: "team mode min < 1", modify: func(in *usecases.CreateTournamentInput) { in.MinTeamSize = ptr(0) }, wantErr: usecases.ErrInvalidTeamSize},
		{name: "description exceeds max length", modify: func(in *usecases.CreateTournamentInput) { in.Description = strings.Repeat("a", 501) }, wantErr: usecases.ErrInvalidTournamentDescription},
		{name: "detail exceeds max length", modify: func(in *usecases.CreateTournamentInput) { in.Detail = strings.Repeat("b", 10001) }, wantErr: usecases.ErrInvalidTournamentDetail},
		{name: "funding goal without deadline", modify: func(in *usecases.CreateTournamentInput) { in.FundingGoalMinor = ptr(int64(100000)) }, wantErr: usecases.ErrInvalidCrowdfunding},
		{name: "funding deadline without goal", modify: func(in *usecases.CreateTournamentInput) { in.FundingDeadline = ptr(now.Add(time.Hour)) }, wantErr: usecases.ErrInvalidCrowdfunding},
		{name: "zero funding goal", modify: func(in *usecases.CreateTournamentInput) {
			in.FundingGoalMinor = ptr(int64(0))
			in.FundingDeadline = ptr(now.Add(time.Hour))
		}, wantErr: usecases.ErrInvalidCrowdfunding},
		{name: "past funding deadline", modify: func(in *usecases.CreateTournamentInput) {
			in.FundingGoalMinor = ptr(int64(100000))
			in.FundingDeadline = ptr(now.Add(-time.Hour))
		}, wantErr: usecases.ErrInvalidCrowdfunding},
		{name: "funding after registration deadline", modify: func(in *usecases.CreateTournamentInput) {
			in.FundingGoalMinor = ptr(int64(100000))
			in.FundingDeadline = in.RegistrationDeadline
		}, wantErr: usecases.ErrInvalidCrowdfunding},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validCreateInput(now)
			tt.modify(&input)
			_, err := usecases.NewTournamentUseCase(&mockTournamentRepository{}, newStubAccountService()).CreateTournament(context.Background(), organizerID, input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestUpdateTournament_Guards(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()

	tests := []struct {
		name     string
		setup    func(*models.Tournament, *mockTournamentRepository)
		callerID uuid.UUID
		input    usecases.UpdateTournamentInput
		wantErr  error
	}{
		{
			name:     "non-owner rejected",
			callerID: uuid.New(),
			input:    usecases.UpdateTournamentInput{Name: ptr("New Name")},
			wantErr:  usecases.ErrNotTournamentOwner,
		},
		{
			name:     "ongoing status rejects non-metadata updates",
			setup:    func(tm *models.Tournament, _ *mockTournamentRepository) { tm.Status = models.TournamentStatusOngoing },
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Capacity: ptr(32)},
			wantErr:  usecases.ErrTournamentCannotBeModified,
		},
		{
			name:     "ongoing status allows general info updates",
			setup:    func(tm *models.Tournament, _ *mockTournamentRepository) { tm.Status = models.TournamentStatusOngoing },
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Name: ptr("Ongoing Name")},
			wantErr:  nil,
		},
		{
			name:     "completed status rejected",
			setup:    func(tm *models.Tournament, _ *mockTournamentRepository) { tm.Status = models.TournamentStatusCompleted },
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Name: ptr("New Name")},
			wantErr:  usecases.ErrTournamentCannotBeModified,
		},
		{
			name: "registration status rejects mode change",
			setup: func(tm *models.Tournament, _ *mockTournamentRepository) {
				tm.Status = models.TournamentStatusRegistration
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{RegistrationMode: ptr("SOLO")},
			wantErr:  usecases.ErrRosterRulesLocked,
		},
		{
			name: "registration status rejects min team size change",
			setup: func(tm *models.Tournament, _ *mockTournamentRepository) {
				tm.Status = models.TournamentStatusRegistration
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{MinTeamSize: ptr(2)},
			wantErr:  usecases.ErrRosterRulesLocked,
		},
		{
			name: "registration status rejects max team size change",
			setup: func(tm *models.Tournament, _ *mockTournamentRepository) {
				tm.Status = models.TournamentStatusRegistration
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{MaxTeamSize: ptr(6)},
			wantErr:  usecases.ErrRosterRulesLocked,
		},
		{
			name: "registration status rejects capacity change",
			setup: func(tm *models.Tournament, _ *mockTournamentRepository) {
				tm.Status = models.TournamentStatusRegistration
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Capacity: ptr(32)},
			wantErr:  usecases.ErrRosterRulesLocked,
		},
		{
			name: "registration permitted",
			setup: func(tm *models.Tournament, _ *mockTournamentRepository) {
				tm.Status = models.TournamentStatusRegistration
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Name: ptr("Closed Cup")},
			wantErr:  nil,
		},
		{
			name: "capacity update in draft permitted",
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Capacity: ptr(8)},
			wantErr:  nil,
		},
		{
			name: "idempotent roster update permitted in registration",
			setup: func(tm *models.Tournament, repo *mockTournamentRepository) {
				tm.Status = models.TournamentStatusRegistration
				tm.RegistrationMode = models.TournamentRegistrationModeTeam
				tm.MinTeamSize, tm.MaxTeamSize = 3, 5
			},
			callerID: ownerID,
			input: usecases.UpdateTournamentInput{
				RegistrationMode: ptr("TEAM"),
				MinTeamSize:      ptr(3),
				MaxTeamSize:      ptr(5),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := baseTournament(ownerID)
			tm.ID = tournamentID
			repo := &mockTournamentRepository{details: tm}
			if tt.setup != nil {
				tt.setup(tm, repo)
			}
			caller := tt.callerID
			if caller == uuid.Nil {
				caller = ownerID
			}

			_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).UpdateTournament(context.Background(), caller, tournamentID, tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestUpdateTournament_ValidationErrors(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()
	deadline := now.Add(24 * time.Hour)
	startAt := now.Add(48 * time.Hour)
	endAt := now.Add(72 * time.Hour)

	tests := []struct {
		name    string
		input   usecases.UpdateTournamentInput
		wantErr error
	}{
		{name: "empty name", input: usecases.UpdateTournamentInput{Name: ptr("")}, wantErr: usecases.ErrInvalidTournamentName},
		{name: "whitespace name", input: usecases.UpdateTournamentInput{Name: ptr("   ")}, wantErr: usecases.ErrInvalidTournamentName},
		{name: "empty game", input: usecases.UpdateTournamentInput{Game: ptr("")}, wantErr: usecases.ErrInvalidTournamentGame},
		{name: "whitespace game", input: usecases.UpdateTournamentInput{Game: ptr("   ")}, wantErr: usecases.ErrInvalidTournamentGame},
		{name: "empty location", input: usecases.UpdateTournamentInput{Location: ptr("")}, wantErr: usecases.ErrInvalidTournamentLocation},
		{name: "whitespace location", input: usecases.UpdateTournamentInput{Location: ptr("   ")}, wantErr: usecases.ErrInvalidTournamentLocation},
		{name: "negative fee", input: usecases.UpdateTournamentInput{EntryFeeMinor: ptr(int64(-1))}, wantErr: usecases.ErrInvalidTournamentFee},
		{name: "zero capacity", input: usecases.UpdateTournamentInput{Capacity: ptr(0)}, wantErr: usecases.ErrInvalidTournamentCapacity},
		{name: "negative capacity", input: usecases.UpdateTournamentInput{Capacity: ptr(-5)}, wantErr: usecases.ErrInvalidTournamentCapacity},
		{name: "invalid mode", input: usecases.UpdateTournamentInput{RegistrationMode: ptr("SQUAD")}, wantErr: usecases.ErrInvalidRegistrationMode},
		{name: "team min > max", input: usecases.UpdateTournamentInput{MinTeamSize: ptr(6), MaxTeamSize: ptr(4)}, wantErr: usecases.ErrInvalidTeamSize},
		{name: "team min < 1", input: usecases.UpdateTournamentInput{MinTeamSize: ptr(0)}, wantErr: usecases.ErrInvalidTeamSize},
		{name: "start before deadline", input: usecases.UpdateTournamentInput{StartsAt: ptr(deadline.Add(-time.Hour))}, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "start equals deadline", input: usecases.UpdateTournamentInput{StartsAt: ptr(deadline)}, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "start after end", input: usecases.UpdateTournamentInput{StartsAt: ptr(endAt.Add(time.Hour))}, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "start equals end", input: usecases.UpdateTournamentInput{StartsAt: ptr(endAt)}, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "deadline after start", input: usecases.UpdateTournamentInput{RegistrationDeadline: ptr(startAt.Add(time.Hour))}, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "deadline equals start", input: usecases.UpdateTournamentInput{RegistrationDeadline: ptr(startAt)}, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "end before start", input: usecases.UpdateTournamentInput{EndsAt: ptr(startAt.Add(-time.Hour))}, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "end equals start", input: usecases.UpdateTournamentInput{EndsAt: ptr(startAt)}, wantErr: usecases.ErrInvalidTournamentDates},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := baseTournament(ownerID)
			tm.ID = tournamentID
			tm.RegistrationDeadline, tm.StartsAt, tm.EndsAt = deadline, startAt, endAt
			repo := &mockTournamentRepository{details: tm}

			_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).UpdateTournament(context.Background(), ownerID, tournamentID, tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestUpdateTournament_Success(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()

	t.Run("updates and trims metadata fields", func(t *testing.T) {
		tm := baseTournament(ownerID)
		tm.ID = tournamentID
		useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{details: tm}, newStubAccountService())

		updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			Name:        ptr("  Trimmed Cup  "),
			Description: ptr("  Trimmed Rules  "),
			Detail:      ptr("  ## Formatted Rules  "),
			Game:        ptr("  Overwatch 2  "),
			Location:      ptr("  Siam Paragon  "),
			EntryFeeMinor: ptr(int64(0)),
			Capacity:      ptr(32),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Trimmed Cup" || updated.Description != "Trimmed Rules" ||
			updated.Detail != "## Formatted Rules" ||
			updated.Game != "Overwatch 2" || updated.Location != "Siam Paragon" ||
			updated.EntryFeeMinor != 0 || updated.Capacity != 32 {
			t.Fatalf("fields were not properly trimmed or updated: %+v", updated)
		}
	})

	t.Run("partial date updates", func(t *testing.T) {
		tm := baseTournament(ownerID)
		tm.ID = tournamentID
		useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{details: tm}, newStubAccountService())

		newDeadline := tm.RegistrationDeadline.Add(-5 * time.Hour)
		newEnd := tm.EndsAt.Add(24 * time.Hour)

		updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			RegistrationDeadline: &newDeadline,
			EndsAt:               &newEnd,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !updated.RegistrationDeadline.Equal(newDeadline) || !updated.EndsAt.Equal(newEnd) {
			t.Fatalf("expected updated dates, got deadline %v, endsAt %v", updated.RegistrationDeadline, updated.EndsAt)
		}
	})

	t.Run("registration mode transitions", func(t *testing.T) {
		// TEAM -> SOLO resets team sizes to 1
		tmTeam := baseTournament(ownerID)
		tmTeam.ID = tournamentID
		tmTeam.RegistrationMode = models.TournamentRegistrationModeTeam
		tmTeam.MinTeamSize, tmTeam.MaxTeamSize = 5, 7

		useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{details: tmTeam}, newStubAccountService())
		soloUpdated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			RegistrationMode: ptr("SOLO"),
		})
		if err != nil || soloUpdated.MinTeamSize != 1 || soloUpdated.MaxTeamSize != 1 || soloUpdated.RegistrationMode != models.TournamentRegistrationModeSolo {
			t.Fatalf("failed transition to SOLO: %+v, err: %v", soloUpdated, err)
		}

		// SOLO -> TEAM normalizes whitespace/case and applies team sizes
		useCase = usecases.NewTournamentUseCase(&mockTournamentRepository{details: soloUpdated}, newStubAccountService())
		teamUpdated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			RegistrationMode: ptr("  team  "),
			MinTeamSize:      ptr(2),
			MaxTeamSize:      ptr(4),
		})
		if err != nil || teamUpdated.MinTeamSize != 2 || teamUpdated.MaxTeamSize != 4 || teamUpdated.RegistrationMode != models.TournamentRegistrationModeTeam {
			t.Fatalf("failed transition to TEAM: %+v, err: %v", teamUpdated, err)
		}
	})
}

func TestUpdateCrowdfundingConfiguration(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()
	makeCampaign := func() *models.Tournament {
		tournament := baseTournament(ownerID)
		tournament.ID = tournamentID
		tournament.Status = models.TournamentStatusDraft
		deadline := now.Add(8 * time.Hour)
		tournament.Crowdfunding = models.CrowdfundingInfo{GoalAmountMinor: 100000, Deadline: &deadline}
		return tournament
	}

	t.Run("owner changes goal and deadline in draft", func(t *testing.T) {
		repo := &mockTournamentRepository{details: makeCampaign()}
		goal := int64(300000)
		deadline := now.Add(12 * time.Hour)
		updated, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).UpdateTournament(
			context.Background(), ownerID, tournamentID,
			usecases.UpdateTournamentInput{FundingGoalMinor: &goal, FundingDeadline: &deadline},
		)
		if err != nil || updated.Crowdfunding.GoalAmountMinor != goal || !updated.Crowdfunding.Deadline.Equal(deadline) || repo.updatedTournament == nil {
			t.Fatalf("funding update failed: tournament=%+v err=%v", updated, err)
		}
	})

	t.Run("registration deadline ignores the funding deadline once registration is open", func(t *testing.T) {
		tournament := makeCampaign()
		tournament.Status = models.TournamentStatusRegistration
		repo := &mockTournamentRepository{details: tournament}
		deadline := now.Add(4 * time.Hour)
		updated, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).UpdateTournament(
			context.Background(), ownerID, tournamentID,
			usecases.UpdateTournamentInput{RegistrationDeadline: &deadline},
		)
		if err != nil || !updated.RegistrationDeadline.Equal(deadline) || repo.updatedTournament == nil {
			t.Fatalf("registration deadline update failed: tournament=%+v err=%v", updated, err)
		}
	})

	tests := []struct {
		name   string
		setup  func(*models.Tournament)
		caller uuid.UUID
		input  usecases.UpdateTournamentInput
		want   error
	}{
		{"not owner", nil, uuid.New(), usecases.UpdateTournamentInput{FundingGoalMinor: ptr(int64(300000))}, usecases.ErrNotTournamentOwner},
		{"description exceeds max length", nil, ownerID, usecases.UpdateTournamentInput{Description: ptr(strings.Repeat("a", 501))}, usecases.ErrInvalidTournamentDescription},
		{"detail exceeds max length", nil, ownerID, usecases.UpdateTournamentInput{Detail: ptr(strings.Repeat("b", 10001))}, usecases.ErrInvalidTournamentDetail},
		{"zero goal", nil, ownerID, usecases.UpdateTournamentInput{FundingGoalMinor: ptr(int64(0))}, usecases.ErrInvalidCrowdfunding},
		{"deadline after registration", nil, ownerID, usecases.UpdateTournamentInput{FundingDeadline: ptr(now.Add(30 * time.Hour))}, usecases.ErrInvalidCrowdfunding},
		{"registration deadline moved before funding deadline", nil, ownerID, usecases.UpdateTournamentInput{RegistrationDeadline: ptr(now.Add(4 * time.Hour))}, usecases.ErrInvalidCrowdfunding},
		{"registration deadline moved before funding deadline while crowdfunding", func(tm *models.Tournament) { tm.Status = models.TournamentStatusCrowdfunding }, ownerID, usecases.UpdateTournamentInput{RegistrationDeadline: ptr(now.Add(4 * time.Hour))}, usecases.ErrInvalidCrowdfunding},
		{"already crowdfunding", func(tm *models.Tournament) { tm.Status = models.TournamentStatusCrowdfunding }, ownerID, usecases.UpdateTournamentInput{FundingGoalMinor: ptr(int64(300000))}, usecases.ErrCrowdfundingClosed},
		{"registration is open", func(tm *models.Tournament) { tm.Status = models.TournamentStatusRegistration }, ownerID, usecases.UpdateTournamentInput{FundingGoalMinor: ptr(int64(300000))}, usecases.ErrCrowdfundingClosed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tournament := makeCampaign()
			if test.setup != nil {
				test.setup(tournament)
			}
			repo := &mockTournamentRepository{details: tournament}
			_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).UpdateTournament(context.Background(), test.caller, tournamentID, test.input)
			if !errors.Is(err, test.want) || repo.updatedTournament != nil {
				t.Fatalf("expected %v without update, got %v", test.want, err)
			}
		})
	}
}

func TestTournamentUseCase_RepositoryErrors(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	dbErr := errors.New("db failure")

	t.Run("CreateTournament propagates Create error", func(t *testing.T) {
		repo := &mockTournamentRepository{createErr: dbErr}
		_, err := usecases.NewTournamentUseCase(repo, newStubAccountService()).CreateTournament(context.Background(), ownerID, validCreateInput(time.Now()))
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected error %v, got %v", dbErr, err)
		}
	})

	tests := []struct {
		name    string
		repo    *mockTournamentRepository
		wantErr error
	}{
		{
			name:    "GetByID not found mapped",
			repo:    &mockTournamentRepository{detailsErr: repositories.ErrTournamentNotFound},
			wantErr: usecases.ErrTournamentNotFound,
		},
		{
			name:    "GetByID arbitrary db error propagated",
			repo:    &mockTournamentRepository{detailsErr: dbErr},
			wantErr: dbErr,
		},
		{
			name: "Update db error propagated",
			repo: &mockTournamentRepository{
				details:   baseTournament(ownerID),
				updateErr: dbErr,
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.repo.details != nil {
				tt.repo.details.ID = tournamentID
			}
			_, err := usecases.NewTournamentUseCase(tt.repo, newStubAccountService()).UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
				Name: ptr("New Name"),
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
