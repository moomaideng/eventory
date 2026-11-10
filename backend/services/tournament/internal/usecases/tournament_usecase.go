package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
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
	ErrInvalidCrowdfunding        = errors.New("crowdfunding requires a positive funding goal and a deadline after now but before the registration deadline")
	ErrCrowdfundingClosed         = errors.New("funding can only be changed while crowdfunding is active")
	ErrTournamentCannotBeModified = errors.New("ongoing or completed tournaments cannot be modified")
	ErrCapacityBelowAcceptedTeams = errors.New("capacity cannot be less than the number of accepted teams")
	ErrRosterRulesLocked          = errors.New("team size and registration mode cannot be modified after teams have locked or been accepted")
	ErrInvalidDashboardPage       = errors.New("invalid dashboard pagination")
	ErrInvalidStatusTransition    = errors.New("this status change is not allowed")
	ErrStatusUnchanged            = errors.New("tournament is already in that status")
	ErrUnknownStatus              = errors.New("unknown tournament status")
	ErrCrowdfundingGoalNotReached = errors.New("funding goal has not been reached")
	ErrRegistrationGateIncomplete = errors.New("event details or registration rules are incomplete")
	ErrCrowdfundingPlanIncomplete = ErrRegistrationGateIncomplete
	ErrRegistrationFormRequired   = errors.New("registration form must be configured before opening registration")
	ErrCrowdfundingSetupIncomplete = errors.New("crowdfunding requires a positive goal, a future deadline, and at least one active sponsorship tier")
	ErrUnresolvedApplicationsExist = repositories.ErrUnresolvedApplicationsExist
	ErrInsufficientAcceptedTeams  = repositories.ErrInsufficientAcceptedTeams
	ErrBracketMissing             = repositories.ErrBracketMissing
	ErrBracketStage1Mismatch      = repositories.ErrBracketStage1Mismatch
	ErrBracketNotSeeded           = repositories.ErrBracketNotSeeded
	ErrUnfinishedMatchesExist     = repositories.ErrUnfinishedMatchesExist
	ErrNoChampionshipWinner       = repositories.ErrNoChampionshipWinner
	ErrTournamentStateChanged     = errors.New("tournament status changed; refresh and try again")
	ErrInvalidTournamentDescription = errors.New("tournament description exceeds maximum length of 500 characters")
	ErrInvalidTournamentDetail      = errors.New("tournament detail exceeds maximum length of 10000 characters")
)

const (
	tournamentDescriptionMaxLength = 500
	tournamentDetailMaxLength      = 10000
)

type CreateTournamentInput struct {
	Name                 string
	Description          string
	Detail               string
	Game                 string
	Location             string
	StartsAt             *time.Time
	EndsAt               *time.Time
	RegistrationDeadline *time.Time
	EntryFee             *int64
	RegistrationMode     *string
	MinTeamSize          *int
	MaxTeamSize          *int
	Capacity             *int
	FundingGoal          *int64
	FundingDeadline      *time.Time
}

type UpdateTournamentInput struct {
	Name                 *string
	Description          *string
	Detail               *string
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
	FundingGoal          *int64
	FundingDeadline      *time.Time
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

// TournamentView is the shared read model: persisted tournament + organizer display name.
type TournamentView struct {
	Tournament    models.Tournament
	OrganizerName string
}

func NewTournamentView(tournament models.Tournament, organizerName string) TournamentView {
	return TournamentView{
		Tournament:    tournament,
		OrganizerName: organizerName,
	}
}

func NewTournamentViews(tournaments []models.Tournament, names map[uuid.UUID]string) []TournamentView {
	views := make([]TournamentView, len(tournaments))
	for i, tournament := range tournaments {
		views[i] = NewTournamentView(tournament, names[tournament.ID])
	}
	return views
}

type TournamentFundingStats struct {
	GoalAmount      int64
	Deadline        *time.Time
	RaisedAmount    int64
	RemainingAmount int64
	SupporterCount  int
	Percentage      float64
}

func NewFundingStats(funding *models.TournamentFunding) TournamentFundingStats {
	stats := TournamentFundingStats{}
	if funding == nil {
		return stats
	}
	stats.GoalAmount = funding.GoalAmount
	stats.Deadline = funding.Deadline
	stats.RaisedAmount = funding.RaisedAmount
	stats.SupporterCount = funding.SupporterCount
	stats.RemainingAmount = max(stats.GoalAmount-stats.RaisedAmount, 0)
	if stats.GoalAmount > 0 {
		stats.Percentage = float64(stats.RaisedAmount) / float64(stats.GoalAmount) * 100
	}
	return stats
}

type TournamentSearchItem struct {
	TournamentView
	RegisteredCount int
}

type TournamentDetailsResult struct {
	TournamentView
	RegisteredCount int
	Funding         TournamentFundingStats
}

type OrganizerDashboardMetrics struct {
	TotalEntries          int
	FormingEntries        int
	LockedEntries         int
	AcceptedEntries       int
	RejectedEntries       int
	ConfirmedParticipants int
	AvailableSpots        int
}

// OrganizerTournamentSummary is the list-row read model. Tournament is the persisted
// row with Teams and Funding cleared; those values live on Metrics and Funding.
type OrganizerTournamentSummary struct {
	Tournament    models.Tournament
	OrganizerName string
	Metrics       OrganizerDashboardMetrics
	Funding       TournamentFundingStats
}

type OrganizerDashboardEntry struct {
	ID        uuid.UUID
	Name      string
	Status    string
	CreatedAt time.Time
	Members   []TeamMemberView
}

func NewOrganizerDashboardEntry(team models.TournamentTeam, members []TeamMemberView) OrganizerDashboardEntry {
	return OrganizerDashboardEntry{
		ID: team.ID, Name: team.Name, Status: string(team.Status), CreatedAt: team.CreatedAt, Members: members,
	}
}

// OrganizerDashboard is one owned tournament plus its enriched registrations.
type OrganizerDashboard struct {
	OrganizerTournamentSummary
	Entries []OrganizerDashboardEntry
}

type OrganizerDashboardList struct {
	Items    []OrganizerTournamentSummary
	Total    int64
	Page     int
	PageSize int
}

type TournamentStatusOverride struct {
	TournamentID uuid.UUID
	FromStatus   models.TournamentStatus
	ToStatus     models.TournamentStatus
	Reason       string
}

type TournamentUseCase struct {
	tournamentRepo repositories.TournamentRepository
	accountSvc     ports.AccountService
	formRepo       repositories.RegistrationFormRepository
}

func NewTournamentUseCase(tournamentRepo repositories.TournamentRepository, accountSvc ports.AccountService) *TournamentUseCase {
	return &TournamentUseCase{tournamentRepo: tournamentRepo, accountSvc: accountSvc}
}

func (u *TournamentUseCase) SetFormRepository(repo repositories.RegistrationFormRepository) {
	u.formRepo = repo
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
		"":                                  true,
		models.TournamentStatusDraft:        true,
		models.TournamentStatusCrowdfunding: true,
		models.TournamentStatusRegistration: true,
		models.TournamentStatusOngoing:      true,
		models.TournamentStatusCompleted:    true,
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
	tournaments := make([]models.Tournament, len(repositoryItems))
	for index, item := range repositoryItems {
		tournaments[index] = item.Tournament
	}
	names, err := organizerNames(ctx, u.accountSvc, tournaments...)
	if err != nil {
		return nil, err
	}
	views := NewTournamentViews(tournaments, names)

	items := make([]TournamentSearchItem, len(repositoryItems))
	for index, item := range repositoryItems {
		items[index] = TournamentSearchItem{
			TournamentView:  views[index],
			RegisteredCount: int(item.RegisteredCount),
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
	registeredCount := len(acceptedTeams)

	names, err := organizerNames(ctx, u.accountSvc, *tournament)
	if err != nil {
		return nil, err
	}

	return &TournamentDetailsResult{
		TournamentView:  NewTournamentView(*tournament, names[tournament.ID]),
		RegisteredCount: registeredCount,
		Funding:         NewFundingStats(tournament.Funding),
	}, nil
}

// CreateTournament initializes and persists a new tournament record for an organizer in DRAFT status.
func (u *TournamentUseCase) CreateTournament(
	ctx context.Context,
	accountID uuid.UUID,
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
	if utf8.RuneCountInString(description) > tournamentDescriptionMaxLength {
		return nil, ErrInvalidTournamentDescription
	}
	detail := strings.TrimSpace(input.Detail)
	if utf8.RuneCountInString(detail) > tournamentDetailMaxLength {
		return nil, ErrInvalidTournamentDetail
	}

	var startsAt, endsAt, regDeadline time.Time
	if input.StartsAt != nil {
		startsAt = *input.StartsAt
	}
	if input.EndsAt != nil {
		endsAt = *input.EndsAt
	}
	if input.RegistrationDeadline != nil {
		regDeadline = *input.RegistrationDeadline
	}
	if err := validateChronologicalDates(regDeadline, startsAt, endsAt); err != nil {
		return nil, err
	}

	var entryFee int64
	if input.EntryFee != nil {
		if *input.EntryFee < 0 {
			return nil, ErrInvalidTournamentFee
		}
		entryFee = *input.EntryFee
	}

	var capacity int
	if input.Capacity != nil {
		if *input.Capacity < 1 {
			return nil, ErrInvalidTournamentCapacity
		}
		capacity = *input.Capacity
	}

	mode := models.TournamentRegistrationModeSolo
	minTeamSize := 1
	maxTeamSize := 1
	if input.RegistrationMode != nil && strings.TrimSpace(*input.RegistrationMode) != "" {
		mode = models.TournamentRegistrationMode(strings.ToUpper(strings.TrimSpace(*input.RegistrationMode)))
		switch mode {
		case models.TournamentRegistrationModeSolo:
			minTeamSize = 1
			maxTeamSize = 1
		case models.TournamentRegistrationModeTeam:
			min := 1
			if input.MinTeamSize != nil {
				min = *input.MinTeamSize
			}
			max := min
			if input.MaxTeamSize != nil {
				max = *input.MaxTeamSize
			}
			if min < 1 || max < min {
				return nil, ErrInvalidTeamSize
			}
			minTeamSize = min
			maxTeamSize = max
		default:
			return nil, ErrInvalidRegistrationMode
		}
	}

	// Always default to THB for now.
	currency := "THB"

	var funding *models.TournamentFunding
	if input.FundingGoal != nil || input.FundingDeadline != nil {
		if input.FundingGoal == nil || input.FundingDeadline == nil ||
			!validFunding(*input.FundingGoal, *input.FundingDeadline, regDeadline) {
			return nil, ErrInvalidCrowdfunding
		}
		funding = &models.TournamentFunding{GoalAmount: *input.FundingGoal, Deadline: input.FundingDeadline}
	}

	tournament := &models.Tournament{
		Staff: []models.TournamentStaff{
			{
				AccountID: accountID,
				Role:      models.TournamentStaffRoleOwner,
			},
		},
		Name:                 name,
		Description:          description,
		Detail:               detail,
		Game:                 game,
		Location:             location,
		StartsAt:             startsAt,
		EndsAt:               endsAt,
		RegistrationDeadline: regDeadline,
		EntryFee:             entryFee,
		Currency:             currency,
		RegistrationMode:     mode,
		MinTeamSize:          minTeamSize,
		MaxTeamSize:          maxTeamSize,
		Capacity:             capacity,
		Status:               models.TournamentStatusDraft,
	}

	if err := u.tournamentRepo.Create(ctx, tournament, funding); err != nil {
		return nil, err
	}
	return tournament, nil
}

func validateChronologicalDates(regDeadline, startsAt, endsAt time.Time) error {
	if !regDeadline.IsZero() && !startsAt.IsZero() && !regDeadline.Before(startsAt) {
		return ErrInvalidTournamentDates
	}
	if !startsAt.IsZero() && !endsAt.IsZero() && !startsAt.Before(endsAt) {
		return ErrInvalidTournamentDates
	}
	if !regDeadline.IsZero() && !endsAt.IsZero() && startsAt.IsZero() && !regDeadline.Before(endsAt) {
		return ErrInvalidTournamentDates
	}
	return nil
}

func validFunding(goal int64, deadline, registrationDeadline time.Time) bool {
	if goal <= 0 || !deadline.After(time.Now()) {
		return false
	}
	if !registrationDeadline.IsZero() && !deadline.Before(registrationDeadline) {
		return false
	}
	return true
}

// UpdateTournament modifies mutable tournament attributes subject to strict safety invariants.
func (u *TournamentUseCase) UpdateTournament(
	ctx context.Context,
	accountID uuid.UUID,
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

	staff, ok := tournament.StaffMember(accountID)
	if !ok || !staff.IsOwner() {
		return nil, ErrNotTournamentOwner
	}

	if tournament.Status == models.TournamentStatusCompleted {
		return nil, ErrTournamentCannotBeModified
	}
	if tournament.Status == models.TournamentStatusOngoing {
		if input.StartsAt != nil || input.EndsAt != nil || input.RegistrationDeadline != nil ||
			input.EntryFee != nil || input.RegistrationMode != nil || input.MinTeamSize != nil ||
			input.MaxTeamSize != nil || input.Capacity != nil || input.FundingGoal != nil ||
			input.FundingDeadline != nil {
			return nil, ErrTournamentCannotBeModified
		}
	}

	// In REGISTRATION status, the registration gate cannot be modified.
	if tournament.Status == models.TournamentStatusRegistration {
		if input.Capacity != nil && *input.Capacity != tournament.Capacity {
			return nil, ErrRosterRulesLocked
		}
		if input.EntryFee != nil && *input.EntryFee != tournament.EntryFee {
			return nil, ErrRosterRulesLocked
		}
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
	}

	// Gate editing allowed in DRAFT and CROWDFUNDING
	if tournament.Status != models.TournamentStatusRegistration && tournament.Status != models.TournamentStatusOngoing {
		if input.Capacity != nil {
			if *input.Capacity < 1 {
				return nil, ErrInvalidTournamentCapacity
			}
			tournament.Capacity = *input.Capacity
		}

		if input.EntryFee != nil {
			if *input.EntryFee < 0 {
				return nil, ErrInvalidTournamentFee
			}
			tournament.EntryFee = *input.EntryFee
		}

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

	if err := validateChronologicalDates(newDeadline, newStartsAt, newEndsAt); err != nil {
		return nil, err
	}
	tournament.RegistrationDeadline = newDeadline
	tournament.StartsAt = newStartsAt
	tournament.EndsAt = newEndsAt

	// Crowdfunding info is only editable in DRAFT status
	if input.FundingGoal != nil || input.FundingDeadline != nil {
		if tournament.Status != models.TournamentStatusDraft {
			return nil, ErrCrowdfundingClosed
		}
		var currentGoal int64
		var currentDeadline time.Time
		if tournament.Funding != nil {
			currentGoal = tournament.Funding.GoalAmount
			if tournament.Funding.Deadline != nil {
				currentDeadline = *tournament.Funding.Deadline
			}
		}
		if input.FundingGoal != nil {
			currentGoal = *input.FundingGoal
		}
		if input.FundingDeadline != nil {
			currentDeadline = *input.FundingDeadline
		}
		if !validFunding(currentGoal, currentDeadline, newDeadline) {
			return nil, ErrInvalidCrowdfunding
		}
		if tournament.Funding == nil {
			tournament.Funding = &models.TournamentFunding{}
		}
		tournament.Funding.GoalAmount = currentGoal
		tournament.Funding.Deadline = &currentDeadline
	} else if tournament.Funding != nil && tournament.Funding.Deadline != nil && !newDeadline.IsZero() &&
		!tournament.Funding.Deadline.Before(newDeadline) {
		return nil, ErrInvalidCrowdfunding
	}

	if input.Name != nil {
		trimmed := strings.TrimSpace(*input.Name)
		if trimmed == "" {
			return nil, ErrInvalidTournamentName
		}
		tournament.Name = trimmed
	}
	if input.Description != nil {
		desc := strings.TrimSpace(*input.Description)
		if utf8.RuneCountInString(desc) > tournamentDescriptionMaxLength {
			return nil, ErrInvalidTournamentDescription
		}
		tournament.Description = desc
	}
	if input.Detail != nil {
		det := strings.TrimSpace(*input.Detail)
		if utf8.RuneCountInString(det) > tournamentDetailMaxLength {
			return nil, ErrInvalidTournamentDetail
		}
		tournament.Detail = det
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

	if err := u.tournamentRepo.Update(ctx, tournament); err != nil {
		if errors.Is(err, repositories.ErrTournamentStateChanged) {
			return nil, ErrTournamentStateChanged
		}
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

func (u *TournamentUseCase) List(ctx context.Context, organizerID uuid.UUID, page, pageSize int) (*OrganizerDashboardList, error) {
	if page < 1 || pageSize < 1 || pageSize > 100 || page > 1000000 {
		return nil, ErrInvalidDashboardPage
	}
	tournaments, total, err := u.tournamentRepo.ListOwned(ctx, organizerID, page, pageSize)
	if err != nil {
		return nil, err
	}
	names, err := organizerNames(ctx, u.accountSvc, tournaments...)
	if err != nil {
		return nil, err
	}
	items := make([]OrganizerTournamentSummary, 0, len(tournaments))
	for _, tournament := range tournaments {
		items = append(items, summarizeDashboard(tournament, names[tournament.ID]))
	}
	return &OrganizerDashboardList{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (u *TournamentUseCase) Get(ctx context.Context, organizerID, tournamentID uuid.UUID) (*OrganizerDashboard, error) {
	tournament, err := u.tournamentRepo.GetOwned(ctx, organizerID, tournamentID)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}
	names, err := organizerNames(ctx, u.accountSvc, *tournament)
	if err != nil {
		return nil, err
	}
	summary := summarizeDashboard(*tournament, names[tournament.ID])

	var allMembers []models.TournamentTeamMember
	for _, team := range tournament.Teams {
		allMembers = append(allMembers, team.Members...)
	}
	accounts, err := accountInfos(ctx, u.accountSvc, allMembers...)
	if err != nil {
		return nil, err
	}
	entries := make([]OrganizerDashboardEntry, 0, len(allMembers))
	for _, team := range tournament.Teams {
		if team.Status == models.TournamentTeamStatusForming || team.Status == models.TournamentTeamStatusPaymentPending {
			continue
		}
		entries = append(entries, NewOrganizerDashboardEntry(team, NewTeamMemberViews(team.Members, accounts)))
	}
	return &OrganizerDashboard{OrganizerTournamentSummary: summary, Entries: entries}, nil
}

func summarizeDashboard(tournament models.Tournament, organizerName string) OrganizerTournamentSummary {
	metrics := OrganizerDashboardMetrics{TotalEntries: len(tournament.Teams)}
	participants := make(map[uuid.UUID]struct{})
	for _, team := range tournament.Teams {
		switch team.Status {
		case models.TournamentTeamStatusForming, models.TournamentTeamStatusPaymentPending:
			metrics.FormingEntries++
		case models.TournamentTeamStatusLocked:
			metrics.LockedEntries++
		case models.TournamentTeamStatusAccepted:
			metrics.AcceptedEntries++
			for _, member := range team.Members {
				participants[member.AccountID] = struct{}{}
			}
		case models.TournamentTeamStatusRejected:
			metrics.RejectedEntries++
		}
	}
	metrics.ConfirmedParticipants = len(participants)
	metrics.AvailableSpots = max(tournament.Capacity-metrics.AcceptedEntries, 0)
	tournament.Teams = nil
	funding := tournament.Funding
	tournament.Funding = nil
	return OrganizerTournamentSummary{
		Tournament:    tournament,
		OrganizerName: organizerName,
		Metrics:       metrics,
		Funding:       NewFundingStats(funding),
	}
}

// allowedTransitions maps each status to the statuses an organizer may manually
// transition it to.
var allowedTransitions = map[models.TournamentStatus][]models.TournamentStatus{
	models.TournamentStatusDraft: {
		models.TournamentStatusCrowdfunding,
		models.TournamentStatusRegistration,
	},
	models.TournamentStatusCrowdfunding: {
		models.TournamentStatusRegistration,
	},
	models.TournamentStatusRegistration: {
		models.TournamentStatusOngoing,
	},
	models.TournamentStatusOngoing: {
		models.TournamentStatusCompleted,
	},
	models.TournamentStatusCompleted: {},
}

func (u *TournamentUseCase) Override(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
	rawStatus, rawReason string,
) (*TournamentStatusOverride, error) {
	return u.OverrideWithForce(ctx, accountID, tournamentID, rawStatus, rawReason, false)
}

func (u *TournamentUseCase) OverrideWithForce(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
	rawStatus, rawReason string,
	force bool,
) (*TournamentStatusOverride, error) {
	to := models.TournamentStatus(strings.ToUpper(strings.TrimSpace(rawStatus)))
	if _, known := allowedTransitions[to]; !known {
		return nil, ErrUnknownStatus
	}

	current, err := u.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		if errors.Is(err, repositories.ErrTournamentNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}

	staff, ok := current.StaffMember(accountID)
	if !ok || !staff.IsOwner() {
		return nil, ErrNotTournamentOwner
	}

	from := current.Status
	if current.Status == to {
		return nil, ErrStatusUnchanged
	}
	if !canTransition(current.Status, to) {
		return nil, ErrInvalidStatusTransition
	}
	if force && (current.Status != models.TournamentStatusCrowdfunding || to != models.TournamentStatusRegistration) {
		return nil, ErrInvalidStatusTransition
	}
	if current.Status == models.TournamentStatusDraft && to == models.TournamentStatusCrowdfunding {
		if current.Funding == nil || current.Funding.GoalAmount <= 0 || current.Funding.Deadline == nil || !current.Funding.Deadline.After(time.Now()) {
			return nil, ErrCrowdfundingSetupIncomplete
		}
		tierCount, err := u.tournamentRepo.CountActiveTiers(ctx, tournamentID)
		if err != nil {
			return nil, err
		}
		if tierCount == 0 {
			return nil, ErrCrowdfundingSetupIncomplete
		}
		if !current.RegistrationDeadline.IsZero() && !current.Funding.Deadline.Before(current.RegistrationDeadline) {
			return nil, ErrCrowdfundingSetupIncomplete
		}
	}
	if current.Status == models.TournamentStatusDraft && to == models.TournamentStatusRegistration {
		if problem := registrationPlanProblem(current, time.Now()); problem != "" {
			return nil, fmt.Errorf("%w: %s", ErrRegistrationGateIncomplete, problem)
		}
		if u.formRepo != nil {
			_, err := u.formRepo.GetFormByTournamentID(ctx, tournamentID)
			if errors.Is(err, repositories.ErrFormNotFound) {
				return nil, ErrRegistrationFormRequired
			}
			if err != nil {
				return nil, err
			}
		}
	}
	if current.Status == models.TournamentStatusCrowdfunding && to == models.TournamentStatusRegistration {
		if problem := registrationPlanProblem(current, time.Now()); problem != "" {
			return nil, fmt.Errorf("%w: %s", ErrRegistrationGateIncomplete, problem)
		}
		if current.Funding != nil && current.Funding.Deadline != nil && !current.Funding.Deadline.Before(current.RegistrationDeadline) {
			return nil, fmt.Errorf("%w: %s", ErrRegistrationGateIncomplete, "set the funding deadline before registration closes")
		}
		if !force && current.Funding != nil && current.Funding.RaisedAmount < current.Funding.GoalAmount {
			return nil, ErrCrowdfundingGoalNotReached
		}
		if u.formRepo != nil {
			_, err := u.formRepo.GetFormByTournamentID(ctx, tournamentID)
			if errors.Is(err, repositories.ErrFormNotFound) {
				return nil, ErrRegistrationFormRequired
			}
			if err != nil {
				return nil, err
			}
		}
	}

	if current.Status == models.TournamentStatusRegistration && to == models.TournamentStatusOngoing {
		if !current.IsRegistrationClosed(time.Now()) {
			return nil, ErrRegistrationStillOpen
		}
		if err := u.tournamentRepo.ValidateBracketReadiness(ctx, tournamentID); err != nil {
			return nil, err
		}
	}

	if current.Status == models.TournamentStatusOngoing && to == models.TournamentStatusCompleted {
		if err := u.tournamentRepo.ValidateCompletionReadiness(ctx, tournamentID); err != nil {
			return nil, err
		}
	}

	if err := u.tournamentRepo.UpdateStatus(ctx, tournamentID, to); err != nil {
		if errors.Is(err, repositories.ErrTournamentNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}

	return &TournamentStatusOverride{
		TournamentID: tournamentID,
		FromStatus:   from,
		ToStatus:     to,
		Reason:       rawReason,
	}, nil
}

func registrationPlanProblem(t *models.Tournament, now time.Time) string {
	if strings.TrimSpace(t.Name) == "" || strings.TrimSpace(t.Game) == "" {
		return "the tournament name or game is missing"
	}
	if strings.TrimSpace(t.Location) == "" {
		return "add an event location"
	}
	if t.Capacity < 1 || t.EntryFee < 0 {
		return "set a valid capacity and entry fee"
	}
	if t.RegistrationDeadline.IsZero() || !t.RegistrationDeadline.After(now) {
		return "set a future registration deadline"
	}
	if t.StartsAt.IsZero() || t.EndsAt.IsZero() || !t.RegistrationDeadline.Before(t.StartsAt) || !t.StartsAt.Before(t.EndsAt) {
		return "set the registration deadline before the event start and the event end after the start"
	}
	switch t.RegistrationMode {
	case models.TournamentRegistrationModeSolo:
		if t.MinTeamSize == 1 && t.MaxTeamSize == 1 {
			return ""
		}
	case models.TournamentRegistrationModeTeam:
		if t.MinTeamSize >= 1 && t.MaxTeamSize >= t.MinTeamSize {
			return ""
		}
	}
	return "set a valid registration mode and team size"
}

func canTransition(from, to models.TournamentStatus) bool {
	return slices.Contains(allowedTransitions[from], to)
}

func AllowedTransitionsFrom(from models.TournamentStatus) []models.TournamentStatus {
	allowed := allowedTransitions[from]
	out := make([]models.TournamentStatus, len(allowed))
	copy(out, allowed)
	return out
}
