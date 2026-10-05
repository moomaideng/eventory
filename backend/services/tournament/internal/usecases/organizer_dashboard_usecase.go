package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var ErrInvalidDashboardPage = errors.New("invalid dashboard pagination")

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
	Tournament models.Tournament
	Metrics    OrganizerDashboardMetrics
	Funding    TournamentFundingStats
}

type OrganizerDashboardEntry struct {
	ID        uuid.UUID
	Name      string
	Status    models.TournamentTeamStatus
	CreatedAt time.Time
	Members   []TeamMemberView
}

func NewOrganizerDashboardEntry(team models.TournamentTeam, members []TeamMemberView) OrganizerDashboardEntry {
	return OrganizerDashboardEntry{
		ID: team.ID, Name: team.Name, Status: team.Status, CreatedAt: team.CreatedAt, Members: members,
	}
}

// OrganizerDashboard is one owned tournament plus its enriched registrations.
// Organizer display name is supplied by the handler from the authenticated profile.
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

type OrganizerDashboardUseCase struct {
	repo       repositories.OrganizerDashboardRepository
	accountSvc ports.AccountService
}

func NewOrganizerDashboardUseCase(repo repositories.OrganizerDashboardRepository, accountSvc ports.AccountService) *OrganizerDashboardUseCase {
	return &OrganizerDashboardUseCase{repo: repo, accountSvc: accountSvc}
}

func (u *OrganizerDashboardUseCase) List(ctx context.Context, organizerID uuid.UUID, page, pageSize int) (*OrganizerDashboardList, error) {
	if page < 1 || pageSize < 1 || pageSize > 100 || page > 1000000 {
		return nil, ErrInvalidDashboardPage
	}
	tournaments, total, err := u.repo.ListOwned(ctx, organizerID, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]OrganizerTournamentSummary, 0, len(tournaments))
	for _, tournament := range tournaments {
		items = append(items, summarizeDashboard(tournament))
	}
	return &OrganizerDashboardList{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (u *OrganizerDashboardUseCase) Get(ctx context.Context, organizerID, tournamentID uuid.UUID) (*OrganizerDashboard, error) {
	tournament, err := u.repo.GetOwned(ctx, organizerID, tournamentID)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}
	summary := summarizeDashboard(*tournament)

	var allMembers []models.TournamentTeamMember
	for _, team := range tournament.Teams {
		allMembers = append(allMembers, team.Members...)
	}
	accounts, err := accountInfos(ctx, u.accountSvc, allMembers...)
	if err != nil {
		return nil, err
	}
	entries := make([]OrganizerDashboardEntry, 0, len(tournament.Teams))
	for _, team := range tournament.Teams {
		entries = append(entries, NewOrganizerDashboardEntry(team, NewTeamMemberViews(team.Members, accounts)))
	}
	return &OrganizerDashboard{OrganizerTournamentSummary: summary, Entries: entries}, nil
}

func summarizeDashboard(tournament models.Tournament) OrganizerTournamentSummary {
	metrics := OrganizerDashboardMetrics{TotalEntries: len(tournament.Teams)}
	participants := make(map[uuid.UUID]struct{})
	for _, team := range tournament.Teams {
		switch team.Status {
		case models.TournamentTeamStatusForming:
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
		Tournament: tournament,
		Metrics:    metrics,
		Funding:    NewFundingStats(funding),
	}
}
