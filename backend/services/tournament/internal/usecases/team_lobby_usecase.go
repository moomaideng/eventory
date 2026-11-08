package usecases

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrNotSoloTournament        = errors.New("this tournament is not configured for solo registration")
	ErrRegistrationNotOpen      = errors.New("tournament registration is not open")
	ErrInvalidTeamName          = errors.New("team name must be between 1 and 120 characters")
	ErrLobbyAccessDenied        = errors.New("only the team captain can manage this lobby")
	ErrInvalidInviteCode        = errors.New("invalid team invite code")
	ErrPaymentUnavailable       = errors.New("payment status could not be verified; roster remains locked, please retry")
	ErrInvalidRegistrationFee   = errors.New("unsupported registration fee or currency")
	ErrTeamLobbyNotFound        = errors.New("team lobby not found")
	ErrTeamLobbyNotForming      = errors.New("team lobby is not forming")
	ErrTeamLobbyFull            = errors.New("team lobby is full")
	ErrRosterBelowMinimum       = errors.New("roster does not meet the tournament minimum")
	ErrAlreadyInTournamentLobby = errors.New("account is already in a team for this tournament")
	ErrCannotRemoveCaptain      = errors.New("captain cannot be removed")
	ErrCannotDisbandLobby       = errors.New("only a forming team can be disbanded")
	ErrRosterChanged            = errors.New("roster changed; refresh and retry")
)

func translateLobbyError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrInvalidRegistrationFee):
		return ErrInvalidRegistrationFee
	case errors.Is(err, repositories.ErrTeamLobbyNotFound):
		return ErrTeamLobbyNotFound
	case errors.Is(err, repositories.ErrTeamLobbyNotForming):
		return ErrTeamLobbyNotForming
	case errors.Is(err, repositories.ErrTeamLobbyFull):
		return ErrTeamLobbyFull
	case errors.Is(err, repositories.ErrRosterBelowMinimum):
		return ErrRosterBelowMinimum
	case errors.Is(err, repositories.ErrAlreadyInTournamentLobby):
		return ErrAlreadyInTournamentLobby
	case errors.Is(err, repositories.ErrTournamentRegistrationClosed):
		return ErrRegistrationNotOpen
	case errors.Is(err, repositories.ErrCannotRemoveCaptain):
		return ErrCannotRemoveCaptain
	case errors.Is(err, repositories.ErrCannotDisbandLobby):
		return ErrCannotDisbandLobby
	case errors.Is(err, repositories.ErrRosterChanged):
		return ErrRosterChanged
	default:
		return err
	}
}

var inviteCodeRegex = regexp.MustCompile(`^[A-Z0-9]{6}$`)

func ValidateInviteCode(code string) bool {
	return inviteCodeRegex.MatchString(code)
}

type TeamMemberView struct {
	RegistrationComplete bool
	Member               models.TournamentTeamMember
	Handle               string
	DisplayName          string
}

func NewTeamMemberViews(members []models.TournamentTeamMember, accounts map[uuid.UUID]ports.AccountInfo) []TeamMemberView {
	views := make([]TeamMemberView, 0, len(members))
	for _, member := range members {
		info := accounts[member.AccountID]
		views = append(views, TeamMemberView{
			Member:      member,
			Handle:      info.Handle,
			DisplayName: info.DisplayName,
		})
	}
	return views
}

type TeamLobbyUseCase struct {
	repo       repositories.TeamLobbyRepository
	accountSvc ports.AccountService
	payments   ports.PaymentService
	formRepo   repositories.RegistrationFormRepository
	registry   *forms.FieldRegistry
	publisher  events.Publisher
}

func NewTeamLobbyUseCase(
	repo repositories.TeamLobbyRepository,
	accountSvc ports.AccountService,
	payments ports.PaymentService,
	formRepo repositories.RegistrationFormRepository,
	registry *forms.FieldRegistry,
	publisher ...events.Publisher,
) *TeamLobbyUseCase {
	var pub events.Publisher = events.NoopPublisher{}
	if len(publisher) > 0 && publisher[0] != nil {
		pub = publisher[0]
	}
	return &TeamLobbyUseCase{
		repo:       repo,
		accountSvc: accountSvc,
		payments:   payments,
		formRepo:   formRepo,
		registry:   registry,
		publisher:  pub,
	}
}

func (u *TeamLobbyUseCase) RefreshSubmission(ctx context.Context, teamID, accountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	team, err := u.captainTeam(ctx, teamID, accountID)
	if err != nil {
		return nil, err
	}
	if u.payments != nil {
		if err := u.Reconcile(ctx, team); err != nil {
			return nil, err
		}
	}
	team, err = u.repo.FindByID(ctx, teamID)
	if err != nil {
		return nil, translateLobbyError(err)
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

func (u *TeamLobbyUseCase) Create(ctx context.Context, tournamentID, captainAccountID uuid.UUID, name string, submitted ...[]models.FieldAnswer) (*TeamLobbyView, error) {
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
	if tournament.Status != models.TournamentStatusRegistration || tournament.IsRegistrationClosed(time.Now()) {
		return nil, ErrRegistrationNotOpen
	}
	answers := firstAnswers(submitted)
	teamID, membershipID := uuid.New(), uuid.New()
	if err := u.saveNewRegistrationAnswers(ctx, tournamentID, captainAccountID, teamID, membershipID, answers); err != nil {
		return nil, translateLobbyError(err)
	}

	inviteCode, err := generateInviteCode()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	team := &models.TournamentTeam{
		ID: teamID, TournamentID: tournamentID, Name: name,
		InviteCode: &inviteCode, Status: models.TournamentTeamStatusForming,
	}
	captainMember := &models.TournamentTeamMember{
		ID: membershipID, AccountID: captainAccountID,
		Role: models.TournamentTeamMemberRoleCaptain, JoinedAt: now,
	}
	created, err := u.repo.Create(ctx, team, captainMember)
	if err != nil {
		return nil, translateLobbyError(err)
	}
	return u.toView(ctx, created)
}

func (u *TeamLobbyUseCase) RegisterSolo(
	ctx context.Context,
	tournamentID, accountID uuid.UUID,
	answers []models.FieldAnswer,
) (*TeamLobbyView, error) {
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
	if tournament.Status != models.TournamentStatusRegistration || tournament.IsRegistrationClosed(time.Now()) {
		return nil, ErrRegistrationNotOpen
	}

	teamID, membershipID := uuid.New(), uuid.New()
	if err := u.saveNewRegistrationAnswers(ctx, tournamentID, accountID, teamID, membershipID, answers); err != nil {
		return nil, err
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
		ID:           teamID,
		TournamentID: tournamentID,
		Name:         name,
		InviteCode:   nil,
		Status:       models.TournamentTeamStatusForming,
	}
	captainMember := &models.TournamentTeamMember{
		ID:        membershipID,
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
		return nil, translateLobbyError(err)
	}

	locked, err := u.repo.Lock(ctx, created.ID)
	if err != nil {
		return nil, translateLobbyError(err)
	}

	if u.payments != nil && locked.Status == models.TournamentTeamStatusPaymentPending {
		if err := u.Reconcile(ctx, locked); err != nil {
			return nil, err
		}
		locked, err = u.repo.FindByID(ctx, created.ID)
		if err != nil {
			return nil, translateLobbyError(err)
		}
	} else if locked.Status == models.TournamentTeamStatusLocked {
		u.publishRegistrationSubmitted(ctx, locked)
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
		return nil, translateLobbyError(err)
	}
	return u.toView(ctx, team)
}

func (u *TeamLobbyUseCase) GetActiveForTournament(ctx context.Context, tournamentID, accountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	team, err := u.repo.FindActiveByTournamentAndAccount(ctx, tournamentID, accountID)
	if err != nil {
		return nil, translateLobbyError(err)
	}
	return u.toView(ctx, team)
}

func (u *TeamLobbyUseCase) Join(ctx context.Context, inviteCode string, memberAccountID uuid.UUID, submitted ...[]models.FieldAnswer) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, memberAccountID); err != nil {
		return nil, err
	}
	normalized := normalizeInviteCode(inviteCode)
	if !ValidateInviteCode(normalized) {
		return nil, ErrInvalidInviteCode
	}
	team, err := u.repo.FindByInviteCode(ctx, normalized)
	if err != nil {
		return nil, translateLobbyError(err)
	}
	if team.Status != models.TournamentTeamStatusForming {
		return nil, ErrTeamLobbyNotForming
	}
	if team.Tournament.Status != models.TournamentStatusRegistration || team.Tournament.IsRegistrationClosed(time.Now()) {
		return nil, ErrRegistrationNotOpen
	}
	if len(team.Members) >= team.Tournament.MaxTeamSize {
		return nil, ErrTeamLobbyFull
	}
	answers := firstAnswers(submitted)
	membershipID := uuid.New()
	if err := u.saveNewRegistrationAnswers(ctx, team.TournamentID, memberAccountID, team.ID, membershipID, answers); err != nil {
		return nil, translateLobbyError(err)
	}
	joined, err := u.repo.Join(ctx, team.ID, memberAccountID, membershipID)
	if err != nil {
		return nil, translateLobbyError(err)
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
		return nil, translateLobbyError(err)
	}
	return u.toView(ctx, team)
}

func (u *TeamLobbyUseCase) Lock(ctx context.Context, teamID, accountID uuid.UUID) (*TeamLobbyView, error) {
	if err := u.ensureActiveAccount(ctx, accountID); err != nil {
		return nil, err
	}
	team, err := u.captainTeam(ctx, teamID, accountID)
	if err != nil {
		return nil, err
	}
	wasForming := team.Status == models.TournamentTeamStatusForming
	if wasForming {
		ready, err := u.memberReadiness(ctx, team)
		if err != nil {
			return nil, err
		}
		for _, complete := range ready {
			if !complete {
				return nil, ErrMemberAnswersIncomplete
			}
		}
	}
	lockedTeam, err := u.repo.Lock(ctx, teamID, team.Members)
	if err != nil {
		return nil, translateLobbyError(err)
	}
	if wasForming && lockedTeam.Status == models.TournamentTeamStatusLocked {
		u.publishRegistrationSubmitted(ctx, lockedTeam)
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
		return nil, translateLobbyError(err)
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
	return translateLobbyError(u.repo.Disband(ctx, teamID))
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
	ready, err := u.memberReadiness(ctx, team)
	if err != nil {
		return nil, err
	}
	for i := range view.Members {
		view.Members[i].RegistrationComplete = ready[view.Members[i].Member.AccountID]
	}
	return &view, nil
}

func (u *TeamLobbyUseCase) captainTeam(ctx context.Context, teamID, accountID uuid.UUID) (*models.TournamentTeam, error) {
	team, err := u.repo.FindByID(ctx, teamID)
	if err != nil {
		return nil, translateLobbyError(err)
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

func (u *TeamLobbyUseCase) Reconcile(ctx context.Context, team *models.TournamentTeam) error {
	if team.Status != models.TournamentTeamStatusPaymentPending {
		return nil
	}
	if u.payments == nil {
		return nil
	}
	if team.PaymentAttemptID == nil || team.PaymentPayerID == nil || team.PaymentExpiresAt == nil {
		return ErrPaymentUnavailable
	}
	desc := "Team registration"
	if team.Tournament.Name != "" {
		desc = "Registration for " + team.Tournament.Name
	}
	request := ports.RegistrationPaymentRequest{
		PayerAccountID: *team.PaymentPayerID, Source: models.PaymentSourceTeamRegistration,
		Reference: team.ID.String() + ":" + team.PaymentAttemptID.String(), IdempotencyKey: team.PaymentAttemptID.String(),
		Description: desc, AmountMinor: team.PaymentAmountMinor, Currency: "THB", ExpiresAt: *team.PaymentExpiresAt,
	}
	payment, err := u.payments.EnsureInvoice(ctx, request)
	if err != nil {
		return ErrPaymentUnavailable
	}
	if payment.Source != request.Source || payment.Reference != request.Reference || !attemptMatches(team, outcomeOf(payment)) || !settleable(payment.Status) {
		return ErrPaymentUnavailable
	}
	if err := u.repo.ApplyPayment(ctx, team, payment.ID, payment.Status); err != nil {
		return err
	}
	if payment.Status == models.PaymentPaid {
		team.Status = models.TournamentTeamStatusLocked
		u.publishRegistrationSubmitted(ctx, team)
	}
	return nil
}

// ApplyPaymentOutcome records the result of a payment event on the
// registration it refers to ("<team id>:<attempt id>"). A team that is gone,
// no longer pending, or on a newer attempt is left alone, so redelivered and
// late events are harmless. An event whose terms differ from the attempt
// returns ErrPaymentMismatch.
func (u *TeamLobbyUseCase) ApplyPaymentOutcome(ctx context.Context, outcome PaymentOutcome) error {
	teamPart, attemptPart, _ := strings.Cut(outcome.Reference, ":")
	teamID, teamErr := uuid.Parse(teamPart)
	attemptID, attemptErr := uuid.Parse(attemptPart)
	if teamErr != nil || attemptErr != nil {
		return fmt.Errorf("%w: registration reference %q", ErrPaymentMismatch, outcome.Reference)
	}
	team, err := u.repo.FindByID(ctx, teamID)
	if errors.Is(err, repositories.ErrTeamLobbyNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if team.Status != models.TournamentTeamStatusPaymentPending || team.PaymentAttemptID == nil || *team.PaymentAttemptID != attemptID {
		return nil
	}
	if !attemptMatches(team, outcome) || !settleable(outcome.Status) {
		return fmt.Errorf("%w: team %s", ErrPaymentMismatch, team.ID)
	}
	if err := u.repo.ApplyPayment(ctx, team, outcome.PaymentID, outcome.Status); err != nil {
		return err
	}
	if outcome.Status == models.PaymentPaid {
		team.Status = models.TournamentTeamStatusLocked
		u.publishRegistrationSubmitted(ctx, team)
	}
	return nil
}

// publishRegistrationSubmitted emits an asynchronous domain event informing
// downstream subscribers (e.g. Notification Service) that a registration
// has been locked / submitted for review.
func (u *TeamLobbyUseCase) publishRegistrationSubmitted(ctx context.Context, team *models.TournamentTeam) {
	if u.publisher == nil || team == nil {
		return
	}

	tourn := team.Tournament
	if tourn.ID == uuid.Nil {
		t, err := u.repo.FindTournamentByID(ctx, team.TournamentID)
		if err == nil && t != nil {
			tourn = *t
		}
	}

	members := team.Members
	if len(members) == 0 {
		t, err := u.repo.FindByID(ctx, team.ID)
		if err == nil && t != nil {
			members = t.Members
			if tourn.ID == uuid.Nil {
				tourn = t.Tournament
			}
		}
	}

	isSolo := tourn.IsSolo()

	var recipientID uuid.UUID
	var recipientRole events.RecipientRole
	var regMode events.RegistrationMode
	var teamID *uuid.UUID
	var teamName *string

	if isSolo {
		regMode = events.RegistrationModeSolo
		recipientRole = events.RecipientRoleApplicant
		if len(members) > 0 {
			recipientID = members[0].AccountID
		}
	} else {
		regMode = events.RegistrationModeTeam
		recipientRole = events.RecipientRoleCaptain
		tID := team.ID
		teamID = &tID
		tName := team.Name
		if tName == "" {
			tName = "Team"
		}
		teamName = &tName

		for _, m := range members {
			if m.Role == models.TournamentTeamMemberRoleCaptain {
				recipientID = m.AccountID
				break
			}
		}
		if recipientID == uuid.Nil && len(members) > 0 {
			recipientID = members[0].AccountID
		}
	}

	var recipientDisplayName, recipientEmail string
	if recipientID != uuid.Nil && u.accountSvc != nil {
		if acc, err := u.accountSvc.GetAccount(ctx, recipientID); err == nil && acc != nil {
			recipientDisplayName = acc.DisplayName
			if recipientDisplayName == "" {
				recipientDisplayName = acc.Handle
			}
			recipientEmail = acc.Email
		}
	}

	if recipientDisplayName == "" {
		recipientDisplayName = "Participant"
	}

	tournName := tourn.Name
	if tournName == "" {
		tournName = "Tournament"
	}

	occurredAt := time.Now().UTC()
	if team.LockedAt != nil && !team.LockedAt.IsZero() {
		occurredAt = *team.LockedAt
	}

	event := events.RegistrationSubmitted{
		SpecVersion:   events.EventSpecVersion,
		ID:            uuid.New(),
		Type:          events.RegistrationSubmittedType,
		Source:        "tournament-service",
		OccurredAt:    occurredAt,
		CorrelationID: uuid.New(),
		Data: events.RegistrationSubmittedData{
			RegistrationID:   team.ID,
			TournamentID:     team.TournamentID,
			TournamentName:   tournName,
			RegistrationMode: regMode,
			TeamID:           teamID,
			TeamName:         teamName,
			Recipient: events.EventRecipient{
				AccountID:   recipientID,
				Email:       recipientEmail,
				DisplayName: recipientDisplayName,
				Role:        recipientRole,
			},
		},
	}

	event.Normalize()
	if err := event.Validate(); err != nil {
		slog.WarnContext(ctx, "skipping registration.submitted publish due to invalid event payload", "error", err, "team_id", team.ID)
		return
	}

	if err := u.publisher.Publish(ctx, events.RegistrationSubmittedType, event); err != nil {
		slog.WarnContext(ctx, "failed to publish registration.submitted event", "error", err, "team_id", team.ID)
	}
}

// attemptMatches reports whether the invoice was issued for the team's current
// payment attempt.
func attemptMatches(team *models.TournamentTeam, p PaymentOutcome) bool {
	return p.PaymentID != uuid.Nil &&
		team.PaymentPayerID != nil && p.PayerAccountID == *team.PaymentPayerID &&
		team.PaymentExpiresAt != nil && p.ExpiresAt.Equal(*team.PaymentExpiresAt) &&
		p.AmountMinor == team.PaymentAmountMinor &&
		p.Currency == "THB" &&
		(team.PaymentID == nil || *team.PaymentID == p.PaymentID)
}
