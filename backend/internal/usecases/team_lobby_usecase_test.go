package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
	"github.com/moomaideng/eventory/internal/usecases"
)

type teamLobbyRepositoryStub struct {
	tournament *models.Tournament
	team       *models.TournamentTeam
	created    *models.TournamentTeam
	captain    *models.TournamentTeamMember
}

func (s *teamLobbyRepositoryStub) FindTournamentByID(_ context.Context, _ uuid.UUID) (*models.Tournament, error) {
	return s.tournament, nil
}
func (s *teamLobbyRepositoryStub) FindByID(_ context.Context, _ uuid.UUID) (*models.TournamentTeam, error) {
	if s.team == nil {
		return nil, repositories.ErrTeamLobbyNotFound
	}
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) FindByInviteCode(_ context.Context, _ string) (*models.TournamentTeam, error) {
	return nil, repositories.ErrTeamLobbyNotFound
}
func (s *teamLobbyRepositoryStub) Create(_ context.Context, team *models.TournamentTeam, captain *models.TournamentTeamMember) (*models.TournamentTeam, error) {
	captain.TournamentTeamID = team.ID
	s.created, s.captain = team, captain
	return team, nil
}
func (s *teamLobbyRepositoryStub) Join(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.TournamentTeam, error) {
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) RegenerateInvite(_ context.Context, _ uuid.UUID, _ string) (*models.TournamentTeam, error) {
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) Lock(_ context.Context, _ uuid.UUID) (*models.TournamentTeam, error) {
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) RemoveMember(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.TournamentTeam, error) {
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) Disband(_ context.Context, _ uuid.UUID) error { return nil }

func TestCreateTeamLobby_InitializesCaptainAndInvite(t *testing.T) {
	tournamentID := uuid.New()
	captainID := uuid.New()
	repo := &teamLobbyRepositoryStub{tournament: &models.Tournament{
		ID: tournamentID, Status: models.TournamentStatusRegistrationOpen,
		RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 5, MaxTeamSize: 5,
	}}
	useCase := usecases.NewTeamLobbyUseCase(repo)

	team, err := useCase.Create(context.Background(), tournamentID, captainID, "  Night Owls  ")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if team.Name != "Night Owls" || len(team.InviteCode) != 6 {
		t.Fatalf("unexpected team: %+v", team)
	}
	if repo.captain == nil || repo.captain.AccountID != captainID || repo.captain.TournamentTeamID != team.ID || repo.captain.Role != models.TournamentTeamMemberRoleCaptain {
		t.Fatalf("captain membership was not initialized: %+v", repo.captain)
	}
}

func TestCreateTeamLobby_RejectsUnavailableTournamentState(t *testing.T) {
	tournamentID := uuid.New()
	captainID := uuid.New()
	tests := []struct {
		name       string
		tournament *models.Tournament
		want       error
	}{
		{
			name: "solo tournament", tournament: &models.Tournament{ID: tournamentID, Status: models.TournamentStatusRegistrationOpen, RegistrationMode: models.TournamentRegistrationModeSolo}, want: usecases.ErrTeamsNotAllowed,
		},
		{
			name: "closed registration", tournament: &models.Tournament{ID: tournamentID, Status: models.TournamentStatusRegistrationClosed, RegistrationMode: models.TournamentRegistrationModeTeam}, want: usecases.ErrRegistrationNotOpen,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := usecases.NewTeamLobbyUseCase(&teamLobbyRepositoryStub{tournament: test.tournament})
			_, err := useCase.Create(context.Background(), tournamentID, captainID, "Night Owls")
			if !errors.Is(err, test.want) {
				t.Fatalf("Create() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestCaptainActions_RejectNonCaptain(t *testing.T) {
	captainID := uuid.New()
	repo := &teamLobbyRepositoryStub{team: &models.TournamentTeam{
		ID: uuid.New(),
		Members: []models.TournamentTeamMember{{
			AccountID: captainID, Role: models.TournamentTeamMemberRoleCaptain,
		}},
	}}
	useCase := usecases.NewTeamLobbyUseCase(repo)

	_, err := useCase.Lock(context.Background(), repo.team.ID, uuid.New())
	if !errors.Is(err, usecases.ErrLobbyAccessDenied) {
		t.Fatalf("Lock() error = %v, want captain access denial", err)
	}
}
