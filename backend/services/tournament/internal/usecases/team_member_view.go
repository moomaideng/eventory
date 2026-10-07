package usecases

import (
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type TeamMemberView struct {
	Member      models.TournamentTeamMember
	Handle      string
	DisplayName string
}

func NewTeamMemberViews(members []models.TournamentTeamMember, accounts map[uuid.UUID]ports.AccountInfo) []TeamMemberView {
	views := make([]TeamMemberView, 0, len(members))
	for _, member := range members {
		info := accounts[member.AccountID]
		views = append(views, TeamMemberView{
			Member:      member,
			Handle:      info.Handle,
			DisplayName: info.DisplayName,
		})
	}
	return views
}
