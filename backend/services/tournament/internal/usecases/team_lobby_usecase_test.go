package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
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
	applied    string
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
func (s *teamLobbyRepositoryStub) RegenerateInvite(_ context.Context, _ uuid.UUID, inviteCode string) (*models.TournamentTeam, error) {
	if s.team != nil {
		s.team.InviteCode = &inviteCode
	}
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) Lock(_ context.Context, _ uuid.UUID) (*models.TournamentTeam, error) {
	if s.team != nil {
		s.team.Status = models.TournamentTeamStatusLocked
	}
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) RemoveMember(_ context.Context, _ uuid.UUID, memberID uuid.UUID) (*models.TournamentTeam, error) {
	if s.team != nil {
		newMembers := make([]models.TournamentTeamMember, 0, len(s.team.Members))
		for _, m := range s.team.Members {
			if m.ID != memberID {
				newMembers = append(newMembers, m)
			}
		}
		s.team.Members = newMembers
	}
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) Disband(_ context.Context, _ uuid.UUID) error {
	s.team = nil
	return nil
}
func (s *teamLobbyRepositoryStub) Pending(_ context.Context) ([]models.TournamentTeam, error) {
	return nil, nil
}
func (s *teamLobbyRepositoryStub) ApplyPayment(_ context.Context, _ *models.TournamentTeam, _ uuid.UUID, status string) error {
	s.applied = status
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
		{
			name: "crowdfunding registration", tournament: &models.Tournament{ID: tournamentID, Status: models.TournamentStatusCrowdfunding, RegistrationMode: models.TournamentRegistrationModeTeam}, want: usecases.ErrRegistrationNotOpen,
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

func TestRegenerateInvite(t *testing.T) {
	captainID := uuid.New()
	teamID := uuid.New()
	oldCode := "OLDCOD"
	team := &models.TournamentTeam{
		ID:         teamID,
		InviteCode: &oldCode,
		Members: []models.TournamentTeamMember{{
			AccountID: captainID, Role: models.TournamentTeamMemberRoleCaptain,
		}},
	}
	repo := &teamLobbyRepositoryStub{team: team}
	useCase := usecases.NewTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

	// Non-captain denied
	_, err := useCase.RegenerateInvite(context.Background(), teamID, uuid.New())
	if !errors.Is(err, usecases.ErrLobbyAccessDenied) {
		t.Fatalf("RegenerateInvite() error = %v, want ErrLobbyAccessDenied", err)
	}

	// Captain succeeds
	updated, err := useCase.RegenerateInvite(context.Background(), teamID, captainID)
	if err != nil {
		t.Fatalf("RegenerateInvite() error = %v", err)
	}
	if updated.Team.InviteCode == nil || *updated.Team.InviteCode == oldCode {
		t.Fatalf("invite code was not regenerated: %v", updated.Team.InviteCode)
	}
	if !usecases.ValidateInviteCode(*updated.Team.InviteCode) {
		t.Fatalf("regenerated code is invalid: %s", *updated.Team.InviteCode)
	}
}

func TestLock_Success(t *testing.T) {
	captainID := uuid.New()
	teamID := uuid.New()
	team := &models.TournamentTeam{
		ID:     teamID,
		Status: models.TournamentTeamStatusForming,
		Members: []models.TournamentTeamMember{{
			AccountID: captainID, Role: models.TournamentTeamMemberRoleCaptain,
		}},
	}
	repo := &teamLobbyRepositoryStub{team: team}
	useCase := usecases.NewTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

	locked, err := useCase.Lock(context.Background(), teamID, captainID)
	if err != nil {
		t.Fatalf("Lock() error = %v", err)
	}
	if locked.Team.Status != models.TournamentTeamStatusLocked {
		t.Fatalf("team status = %v, want %v", locked.Team.Status, models.TournamentTeamStatusLocked)
	}
}

func TestRemoveMember(t *testing.T) {
	captainID := uuid.New()
	memberID := uuid.New()
	memberAccountID := uuid.New()
	teamID := uuid.New()
	team := &models.TournamentTeam{
		ID: teamID,
		Members: []models.TournamentTeamMember{
			{ID: uuid.New(), AccountID: captainID, Role: models.TournamentTeamMemberRoleCaptain},
			{ID: memberID, AccountID: memberAccountID, Role: models.TournamentTeamMemberRoleMember},
		},
	}
	repo := &teamLobbyRepositoryStub{team: team}
	accounts := newStubAccountService().
		withAccount(captainID, "captain", "Captain").
		withAccount(memberAccountID, "member", "Member")
	useCase := usecases.NewTeamLobbyUseCase(repo, accounts)

	// Non-captain denied
	_, err := useCase.RemoveMember(context.Background(), teamID, memberID, memberAccountID)
	if !errors.Is(err, usecases.ErrLobbyAccessDenied) {
		t.Fatalf("RemoveMember() error = %v, want ErrLobbyAccessDenied", err)
	}

	// Captain succeeds
	updated, err := useCase.RemoveMember(context.Background(), teamID, memberID, captainID)
	if err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}
	if len(updated.Team.Members) != 1 || updated.Team.Members[0].ID == memberID {
		t.Fatalf("member was not removed: %+v", updated.Team.Members)
	}
}

func TestDisband(t *testing.T) {
	captainID := uuid.New()
	teamID := uuid.New()
	team := &models.TournamentTeam{
		ID: teamID,
		Members: []models.TournamentTeamMember{{
			AccountID: captainID, Role: models.TournamentTeamMemberRoleCaptain,
		}},
	}
	repo := &teamLobbyRepositoryStub{team: team}
	useCase := usecases.NewTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

	// Non-captain denied
	err := useCase.Disband(context.Background(), teamID, uuid.New())
	if !errors.Is(err, usecases.ErrLobbyAccessDenied) {
		t.Fatalf("Disband() error = %v, want ErrLobbyAccessDenied", err)
	}

	// Captain succeeds
	if err := useCase.Disband(context.Background(), teamID, captainID); err != nil {
		t.Fatalf("Disband() error = %v", err)
	}
	if repo.team != nil {
		t.Fatalf("team was not disbanded: %+v", repo.team)
	}
}

type registrationPaymentStub struct {
	mutate func(*ports.RegistrationPayment)
	err    error
}

func (p registrationPaymentStub) EnsureInvoice(_ context.Context, r ports.RegistrationPaymentRequest) (*ports.RegistrationPayment, error) {
	payment := &ports.RegistrationPayment{ID: uuid.New(), PayerAccountID: r.PayerAccountID, Source: r.Source, Reference: r.Reference, AmountMinor: r.AmountMinor, Currency: r.Currency, ExpiresAt: r.ExpiresAt, Status: "PAID"}
	if p.mutate != nil {
		p.mutate(payment)
	}
	return payment, p.err
}

func TestSubmissionRequiresAuthoritativeMatchingInvoice(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*ports.RegistrationPayment)
		err    error
	}{
		{name: "unavailable", err: errors.New("timeout after creating invoice")},
		{name: "wrong payer", mutate: func(p *ports.RegistrationPayment) { p.PayerAccountID = uuid.New() }},
		{name: "wrong amount", mutate: func(p *ports.RegistrationPayment) { p.AmountMinor++ }},
		{name: "wrong team or attempt", mutate: func(p *ports.RegistrationPayment) { p.Reference = uuid.NewString() }},
		{name: "wrong currency", mutate: func(p *ports.RegistrationPayment) { p.Currency = "USD" }},
		{name: "wrong source", mutate: func(p *ports.RegistrationPayment) { p.Source = "OTHER" }},
		{name: "wrong deadline", mutate: func(p *ports.RegistrationPayment) { p.ExpiresAt = p.ExpiresAt.Add(time.Second) }},
		{name: "unknown status", mutate: func(p *ports.RegistrationPayment) { p.Status = "UNKNOWN" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			attempt, payer := uuid.New(), uuid.New()
			expires := time.Now().Add(time.Minute)
			team := &models.TournamentTeam{ID: uuid.New(), Status: models.TournamentTeamStatusPaymentPending, PaymentAttemptID: &attempt, PaymentPayerID: &payer, PaymentExpiresAt: &expires, PaymentAmountMinor: 10000}
			repo := &teamLobbyRepositoryStub{}
			uc := usecases.NewTeamLobbyUseCase(repo, nil, registrationPaymentStub{tc.mutate, tc.err})
			if err := uc.Reconcile(context.Background(), team); !errors.Is(err, usecases.ErrPaymentUnavailable) {
				t.Fatalf("error = %v", err)
			}
			if repo.applied != "" {
				t.Fatal("unverified payment changed roster")
			}
			// A later retry of the same durable attempt can complete after an outage.
			uc = usecases.NewTeamLobbyUseCase(repo, nil, registrationPaymentStub{})
			if err := uc.Reconcile(context.Background(), team); err != nil {
				t.Fatal(err)
			}
			if repo.applied != "PAID" {
				t.Fatal("paid attempt was not submitted")
			}
		})
	}
}
