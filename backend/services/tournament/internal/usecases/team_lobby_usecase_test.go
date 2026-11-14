package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type teamLobbyRepositoryStub struct {
	joined     bool
	locked     bool
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
	if s.team != nil {
		return s.team, nil
	}
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
	if s.tournament != nil {
		team.Tournament = *s.tournament
	}
	return team, nil
}
func (s *teamLobbyRepositoryStub) Join(_ context.Context, _ uuid.UUID, accountID uuid.UUID, membershipID ...uuid.UUID) (*models.TournamentTeam, error) {
	s.joined = true
	if len(membershipID) > 0 {
		s.team.Members = append(s.team.Members, models.TournamentTeamMember{ID: membershipID[0], AccountID: accountID, TournamentTeamID: s.team.ID, Role: models.TournamentTeamMemberRoleMember})
	}
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) RegenerateInvite(_ context.Context, _ uuid.UUID, inviteCode string) (*models.TournamentTeam, error) {
	if s.team != nil {
		s.team.InviteCode = &inviteCode
	}
	return s.team, nil
}
func (s *teamLobbyRepositoryStub) Lock(_ context.Context, _ uuid.UUID, _ ...[]models.TournamentTeamMember) (*models.TournamentTeam, error) {
	s.locked = true
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
func (s *teamLobbyRepositoryStub) ApplyPayment(_ context.Context, _ *models.TournamentTeam, _ uuid.UUID, status models.PaymentStatus) error {
	s.applied = string(status)
	return nil
}

type registrationFormRepositoryStub struct {
	responses   map[uuid.UUID]*models.FormResponse
	batchCalls  int
	batchError  error
	upsertError error

	form     *models.RegistrationForm
	response *models.FormResponse
	upserted *models.FormResponse
}

func (s *registrationFormRepositoryStub) EnsureIndexes(_ context.Context) error { return nil }
func (s *registrationFormRepositoryStub) GetFormByTournamentID(_ context.Context, _ uuid.UUID) (*models.RegistrationForm, error) {
	if s.form == nil {
		return nil, repositories.ErrFormNotFound
	}
	return s.form, nil
}
func (s *registrationFormRepositoryStub) UpsertForm(_ context.Context, form *models.RegistrationForm) (*models.RegistrationForm, error) {
	s.form = form
	return form, nil
}
func (s *registrationFormRepositoryStub) UpsertResponse(_ context.Context, resp *models.FormResponse) (*models.FormResponse, error) {
	if s.upsertError != nil {
		return nil, s.upsertError
	}
	if s.responses == nil {
		s.responses = make(map[uuid.UUID]*models.FormResponse)
	}

	s.responses[resp.AccountID] = resp
	s.upserted = resp
	s.response = resp
	return resp, nil
}
func (s *registrationFormRepositoryStub) GetResponseByAccount(_ context.Context, _ uuid.UUID, accountID uuid.UUID) (*models.FormResponse, error) {
	if response, exists := s.responses[accountID]; exists {
		return response, nil
	}
	return nil, repositories.ErrResponseNotFound
}
func (s *registrationFormRepositoryStub) ListResponsesByTournamentID(_ context.Context, _ uuid.UUID) ([]models.FormResponse, error) {
	return nil, nil
}

func newTestTeamLobbyUseCase(
	repo repositories.TeamLobbyRepository,
	accountSvc ports.AccountService,
	payments ...ports.PaymentService,
) *usecases.TeamLobbyUseCase {
	var p ports.PaymentService
	if len(payments) > 0 {
		p = payments[0]
	}
	return usecases.NewTeamLobbyUseCase(repo, accountSvc, p, nil, nil)
}

func TestCreateTeamLobby_InitializesCaptainAndInvite(t *testing.T) {
	tournamentID := uuid.New()
	captainID := uuid.New()
	repo := &teamLobbyRepositoryStub{tournament: &models.Tournament{
		ID: tournamentID, Status: models.TournamentStatusRegistration, RegistrationDeadline: time.Now().Add(time.Hour),
		RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 5, MaxTeamSize: 5,
	}}
	useCase := newTestTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

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
			name: "closed registration", tournament: &models.Tournament{ID: tournamentID, Status: models.TournamentStatusRegistration, RegistrationMode: models.TournamentRegistrationModeTeam, RegistrationDeadline: time.Now().Add(-time.Hour)}, want: usecases.ErrRegistrationNotOpen,
		},
		{
			name: "crowdfunding registration", tournament: &models.Tournament{ID: tournamentID, Status: models.TournamentStatusCrowdfunding, RegistrationMode: models.TournamentRegistrationModeTeam}, want: usecases.ErrRegistrationNotOpen,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := newTestTeamLobbyUseCase(&teamLobbyRepositoryStub{tournament: test.tournament}, newStubAccountService().withAccount(captainID, "captain", "Captain"))
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
		Status:               models.TournamentStatusRegistration,
		RegistrationDeadline: time.Now().Add(time.Hour),
		MinTeamSize:      1,
		MaxTeamSize:      1,
		RegistrationMode: models.TournamentRegistrationModeSolo,
	}
	repo := &teamLobbyRepositoryStub{tournament: tournament}
	accounts := newStubAccountService().withAccount(playerID, "playerone", "Player One")
	useCase := newTestTeamLobbyUseCase(repo, accounts)

	got, err := useCase.RegisterSolo(context.Background(), tournamentID, playerID, nil)
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

func TestRegisterSolo_WithFormValidation(t *testing.T) {
	tournamentID := uuid.New()
	playerID := uuid.New()
	tournament := &models.Tournament{
		ID:               tournamentID,
		Status:               models.TournamentStatusRegistration,
		RegistrationDeadline: time.Now().Add(time.Hour),
		MinTeamSize:      1,
		MaxTeamSize:      1,
		RegistrationMode: models.TournamentRegistrationModeSolo,
	}
	repo := &teamLobbyRepositoryStub{tournament: tournament}
	accounts := newStubAccountService().withAccount(playerID, "playerone", "Player One")
	formRepo := &registrationFormRepositoryStub{
		form: &models.RegistrationForm{
			TournamentID: tournamentID,
			Title:        "Survey",
			Fields: []models.FormField{
				{
					ID:       "discord_id",
					Type:     models.FieldTypeShortText,
					Label:    "Discord ID",
					Required: true,
				},
			},
		},
	}
	registry := forms.NewDefaultRegistry()
	useCase := usecases.NewTeamLobbyUseCase(repo, accounts, nil, formRepo, registry)

	// Missing required answer
	_, err := useCase.RegisterSolo(context.Background(), tournamentID, playerID, nil)
	if !errors.Is(err, forms.ErrFieldRequired) {
		t.Fatalf("RegisterSolo() expected ErrFieldRequired, got %v", err)
	}

	// Valid answer
	answers := []models.FieldAnswer{{FieldID: "discord_id", Value: "gamer#1234"}}
	got, err := useCase.RegisterSolo(context.Background(), tournamentID, playerID, answers)
	if err != nil {
		t.Fatalf("RegisterSolo() with answers error = %v", err)
	}
	if got.Team == nil {
		t.Fatal("RegisterSolo() returned nil team")
	}
	if formRepo.upserted == nil || len(formRepo.upserted.Answers) != 1 {
		t.Fatalf("expected answers to be upserted in formRepo, got %+v", formRepo.upserted)
	}
}

func TestRegisterSolo_RejectsNonSoloTournament(t *testing.T) {
	tournamentID := uuid.New()
	playerID := uuid.New()
	tournament := &models.Tournament{
		ID:               tournamentID,
		Status:               models.TournamentStatusRegistration,
		RegistrationDeadline: time.Now().Add(time.Hour),
		MinTeamSize:      5,
		MaxTeamSize:      5,
		RegistrationMode: models.TournamentRegistrationModeTeam,
	}
	repo := &teamLobbyRepositoryStub{tournament: tournament}
	accounts := newStubAccountService().withAccount(playerID, "playerone", "Player One")
	useCase := newTestTeamLobbyUseCase(repo, accounts)

	_, err := useCase.RegisterSolo(context.Background(), tournamentID, playerID, nil)
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
	useCase := newTestTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

	_, err := useCase.Lock(context.Background(), repo.team.ID, uuid.New())
	if !errors.Is(err, usecases.ErrLobbyAccessDenied) {
		t.Fatalf("Lock() error = %v, want captain access denial", err)
	}
}

func TestGetActiveForTournament_ReturnsActiveTeam(t *testing.T) {
	team := &models.TournamentTeam{ID: uuid.New(), Name: "Night Owls"}
	accountID := uuid.New()
	useCase := newTestTeamLobbyUseCase(&teamLobbyRepositoryStub{activeTeam: team}, newStubAccountService().withAccount(accountID, "player", "Player"))

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
	useCase := newTestTeamLobbyUseCase(repo, newStubAccountService().withAccount(uuid.New(), "user", "User"))

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
	useCase := newTestTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

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
	useCase := newTestTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

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
	useCase := newTestTeamLobbyUseCase(repo, accounts)

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
	useCase := newTestTeamLobbyUseCase(repo, newStubAccountService().withAccount(captainID, "captain", "Captain"))

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

func (p registrationPaymentStub) GetPayment(_ context.Context, id uuid.UUID) (*ports.RegistrationPayment, error) {
	return &ports.RegistrationPayment{ID: id, Status: models.PaymentPaid}, p.err
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
			uc := newTestTeamLobbyUseCase(repo, nil, registrationPaymentStub{tc.mutate, tc.err})
			if err := uc.Reconcile(context.Background(), team); !errors.Is(err, usecases.ErrPaymentUnavailable) {
				t.Fatalf("error = %v", err)
			}
			if repo.applied != "" {
				t.Fatal("unverified payment changed roster")
			}
			// A later retry of the same durable attempt can complete after an outage.
			uc = newTestTeamLobbyUseCase(repo, nil, registrationPaymentStub{})
			if err := uc.Reconcile(context.Background(), team); err != nil {
				t.Fatal(err)
			}
			if repo.applied != "PAID" {
				t.Fatal("paid attempt was not submitted")
			}
		})
	}
}

func TestReconcile_ClearedAttemptWaitsForExplicitRetry(t *testing.T) {
	team := &models.TournamentTeam{ID: uuid.New(), Status: models.TournamentTeamStatusPaymentPending}
	u := newTestTeamLobbyUseCase(&teamLobbyRepositoryStub{team: team}, nil,
		registrationPaymentStub{err: errors.New("invoice must not be requested before retry")})
	if err := u.Reconcile(context.Background(), team); err != nil {
		t.Fatalf("cleared attempt should remain pending: %v", err)
	}
	if team.Status != models.TournamentTeamStatusPaymentPending {
		t.Fatal("refresh must not unfreeze the roster")
	}
	// An inconsistent partial attempt must still fail closed.
	paymentID := uuid.New()
	team.PaymentID = &paymentID
	if err := u.Reconcile(context.Background(), team); !errors.Is(err, usecases.ErrPaymentUnavailable) {
		t.Fatalf("incomplete payment metadata: %v", err)
	}
}

func TestApplyPaymentOutcome(t *testing.T) {
	attempt, payer := uuid.New(), uuid.New()
	expires := time.Now().Add(time.Minute).UTC().Truncate(time.Microsecond)
	pending := func() *models.TournamentTeam {
		return &models.TournamentTeam{ID: uuid.New(), Status: models.TournamentTeamStatusPaymentPending, PaymentAttemptID: &attempt, PaymentPayerID: &payer, PaymentExpiresAt: &expires, PaymentAmountMinor: 10000}
	}
	outcome := func(team *models.TournamentTeam, status models.PaymentStatus) usecases.PaymentOutcome {
		return usecases.PaymentOutcome{Status: status, PaymentID: uuid.New(), PayerAccountID: payer, Reference: team.ID.String() + ":" + attempt.String(), AmountMinor: 10000, Currency: "THB", ExpiresAt: expires}
	}
	for _, tc := range []struct {
		name        string
		team        *models.TournamentTeam // nil: no such team
		change      func(*usecases.PaymentOutcome)
		wantApplied string
		wantErr     error
	}{
		{"paid locks the roster", pending(), nil, "PAID", nil},
		{"expired clears the attempt", pending(), func(o *usecases.PaymentOutcome) { o.Status = models.PaymentExpired }, "EXPIRED", nil},
		{"cancelled clears the attempt", pending(), func(o *usecases.PaymentOutcome) { o.Status = models.PaymentCancelled }, "CANCELLED", nil},
		{"deleted team is ignored", nil, nil, "", nil},
		{"settled team is left alone", &models.TournamentTeam{ID: uuid.New(), Status: models.TournamentTeamStatusLocked}, nil, "", nil},
		{"event for an older attempt is ignored", pending(), func(o *usecases.PaymentOutcome) {
			o.Reference = strings.Split(o.Reference, ":")[0] + ":" + uuid.NewString()
		}, "", nil},
		{"malformed reference is rejected", pending(), func(o *usecases.PaymentOutcome) { o.Reference = "not-a-team" }, "", usecases.ErrPaymentMismatch},
		{"wrong payer is rejected", pending(), func(o *usecases.PaymentOutcome) { o.PayerAccountID = uuid.New() }, "", usecases.ErrPaymentMismatch},
		{"wrong amount is rejected", pending(), func(o *usecases.PaymentOutcome) { o.AmountMinor++ }, "", usecases.ErrPaymentMismatch},
		{"wrong currency is rejected", pending(), func(o *usecases.PaymentOutcome) { o.Currency = "USD" }, "", usecases.ErrPaymentMismatch},
		{"wrong deadline is rejected", pending(), func(o *usecases.PaymentOutcome) { o.ExpiresAt = o.ExpiresAt.Add(time.Second) }, "", usecases.ErrPaymentMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			team := tc.team
			if team == nil {
				team = pending()
			}
			o := outcome(team, models.PaymentPaid)
			if tc.change != nil {
				tc.change(&o)
			}
			repo := &teamLobbyRepositoryStub{team: tc.team}
			err := newTestTeamLobbyUseCase(repo, nil, registrationPaymentStub{}).ApplyPaymentOutcome(context.Background(), o)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if repo.applied != tc.wantApplied {
				t.Fatalf("applied = %q, want %q", repo.applied, tc.wantApplied)
			}
		})
	}
}

func (s *teamLobbyRepositoryStub) LockRegistration(_ context.Context, _ uuid.UUID) (repositories.TeamRegistrationLock, error) {
	if s.team == nil {
		return nil, repositories.ErrTeamLobbyNotFound
	}
	return &registrationLockStub{team: s.team}, nil
}

type registrationLockStub struct{ team *models.TournamentTeam }

func (s *registrationLockStub) Team() *models.TournamentTeam { return s.team }
func (s *registrationLockStub) Release() error               { return nil }

func (s *registrationFormRepositoryStub) GetResponsesByAccounts(_ context.Context, tournamentID uuid.UUID, accountIDs []uuid.UUID) ([]models.FormResponse, error) {
	s.batchCalls++
	if s.batchError != nil {
		return nil, s.batchError
	}
	responses := make([]models.FormResponse, 0, len(accountIDs))
	for _, id := range accountIDs {
		if r := s.responses[id]; r != nil && r.TournamentID == tournamentID {
			responses = append(responses, *r)
		}
	}
	return responses, nil
}

func TestLock_PublishesRegistrationSubmitted_ForFreeTournament(t *testing.T) {
	tournamentID := uuid.New()
	captainID := uuid.New()
	teamID := uuid.New()
	tournament := &models.Tournament{
		ID:                   tournamentID,
		Name:                 "Championship 2026",
		Status:               models.TournamentStatusRegistration,
		RegistrationDeadline: time.Now().Add(time.Hour),
		RegistrationMode:     models.TournamentRegistrationModeTeam,
		MinTeamSize:          1,
		MaxTeamSize:          5,
		EntryFeeMinor:        0,
	}
	team := &models.TournamentTeam{
		ID:           teamID,
		TournamentID: tournamentID,
		Tournament:   *tournament,
		Name:         "Falcon Squad",
		Status:       models.TournamentTeamStatusForming,
		Members: []models.TournamentTeamMember{{
			ID:               uuid.New(),
			TournamentTeamID: teamID,
			AccountID:        captainID,
			Role:             models.TournamentTeamMemberRoleCaptain,
		}},
	}
	repo := &teamLobbyRepositoryStub{
		tournament: tournament,
		team:       team,
	}
	accountSvc := newStubAccountService().withAccountEmail(captainID, "captain", "Falcon Captain", "captain@example.com")
	pub := &mockPublisher{}

	uc := usecases.NewTeamLobbyUseCase(repo, accountSvc, nil, nil, nil, pub)

	view, err := uc.Lock(context.Background(), teamID, captainID)
	if err != nil {
		t.Fatalf("Lock() failed: %v", err)
	}
	if view == nil || view.Team.Status != models.TournamentTeamStatusLocked {
		t.Fatalf("expected locked team view, got: %+v", view)
	}

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(pub.events))
	}
	if pub.events[0].routingKey != events.RegistrationSubmittedType {
		t.Fatalf("expected routing key %q, got %q", events.RegistrationSubmittedType, pub.events[0].routingKey)
	}
	evt, ok := pub.events[0].event.(events.RegistrationSubmitted)
	if !ok {
		t.Fatalf("expected events.RegistrationSubmitted, got %T", pub.events[0].event)
	}
	if evt.Data.RegistrationID != teamID {
		t.Errorf("expected RegistrationID %s, got %s", teamID, evt.Data.RegistrationID)
	}
	if evt.Data.TournamentID != tournamentID {
		t.Errorf("expected TournamentID %s, got %s", tournamentID, evt.Data.TournamentID)
	}
	if evt.Data.TournamentName != "Championship 2026" {
		t.Errorf("expected TournamentName %q, got %q", "Championship 2026", evt.Data.TournamentName)
	}
	if evt.Data.RegistrationMode != events.RegistrationModeTeam {
		t.Errorf("expected RegistrationMode %q, got %q", events.RegistrationModeTeam, evt.Data.RegistrationMode)
	}
	if evt.Data.TeamID == nil || *evt.Data.TeamID != teamID {
		t.Errorf("expected TeamID %s, got %v", teamID, evt.Data.TeamID)
	}
	if evt.Data.TeamName == nil || *evt.Data.TeamName != "Falcon Squad" {
		t.Errorf("expected TeamName %q, got %v", "Falcon Squad", evt.Data.TeamName)
	}
	if evt.Data.Recipient.AccountID != captainID {
		t.Errorf("expected recipient AccountID %s, got %s", captainID, evt.Data.Recipient.AccountID)
	}
	if evt.Data.Recipient.Role != events.RecipientRoleCaptain {
		t.Errorf("expected recipient role %q, got %q", events.RecipientRoleCaptain, evt.Data.Recipient.Role)
	}
	if evt.Data.Recipient.Email != "captain@example.com" {
		t.Errorf("expected recipient email %q, got %q", "captain@example.com", evt.Data.Recipient.Email)
	}
	if evt.Data.Recipient.DisplayName != "Falcon Captain" {
		t.Errorf("expected recipient DisplayName %q, got %q", "Falcon Captain", evt.Data.Recipient.DisplayName)
	}
}

func TestRegisterSolo_PublishesRegistrationSubmitted(t *testing.T) {
	tournamentID := uuid.New()
	playerID := uuid.New()
	tournament := &models.Tournament{
		ID:                   tournamentID,
		Name:                 "Solo Clash",
		Status:               models.TournamentStatusRegistration,
		RegistrationDeadline: time.Now().Add(time.Hour),
		RegistrationMode:     models.TournamentRegistrationModeSolo,
		MinTeamSize:          1,
		MaxTeamSize:          1,
		EntryFeeMinor:        0,
	}
	repo := &teamLobbyRepositoryStub{
		tournament: tournament,
	}
	accountSvc := newStubAccountService().withAccountEmail(playerID, "soloplayer", "Solo Player", "solo@example.com")
	pub := &mockPublisher{}

	uc := usecases.NewTeamLobbyUseCase(repo, accountSvc, nil, nil, nil, pub)

	view, err := uc.RegisterSolo(context.Background(), tournamentID, playerID, nil)
	if err != nil {
		t.Fatalf("RegisterSolo() failed: %v", err)
	}
	if view == nil || view.Team.Status != models.TournamentTeamStatusLocked {
		t.Fatalf("expected locked solo team view, got: %+v", view)
	}

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(pub.events))
	}
	if pub.events[0].routingKey != events.RegistrationSubmittedType {
		t.Fatalf("expected routing key %q, got %q", events.RegistrationSubmittedType, pub.events[0].routingKey)
	}
	evt, ok := pub.events[0].event.(events.RegistrationSubmitted)
	if !ok {
		t.Fatalf("expected events.RegistrationSubmitted, got %T", pub.events[0].event)
	}
	if evt.Data.RegistrationMode != events.RegistrationModeSolo {
		t.Errorf("expected RegistrationMode %q, got %q", events.RegistrationModeSolo, evt.Data.RegistrationMode)
	}
	if evt.Data.Recipient.Role != events.RecipientRoleApplicant {
		t.Errorf("expected recipient role %q, got %q", events.RecipientRoleApplicant, evt.Data.Recipient.Role)
	}
	if evt.Data.Recipient.Email != "solo@example.com" {
		t.Errorf("expected email %q, got %q", "solo@example.com", evt.Data.Recipient.Email)
	}
	if evt.Data.TeamID != nil {
		t.Errorf("expected nil TeamID for solo, got %v", evt.Data.TeamID)
	}
	if evt.Data.TeamName != nil {
		t.Errorf("expected nil TeamName for solo, got %v", evt.Data.TeamName)
	}
}

func TestReconcile_PublishesRegistrationSubmitted_WhenPaid(t *testing.T) {
	tournamentID := uuid.New()
	teamID := uuid.New()
	payerID := uuid.New()
	attemptID := uuid.New()
	expires := time.Now().Add(time.Minute)

	tournament := models.Tournament{
		ID:                   tournamentID,
		Name:                 "Pro League",
		Status:               models.TournamentStatusRegistration,
		RegistrationDeadline: time.Now().Add(time.Hour),
		RegistrationMode:     models.TournamentRegistrationModeTeam,
		MinTeamSize:          2,
		MaxTeamSize:          5,
	}
	team := &models.TournamentTeam{
		ID:                  teamID,
		TournamentID:        tournamentID,
		Tournament:          tournament,
		Name:                "Shadow Hunters",
		Status:              models.TournamentTeamStatusPaymentPending,
		PaymentAttemptID:    &attemptID,
		PaymentPayerID:      &payerID,
		PaymentExpiresAt:    &expires,
		PaymentAmountMinor:  50000,
		Members: []models.TournamentTeamMember{{
			ID:               uuid.New(),
			TournamentTeamID: teamID,
			AccountID:        payerID,
			Role:             models.TournamentTeamMemberRoleCaptain,
		}},
	}
	repo := &teamLobbyRepositoryStub{
		tournament: &tournament,
		team:       team,
	}
	accountSvc := newStubAccountService().withAccountEmail(payerID, "shadow_lead", "Shadow Lead", "lead@shadow.org")
	pub := &mockPublisher{}

	uc := usecases.NewTeamLobbyUseCase(repo, accountSvc, registrationPaymentStub{}, nil, nil, pub)

	if err := uc.Reconcile(context.Background(), team); err != nil {
		t.Fatalf("Reconcile() failed: %v", err)
	}

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(pub.events))
	}
	evt, ok := pub.events[0].event.(events.RegistrationSubmitted)
	if !ok {
		t.Fatalf("expected events.RegistrationSubmitted, got %T", pub.events[0].event)
	}
	if evt.Data.TournamentName != "Pro League" {
		t.Errorf("expected TournamentName %q, got %q", "Pro League", evt.Data.TournamentName)
	}
	if evt.Data.Recipient.Email != "lead@shadow.org" {
		t.Errorf("expected email %q, got %q", "lead@shadow.org", evt.Data.Recipient.Email)
	}
}

func TestApplyPaymentOutcome_PublishesRegistrationSubmitted_WhenPaid(t *testing.T) {
	tournamentID := uuid.New()
	teamID := uuid.New()
	payerID := uuid.New()
	attemptID := uuid.New()
	expires := time.Now().Add(time.Minute).UTC().Truncate(time.Microsecond)

	tournament := models.Tournament{
		ID:                   tournamentID,
		Name:                 "Masters Cup",
		Status:               models.TournamentStatusRegistration,
		RegistrationDeadline: time.Now().Add(time.Hour),
		RegistrationMode:     models.TournamentRegistrationModeTeam,
		MinTeamSize:          2,
		MaxTeamSize:          5,
	}
	pendingTeam := func() *models.TournamentTeam {
		return &models.TournamentTeam{
			ID:                  teamID,
			TournamentID:        tournamentID,
			Tournament:          tournament,
			Name:                "Titans",
			Status:              models.TournamentTeamStatusPaymentPending,
			PaymentAttemptID:    &attemptID,
			PaymentPayerID:      &payerID,
			PaymentExpiresAt:    &expires,
			PaymentAmountMinor:  10000,
			Members: []models.TournamentTeamMember{{
				ID:               uuid.New(),
				TournamentTeamID: teamID,
				AccountID:        payerID,
				Role:             models.TournamentTeamMemberRoleCaptain,
			}},
		}
	}

	repo := &teamLobbyRepositoryStub{
		tournament: &tournament,
		team:       pendingTeam(),
	}
	accountSvc := newStubAccountService().withAccountEmail(payerID, "titan_cap", "Titan Captain", "cap@titans.com")
	pub := &mockPublisher{}

	uc := usecases.NewTeamLobbyUseCase(repo, accountSvc, nil, nil, nil, pub)

	paidOutcome := usecases.PaymentOutcome{
		Status:         models.PaymentPaid,
		PaymentID:      uuid.New(),
		PayerAccountID: payerID,
		Reference:      teamID.String() + ":" + attemptID.String(),
		AmountMinor:    10000,
		Currency:       "THB",
		ExpiresAt:      expires,
	}

	if err := uc.ApplyPaymentOutcome(context.Background(), paidOutcome); err != nil {
		t.Fatalf("ApplyPaymentOutcome() failed: %v", err)
	}

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(pub.events))
	}
	evt := pub.events[0].event.(events.RegistrationSubmitted)
	if evt.Data.RegistrationID != teamID {
		t.Errorf("expected RegistrationID %s, got %s", teamID, evt.Data.RegistrationID)
	}

	// Calling again when already locked/processed should be a no-op and not republish
	if err := uc.ApplyPaymentOutcome(context.Background(), paidOutcome); err != nil {
		t.Fatalf("second ApplyPaymentOutcome() failed: %v", err)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected still 1 published event after idempotent call, got %d", len(pub.events))
	}
}
