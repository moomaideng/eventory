package usecases

import (
	"context"
	"errors"
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
	ErrTournamentCannotBeModified = errors.New("ongoing, completed, or cancelled tournaments cannot be modified")
	ErrCapacityBelowAcceptedTeams = errors.New("capacity cannot be less than the number of accepted teams")
	ErrRosterRulesLocked          = errors.New("team size and registration mode cannot be modified after teams have locked or been accepted")
	ErrInvalidDashboardPage       = errors.New("invalid dashboard pagination")
	ErrInvalidStatusTransition    = errors.New("this status change is not allowed")
	ErrStatusUnchanged            = errors.New("tournament is already in that status")
	ErrOverrideReasonRequired     = errors.New("a reason between 10 and 500 characters is required")
	ErrUnknownStatus              = errors.New("unknown tournament status")
)

const (
	overrideReasonMinLength = 10
	overrideReasonMaxLength = 500
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
	Change       models.TournamentStatusChange
}

// TournamentStatusDetail describes where a tournament stands and where an
// organizer may move it next.
type TournamentStatusDetail struct {
	CurrentStatus      models.TournamentStatus
	AllowedTransitions []models.TournamentStatus
	Changes            []models.TournamentStatusChange
}

type TournamentUseCase struct {
	tournamentRepo repositories.TournamentRepository
	accountSvc     ports.AccountService
}

func NewTournamentUseCase(tournamentRepo repositories.TournamentRepository, accountSvc ports.AccountService) *TournamentUseCase {
	return &TournamentUseCase{tournamentRepo: tournamentRepo, accountSvc: accountSvc}
}

type OrganizerDashboardUseCase = TournamentUseCase

func NewOrganizerDashboardUseCase(repo repositories.OrganizerDashboardRepository, accountSvc ports.AccountService) *OrganizerDashboardUseCase {
	return NewTournamentUseCase(repo, accountSvc)
}

type TournamentStatusUseCase = TournamentUseCase

func NewTournamentStatusUseCase(repo repositories.TournamentStatusRepository, accountSvc ports.AccountService) *TournamentStatusUseCase {
	return NewTournamentUseCase(repo, accountSvc)
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
		"":                                        true,
		models.TournamentStatusDraft:              true,
		models.TournamentStatusCrowdfunding:       true,
		models.TournamentStatusRegistrationOpen:   true,
		models.TournamentStatusRegistrationClosed: true,
		models.TournamentStatusOngoing:            true,
		models.TournamentStatusCompleted:          true,
		models.TournamentStatusCancelled:          true,
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

// CreateTournament initializes and persists a new tournament record for an organizer.
// It applies default status REGISTRATION_OPEN, published=true, hardcoded currency "THB",
// automatically coerces solo tournaments to min/max team size of 1, and creates an initial
// funding stub. It strictly validates that registrationDeadline < startsAt < endsAt.
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
		Staff: []models.TournamentStaff{
			{
				AccountID: accountID,
				Role:      models.TournamentStaffRoleOwner,
			},
		},
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

	// CANCELLED is terminal too (US3-3): cancelling refunds entrants, so the
	// tournament must not be re-configured afterwards.
	if tournament.Status == models.TournamentStatusOngoing ||
		tournament.Status == models.TournamentStatusCompleted ||
		tournament.Status == models.TournamentStatusCancelled {
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
// override it to.
var allowedTransitions = map[models.TournamentStatus][]models.TournamentStatus{
	models.TournamentStatusDraft: {
		models.TournamentStatusRegistrationOpen,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusCrowdfunding: {
		models.TournamentStatusRegistrationOpen,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusRegistrationOpen: {
		models.TournamentStatusRegistrationClosed,
		models.TournamentStatusOngoing, // skip ahead when registration is cut short
		models.TournamentStatusCompleted,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusRegistrationClosed: {
		models.TournamentStatusRegistrationOpen, // reopen after a delay
		models.TournamentStatusOngoing,
		models.TournamentStatusCompleted,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusOngoing: {
		models.TournamentStatusRegistrationClosed, // started early by mistake
		models.TournamentStatusCompleted,
		models.TournamentStatusCancelled,
	},
	models.TournamentStatusCompleted: {}, // terminal
	models.TournamentStatusCancelled: {}, // terminal
}

func (u *TournamentUseCase) Override(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
	rawStatus, rawReason string,
) (*TournamentStatusOverride, error) {
	to := models.TournamentStatus(strings.ToUpper(strings.TrimSpace(rawStatus)))
	if _, known := allowedTransitions[to]; !known {
		return nil, ErrUnknownStatus
	}

	reason := strings.TrimSpace(rawReason)
	if n := utf8.RuneCountInString(reason); n < overrideReasonMinLength || n > overrideReasonMaxLength {
		return nil, ErrOverrideReasonRequired
	}

	actor, err := u.accountSvc.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	actorDisplayName := actor.DisplayName

	var from models.TournamentStatus
	change, err := u.tournamentRepo.ApplyOverride(ctx, accountID, actorDisplayName, tournamentID, reason,
		func(current *models.Tournament) (models.TournamentStatus, error) {
			from = current.Status
			if current.Status == to {
				return "", ErrStatusUnchanged
			}
			if !canTransition(current.Status, to) {
				return "", ErrInvalidStatusTransition
			}
			return to, nil
		})
	if err != nil {
		if errors.Is(err, repositories.ErrTournamentNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}

	return &TournamentStatusOverride{
		TournamentID: tournamentID, FromStatus: from, ToStatus: to,
		Reason: reason, Change: *change,
	}, nil
}

func (u *TournamentUseCase) History(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
) (*TournamentStatusDetail, error) {
	history, err := u.tournamentRepo.ListHistory(ctx, accountID, tournamentID)
	if err != nil {
		if errors.Is(err, repositories.ErrTournamentNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}
	return &TournamentStatusDetail{
		CurrentStatus:      history.CurrentStatus,
		AllowedTransitions: AllowedTransitionsFrom(history.CurrentStatus),
		Changes:            history.Changes,
	}, nil
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
