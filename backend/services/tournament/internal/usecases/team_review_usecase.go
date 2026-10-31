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
}

func NewTeamReviewUseCase(
	repo repositories.TeamReviewRepository,
	accountSvc ports.AccountService,
	payments ports.RegistrationRefundService,
) *TeamReviewUseCase {
	return &TeamReviewUseCase{repo: repo, accountSvc: accountSvc, payments: payments}
}

// Review accepts or rejects an application, recording who decided and why.
//
// NOTE(design): Review is a one-time decision. Only applications waiting in
// LOCKED status can be reviewed. Once decided (ACCEPTED or REJECTED), the decision
// is final and cannot be switched.
//
// TODO(sprint-next): notify the applicant. No notification system exists yet.
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

func (u *TeamReviewUseCase) RetryRefunds(ctx context.Context) error {
	teams, err := u.repo.PendingRefunds(ctx)
	if err != nil {
		return err
	}
	for _, team := range teams {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := u.refund(ctx, team.ID, *team.PaymentID); err != nil {
			slog.WarnContext(ctx, "retry rejected registration refund", "entry_id", team.ID, "error", err)
		}
	}
	return nil
}

func (u *TeamReviewUseCase) RunRefunds(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		if err := u.RetryRefunds(ctx); err != nil && ctx.Err() == nil {
			slog.ErrorContext(ctx, "list rejected registration refunds", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
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
