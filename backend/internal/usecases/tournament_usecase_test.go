package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
	"github.com/moomaideng/eventory/internal/usecases"
)

type mockTournamentRepository struct {
	filters               repositories.TournamentFilters
	items                 []repositories.TournamentSearchItem
	total                 int64
	err                   error
	details               *models.Tournament
	detailsErr            error
	createdTournament     *models.Tournament
	createdFunding        *models.TournamentFunding
	createErr             error
	updatedTournament     *models.Tournament
	updateErr             error
	acceptedCount         int64
	lockedOrAcceptedCount int64
	teamCountErr          error
}

func (m *mockTournamentRepository) GetPublishedByID(_ context.Context, _ uuid.UUID) (*models.Tournament, error) {
	return m.details, m.detailsErr
}

func (m *mockTournamentRepository) GetByID(_ context.Context, _ uuid.UUID) (*models.Tournament, error) {
	return m.details, m.detailsErr
}

func (m *mockTournamentRepository) Create(_ context.Context, tournament *models.Tournament, funding *models.TournamentFunding) error {
	m.createdTournament = tournament
	m.createdFunding = funding
	return m.createErr
}

func (m *mockTournamentRepository) Update(_ context.Context, tournament *models.Tournament) error {
	m.updatedTournament = tournament
	return m.updateErr
}

func (m *mockTournamentRepository) GetActiveTeamCount(_ context.Context, _ uuid.UUID) (int64, int64, error) {
	return m.acceptedCount, m.lockedOrAcceptedCount, m.teamCountErr
}

func (m *mockTournamentRepository) Search(_ context.Context, filters repositories.TournamentFilters) ([]repositories.TournamentSearchItem, int64, error) {
	m.filters = filters
	return m.items, m.total, m.err
}

func ptr[T any](v T) *T { return &v }

func baseTournament(ownerID uuid.UUID) *models.Tournament {
	now := time.Now()
	return &models.Tournament{
		ID:                   uuid.New(),
		OrganizerID:          ownerID,
		Name:                 "Original Name",
		Description:          "Original Description",
		Game:                 "Original Game",
		Location:             "Original Location",
		Status:               models.TournamentStatusRegistrationOpen,
		RegistrationMode:     models.TournamentRegistrationModeTeam,
		MinTeamSize:          3,
		MaxTeamSize:          5,
		Capacity:             16,
		RegistrationDeadline: now.Add(24 * time.Hour),
		StartsAt:             now.Add(48 * time.Hour),
		EndsAt:               now.Add(72 * time.Hour),
		EntryFee:             100,
		Currency:             "THB",
		Published:            true,
	}
}

func validCreateInput(now time.Time) usecases.CreateTournamentInput {
	return usecases.CreateTournamentInput{
		Name:                 "Valid Cup",
		Description:          "Tournament rules",
		Game:                 "Chess",
		Location:             "Online",
		StartsAt:             now.Add(48 * time.Hour),
		EndsAt:               now.Add(72 * time.Hour),
		RegistrationDeadline: now.Add(24 * time.Hour),
		RegistrationMode:     "TEAM",
		MinTeamSize:          3,
		MaxTeamSize:          5,
		Capacity:             16,
		EntryFee:             100,
	}
}

func TestSearchTournaments_NormalizesFilters(t *testing.T) {
	repo := &mockTournamentRepository{
		items: []repositories.TournamentSearchItem{{
			Tournament: models.Tournament{ID: uuid.New()}, RegisteredCount: 3,
		}},
		total: 5,
	}
	useCase := usecases.NewTournamentUseCase(repo)

	result, err := useCase.Search(context.Background(), usecases.SearchTournamentsInput{
		Query: "  valorant  ", StartFrom: "2026-09-01", StartTo: "2026-09-30",
		MaxEntryFee: ptr(int64(500)), Status: "registration_open",
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
	}{
		{name: "reversed dates", input: usecases.SearchTournamentsInput{StartFrom: "2026-10-01", StartTo: "2026-09-01"}},
		{name: "invalid date", input: usecases.SearchTournamentsInput{StartFrom: "09/01/2026"}},
		{name: "negative fee", input: usecases.SearchTournamentsInput{MaxEntryFee: ptr(int64(-1))}},
		{name: "reversed fees", input: usecases.SearchTournamentsInput{MinEntryFee: ptr(int64(501)), MaxEntryFee: ptr(int64(500))}},
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

func TestCreateTournament_Success(t *testing.T) {
	organizer := &models.OrganizerProfile{ID: uuid.New(), OrganizerName: "Test Org"}
	now := time.Now()

	t.Run("solo tournament coerces team size and creates funding stub", func(t *testing.T) {
		repo := &mockTournamentRepository{}
		useCase := usecases.NewTournamentUseCase(repo)
		in := validCreateInput(now)
		in.RegistrationMode = "SOLO"
		in.MinTeamSize, in.MaxTeamSize = 5, 5

		got, err := useCase.CreateTournament(context.Background(), organizer, in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.OrganizerID != organizer.ID || got.Currency != "THB" || !got.Published ||
			got.Status != models.TournamentStatusRegistrationOpen ||
			got.MinTeamSize != 1 || got.MaxTeamSize != 1 ||
			repo.createdFunding == nil || repo.createdFunding.GoalAmount != 1000 {
			t.Fatalf("unexpected tournament state: %+v", got)
		}
	})

	t.Run("team tournament preserves team sizes", func(t *testing.T) {
		repo := &mockTournamentRepository{}
		got, err := usecases.NewTournamentUseCase(repo).CreateTournament(context.Background(), organizer, validCreateInput(now))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.RegistrationMode != models.TournamentRegistrationModeTeam || got.MinTeamSize != 3 || got.MaxTeamSize != 5 {
			t.Fatalf("unexpected team tournament state: %+v", got)
		}
	})
}

func TestCreateTournament_ValidationErrors(t *testing.T) {
	organizer := &models.OrganizerProfile{ID: uuid.New()}
	now := time.Now()

	tests := []struct {
		name    string
		modify  func(*usecases.CreateTournamentInput)
		wantErr error
	}{
		{name: "empty name", modify: func(in *usecases.CreateTournamentInput) { in.Name = "   " }, wantErr: usecases.ErrInvalidTournamentName},
		{name: "empty game", modify: func(in *usecases.CreateTournamentInput) { in.Game = "" }, wantErr: usecases.ErrInvalidTournamentGame},
		{name: "empty location", modify: func(in *usecases.CreateTournamentInput) { in.Location = "  " }, wantErr: usecases.ErrInvalidTournamentLocation},
		{name: "deadline after start", modify: func(in *usecases.CreateTournamentInput) { in.RegistrationDeadline = in.StartsAt.Add(time.Hour) }, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "deadline equals start", modify: func(in *usecases.CreateTournamentInput) { in.RegistrationDeadline = in.StartsAt }, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "start after end", modify: func(in *usecases.CreateTournamentInput) { in.StartsAt = in.EndsAt.Add(time.Hour) }, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "start equals end", modify: func(in *usecases.CreateTournamentInput) { in.StartsAt = in.EndsAt }, wantErr: usecases.ErrInvalidTournamentDates},
		{name: "negative fee", modify: func(in *usecases.CreateTournamentInput) { in.EntryFee = -10 }, wantErr: usecases.ErrInvalidTournamentFee},
		{name: "zero capacity", modify: func(in *usecases.CreateTournamentInput) { in.Capacity = 0 }, wantErr: usecases.ErrInvalidTournamentCapacity},
		{name: "invalid mode", modify: func(in *usecases.CreateTournamentInput) { in.RegistrationMode = "SQUAD" }, wantErr: usecases.ErrInvalidRegistrationMode},
		{name: "team mode min > max", modify: func(in *usecases.CreateTournamentInput) { in.MinTeamSize, in.MaxTeamSize = 5, 3 }, wantErr: usecases.ErrInvalidTeamSize},
		{name: "team mode min < 1", modify: func(in *usecases.CreateTournamentInput) { in.MinTeamSize = 0 }, wantErr: usecases.ErrInvalidTeamSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validCreateInput(now)
			tt.modify(&input)
			_, err := usecases.NewTournamentUseCase(&mockTournamentRepository{}).CreateTournament(context.Background(), organizer, input)
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
			name:     "ongoing status rejected",
			setup:    func(tm *models.Tournament, _ *mockTournamentRepository) { tm.Status = models.TournamentStatusOngoing },
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Name: ptr("New Name")},
			wantErr:  usecases.ErrTournamentCannotBeModified,
		},
		{
			name:     "completed status rejected",
			setup:    func(tm *models.Tournament, _ *mockTournamentRepository) { tm.Status = models.TournamentStatusCompleted },
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Name: ptr("New Name")},
			wantErr:  usecases.ErrTournamentCannotBeModified,
		},
		{
			name: "capacity below accepted count rejected",
			setup: func(_ *models.Tournament, repo *mockTournamentRepository) {
				repo.acceptedCount = 8
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Capacity: ptr(5)},
			wantErr:  usecases.ErrCapacityBelowAcceptedTeams,
		},
		{
			name: "mode change rejected when roster locked",
			setup: func(_ *models.Tournament, repo *mockTournamentRepository) {
				repo.lockedOrAcceptedCount = 2
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{RegistrationMode: ptr("SOLO")},
			wantErr:  usecases.ErrRosterRulesLocked,
		},
		{
			name: "min team size change rejected when roster locked",
			setup: func(_ *models.Tournament, repo *mockTournamentRepository) {
				repo.lockedOrAcceptedCount = 2
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{MinTeamSize: ptr(2)},
			wantErr:  usecases.ErrRosterRulesLocked,
		},
		{
			name: "max team size change rejected when roster locked",
			setup: func(_ *models.Tournament, repo *mockTournamentRepository) {
				repo.lockedOrAcceptedCount = 2
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{MaxTeamSize: ptr(6)},
			wantErr:  usecases.ErrRosterRulesLocked,
		},
		{
			name: "registration closed permitted",
			setup: func(tm *models.Tournament, _ *mockTournamentRepository) {
				tm.Status = models.TournamentStatusRegistrationClosed
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Name: ptr("Closed Cup")},
			wantErr:  nil,
		},
		{
			name: "capacity equal to accepted count permitted",
			setup: func(_ *models.Tournament, repo *mockTournamentRepository) {
				repo.acceptedCount = 8
			},
			callerID: ownerID,
			input:    usecases.UpdateTournamentInput{Capacity: ptr(8)},
			wantErr:  nil,
		},
		{
			name: "idempotent roster update permitted when locked",
			setup: func(tm *models.Tournament, repo *mockTournamentRepository) {
				repo.lockedOrAcceptedCount = 2
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

			_, err := usecases.NewTournamentUseCase(repo).UpdateTournament(context.Background(), caller, tournamentID, tt.input)
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
		{name: "negative fee", input: usecases.UpdateTournamentInput{EntryFee: ptr(int64(-1))}, wantErr: usecases.ErrInvalidTournamentFee},
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

			_, err := usecases.NewTournamentUseCase(repo).UpdateTournament(context.Background(), ownerID, tournamentID, tt.input)
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
		useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{details: tm})

		updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			Name:        ptr("  Trimmed Cup  "),
			Description: ptr("  Trimmed Rules  "),
			Game:        ptr("  Overwatch 2  "),
			Location:    ptr("  Siam Paragon  "),
			EntryFee:    ptr(int64(0)),
			Capacity:    ptr(32),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Trimmed Cup" || updated.Description != "Trimmed Rules" ||
			updated.Game != "Overwatch 2" || updated.Location != "Siam Paragon" ||
			updated.EntryFee != 0 || updated.Capacity != 32 {
			t.Fatalf("fields were not properly trimmed or updated: %+v", updated)
		}
	})

	t.Run("partial date updates", func(t *testing.T) {
		tm := baseTournament(ownerID)
		tm.ID = tournamentID
		useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{details: tm})

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

		useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{details: tmTeam})
		soloUpdated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			RegistrationMode: ptr("SOLO"),
		})
		if err != nil || soloUpdated.MinTeamSize != 1 || soloUpdated.MaxTeamSize != 1 || soloUpdated.RegistrationMode != models.TournamentRegistrationModeSolo {
			t.Fatalf("failed transition to SOLO: %+v, err: %v", soloUpdated, err)
		}

		// SOLO -> TEAM normalizes whitespace/case and applies team sizes
		useCase = usecases.NewTournamentUseCase(&mockTournamentRepository{details: soloUpdated})
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

func TestTournamentUseCase_RepositoryErrors(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	dbErr := errors.New("db failure")

	t.Run("CreateTournament propagates Create error", func(t *testing.T) {
		repo := &mockTournamentRepository{createErr: dbErr}
		_, err := usecases.NewTournamentUseCase(repo).CreateTournament(context.Background(), &models.OrganizerProfile{ID: ownerID}, validCreateInput(time.Now()))
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
			name: "GetActiveTeamCount db error propagated",
			repo: &mockTournamentRepository{
				details:      baseTournament(ownerID),
				teamCountErr: dbErr,
			},
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
			_, err := usecases.NewTournamentUseCase(tt.repo).UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
				Name: ptr("New Name"),
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
