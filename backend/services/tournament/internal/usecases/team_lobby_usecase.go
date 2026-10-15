package usecases

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrNotSoloTournament     = errors.New("this tournament is not configured for solo registration")
	ErrRegistrationNotOpen   = errors.New("tournament registration is not open")
	ErrInvalidTeamName       = errors.New("team name must be between 1 and 120 characters")
	ErrLobbyAccessDenied     = errors.New("only the team captain can manage this lobby")
	ErrInvalidInviteCode     = errors.New("invalid team invite code")
)

var inviteCodeRegex = regexp.MustCompile(`^[A-Z0-9]{6}$`)

func ValidateInviteCode(code string) bool {
	return inviteCodeRegex.MatchString(code)
}

type TeamLobbyUseCase struct {
	submission *TeamSubmissionUseCase
	repo       TeamLobbyRepository
	accountSvc ports.AccountService
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

func NewTeamLobbyUseCase(repo TeamLobbyRepository, accountSvc ports.AccountService) *TeamLobbyUseCase {
	return &TeamLobbyUseCase{repo: repo, accountSvc: accountSvc}
}

func (u *TeamLobbyUseCase) WithSubmission(submission *TeamSubmissionUseCase) *TeamLobbyUseCase {
	u.submission = submission
	return u
}

func (u *TeamLobbyUseCase) RefreshSubmission(ctx context.Context, teamID, accountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	team, err := u.captainTeam(ctx, teamID, accountID)
	if err != nil {
		return nil, err
	}
	if u.submission != nil {
		if err := u.submission.Reconcile(ctx, team); err != nil {
			return nil, err
		}
	}
	team, err = u.repo.FindByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return u.toView(ctx, team)
}

type TeamLobbyView struct {
	Team          *models.TournamentTeam
	OrganizerName string
	Members       []TeamMemberView
}

func NewTeamLobbyView(team *models.TournamentTeam, organizerName string, members []TeamMemberView) TeamLobbyView {
	return TeamLobbyView{
		Team:          team,
		OrganizerName: organizerName,
		Members:       members,
	}
}

func (u *TeamLobbyUseCase) ensureActiveAccount(ctx context.Context, accountID uuid.UUID) error {
	_, err := requireActiveAccount(ctx, u.accountSvc, accountID)
	return err
}

func (u *TeamLobbyUseCase) Create(ctx context.Context, tournamentID, captainAccountID uuid.UUID, name string) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, captainAccountID); err != nil {
		return nil, err
	}
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

	inviteCode, err := generateInviteCode()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	team := &models.TournamentTeam{
		ID: uuid.New(), TournamentID: tournamentID, Name: name,
		InviteCode: &inviteCode, Status: models.TournamentTeamStatusForming,
	}
	captainMember := &models.TournamentTeamMember{
		ID: uuid.New(), AccountID: captainAccountID,
		Role: models.TournamentTeamMemberRoleCaptain, JoinedAt: now,
	}
	created, err := u.repo.Create(ctx, team, captainMember)
	if err != nil {
		return nil, err
	}
	return u.toView(ctx, created)
}

func (u *TeamLobbyUseCase) RegisterSolo(ctx context.Context, tournamentID, accountID uuid.UUID) (*TeamLobbyView, error) {
	account, err := requireActiveAccount(ctx, u.accountSvc, accountID)
	if err != nil {
		return nil, err
	}
	tournament, err := u.repo.FindTournamentByID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if tournament == nil {
		return nil, ErrTournamentNotFound
	}
	if !tournament.IsSolo() {
		return nil, ErrNotSoloTournament
	}
	if tournament.Status != models.TournamentStatusRegistrationOpen {
		return nil, ErrRegistrationNotOpen
	}

	name := strings.TrimSpace(account.DisplayName)
	if name == "" {
		name = strings.TrimSpace(account.Handle)
	}
	if name == "" {
		name = "Player " + accountID.String()[:8]
	}

	now := time.Now().UTC()
	team := &models.TournamentTeam{
		ID:           uuid.New(),
		TournamentID: tournamentID,
		Name:         name,
		InviteCode:   nil,
		Status:       models.TournamentTeamStatusForming,
	}
	captainMember := &models.TournamentTeamMember{
		ID:        uuid.New(),
		AccountID: accountID,
		Role:      models.TournamentTeamMemberRoleCaptain,
		JoinedAt:  now,
	}
	created, err := u.repo.Create(ctx, team, captainMember)
	if err != nil && account.Handle != "" && account.Handle != name {
		team.Name = name + " (@" + account.Handle + ")"
		created, err = u.repo.Create(ctx, team, captainMember)
	}
	if err != nil {
		return nil, err
	}

	locked, err := u.repo.Lock(ctx, created.ID)
	if err != nil {
		return nil, err
	}

	if u.submission != nil && locked.Status == models.TournamentTeamStatusPaymentPending {
		if err := u.submission.Reconcile(ctx, locked); err != nil {
			return nil, err
		}
		locked, err = u.repo.FindByID(ctx, created.ID)
		if err != nil {
			return nil, err
		}
	}

	return u.toView(ctx, locked)
}

func (u *TeamLobbyUseCase) GetByInviteCode(ctx context.Context, inviteCode string, viewerAccountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, viewerAccountID); err != nil {
		return nil, err
	}
	normalized := normalizeInviteCode(inviteCode)
	if !ValidateInviteCode(normalized) {
		return nil, ErrInvalidInviteCode
	}
	team, err := u.repo.FindByInviteCode(ctx, normalized)
	if err != nil {
		return nil, err
	}
	return u.toView(ctx, team)
}

func (u *TeamLobbyUseCase) GetActiveForTournament(ctx context.Context, tournamentID, accountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	team, err := u.repo.FindActiveByTournamentAndAccount(ctx, tournamentID, accountID)
	if err != nil {
		return nil, err
	}
	return u.toView(ctx, team)
}

func (u *TeamLobbyUseCase) Join(ctx context.Context, inviteCode string, memberAccountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, memberAccountID); err != nil {
		return nil, err
	}
	normalized := normalizeInviteCode(inviteCode)
	if !ValidateInviteCode(normalized) {
		return nil, ErrInvalidInviteCode
	}
	team, err := u.repo.FindByInviteCode(ctx, normalized)
	if err != nil {
		return nil, err
	}
	joined, err := u.repo.Join(ctx, team.ID, memberAccountID)
	if err != nil {
		return nil, err
	}
	return u.toView(ctx, joined)
}

func (u *TeamLobbyUseCase) RegenerateInvite(ctx context.Context, teamID, accountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	if _, err := u.captainTeam(ctx, teamID, accountID); err != nil {
		return nil, err
	}
	inviteCode, err := generateInviteCode()
	if err != nil {
		return nil, err
	}
	team, err := u.repo.RegenerateInvite(ctx, teamID, inviteCode)
	if err != nil {
		return nil, err
	}
	return u.toView(ctx, team)
}

func (u *TeamLobbyUseCase) Lock(ctx context.Context, teamID, accountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	if _, err := u.captainTeam(ctx, teamID, accountID); err != nil {
		return nil, err
	}
	_, err := u.repo.Lock(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return u.RefreshSubmission(ctx, teamID, accountID)
}

func (u *TeamLobbyUseCase) RemoveMember(ctx context.Context, teamID, memberID, accountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	if _, err := u.captainTeam(ctx, teamID, accountID); err != nil {
		return nil, err
	}
	team, err := u.repo.RemoveMember(ctx, teamID, memberID)
	if err != nil {
		return nil, err
	}
	return u.toView(ctx, team)
}

func (u *TeamLobbyUseCase) Disband(ctx context.Context, teamID, accountID uuid.UUID) error {
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return err
	}
	if _, err := u.captainTeam(ctx, teamID, accountID); err != nil {
		return err
	}
	return u.repo.Disband(ctx, teamID)
}

func (u *TeamLobbyUseCase) toView(ctx context.Context, team *models.TournamentTeam) (*TeamLobbyView, error) {
	names, err := organizerNames(ctx, u.accountSvc, team.Tournament)
	if err != nil {
		return nil, err
	}
	accounts, err := accountInfos(ctx, u.accountSvc, team.Members...)
	if err != nil {
		return nil, err
	}
	view := NewTeamLobbyView(team, names[team.Tournament.ID], NewTeamMemberViews(team.Members, accounts))
	return &view, nil
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
