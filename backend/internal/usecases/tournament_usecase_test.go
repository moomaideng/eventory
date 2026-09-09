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

func (m *mockTournamentRepository) GetPublishedByID(
	_ context.Context,
	_ uuid.UUID,
) (*models.Tournament, error) {
	return m.details, m.detailsErr
}

func (m *mockTournamentRepository) GetByID(
	_ context.Context,
	_ uuid.UUID,
) (*models.Tournament, error) {
	return m.details, m.detailsErr
}

func (m *mockTournamentRepository) Create(
	_ context.Context,
	tournament *models.Tournament,
	funding *models.TournamentFunding,
) error {
	m.createdTournament = tournament
	m.createdFunding = funding
	return m.createErr
}

func (m *mockTournamentRepository) Update(
	_ context.Context,
	tournament *models.Tournament,
) error {
	m.updatedTournament = tournament
	return m.updateErr
}

func (m *mockTournamentRepository) GetActiveTeamCount(
	_ context.Context,
	_ uuid.UUID,
) (int64, int64, error) {
	return m.acceptedCount, m.lockedOrAcceptedCount, m.teamCountErr
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

func stringPointer(value string) *string {
	return &value
}

func intPointer(value int) *int {
	return &value
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func TestCreateTournament_Success_Solo(t *testing.T) {
	repo := &mockTournamentRepository{}
	useCase := usecases.NewTournamentUseCase(repo)

	organizer := &models.OrganizerProfile{
		ID:            uuid.New(),
		OrganizerName: "Test Organizer",
	}

	now := time.Now()
	deadline := now.Add(24 * time.Hour)
	startAt := now.Add(48 * time.Hour)
	endAt := now.Add(72 * time.Hour)

	input := usecases.CreateTournamentInput{
		Name:                 "Solo Showdown",
		Description:          "A 1v1 duel",
		Game:                 "Tekken 8",
		Location:             "Online",
		StartsAt:             startAt,
		EndsAt:               endAt,
		RegistrationDeadline: deadline,
		EntryFee:             200,
		RegistrationMode:     "SOLO",
		MinTeamSize:          5, // Should be coerced to 1
		MaxTeamSize:          5, // Should be coerced to 1
		Capacity:             32,
	}

	tournament, err := useCase.CreateTournament(context.Background(), organizer, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tournament.OrganizerID != organizer.ID {
		t.Fatalf("expected organizer id %v, got %v", organizer.ID, tournament.OrganizerID)
	}
	if tournament.Currency != "THB" {
		t.Fatalf("expected currency THB, got %s", tournament.Currency)
	}
	if tournament.Status != models.TournamentStatusRegistrationOpen {
		t.Fatalf("expected status REGISTRATION_OPEN, got %s", tournament.Status)
	}
	if !tournament.Published {
		t.Fatal("expected published to be true")
	}
	if tournament.MinTeamSize != 1 || tournament.MaxTeamSize != 1 {
		t.Fatalf("expected team sizes coerced to 1 for SOLO, got min=%d max=%d", tournament.MinTeamSize, tournament.MaxTeamSize)
	}
	if repo.createdFunding == nil {
		t.Fatal("expected default funding stub to be created")
	}
}

func TestCreateTournament_Success_Team(t *testing.T) {
	repo := &mockTournamentRepository{}
	useCase := usecases.NewTournamentUseCase(repo)

	organizer := &models.OrganizerProfile{
		ID:            uuid.New(),
		OrganizerName: "Team Organizer",
	}

	now := time.Now()
	input := usecases.CreateTournamentInput{
		Name:                 "5v5 Cup",
		Description:          "Team battle",
		Game:                 "Valorant",
		Location:             "Bangkok",
		StartsAt:             now.Add(48 * time.Hour),
		EndsAt:               now.Add(72 * time.Hour),
		RegistrationDeadline: now.Add(24 * time.Hour),
		EntryFee:             500,
		RegistrationMode:     "TEAM",
		MinTeamSize:          5,
		MaxTeamSize:          7,
		Capacity:             16,
	}

	tournament, err := useCase.CreateTournament(context.Background(), organizer, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tournament.MinTeamSize != 5 || tournament.MaxTeamSize != 7 {
		t.Fatalf("expected team sizes 5-7, got min=%d max=%d", tournament.MinTeamSize, tournament.MaxTeamSize)
	}
	if tournament.RegistrationMode != models.TournamentRegistrationModeTeam {
		t.Fatalf("expected mode TEAM, got %s", tournament.RegistrationMode)
	}
}

func TestCreateTournament_ValidationErrors(t *testing.T) {
	organizer := &models.OrganizerProfile{ID: uuid.New()}
	now := time.Now()

	validInput := usecases.CreateTournamentInput{
		Name:                 "Valid Tournament",
		Game:                 "Chess",
		Location:             "Online",
		StartsAt:             now.Add(48 * time.Hour),
		EndsAt:               now.Add(72 * time.Hour),
		RegistrationDeadline: now.Add(24 * time.Hour),
		RegistrationMode:     "SOLO",
		Capacity:             10,
	}

	tests := []struct {
		name        string
		modify      func(in *usecases.CreateTournamentInput)
		expectedErr error
	}{
		{
			name:        "empty name",
			modify:      func(in *usecases.CreateTournamentInput) { in.Name = "   " },
			expectedErr: usecases.ErrInvalidTournamentName,
		},
		{
			name:        "empty game",
			modify:      func(in *usecases.CreateTournamentInput) { in.Game = "" },
			expectedErr: usecases.ErrInvalidTournamentGame,
		},
		{
			name:        "empty location",
			modify:      func(in *usecases.CreateTournamentInput) { in.Location = "  " },
			expectedErr: usecases.ErrInvalidTournamentLocation,
		},
		{
			name:        "deadline after start",
			modify:      func(in *usecases.CreateTournamentInput) { in.RegistrationDeadline = in.StartsAt.Add(time.Hour) },
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name:        "start after end",
			modify:      func(in *usecases.CreateTournamentInput) { in.StartsAt = in.EndsAt.Add(time.Hour) },
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name:        "negative fee",
			modify:      func(in *usecases.CreateTournamentInput) { in.EntryFee = -10 },
			expectedErr: usecases.ErrInvalidTournamentFee,
		},
		{
			name:        "zero capacity",
			modify:      func(in *usecases.CreateTournamentInput) { in.Capacity = 0 },
			expectedErr: usecases.ErrInvalidTournamentCapacity,
		},
		{
			name:        "invalid registration mode",
			modify:      func(in *usecases.CreateTournamentInput) { in.RegistrationMode = "SQUAD" },
			expectedErr: usecases.ErrInvalidRegistrationMode,
		},
		{
			name: "team mode with min > max",
			modify: func(in *usecases.CreateTournamentInput) {
				in.RegistrationMode = "TEAM"
				in.MinTeamSize = 5
				in.MaxTeamSize = 3
			},
			expectedErr: usecases.ErrInvalidTeamSize,
		},
		{
			name: "team mode with min < 1",
			modify: func(in *usecases.CreateTournamentInput) {
				in.RegistrationMode = "TEAM"
				in.MinTeamSize = 0
				in.MaxTeamSize = 3
			},
			expectedErr: usecases.ErrInvalidTeamSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validInput
			tt.modify(&input)
			useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{})
			_, err := useCase.CreateTournament(context.Background(), organizer, input)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestUpdateTournament_OwnershipGuard(t *testing.T) {
	ownerID := uuid.New()
	differentOrganizerID := uuid.New()
	tournamentID := uuid.New()

	repo := &mockTournamentRepository{
		details: &models.Tournament{
			ID:          tournamentID,
			OrganizerID: ownerID,
			Status:      models.TournamentStatusRegistrationOpen,
		},
	}
	useCase := usecases.NewTournamentUseCase(repo)

	_, err := useCase.UpdateTournament(context.Background(), differentOrganizerID, tournamentID, usecases.UpdateTournamentInput{
		Name: stringPointer("New Title"),
	})
	if !errors.Is(err, usecases.ErrNotTournamentOwner) {
		t.Fatalf("expected ErrNotTournamentOwner, got %v", err)
	}
}

func TestUpdateTournament_StatusCheck(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()

	repo := &mockTournamentRepository{
		details: &models.Tournament{
			ID:          tournamentID,
			OrganizerID: ownerID,
			Status:      models.TournamentStatusOngoing,
		},
	}
	useCase := usecases.NewTournamentUseCase(repo)

	_, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
		Name: stringPointer("New Title"),
	})
	if !errors.Is(err, usecases.ErrTournamentCannotBeModified) {
		t.Fatalf("expected ErrTournamentCannotBeModified for ONGOING tournament, got %v", err)
	}
}

func TestUpdateTournament_CapacityGuard(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()

	repo := &mockTournamentRepository{
		details: &models.Tournament{
			ID:                   tournamentID,
			OrganizerID:          ownerID,
			Status:               models.TournamentStatusRegistrationOpen,
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          5,
			MaxTeamSize:          5,
			Capacity:             16,
			RegistrationDeadline: time.Now().Add(24 * time.Hour),
			StartsAt:             time.Now().Add(48 * time.Hour),
			EndsAt:               time.Now().Add(72 * time.Hour),
		},
		acceptedCount:         8,
		lockedOrAcceptedCount: 8,
	}
	useCase := usecases.NewTournamentUseCase(repo)

	// Trying to reduce capacity below 8 accepted teams
	_, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
		Capacity: intPointer(5),
	})
	if !errors.Is(err, usecases.ErrCapacityBelowAcceptedTeams) {
		t.Fatalf("expected ErrCapacityBelowAcceptedTeams, got %v", err)
	}

	// Capacity >= 8 should succeed
	updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
		Capacity: intPointer(10),
	})
	if err != nil {
		t.Fatalf("expected successful capacity update, got %v", err)
	}
	if updated.Capacity != 10 {
		t.Fatalf("expected capacity 10, got %d", updated.Capacity)
	}
}

func TestUpdateTournament_RosterRulesLocked(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()

	repo := &mockTournamentRepository{
		details: &models.Tournament{
			ID:                   tournamentID,
			OrganizerID:          ownerID,
			Status:               models.TournamentStatusRegistrationOpen,
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          5,
			MaxTeamSize:          5,
			Capacity:             16,
			RegistrationDeadline: time.Now().Add(24 * time.Hour),
			StartsAt:             time.Now().Add(48 * time.Hour),
			EndsAt:               time.Now().Add(72 * time.Hour),
		},
		lockedOrAcceptedCount: 2, // Teams have locked or been accepted
	}
	useCase := usecases.NewTournamentUseCase(repo)

	// Attempting to change mode
	_, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
		RegistrationMode: stringPointer("SOLO"),
	})
	if !errors.Is(err, usecases.ErrRosterRulesLocked) {
		t.Fatalf("expected ErrRosterRulesLocked when changing mode, got %v", err)
	}

	// Attempting to change minTeamSize
	_, err = useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
		MinTeamSize: intPointer(3),
	})
	if !errors.Is(err, usecases.ErrRosterRulesLocked) {
		t.Fatalf("expected ErrRosterRulesLocked when changing minTeamSize, got %v", err)
	}

	// Attempting to change maxTeamSize
	_, err = useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
		MaxTeamSize: intPointer(7),
	})
	if !errors.Is(err, usecases.ErrRosterRulesLocked) {
		t.Fatalf("expected ErrRosterRulesLocked when changing maxTeamSize, got %v", err)
	}
}

func TestUpdateTournament_Success_EditableRoster(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()

	repo := &mockTournamentRepository{
		details: &models.Tournament{
			ID:                   tournamentID,
			OrganizerID:          ownerID,
			Name:                 "Initial Name",
			Game:                 "Valorant",
			Location:             "Online",
			Status:               models.TournamentStatusRegistrationOpen,
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          5,
			MaxTeamSize:          5,
			Capacity:             16,
			RegistrationDeadline: now.Add(24 * time.Hour),
			StartsAt:             now.Add(48 * time.Hour),
			EndsAt:               now.Add(72 * time.Hour),
		},
		lockedOrAcceptedCount: 0, // No locked/accepted teams yet
	}
	useCase := usecases.NewTournamentUseCase(repo)

	newDeadline := now.Add(30 * time.Hour)
	newStart := now.Add(50 * time.Hour)
	newEnd := now.Add(80 * time.Hour)

	updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
		Name:                 stringPointer("Updated Tournament Name"),
		MinTeamSize:          intPointer(3),
		MaxTeamSize:          intPointer(6),
		RegistrationDeadline: &newDeadline,
		StartsAt:             &newStart,
		EndsAt:               &newEnd,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Name != "Updated Tournament Name" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}
	if updated.MinTeamSize != 3 || updated.MaxTeamSize != 6 {
		t.Fatalf("expected team sizes 3-6, got %d-%d", updated.MinTeamSize, updated.MaxTeamSize)
	}
	if !updated.RegistrationDeadline.Equal(newDeadline) || !updated.StartsAt.Equal(newStart) {
		t.Fatal("expected dates to be updated")
	}
}

func TestCreateTournament_DateBoundaries(t *testing.T) {
	organizer := &models.OrganizerProfile{ID: uuid.New()}
	now := time.Now()

	tests := []struct {
		name                 string
		registrationDeadline time.Time
		startsAt             time.Time
		endsAt               time.Time
	}{
		{
			name:                 "deadline equals startsAt",
			registrationDeadline: now.Add(24 * time.Hour),
			startsAt:             now.Add(24 * time.Hour),
			endsAt:               now.Add(48 * time.Hour),
		},
		{
			name:                 "startsAt equals endsAt",
			registrationDeadline: now.Add(24 * time.Hour),
			startsAt:             now.Add(48 * time.Hour),
			endsAt:               now.Add(48 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := usecases.NewTournamentUseCase(&mockTournamentRepository{})
			_, err := useCase.CreateTournament(context.Background(), organizer, usecases.CreateTournamentInput{
				Name:                 "Boundary Cup",
				Game:                 "Chess",
				Location:             "Online",
				StartsAt:             tt.startsAt,
				EndsAt:               tt.endsAt,
				RegistrationDeadline: tt.registrationDeadline,
				RegistrationMode:     "SOLO",
				Capacity:             8,
			})
			if !errors.Is(err, usecases.ErrInvalidTournamentDates) {
				t.Fatalf("expected ErrInvalidTournamentDates, got %v", err)
			}
		})
	}
}

func TestUpdateTournament_ValidationErrors(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()

	baseTournament := func() *models.Tournament {
		return &models.Tournament{
			ID:                   tournamentID,
			OrganizerID:          ownerID,
			Name:                 "Original Name",
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
		}
	}

	tests := []struct {
		name        string
		input       usecases.UpdateTournamentInput
		expectedErr error
	}{
		{
			name:        "empty name string",
			input:       usecases.UpdateTournamentInput{Name: stringPointer("")},
			expectedErr: usecases.ErrInvalidTournamentName,
		},
		{
			name:        "whitespace name string",
			input:       usecases.UpdateTournamentInput{Name: stringPointer("   ")},
			expectedErr: usecases.ErrInvalidTournamentName,
		},
		{
			name:        "empty game string",
			input:       usecases.UpdateTournamentInput{Game: stringPointer("")},
			expectedErr: usecases.ErrInvalidTournamentGame,
		},
		{
			name:        "whitespace game string",
			input:       usecases.UpdateTournamentInput{Game: stringPointer("   ")},
			expectedErr: usecases.ErrInvalidTournamentGame,
		},
		{
			name:        "empty location string",
			input:       usecases.UpdateTournamentInput{Location: stringPointer("")},
			expectedErr: usecases.ErrInvalidTournamentLocation,
		},
		{
			name:        "whitespace location string",
			input:       usecases.UpdateTournamentInput{Location: stringPointer("   ")},
			expectedErr: usecases.ErrInvalidTournamentLocation,
		},
		{
			name:        "negative entry fee",
			input:       usecases.UpdateTournamentInput{EntryFee: int64Pointer(-1)},
			expectedErr: usecases.ErrInvalidTournamentFee,
		},
		{
			name:        "zero capacity",
			input:       usecases.UpdateTournamentInput{Capacity: intPointer(0)},
			expectedErr: usecases.ErrInvalidTournamentCapacity,
		},
		{
			name:        "negative capacity",
			input:       usecases.UpdateTournamentInput{Capacity: intPointer(-5)},
			expectedErr: usecases.ErrInvalidTournamentCapacity,
		},
		{
			name:        "invalid registration mode",
			input:       usecases.UpdateTournamentInput{RegistrationMode: stringPointer("SQUAD")},
			expectedErr: usecases.ErrInvalidRegistrationMode,
		},
		{
			name: "team mode with min > max",
			input: usecases.UpdateTournamentInput{
				MinTeamSize: intPointer(6),
				MaxTeamSize: intPointer(4),
			},
			expectedErr: usecases.ErrInvalidTeamSize,
		},
		{
			name: "team mode with min < 1",
			input: usecases.UpdateTournamentInput{
				MinTeamSize: intPointer(0),
			},
			expectedErr: usecases.ErrInvalidTeamSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTournamentRepository{
				details:               baseTournament(),
				lockedOrAcceptedCount: 0,
			}
			useCase := usecases.NewTournamentUseCase(repo)

			_, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, tt.input)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestUpdateTournament_PartialDateUpdates(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()

	initDeadline := now.Add(24 * time.Hour)
	initStartsAt := now.Add(48 * time.Hour)
	initEndsAt := now.Add(72 * time.Hour)

	baseTournament := func() *models.Tournament {
		return &models.Tournament{
			ID:                   tournamentID,
			OrganizerID:          ownerID,
			Name:                 "Date Test Cup",
			Game:                 "Chess",
			Location:             "Online",
			Status:               models.TournamentStatusRegistrationOpen,
			RegistrationMode:     models.TournamentRegistrationModeSolo,
			MinTeamSize:          1,
			MaxTeamSize:          1,
			Capacity:             16,
			RegistrationDeadline: initDeadline,
			StartsAt:             initStartsAt,
			EndsAt:               initEndsAt,
		}
	}

	tests := []struct {
		name        string
		input       usecases.UpdateTournamentInput
		expectErr   bool
		expectedErr error
		verify      func(t *testing.T, updated *models.Tournament)
	}{
		{
			name: "only StartsAt moved before existing RegistrationDeadline",
			input: usecases.UpdateTournamentInput{
				StartsAt: timePointer(initDeadline.Add(-1 * time.Hour)),
			},
			expectErr:   true,
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name: "only StartsAt moved equal to existing RegistrationDeadline",
			input: usecases.UpdateTournamentInput{
				StartsAt: timePointer(initDeadline),
			},
			expectErr:   true,
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name: "only StartsAt moved after existing EndsAt",
			input: usecases.UpdateTournamentInput{
				StartsAt: timePointer(initEndsAt.Add(1 * time.Hour)),
			},
			expectErr:   true,
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name: "only StartsAt moved equal to existing EndsAt",
			input: usecases.UpdateTournamentInput{
				StartsAt: timePointer(initEndsAt),
			},
			expectErr:   true,
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name: "only RegistrationDeadline moved after existing StartsAt",
			input: usecases.UpdateTournamentInput{
				RegistrationDeadline: timePointer(initStartsAt.Add(1 * time.Hour)),
			},
			expectErr:   true,
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name: "only RegistrationDeadline moved equal to existing StartsAt",
			input: usecases.UpdateTournamentInput{
				RegistrationDeadline: timePointer(initStartsAt),
			},
			expectErr:   true,
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name: "only EndsAt moved before existing StartsAt",
			input: usecases.UpdateTournamentInput{
				EndsAt: timePointer(initStartsAt.Add(-1 * time.Hour)),
			},
			expectErr:   true,
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name: "only EndsAt moved equal to existing StartsAt",
			input: usecases.UpdateTournamentInput{
				EndsAt: timePointer(initStartsAt),
			},
			expectErr:   true,
			expectedErr: usecases.ErrInvalidTournamentDates,
		},
		{
			name: "valid partial update: only RegistrationDeadline moved earlier",
			input: usecases.UpdateTournamentInput{
				RegistrationDeadline: timePointer(initDeadline.Add(-5 * time.Hour)),
			},
			expectErr: false,
			verify: func(t *testing.T, updated *models.Tournament) {
				expected := initDeadline.Add(-5 * time.Hour)
				if !updated.RegistrationDeadline.Equal(expected) {
					t.Fatalf("expected deadline %v, got %v", expected, updated.RegistrationDeadline)
				}
				if !updated.StartsAt.Equal(initStartsAt) || !updated.EndsAt.Equal(initEndsAt) {
					t.Fatal("existing start and end times should remain unchanged")
				}
			},
		},
		{
			name: "valid partial update: only EndsAt extended",
			input: usecases.UpdateTournamentInput{
				EndsAt: timePointer(initEndsAt.Add(24 * time.Hour)),
			},
			expectErr: false,
			verify: func(t *testing.T, updated *models.Tournament) {
				expected := initEndsAt.Add(24 * time.Hour)
				if !updated.EndsAt.Equal(expected) {
					t.Fatalf("expected endsAt %v, got %v", expected, updated.EndsAt)
				}
				if !updated.RegistrationDeadline.Equal(initDeadline) || !updated.StartsAt.Equal(initStartsAt) {
					t.Fatal("existing deadline and start times should remain unchanged")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTournamentRepository{
				details:               baseTournament(),
				lockedOrAcceptedCount: 0,
			}
			useCase := usecases.NewTournamentUseCase(repo)

			updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, tt.input)
			if tt.expectErr {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if tt.verify != nil {
					tt.verify(t, updated)
				}
			}
		})
	}
}

func TestUpdateTournament_RegistrationModeTransitions(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()

	t.Run("transition TEAM to SOLO resets team sizes to 1", func(t *testing.T) {
		repo := &mockTournamentRepository{
			details: &models.Tournament{
				ID:                   tournamentID,
				OrganizerID:          ownerID,
				Status:               models.TournamentStatusRegistrationOpen,
				RegistrationMode:     models.TournamentRegistrationModeTeam,
				MinTeamSize:          5,
				MaxTeamSize:          7,
				Capacity:             16,
				RegistrationDeadline: now.Add(24 * time.Hour),
				StartsAt:             now.Add(48 * time.Hour),
				EndsAt:               now.Add(72 * time.Hour),
			},
			lockedOrAcceptedCount: 0,
		}
		useCase := usecases.NewTournamentUseCase(repo)

		updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			RegistrationMode: stringPointer("SOLO"),
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.RegistrationMode != models.TournamentRegistrationModeSolo {
			t.Fatalf("expected mode SOLO, got %s", updated.RegistrationMode)
		}
		if updated.MinTeamSize != 1 || updated.MaxTeamSize != 1 {
			t.Fatalf("expected team sizes reset to 1, got min=%d max=%d", updated.MinTeamSize, updated.MaxTeamSize)
		}
	})

	t.Run("transition SOLO to TEAM with new team sizes", func(t *testing.T) {
		repo := &mockTournamentRepository{
			details: &models.Tournament{
				ID:                   tournamentID,
				OrganizerID:          ownerID,
				Status:               models.TournamentStatusRegistrationOpen,
				RegistrationMode:     models.TournamentRegistrationModeSolo,
				MinTeamSize:          1,
				MaxTeamSize:          1,
				Capacity:             16,
				RegistrationDeadline: now.Add(24 * time.Hour),
				StartsAt:             now.Add(48 * time.Hour),
				EndsAt:               now.Add(72 * time.Hour),
			},
			lockedOrAcceptedCount: 0,
		}
		useCase := usecases.NewTournamentUseCase(repo)

		updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			RegistrationMode: stringPointer("TEAM"),
			MinTeamSize:      intPointer(3),
			MaxTeamSize:      intPointer(5),
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.RegistrationMode != models.TournamentRegistrationModeTeam {
			t.Fatalf("expected mode TEAM, got %s", updated.RegistrationMode)
		}
		if updated.MinTeamSize != 3 || updated.MaxTeamSize != 5 {
			t.Fatalf("expected team sizes 3-5, got min=%d max=%d", updated.MinTeamSize, updated.MaxTeamSize)
		}
	})

	t.Run("normalizes registration mode casing and whitespace", func(t *testing.T) {
		repo := &mockTournamentRepository{
			details: &models.Tournament{
				ID:                   tournamentID,
				OrganizerID:          ownerID,
				Status:               models.TournamentStatusRegistrationOpen,
				RegistrationMode:     models.TournamentRegistrationModeSolo,
				MinTeamSize:          1,
				MaxTeamSize:          1,
				Capacity:             16,
				RegistrationDeadline: now.Add(24 * time.Hour),
				StartsAt:             now.Add(48 * time.Hour),
				EndsAt:               now.Add(72 * time.Hour),
			},
			lockedOrAcceptedCount: 0,
		}
		useCase := usecases.NewTournamentUseCase(repo)

		updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			RegistrationMode: stringPointer("  team  "),
			MinTeamSize:      intPointer(2),
			MaxTeamSize:      intPointer(4),
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.RegistrationMode != models.TournamentRegistrationModeTeam {
			t.Fatalf("expected mode TEAM, got %s", updated.RegistrationMode)
		}
	})

	t.Run("idempotent roster update allowed even when teams locked or accepted", func(t *testing.T) {
		repo := &mockTournamentRepository{
			details: &models.Tournament{
				ID:                   tournamentID,
				OrganizerID:          ownerID,
				Name:                 "Current Name",
				Status:               models.TournamentStatusRegistrationOpen,
				RegistrationMode:     models.TournamentRegistrationModeTeam,
				MinTeamSize:          5,
				MaxTeamSize:          5,
				Capacity:             16,
				RegistrationDeadline: now.Add(24 * time.Hour),
				StartsAt:             now.Add(48 * time.Hour),
				EndsAt:               now.Add(72 * time.Hour),
			},
			lockedOrAcceptedCount: 3,
		}
		useCase := usecases.NewTournamentUseCase(repo)

		updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			Name:             stringPointer("Updated Name"),
			RegistrationMode: stringPointer("TEAM"),
			MinTeamSize:      intPointer(5),
			MaxTeamSize:      intPointer(5),
		})
		if err != nil {
			t.Fatalf("expected no error when submitting identical roster values, got %v", err)
		}
		if updated.Name != "Updated Name" {
			t.Fatalf("expected updated name, got %s", updated.Name)
		}
	})
}

func TestUpdateTournament_CapacityBoundary(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()

	repo := &mockTournamentRepository{
		details: &models.Tournament{
			ID:                   tournamentID,
			OrganizerID:          ownerID,
			Status:               models.TournamentStatusRegistrationOpen,
			RegistrationMode:     models.TournamentRegistrationModeSolo,
			MinTeamSize:          1,
			MaxTeamSize:          1,
			Capacity:             16,
			RegistrationDeadline: now.Add(24 * time.Hour),
			StartsAt:             now.Add(48 * time.Hour),
			EndsAt:               now.Add(72 * time.Hour),
		},
		acceptedCount:         8,
		lockedOrAcceptedCount: 8,
	}
	useCase := usecases.NewTournamentUseCase(repo)

	// Setting capacity to exactly acceptedCount (8) should be permitted
	updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
		Capacity: intPointer(8),
	})
	if err != nil {
		t.Fatalf("expected capacity equal to accepted count to succeed, got %v", err)
	}
	if updated.Capacity != 8 {
		t.Fatalf("expected capacity 8, got %d", updated.Capacity)
	}
}

func TestUpdateTournament_StatusCheck_CompletedAndClosed(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()

	t.Run("rejects modification for COMPLETED tournament", func(t *testing.T) {
		repo := &mockTournamentRepository{
			details: &models.Tournament{
				ID:          tournamentID,
				OrganizerID: ownerID,
				Status:      models.TournamentStatusCompleted,
			},
		}
		useCase := usecases.NewTournamentUseCase(repo)

		_, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			Name: stringPointer("New Title"),
		})
		if !errors.Is(err, usecases.ErrTournamentCannotBeModified) {
			t.Fatalf("expected ErrTournamentCannotBeModified for COMPLETED tournament, got %v", err)
		}
	})

	t.Run("permits modification for REGISTRATION_CLOSED tournament", func(t *testing.T) {
		repo := &mockTournamentRepository{
			details: &models.Tournament{
				ID:                   tournamentID,
				OrganizerID:          ownerID,
				Name:                 "Closed Tournament",
				Game:                 "Valorant",
				Location:             "Online",
				Status:               models.TournamentStatusRegistrationClosed,
				RegistrationMode:     models.TournamentRegistrationModeSolo,
				MinTeamSize:          1,
				MaxTeamSize:          1,
				Capacity:             16,
				RegistrationDeadline: now.Add(24 * time.Hour),
				StartsAt:             now.Add(48 * time.Hour),
				EndsAt:               now.Add(72 * time.Hour),
			},
			acceptedCount:         4,
			lockedOrAcceptedCount: 4,
		}
		useCase := usecases.NewTournamentUseCase(repo)

		updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			Name: stringPointer("Updated Closed Tournament Name"),
		})
		if err != nil {
			t.Fatalf("expected registration closed tournament to be modifiable, got %v", err)
		}
		if updated.Name != "Updated Closed Tournament Name" {
			t.Fatalf("expected name to be updated, got %s", updated.Name)
		}
	})
}

func TestUpdateTournament_TrimsStringInputs(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()

	repo := &mockTournamentRepository{
		details: &models.Tournament{
			ID:                   tournamentID,
			OrganizerID:          ownerID,
			Name:                 "Original Name",
			Description:          "Original Description",
			Game:                 "Original Game",
			Location:             "Original Location",
			Status:               models.TournamentStatusRegistrationOpen,
			RegistrationMode:     models.TournamentRegistrationModeSolo,
			MinTeamSize:          1,
			MaxTeamSize:          1,
			Capacity:             16,
			RegistrationDeadline: now.Add(24 * time.Hour),
			StartsAt:             now.Add(48 * time.Hour),
			EndsAt:               now.Add(72 * time.Hour),
			EntryFee:             100,
		},
		lockedOrAcceptedCount: 0,
	}
	useCase := usecases.NewTournamentUseCase(repo)

	updated, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
		Name:        stringPointer("  Trimmed Cup  "),
		Description: stringPointer("  Trimmed Rules  "),
		Game:        stringPointer("  Overwatch 2  "),
		Location:    stringPointer("  Siam Paragon  "),
		EntryFee:    int64Pointer(0),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Name != "Trimmed Cup" {
		t.Fatalf("expected trimmed name, got %q", updated.Name)
	}
	if updated.Description != "Trimmed Rules" {
		t.Fatalf("expected trimmed description, got %q", updated.Description)
	}
	if updated.Game != "Overwatch 2" {
		t.Fatalf("expected trimmed game, got %q", updated.Game)
	}
	if updated.Location != "Siam Paragon" {
		t.Fatalf("expected trimmed location, got %q", updated.Location)
	}
	if updated.EntryFee != 0 {
		t.Fatalf("expected entry fee 0, got %d", updated.EntryFee)
	}
}

func TestTournamentUseCase_RepositoryErrorPropagation(t *testing.T) {
	ownerID := uuid.New()
	tournamentID := uuid.New()
	now := time.Now()
	dbErr := errors.New("database connection failed")

	t.Run("CreateTournament propagates repository Create error", func(t *testing.T) {
		repo := &mockTournamentRepository{createErr: dbErr}
		useCase := usecases.NewTournamentUseCase(repo)

		_, err := useCase.CreateTournament(context.Background(), &models.OrganizerProfile{ID: ownerID}, usecases.CreateTournamentInput{
			Name:                 "Valid Tournament",
			Game:                 "Chess",
			Location:             "Online",
			StartsAt:             now.Add(48 * time.Hour),
			EndsAt:               now.Add(72 * time.Hour),
			RegistrationDeadline: now.Add(24 * time.Hour),
			RegistrationMode:     "SOLO",
			Capacity:             10,
		})
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected error %v, got %v", dbErr, err)
		}
	})

	t.Run("UpdateTournament maps ErrTournamentNotFound from GetByID", func(t *testing.T) {
		repo := &mockTournamentRepository{detailsErr: repositories.ErrTournamentNotFound}
		useCase := usecases.NewTournamentUseCase(repo)

		_, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			Name: stringPointer("New Name"),
		})
		if !errors.Is(err, usecases.ErrTournamentNotFound) {
			t.Fatalf("expected ErrTournamentNotFound, got %v", err)
		}
	})

	t.Run("UpdateTournament propagates arbitrary DB error from GetByID", func(t *testing.T) {
		repo := &mockTournamentRepository{detailsErr: dbErr}
		useCase := usecases.NewTournamentUseCase(repo)

		_, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			Name: stringPointer("New Name"),
		})
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected error %v, got %v", dbErr, err)
		}
	})

	t.Run("UpdateTournament propagates DB error from GetActiveTeamCount", func(t *testing.T) {
		repo := &mockTournamentRepository{
			details: &models.Tournament{
				ID:          tournamentID,
				OrganizerID: ownerID,
				Status:      models.TournamentStatusRegistrationOpen,
			},
			teamCountErr: dbErr,
		}
		useCase := usecases.NewTournamentUseCase(repo)

		_, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			Name: stringPointer("New Name"),
		})
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected error %v, got %v", dbErr, err)
		}
	})

	t.Run("UpdateTournament propagates DB error from Update", func(t *testing.T) {
		repo := &mockTournamentRepository{
			details: &models.Tournament{
				ID:                   tournamentID,
				OrganizerID:          ownerID,
				Status:               models.TournamentStatusRegistrationOpen,
				RegistrationMode:     models.TournamentRegistrationModeSolo,
				MinTeamSize:          1,
				MaxTeamSize:          1,
				Capacity:             16,
				RegistrationDeadline: now.Add(24 * time.Hour),
				StartsAt:             now.Add(48 * time.Hour),
				EndsAt:               now.Add(72 * time.Hour),
			},
			updateErr: dbErr,
		}
		useCase := usecases.NewTournamentUseCase(repo)

		_, err := useCase.UpdateTournament(context.Background(), ownerID, tournamentID, usecases.UpdateTournamentInput{
			Name: stringPointer("New Name"),
		})
		if !errors.Is(err, dbErr) {
			t.Fatalf("expected error %v, got %v", dbErr, err)
		}
	})
}

