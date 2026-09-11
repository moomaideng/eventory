package usecases

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
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

type OrganizerDashboard struct {
	Tournament models.Tournament
	Metrics    OrganizerDashboardMetrics
	Funding    TournamentFundingStats
}

type OrganizerDashboardList struct {
	Items    []OrganizerDashboard
	Total    int64
	Page     int
	PageSize int
}

type OrganizerDashboardUseCase struct {
	repo repositories.OrganizerDashboardRepository
}

func NewOrganizerDashboardUseCase(repo repositories.OrganizerDashboardRepository) *OrganizerDashboardUseCase {
	return &OrganizerDashboardUseCase{repo: repo}
}

func (u *OrganizerDashboardUseCase) List(ctx context.Context, accountID uuid.UUID, page, pageSize int) (*OrganizerDashboardList, error) {
	if page < 1 || pageSize < 1 || pageSize > 100 || page > 1000000 {
		return nil, ErrInvalidDashboardPage
	}
	tournaments, total, err := u.repo.ListOwned(ctx, accountID, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]OrganizerDashboard, 0, len(tournaments))
	for _, tournament := range tournaments {
		items = append(items, summarizeDashboard(tournament))
	}
	return &OrganizerDashboardList{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (u *OrganizerDashboardUseCase) Get(ctx context.Context, accountID, tournamentID uuid.UUID) (*OrganizerDashboard, error) {
	tournament, err := u.repo.GetOwned(ctx, accountID, tournamentID)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}
	result := summarizeDashboard(*tournament)
	return &result, nil
}

func summarizeDashboard(tournament models.Tournament) OrganizerDashboard {
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
	funding := TournamentFundingStats{}
	if tournament.Funding != nil {
		funding.GoalAmount = tournament.Funding.GoalAmount
		funding.RaisedAmount = tournament.Funding.RaisedAmount
		funding.SupporterCount = tournament.Funding.SupporterCount
		funding.RemainingAmount = max(funding.GoalAmount-funding.RaisedAmount, 0)
		if funding.GoalAmount > 0 {
			funding.Percentage = float64(funding.RaisedAmount) / float64(funding.GoalAmount) * 100
		}
	}
	return OrganizerDashboard{Tournament: tournament, Metrics: metrics, Funding: funding}
}
