package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type teamLobbyRepositoryStub struct {
	tournament *models.Tournament
	team       *models.TournamentTeam
	activeTeam *models.TournamentTeam
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
func (s *teamLobbyRepositoryStub) FindActiveByTournamentAndAccount(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.TournamentTeam, error) {
	if s.activeTeam == nil {
		return nil, repositories.ErrTeamLobbyNotFound
	}
	return s.activeTeam, nil
}
func (s *teamLobbyRepositoryStub) Create(_ context.Context, team *models.TournamentTeam, captain *models.TournamentTeamMember) (*models.TournamentTeam, error) {
	captain.TournamentTeamID = team.ID
	team.Members = []models.TournamentTeamMember{*captain}
	s.created, s.captain, s.team = team, captain, team
	return team, nil
}
func (s *teamLobbyRepositoryStub) Join(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.TournamentTeam, error) {
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) RegenerateInvite(_ context.Context, _ uuid.UUID, _ string) (*models.TournamentTeam, error) {
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) Lock(_ context.Context, _ uuid.UUID) (*models.TournamentTeam, error) {
	if s.team != nil {
		s.team.Status = models.TournamentTeamStatusLocked
	}
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) RemoveMember(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*models.TournamentTeam, error) {
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) Disband(_ context.Context, _ uuid.UUID) error { return nil }
func (s *teamLobbyRepositoryStub) Pending(_ context.Context) ([]models.TournamentTeam, error) {
	return nil, nil
}
func (s *teamLobbyRepositoryStub) ApplyPayment(_ context.Context, _ *models.TournamentTeam, _ uuid.UUID, _ string) error {
	return nil
}

func TestCreateTeamLobby_InitializesCaptainAndInvite(t *testing.T) {
	tournamentID := uuid.New()
	captainID := uuid.New()
	repo := &teamLobbyRepositoryStub{tournament: &models.Tournament{
		ID: tournamentID, Status: models.TournamentStatusRegistrationOpen,
		RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 5, MaxTeamSize: 5,
	}}
	useCase := usecases.NewTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

	team, err := useCase.Create(context.Background(), tournamentID, captainID, "  Night Owls  ")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if team.Team.Name != "Night Owls" || team.Team.InviteCode == nil || len(*team.Team.InviteCode) != 6 {
		t.Fatalf("unexpected team: %+v", team)
	}
	if !usecases.ValidateInviteCode(*team.Team.InviteCode) {
		t.Fatalf("invalid invite code: %s", *team.Team.InviteCode)
	}
	if repo.captain == nil || repo.captain.AccountID != captainID || repo.captain.TournamentTeamID != team.Team.ID || repo.captain.Role != models.TournamentTeamMemberRoleCaptain {
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
			name: "tournament not found", tournament: nil, want: usecases.ErrTournamentNotFound,
		},
		{
			name: "closed registration", tournament: &models.Tournament{ID: tournamentID, Status: models.TournamentStatusRegistrationClosed, RegistrationMode: models.TournamentRegistrationModeTeam}, want: usecases.ErrRegistrationNotOpen,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := usecases.NewTeamLobbyUseCase(&teamLobbyRepositoryStub{tournament: test.tournament}, newStubAccountService().withAccount(captainID, "captain", "Captain"))
			_, err := useCase.Create(context.Background(), tournamentID, captainID, "Night Owls")
			if !errors.Is(err, test.want) {
				t.Fatalf("Create() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestRegisterSolo_Success(t *testing.T) {
	tournamentID := uuid.New()
	playerID := uuid.New()
	tournament := &models.Tournament{
		ID:               tournamentID,
		Status:           models.TournamentStatusRegistrationOpen,
		MinTeamSize:      1,
		MaxTeamSize:      1,
		RegistrationMode: models.TournamentRegistrationModeSolo,
	}
	repo := &teamLobbyRepositoryStub{tournament: tournament}
	accounts := newStubAccountService().withAccount(playerID, "playerone", "Player One")
	useCase := usecases.NewTeamLobbyUseCase(repo, accounts)

	got, err := useCase.RegisterSolo(context.Background(), tournamentID, playerID)
	if err != nil {
		t.Fatalf("RegisterSolo() error = %v", err)
	}
	if got.Team == nil {
		t.Fatal("RegisterSolo() returned nil team")
	}
	if repo.created.Name != "Player One" {
		t.Fatalf("RegisterSolo() team name = %q, want %q", repo.created.Name, "Player One")
	}
	if repo.created.InviteCode != nil {
		t.Fatalf("RegisterSolo() team invite code = %v, want nil", repo.created.InviteCode)
	}
	if repo.captain.AccountID != playerID || repo.captain.Role != models.TournamentTeamMemberRoleCaptain {
		t.Fatalf("RegisterSolo() captain = %+v", repo.captain)
	}
}

func TestRegisterSolo_RejectsNonSoloTournament(t *testing.T) {
	tournamentID := uuid.New()
	playerID := uuid.New()
	tournament := &models.Tournament{
		ID:               tournamentID,
		Status:           models.TournamentStatusRegistrationOpen,
		MinTeamSize:      5,
		MaxTeamSize:      5,
		RegistrationMode: models.TournamentRegistrationModeTeam,
	}
	repo := &teamLobbyRepositoryStub{tournament: tournament}
	accounts := newStubAccountService().withAccount(playerID, "playerone", "Player One")
	useCase := usecases.NewTeamLobbyUseCase(repo, accounts)

	_, err := useCase.RegisterSolo(context.Background(), tournamentID, playerID)
	if !errors.Is(err, usecases.ErrNotSoloTournament) {
		t.Fatalf("RegisterSolo() error = %v, want %v", err, usecases.ErrNotSoloTournament)
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
	useCase := usecases.NewTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

	_, err := useCase.Lock(context.Background(), repo.team.ID, uuid.New())
	if !errors.Is(err, usecases.ErrLobbyAccessDenied) {
		t.Fatalf("Lock() error = %v, want captain access denial", err)
	}
}

func TestGetActiveForTournament_ReturnsActiveTeam(t *testing.T) {
	team := &models.TournamentTeam{ID: uuid.New(), Name: "Night Owls"}
	accountID := uuid.New()
	useCase := usecases.NewTeamLobbyUseCase(&teamLobbyRepositoryStub{activeTeam: team}, newStubAccountService().withAccount(accountID, "player", "Player"))

	got, err := useCase.GetActiveForTournament(context.Background(), uuid.New(), accountID)
	if err != nil {
		t.Fatalf("GetActiveForTournament() error = %v", err)
	}
	if got.Team.ID != team.ID {
		t.Fatalf("GetActiveForTournament() = %+v, want %+v", got.Team, team)
	}
}

func TestGetByInviteCode_ValidatesFormat(t *testing.T) {
	repo := &teamLobbyRepositoryStub{}
	useCase := usecases.NewTeamLobbyUseCase(repo, newStubAccountService().withAccount(uuid.New(), "user", "User"))

	_, err := useCase.GetByInviteCode(context.Background(), "INVALID-CODE", uuid.New())
	if !errors.Is(err, usecases.ErrInvalidInviteCode) {
		t.Fatalf("expected ErrInvalidInviteCode, got %v", err)
	}

	_, err = useCase.Join(context.Background(), "123", uuid.New())
	if !errors.Is(err, usecases.ErrInvalidInviteCode) {
		t.Fatalf("expected ErrInvalidInviteCode, got %v", err)
	}
}
