package repositories

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

// SaveMemberAnswers serializes answer edits with roster locking. A rejected
// conditional update leaves both the previous answers and their version intact.
func (r *teamLobbyRepository) SaveMemberAnswers(ctx context.Context, teamID, accountID uuid.UUID, answers []models.FieldAnswer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		team, err := r.lockTeam(tx, teamID)
		if err != nil {
			return err
		}
		if team.Status != models.TournamentTeamStatusForming {
			return ErrTeamLobbyNotForming
		}
		var member models.TournamentTeamMember
		if err := tx.Where("tournament_team_id = ? AND account_id = ?", teamID, accountID).First(&member).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTeamLobbyNotFound
			}
			return err
		}
		member.RegistrationAnswers = answers
		member.RegistrationSubmitted = true
		member.RegistrationVersion++
		return tx.Save(&member).Error
	})
}

func (r *teamLobbyRepository) MemberResponses(ctx context.Context, tournamentID uuid.UUID) ([]models.FormResponse, error) {
	var members []models.TournamentTeamMember
	err := r.db.WithContext(ctx).Model(&models.TournamentTeamMember{}).
		Joins("JOIN tournament_teams ON tournament_teams.id = tournament_team_members.tournament_team_id").
		Where("tournament_teams.tournament_id = ? AND registration_submitted = true", tournamentID).
		Order("CASE WHEN tournament_teams.status = 'REJECTED' THEN 0 ELSE 1 END, tournament_team_members.joined_at ASC, tournament_team_members.id ASC").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	responses := make([]models.FormResponse, 0, len(members))
	for _, member := range members {
		teamID := member.TournamentTeamID
		responses = append(responses, models.FormResponse{TeamID: &teamID, TournamentID: tournamentID, AccountID: member.AccountID, Answers: member.RegistrationAnswers})
	}
	return responses, nil
}
