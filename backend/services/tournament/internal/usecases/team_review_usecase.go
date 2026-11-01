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
	models.TournamentStatusRegistrationOpen,
	models.TournamentStatusRegistrationClosed,
}

// ReviewResult is what one decision looks like to callers.
type ReviewResult struct {
	ID                  uuid.UUID
	TournamentID        uuid.UUID
	EntryID             uuid.UUID
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
	accountID, tournamentID, entryID uuid.UUID,
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

	review, paymentID, err := u.repo.ApplyReview(ctx, accountID, reviewer.DisplayName, tournamentID, entryID, reason,
		func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error) {
			if err := checkReview(tournament, team.Status, to, acceptedEntries); err != nil {
				return "", err
			}
			return to, nil
		})
	if err != nil {
		return nil, translateReviewError(err)
	}
	if to == models.TournamentTeamStatusRejected && paymentID != nil {
		if err := u.refund(ctx, entryID, *paymentID); err != nil {
			slog.WarnContext(ctx, "refund rejected registration will retry", "entry_id", entryID, "error", err)
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

func (u *TeamReviewUseCase) refund(ctx context.Context, entryID, paymentID uuid.UUID) error {
	payment, err := u.payments.RefundRegistration(ctx, paymentID, entryID)
	if err != nil {
		return err
	}
	if payment == nil || payment.ID != paymentID || payment.Status != "REFUNDED" {
		return ErrPaymentUnavailable
	}
	return u.repo.MarkRefunded(ctx, entryID, paymentID)
}

func (u *TeamReviewUseCase) PendingRefunds(ctx context.Context) ([]models.TournamentTeam, error) {
	return u.repo.PendingRefunds(ctx)
}

func (u *TeamReviewUseCase) RetryRefund(ctx context.Context, team *models.TournamentTeam) error {
	if team.PaymentID == nil {
		return ErrPaymentUnavailable
	}
	return u.refund(ctx, team.ID, *team.PaymentID)
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
