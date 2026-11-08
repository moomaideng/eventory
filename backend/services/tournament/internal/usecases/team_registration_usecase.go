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
var ErrNotTeamMember = errors.New("only a current team member can submit their own registration answers")

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

func (u *TeamLobbyUseCase) saveNewRegistrationAnswers(ctx context.Context, tournamentID, accountID, teamID, registrationID uuid.UUID, answers []models.FieldAnswer) error {
	if err := u.validateRegistrationAnswers(ctx, tournamentID, answers); err != nil {
		return err
	}
	if _, err := u.repo.FindActiveByTournamentAndAccount(ctx, tournamentID, accountID); err == nil {
		return ErrAlreadyInTournamentLobby
	} else if !errors.Is(err, repositories.ErrTeamLobbyNotFound) {
		return translateLobbyError(err)
	}
	return u.storeRegistrationAnswers(ctx, tournamentID, accountID, teamID, registrationID, answers)
}

func (u *TeamLobbyUseCase) storeRegistrationAnswers(ctx context.Context, tournamentID, accountID, teamID, registrationID uuid.UUID, answers []models.FieldAnswer) error {
	if u.formRepo == nil {
		return nil
	}
	_, err := u.formRepo.CreateResponse(ctx, &models.FormResponse{
		TournamentID: tournamentID, AccountID: accountID, TeamID: teamID,
		RegistrationID: registrationID, Answers: answers,
	})
	return err
}

func (u *TeamLobbyUseCase) memberReadiness(ctx context.Context, team *models.TournamentTeam) (map[uuid.UUID]bool, error) {
	fields, err := u.registrationFields(ctx, team.TournamentID)
	if err != nil {
		return nil, err
	}
	ready := make(map[uuid.UUID]bool, len(team.Members))
	if len(fields) == 0 {
		for _, member := range team.Members {
			ready[member.AccountID] = true
		}
		return ready, nil
	}
	accountIDs := make([]uuid.UUID, 0, len(team.Members))
	for _, member := range team.Members {
		accountIDs = append(accountIDs, member.AccountID)
	}
	responses, err := u.formRepo.GetResponsesByAccounts(ctx, team.TournamentID, accountIDs)
	if err != nil {
		return nil, err
	}
	byRegistration := make(map[uuid.UUID][]models.FieldAnswer, len(responses))
	legacy := make(map[uuid.UUID][]models.FieldAnswer)
	for _, response := range responses {
		if response.RegistrationID == uuid.Nil {
			legacy[response.AccountID] = response.Answers
		} else {
			byRegistration[response.RegistrationID] = response.Answers
		}
	}
	for _, member := range team.Members {
		answers, exists := byRegistration[member.ID]
		if !exists {
			answers = legacy[member.AccountID]
		}
		ready[member.AccountID] = u.registry.ValidateFormAnswers(fields, answers) == nil
	}
	return ready, nil
}

func (u *TeamLobbyUseCase) SaveRegistrationAnswers(ctx context.Context, teamID, accountID uuid.UUID, answers []models.FieldAnswer) (*TeamLobbyView, error) {
	answers = firstAnswers([][]models.FieldAnswer{answers})
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	registration, err := u.repo.LockRegistration(ctx, teamID)
	if err != nil {
		return nil, translateLobbyError(err)
	}
	defer registration.Release()
	team := registration.Team()
	member := false
	var registrationID uuid.UUID
	for _, entry := range team.Members {
		if entry.AccountID == accountID {
			member = true
			registrationID = entry.ID
		}
	}
	if !member {
		return nil, ErrNotTeamMember
	}
	if team.Status != models.TournamentTeamStatusForming {
		return nil, ErrTeamLobbyNotForming
	}
	if team.Tournament.Status != models.TournamentStatusRegistration || !time.Now().Before(team.Tournament.RegistrationDeadline) {
		return nil, ErrRegistrationNotOpen
	}
	if err := u.validateRegistrationAnswers(ctx, team.TournamentID, answers); err != nil {
		return nil, err
	}
	if u.formRepo != nil {
		_, err := u.formRepo.GetResponseByAccount(ctx, team.TournamentID, accountID, registrationID)
		if err == nil {
			return nil, repositories.ErrResponseAlreadyExists
		}
		if !errors.Is(err, repositories.ErrResponseNotFound) {
			return nil, err
		}
	}
	if err := u.storeRegistrationAnswers(ctx, team.TournamentID, accountID, team.ID, registrationID, answers); err != nil {
		return nil, err
	}
	return u.toView(ctx, team)
}
