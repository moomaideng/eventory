package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
)

var (
	ErrInvalidTournamentFilters   = errors.New("invalid tournament filters")
	ErrTournamentNotFound         = errors.New("tournament not found")
	ErrNotTournamentOwner         = errors.New("only the organizing owner can modify this tournament")
	ErrInvalidTournamentName      = errors.New("tournament name is required")
	ErrInvalidTournamentGame      = errors.New("game name is required")
	ErrInvalidTournamentLocation  = errors.New("location is required")
	ErrInvalidTournamentDates     = errors.New("registration deadline must be before start date, and start date must be before end date")
	ErrInvalidTournamentFee       = errors.New("entry fee cannot be negative")
	ErrInvalidTournamentCapacity  = errors.New("tournament capacity must be at least 1")
	ErrInvalidRegistrationMode    = errors.New("invalid registration mode: must be SOLO or TEAM")
	ErrInvalidTeamSize            = errors.New("minimum team size must be at least 1 and less than or equal to maximum team size")
	ErrTournamentCannotBeModified = errors.New("ongoing or completed tournaments cannot be modified")
	ErrCapacityBelowAcceptedTeams = errors.New("capacity cannot be less than the number of accepted teams")
	ErrRosterRulesLocked          = errors.New("team size and registration mode cannot be modified after teams have locked or been accepted")
)

type CreateTournamentInput struct {
	Name                 string
	Description          string
	Game                 string
	Location             string
	StartsAt             time.Time
	EndsAt               time.Time
	RegistrationDeadline time.Time
	EntryFee             int64
	RegistrationMode     string
	MinTeamSize          int
	MaxTeamSize          int
	Capacity             int
}

type UpdateTournamentInput struct {
	Name                 *string
	Description          *string
	Game                 *string
	Location             *string
	StartsAt             *time.Time
	EndsAt               *time.Time
	RegistrationDeadline *time.Time
	EntryFee             *int64
	RegistrationMode     *string
	MinTeamSize          *int
	MaxTeamSize          *int
	Capacity             *int
}

var catalogLocation = time.FixedZone("Asia/Bangkok", 7*60*60)

type SearchTournamentsInput struct {
	Query       string
	StartFrom   string
	StartTo     string
	MinEntryFee *int64
	MaxEntryFee *int64
	Status      string
	Sort        string
	Page        int
	PageSize    int
}

type TournamentSearchResult struct {
	Items    []TournamentSearchItem
	Total    int64
	Page     int
	PageSize int
}

type TournamentSearchItem struct {
	Tournament      models.Tournament
	RegisteredCount int
}

type TournamentFundingStats struct {
	GoalAmount      int64
	RaisedAmount    int64
	RemainingAmount int64
	SupporterCount  int
	Percentage      float64
}

type TournamentDetailsResult struct {
	Tournament      models.Tournament
	RegisteredCount int
	Funding         TournamentFundingStats
}

type TournamentUseCase struct {
	tournamentRepo repositories.TournamentRepository
}

func NewTournamentUseCase(tournamentRepo repositories.TournamentRepository) *TournamentUseCase {
	return &TournamentUseCase{tournamentRepo: tournamentRepo}
}

// validate query
func (u *TournamentUseCase) Search(
	ctx context.Context,
	input SearchTournamentsInput,
) (*TournamentSearchResult, error) {
	query := strings.TrimSpace(input.Query)
	if len(query) > 100 {
		return nil, ErrInvalidTournamentFilters
	}

	startFrom, err := parseCatalogDate(input.StartFrom, false)
	if err != nil {
		return nil, ErrInvalidTournamentFilters
	}
	startTo, err := parseCatalogDate(input.StartTo, true)
	if err != nil {
		return nil, ErrInvalidTournamentFilters
	}
	if startFrom != nil && startTo != nil && startFrom.After(*startTo) {
		return nil, ErrInvalidTournamentFilters
	}
	if input.MinEntryFee != nil && *input.MinEntryFee < 0 ||
		input.MaxEntryFee != nil && *input.MaxEntryFee < 0 ||
		input.MinEntryFee != nil && input.MaxEntryFee != nil && *input.MinEntryFee > *input.MaxEntryFee {
		return nil, ErrInvalidTournamentFilters
	}

	status := models.TournamentStatus(strings.ToUpper(strings.TrimSpace(input.Status)))
	validStatuses := map[models.TournamentStatus]bool{
		"":                                      true,
		models.TournamentStatusRegistrationOpen: true,
		models.TournamentStatusRegistrationClosed: true,
		models.TournamentStatusOngoing:            true,
		models.TournamentStatusCompleted:          true,
	}
	if !validStatuses[status] {
		return nil, ErrInvalidTournamentFilters
	}

	sort := strings.ToLower(strings.TrimSpace(input.Sort))
	if sort == "" {
		sort = "start_asc"
	}
	validSorts := map[string]bool{
		"start_asc": true, "start_desc": true, "fee_asc": true, "fee_desc": true,
	}
	if !validSorts[sort] {
		return nil, ErrInvalidTournamentFilters
	}

	page := input.Page
	if page == 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 12
	}
	if page < 1 || pageSize < 1 || pageSize > 100 {
		return nil, ErrInvalidTournamentFilters
	}

	repositoryItems, total, err := u.tournamentRepo.Search(ctx, repositories.TournamentFilters{
		Query: query, StartFrom: startFrom, StartTo: startTo,
		MinEntryFee: input.MinEntryFee, MaxEntryFee: input.MaxEntryFee,
		Status: status, Sort: sort, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]TournamentSearchItem, len(repositoryItems))
	for index, item := range repositoryItems {
		items[index] = TournamentSearchItem{
			Tournament: item.Tournament, RegisteredCount: int(item.RegisteredCount),
		}
	}

	return &TournamentSearchResult{
		Items: items, Total: total, Page: page, PageSize: pageSize,
	}, nil
}

func (u *TournamentUseCase) GetDetails(
	ctx context.Context,
	id uuid.UUID,
) (*TournamentDetailsResult, error) {
	tournament, err := u.tournamentRepo.GetPublishedByID(ctx, id)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}
	acceptedTeams := make([]models.TournamentTeam, 0, len(tournament.Teams))
	for _, team := range tournament.Teams {
		if team.Status == models.TournamentTeamStatusAccepted {
			acceptedTeams = append(acceptedTeams, team)
		}
	}
	tournament.Teams = acceptedTeams

	stats := TournamentFundingStats{}
	if tournament.Funding != nil {
		stats.GoalAmount = tournament.Funding.GoalAmount
		stats.RaisedAmount = tournament.Funding.RaisedAmount
		stats.SupporterCount = tournament.Funding.SupporterCount
	}
	// handle negative & remain case
	stats.RemainingAmount = max(stats.GoalAmount-stats.RaisedAmount, 0)
	if stats.GoalAmount > 0 {
		stats.Percentage = float64(stats.RaisedAmount) / float64(stats.GoalAmount) * 100
	}

	return &TournamentDetailsResult{
		Tournament: *tournament, RegisteredCount: len(acceptedTeams), Funding: stats,
	}, nil
}

// CreateTournament initializes and persists a new tournament record for an organizer.
// It applies default status REGISTRATION_OPEN, published=true, hardcoded currency "THB",
// automatically coerces solo tournaments to min/max team size of 1, and creates an initial
// funding stub. It strictly validates that registrationDeadline < startsAt < endsAt.
func (u *TournamentUseCase) CreateTournament(
	ctx context.Context,
	organizer *models.OrganizerProfile,
	input CreateTournamentInput,
) (*models.Tournament, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrInvalidTournamentName
	}
	game := strings.TrimSpace(input.Game)
	if game == "" {
		return nil, ErrInvalidTournamentGame
	}
	location := strings.TrimSpace(input.Location)
	if location == "" {
		return nil, ErrInvalidTournamentLocation
	}
	description := strings.TrimSpace(input.Description)

	if !input.RegistrationDeadline.Before(input.StartsAt) || !input.StartsAt.Before(input.EndsAt) {
		return nil, ErrInvalidTournamentDates
	}
	if input.EntryFee < 0 {
		return nil, ErrInvalidTournamentFee
	}
	if input.Capacity < 1 {
		return nil, ErrInvalidTournamentCapacity
	}

	mode := models.TournamentRegistrationMode(strings.ToUpper(strings.TrimSpace(input.RegistrationMode)))
	var minTeamSize, maxTeamSize int
	switch mode {
	case models.TournamentRegistrationModeSolo:
		minTeamSize = 1
		maxTeamSize = 1
	case models.TournamentRegistrationModeTeam:
		if input.MinTeamSize < 1 || input.MaxTeamSize < input.MinTeamSize {
			return nil, ErrInvalidTeamSize
		}
		minTeamSize = input.MinTeamSize
		maxTeamSize = input.MaxTeamSize
	default:
		return nil, ErrInvalidRegistrationMode
	}

	// Always default to THB for now.
	// In the future, currency selection may be expanded based on team consensus.
	currency := "THB"

	// [TODO]: Initialize default funding stub for now; funding features and schemas
	// are currently being finalized by the team.
	funding := &models.TournamentFunding{
		GoalAmount:     1000,
		RaisedAmount:   500,
		SupporterCount: 1,
	}

	tournament := &models.Tournament{
		OrganizerID:          organizer.ID,
		Organizer:            *organizer,
		Name:                 name,
		Description:          description,
		Game:                 game,
		Location:             location,
		StartsAt:             input.StartsAt,
		EndsAt:               input.EndsAt,
		RegistrationDeadline: input.RegistrationDeadline,
		EntryFee:             input.EntryFee,
		Currency:             currency,
		RegistrationMode:     mode,
		MinTeamSize:          minTeamSize,
		MaxTeamSize:          maxTeamSize,
		Capacity:             input.Capacity,
		Status:               models.TournamentStatusRegistrationOpen,
		Published:            true,
	}

	if err := u.tournamentRepo.Create(ctx, tournament, funding); err != nil {
		return nil, err
	}

	return tournament, nil
}

// UpdateTournament modifies mutable tournament attributes subject to strict safety invariants.
func (u *TournamentUseCase) UpdateTournament(
	ctx context.Context,
	organizerID uuid.UUID,
	tournamentID uuid.UUID,
	input UpdateTournamentInput,
) (*models.Tournament, error) {
	tournament, err := u.tournamentRepo.GetByID(ctx, tournamentID)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	if tournament.OrganizerID != organizerID {
		return nil, ErrNotTournamentOwner
	}

	if tournament.Status == models.TournamentStatusOngoing || tournament.Status == models.TournamentStatusCompleted {
		return nil, ErrTournamentCannotBeModified
	}

	acceptedCount, lockedOrAcceptedCount, err := u.tournamentRepo.GetActiveTeamCount(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	if input.Capacity != nil {
		if *input.Capacity < 1 {
			return nil, ErrInvalidTournamentCapacity
		}
		if int64(*input.Capacity) < acceptedCount {
			return nil, ErrCapacityBelowAcceptedTeams
		}
		tournament.Capacity = *input.Capacity
	}

	if lockedOrAcceptedCount > 0 {
		if input.RegistrationMode != nil {
			normalized := models.TournamentRegistrationMode(strings.ToUpper(strings.TrimSpace(*input.RegistrationMode)))
			if normalized != tournament.RegistrationMode {
				return nil, ErrRosterRulesLocked
			}
		}
		if input.MinTeamSize != nil && *input.MinTeamSize != tournament.MinTeamSize {
			return nil, ErrRosterRulesLocked
		}
		if input.MaxTeamSize != nil && *input.MaxTeamSize != tournament.MaxTeamSize {
			return nil, ErrRosterRulesLocked
		}
	} else {
		newMode := tournament.RegistrationMode
		if input.RegistrationMode != nil {
			normalized := models.TournamentRegistrationMode(strings.ToUpper(strings.TrimSpace(*input.RegistrationMode)))
			if normalized != models.TournamentRegistrationModeSolo && normalized != models.TournamentRegistrationModeTeam {
				return nil, ErrInvalidRegistrationMode
			}
			newMode = normalized
		}

		newMin := tournament.MinTeamSize
		if input.MinTeamSize != nil {
			newMin = *input.MinTeamSize
		}

		newMax := tournament.MaxTeamSize
		if input.MaxTeamSize != nil {
			newMax = *input.MaxTeamSize
		}

		if newMode == models.TournamentRegistrationModeSolo {
			newMin = 1
			newMax = 1
		} else {
			if newMin < 1 || newMax < newMin {
				return nil, ErrInvalidTeamSize
			}
		}

		tournament.RegistrationMode = newMode
		tournament.MinTeamSize = newMin
		tournament.MaxTeamSize = newMax
	}

	newDeadline := tournament.RegistrationDeadline
	if input.RegistrationDeadline != nil {
		newDeadline = *input.RegistrationDeadline
	}
	newStartsAt := tournament.StartsAt
	if input.StartsAt != nil {
		newStartsAt = *input.StartsAt
	}
	newEndsAt := tournament.EndsAt
	if input.EndsAt != nil {
		newEndsAt = *input.EndsAt
	}

	if !newDeadline.Before(newStartsAt) || !newStartsAt.Before(newEndsAt) {
		return nil, ErrInvalidTournamentDates
	}
	tournament.RegistrationDeadline = newDeadline
	tournament.StartsAt = newStartsAt
	tournament.EndsAt = newEndsAt

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, ErrInvalidTournamentName
		}
		tournament.Name = trimmed
	}
	if input.Description != nil {
		tournament.Description = strings.TrimSpace(*input.Description)
	}
	if input.Game != nil {
		trimmed := strings.TrimSpace(*input.Game)
		if trimmed == "" {
			return nil, ErrInvalidTournamentGame
		}
		tournament.Game = trimmed
	}
	if input.Location != nil {
		trimmed := strings.TrimSpace(*input.Location)
		if trimmed == "" {
			return nil, ErrInvalidTournamentLocation
		}
		tournament.Location = trimmed
	}
	if input.EntryFee != nil {
		if *input.EntryFee < 0 {
			return nil, ErrInvalidTournamentFee
		}
		tournament.EntryFee = *input.EntryFee
	}

	if err := u.tournamentRepo.Update(ctx, tournament); err != nil {
		return nil, err
	}

	return tournament, nil
}

func parseCatalogDate(value string, endOfDay bool) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, catalogLocation)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return &parsed, nil
}
