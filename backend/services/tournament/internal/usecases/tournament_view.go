package usecases

import (
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
)

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
		views[i] = NewTournamentView(tournament, names[tournament.OrganizerID])
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
