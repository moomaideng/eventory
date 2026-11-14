package usecases

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type ReviewDecision string

const (
	ReviewDecisionAccept ReviewDecision = "ACCEPT"
	ReviewDecisionReject ReviewDecision = "REJECT"
)

var (
	ErrUnknownReviewDecision     = errors.New("unknown review decision")
	ErrReviewReasonTooLong       = errors.New("reason must be at most 500 characters")
	ErrReviewNotAllowedInStage   = errors.New("applications cannot be reviewed in the tournament's current stage")
	ErrApplicationNotReviewable  = errors.New("only locked applications can be reviewed")
	ErrApplicationAlreadyDecided = errors.New("application already has that decision")
	ErrTournamentFull            = errors.New("tournament has reached its capacity")
	ErrTeamEntryNotFound         = errors.New("team entry not found")
)

const reviewReasonMaxLength = 500

// reviewableStages are the tournament statuses in which organizers may decide
// on applications. Applications arrive while registration is open, and the
// organizer may keep working through them after it closes.
var reviewableStages = []models.TournamentStatus{
	models.TournamentStatusRegistration,
}

// ReviewResult is what one decision looks like to callers.
type ReviewResult struct {
	ID                  uuid.UUID
	TournamentID        uuid.UUID
	EntryID             string
	FromStatus          string
	ToStatus            string
	Reason              string
	ReviewerAccountID   uuid.UUID
	ReviewerDisplayName string
	CreatedAt           time.Time
}

// TeamReviewUseCase reviews applications. Solo players register as
// one-person teams, so one set of rules covers both kinds of application.
type TeamReviewUseCase struct {
	repo       repositories.TeamReviewRepository
	accountSvc ports.AccountService
	payments   ports.RegistrationRefundService
	publisher  events.Publisher
}

func NewTeamReviewUseCase(
	repo repositories.TeamReviewRepository,
	accountSvc ports.AccountService,
	payments ports.RegistrationRefundService,
	publisher ...events.Publisher,
) *TeamReviewUseCase {
	var pub events.Publisher = events.NoopPublisher{}
	if len(publisher) > 0 && publisher[0] != nil {
		pub = publisher[0]
	}
	return &TeamReviewUseCase{
		repo:       repo,
		accountSvc: accountSvc,
		payments:   payments,
		publisher:  pub,
	}
}

// Review accepts or rejects an application, recording who decided and why.
//
// NOTE(design): Review is a one-time decision. Only applications waiting in
// LOCKED status can be reviewed. Once decided (ACCEPTED or REJECTED), the decision
// is final and cannot be switched.
func (u *TeamReviewUseCase) Review(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
	entryID string,
	rawDecision, rawReason string,
) (*ReviewResult, error) {
	var to models.TournamentTeamStatus
	switch ReviewDecision(strings.ToUpper(strings.TrimSpace(rawDecision))) {
	case ReviewDecisionAccept:
		to = models.TournamentTeamStatusAccepted
	case ReviewDecisionReject:
		to = models.TournamentTeamStatusRejected
	default:
		return nil, ErrUnknownReviewDecision
	}

	reason := strings.TrimSpace(rawReason)
	if utf8.RuneCountInString(reason) > reviewReasonMaxLength {
		return nil, ErrReviewReasonTooLong
	}

	reviewer, err := u.accountSvc.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	review, paymentID, err := u.applyReview(ctx, accountID, reviewer.DisplayName, tournamentID, entryID, to, reason)
	if err != nil {
		return nil, translateReviewError(err)
	}
	if to == models.TournamentTeamStatusRejected && paymentID != nil {
		if err := u.refund(ctx, tournamentID, entryID, *paymentID); err != nil {
			slog.WarnContext(ctx, "refund rejected registration will retry", "tournament_id", tournamentID, "entry_id", entryID, "error", err)
		}
	}

	// Determine notification recipient: team captain or solo applicant
	var recipientID uuid.UUID
	for _, member := range review.Team.Members {
		if member.Role == models.TournamentTeamMemberRoleCaptain {
			recipientID = member.AccountID
			break
		}
	}
	if recipientID == uuid.Nil && len(review.Team.Members) > 0 {
		recipientID = review.Team.Members[0].AccountID
	}

	var recipientName, recipientEmail string
	if recipientID != uuid.Nil {
		if acc, err := u.accountSvc.GetAccount(ctx, recipientID); err == nil && acc != nil {
			recipientName = acc.DisplayName
			recipientEmail = acc.Email
		}
	}

	eventType := events.RegistrationApprovedType
	if to == models.TournamentTeamStatusRejected {
		eventType = events.RegistrationRejectedType
	}

	tournName := review.Tournament.Name
	if tournName == "" {
		tournName = "Tournament"
	}
	teamName := review.Team.Name
	if teamName == "" {
		teamName = "Team"
	}

	event := events.RegistrationReviewed{
		SpecVersion:   events.EventSpecVersion,
		ID:            uuid.New(),
		Type:          eventType,
		Source:        "tournament-service",
		OccurredAt:    review.CreatedAt,
		CorrelationID: uuid.New(),
		Data: events.RegistrationReviewedData{
			TournamentID:   review.TournamentID,
			TournamentName: tournName,
			TeamID:         review.TournamentTeamID,
			TeamName:       teamName,
			Recipient: events.EventRecipient{
				AccountID:   recipientID,
				DisplayName: recipientName,
				Email:       recipientEmail,
			},
			Reason: reason,
		},
	}
	_ = u.publisher.Publish(ctx, eventType, event)

	return &ReviewResult{
		ID:                  review.ID,
		TournamentID:        review.TournamentID,
		EntryID:             review.TournamentTeamID,
		FromStatus:          string(review.FromStatus),
		ToStatus:            string(review.ToStatus),
		Reason:              review.Reason,
		ReviewerAccountID:   review.ReviewerAccountID,
		ReviewerDisplayName: review.ReviewerDisplayName,
		CreatedAt:           review.CreatedAt,
	}, nil
}

func (u *TeamReviewUseCase) applyReview(
	ctx context.Context,
	accountID uuid.UUID,
	reviewerDisplayName string,
	tournamentID uuid.UUID,
	entryID string,
	to models.TournamentTeamStatus,
	reason string,
) (*models.TournamentTeamReview, *uuid.UUID, error) {
	var review *models.TournamentTeamReview
	var paymentID *uuid.UUID

	err := u.repo.InTransaction(ctx, func(tx repositories.TeamReviewTransaction) error {
		tournament, err := tx.LockTournament(tournamentID)
		if err != nil {
			return err
		}

		staff, ok := tournament.StaffMember(accountID)
		if !ok || (!staff.IsOwner() && !staff.IsAdmin()) {
			return ErrTournamentNotFound
		}

		team, acceptedEntries, err := tx.LockTeamAndAcceptedCount(tournamentID, entryID)
		if err != nil {
			return err
		}

		if err := checkReview(tournament, team.Status, to, acceptedEntries); err != nil {
			return err
		}

		from := team.Status
		paymentID = team.PaymentID
		review = &models.TournamentTeamReview{
			TournamentID:        tournament.ID,
			Tournament:          *tournament,
			TournamentTeamID:    team.ID,
			Team:                *team,
			FromStatus:          from,
			ToStatus:            to,
			Reason:              reason,
			ReviewerAccountID:   accountID,
			ReviewerDisplayName: reviewerDisplayName,
		}
		return tx.SaveReview(team, to, review)
	})
	if err != nil {
		return nil, nil, err
	}
	return review, paymentID, nil
}

func (u *TeamReviewUseCase) refund(ctx context.Context, tournamentID uuid.UUID, entryID string, paymentID uuid.UUID) error {
	payment, err := u.payments.RefundRegistration(ctx, paymentID, tournamentID.String()+":"+entryID)
	if err != nil {
		return err
	}
	if payment == nil || payment.ID != paymentID || payment.Status != models.PaymentRefunded {
		return ErrPaymentUnavailable
	}
	return u.repo.MarkRefunded(ctx, tournamentID, entryID, paymentID)
}

func (u *TeamReviewUseCase) PendingRefunds(ctx context.Context) ([]models.TournamentTeam, error) {
	return u.repo.PendingRefunds(ctx)
}

func (u *TeamReviewUseCase) RetryRefund(ctx context.Context, team *models.TournamentTeam) error {
	if team.PaymentID == nil {
		return ErrPaymentUnavailable
	}
	return u.refund(ctx, team.TournamentID, team.ID, *team.PaymentID)
}

// checkReview holds the decision rules. It runs inside the repository's
// transaction, while the rows are locked.
//
// NOTE(design): Review is a one-time decision: only applications in LOCKED status
// can be reviewed. Once decided (ACCEPTED or REJECTED), the decision is final
// to prevent flip-flopping and maintain audit integrity.
func checkReview(
	tournament *models.Tournament,
	from, to models.TournamentTeamStatus,
	acceptedEntries int,
) error {
	if !slices.Contains(reviewableStages, tournament.Status) {
		return ErrReviewNotAllowedInStage
	}
	if from != models.TournamentTeamStatusLocked {
		return ErrApplicationNotReviewable
	}
	if from == to {
		return ErrApplicationAlreadyDecided
	}
	// acceptedEntries never includes this application here (its status differs
	// from the target), so a full tournament has no spot left for it.
	if to == models.TournamentTeamStatusAccepted && acceptedEntries >= tournament.Capacity {
		return ErrTournamentFull
	}
	return nil
}

func translateReviewError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrTournamentNotFound):
		return ErrTournamentNotFound
	case errors.Is(err, repositories.ErrTeamEntryNotFound):
		return ErrTeamEntryNotFound
	}
	return err
}
