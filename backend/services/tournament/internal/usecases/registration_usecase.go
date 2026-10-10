package usecases

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/registration"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type RegistrationUseCase struct {
	repo        repositories.TeamLobbyRepository
	tournaments repositories.TournamentRepository
	accounts    ports.AccountService
}

func NewRegistrationUseCase(repo repositories.TeamLobbyRepository, tournaments repositories.TournamentRepository, accounts ports.AccountService) *RegistrationUseCase {
	return &RegistrationUseCase{repo: repo, tournaments: tournaments, accounts: accounts}
}

func (u *RegistrationUseCase) GetForm(ctx context.Context, id, viewer uuid.UUID) (*models.RegistrationForm, error) {
	tournament, err := u.tournaments.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !tournament.Published {
		organizer, err := u.accounts.GetOrganizerProfileByAccountID(ctx, viewer)
		if err != nil || organizer.ID != tournament.OrganizerID {
			return nil, repositories.ErrFormForbidden
		}
	}
	return u.repo.FindRegistrationForm(ctx, id)
}

func (u *RegistrationUseCase) SaveForm(ctx context.Context, id, viewer uuid.UUID, form *models.RegistrationForm) error {
	if _, err := requireActiveAccount(ctx, u.accounts, viewer); err != nil {
		return err
	}
	if err := registration.ValidateForm(form); err != nil {
		return err
	}
	organizer, err := u.accounts.GetOrganizerProfileByAccountID(ctx, viewer)
	if err != nil {
		return err
	}
	return u.repo.SaveRegistrationForm(ctx, id, organizer.ID, form)
}

func (u *RegistrationUseCase) GetAnswers(ctx context.Context, memberID, viewer uuid.UUID) (*models.TournamentTeamMember, error) {
	organizerID := uuid.Nil
	organizer, err := u.accounts.GetOrganizerProfileByAccountID(ctx, viewer)
	if err == nil {
		organizerID = organizer.ID
	} else if !errors.Is(err, ports.ErrOrganizerNotFound) {
		return nil, err
	}
	return u.repo.FindMemberRegistration(ctx, memberID, viewer, organizerID)
}
