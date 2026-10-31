package usecases

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
	"strings"
	"time"
)

var ErrMemberAnswersIncomplete = errors.New("every member must complete the current registration questions before the team can be submitted")
var ErrNotTeamMember = errors.New("only a current team member can update their own registration answers")

func firstAnswers(submitted [][]models.FieldAnswer) []models.FieldAnswer {
	if len(submitted) == 0 {
		return nil
	}
	answers := append([]models.FieldAnswer(nil), submitted[0]...)
	for i := range answers {
		answers[i].FieldID = strings.TrimSpace(answers[i].FieldID)
	}
	return answers
}

func (u *TeamLobbyUseCase) registrationFields(ctx context.Context, tournamentID uuid.UUID) ([]models.FormField, error) {
	if u.formRepo == nil || u.registry == nil {
		return nil, nil
	}
	form, err := u.formRepo.GetFormByTournamentID(ctx, tournamentID)
	if errors.Is(err, repositories.ErrFormNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return form.Fields, nil
}

func (u *TeamLobbyUseCase) validateRegistrationAnswers(ctx context.Context, tournamentID uuid.UUID, answers []models.FieldAnswer) error {
	fields, err := u.registrationFields(ctx, tournamentID)
	if err != nil {
		return err
	}
	if u.registry == nil {
		return nil
	}
	return u.registry.ValidateFormAnswers(fields, answers)
}

func (u *TeamLobbyUseCase) memberReadiness(ctx context.Context, team *models.TournamentTeam) (map[uuid.UUID]bool, error) {
	fields, err := u.registrationFields(ctx, team.TournamentID)
	if err != nil {
		return nil, err
	}
	ready := make(map[uuid.UUID]bool, len(team.Members))
	for _, member := range team.Members {
		answers := member.RegistrationAnswers
		// Old memberships keep their existing Mongo responses until the member edits.
		if !member.RegistrationSubmitted && len(fields) > 0 && u.formRepo != nil {
			response, err := u.formRepo.GetResponseByAccount(ctx, team.TournamentID, member.AccountID)
			if err != nil && !errors.Is(err, repositories.ErrResponseNotFound) {
				return nil, err
			}
			if err == nil {
				answers = response.Answers
			}
		}
		ready[member.AccountID] = u.registry == nil || u.registry.ValidateFormAnswers(fields, answers) == nil
	}
	return ready, nil
}

func (u *TeamLobbyUseCase) SaveRegistrationAnswers(ctx context.Context, teamID, accountID uuid.UUID, answers []models.FieldAnswer) (*TeamLobbyView, error) {
	answers = firstAnswers([][]models.FieldAnswer{answers})
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	team, err := u.repo.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	member := false
	for _, entry := range team.Members {
		if entry.AccountID == accountID {
			member = true
		}
	}
	if !member {
		return nil, ErrNotTeamMember
	}
	if team.Status != models.TournamentTeamStatusForming {
		return nil, repositories.ErrTeamLobbyNotForming
	}
	if team.Tournament.Status != models.TournamentStatusRegistrationOpen || !time.Now().Before(team.Tournament.RegistrationDeadline) {
		return nil, ErrRegistrationNotOpen
	}
	if err := u.validateRegistrationAnswers(ctx, team.TournamentID, answers); err != nil {
		return nil, err
	}
	if err := u.repo.SaveMemberAnswers(ctx, teamID, accountID, answers); err != nil {
		return nil, err
	}
	team, err = u.repo.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return u.toView(ctx, team)
}
