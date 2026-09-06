package usecases

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
)

var (
	ErrTeamsNotAllowed       = errors.New("this tournament does not allow team registration")
	ErrRegistrationNotOpen   = errors.New("tournament registration is not open")
	ErrInvalidTeamName       = errors.New("team name must be between 1 and 120 characters")
	ErrLobbyAccessDenied     = errors.New("only the team captain can manage this lobby")
	ErrInviteCodeUnavailable = errors.New("unable to create a unique invite code")
)

type TeamLobbyUseCase struct {
	repo TeamLobbyRepository
}

type TeamLobbyRepository interface {
	FindTournamentByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.TournamentTeam, error)
	FindByInviteCode(ctx context.Context, inviteCode string) (*models.TournamentTeam, error)
	FindActiveByTournamentAndAccount(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.TournamentTeam, error)
	Create(ctx context.Context, team *models.TournamentTeam, captain *models.TournamentTeamMember) (*models.TournamentTeam, error)
	Join(ctx context.Context, teamID, accountID uuid.UUID) (*models.TournamentTeam, error)
	RegenerateInvite(ctx context.Context, teamID uuid.UUID, inviteCode string) (*models.TournamentTeam, error)
	Lock(ctx context.Context, teamID uuid.UUID) (*models.TournamentTeam, error)
	RemoveMember(ctx context.Context, teamID, memberID uuid.UUID) (*models.TournamentTeam, error)
	Disband(ctx context.Context, teamID uuid.UUID) error
}

func NewTeamLobbyUseCase(repo TeamLobbyRepository) *TeamLobbyUseCase {
	return &TeamLobbyUseCase{repo: repo}
}

func (u *TeamLobbyUseCase) Create(ctx context.Context, tournamentID, captainID uuid.UUID, name string) (*models.TournamentTeam, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 120 {
		return nil, ErrInvalidTeamName
	}
	tournament, err := u.repo.FindTournamentByID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if tournament == nil {
		return nil, ErrTournamentNotFound
	}
	if tournament.Status != models.TournamentStatusRegistrationOpen {
		return nil, ErrRegistrationNotOpen
	}
	if tournament.RegistrationMode != models.TournamentRegistrationModeTeam && tournament.RegistrationMode != models.TournamentRegistrationModeBoth {
		return nil, ErrTeamsNotAllowed
	}

	inviteCode, err := u.newInviteCode(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	team := &models.TournamentTeam{
		ID: uuid.New(), TournamentID: tournamentID, Name: name,
		InviteCode: inviteCode, Status: models.TournamentTeamStatusForming,
	}
	captain := &models.TournamentTeamMember{
		ID: uuid.New(), AccountID: captainID, Role: models.TournamentTeamMemberRoleCaptain, JoinedAt: now,
	}
	return u.repo.Create(ctx, team, captain)
}

func (u *TeamLobbyUseCase) GetByInviteCode(ctx context.Context, inviteCode string) (*models.TournamentTeam, error) {
	return u.repo.FindByInviteCode(ctx, normalizeInviteCode(inviteCode))
}

func (u *TeamLobbyUseCase) GetActiveForTournament(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.TournamentTeam, error) {
	return u.repo.FindActiveByTournamentAndAccount(ctx, tournamentID, accountID)
}

func (u *TeamLobbyUseCase) Join(ctx context.Context, inviteCode string, accountID uuid.UUID) (*models.TournamentTeam, error) {
	team, err := u.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, err
	}
	return u.repo.Join(ctx, team.ID, accountID)
}

func (u *TeamLobbyUseCase) RegenerateInvite(ctx context.Context, teamID, accountID uuid.UUID) (*models.TournamentTeam, error) {
	if _, err := u.captainTeam(ctx, teamID, accountID); err != nil {
		return nil, err
	}
	inviteCode, err := u.newInviteCode(ctx)
	if err != nil {
		return nil, err
	}
	return u.repo.RegenerateInvite(ctx, teamID, inviteCode)
}

func (u *TeamLobbyUseCase) Lock(ctx context.Context, teamID, accountID uuid.UUID) (*models.TournamentTeam, error) {
	if _, err := u.captainTeam(ctx, teamID, accountID); err != nil {
		return nil, err
	}
	return u.repo.Lock(ctx, teamID)
}

func (u *TeamLobbyUseCase) RemoveMember(ctx context.Context, teamID, memberID, accountID uuid.UUID) (*models.TournamentTeam, error) {
	if _, err := u.captainTeam(ctx, teamID, accountID); err != nil {
		return nil, err
	}
	return u.repo.RemoveMember(ctx, teamID, memberID)
}

func (u *TeamLobbyUseCase) Disband(ctx context.Context, teamID, accountID uuid.UUID) error {
	if _, err := u.captainTeam(ctx, teamID, accountID); err != nil {
		return err
	}
	return u.repo.Disband(ctx, teamID)
}

func (u *TeamLobbyUseCase) captainTeam(ctx context.Context, teamID, accountID uuid.UUID) (*models.TournamentTeam, error) {
	team, err := u.repo.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	for _, member := range team.Members {
		if member.AccountID == accountID && member.Role == models.TournamentTeamMemberRoleCaptain {
			return team, nil
		}
	}
	return nil, ErrLobbyAccessDenied
}

func (u *TeamLobbyUseCase) newInviteCode(ctx context.Context) (string, error) {
	for range 10 {
		inviteCode, err := generateInviteCode()
		if err != nil {
			return "", err
		}
		_, err = u.repo.FindByInviteCode(ctx, inviteCode)
		if errors.Is(err, repositories.ErrTeamLobbyNotFound) {
			return inviteCode, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", ErrInviteCodeUnavailable
}

func normalizeInviteCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func generateInviteCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for index := range code {
		randomIndex, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		code[index] = alphabet[randomIndex.Int64()]
	}
	return string(code), nil
}
