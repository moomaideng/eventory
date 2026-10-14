package usecases

import (
	"context"
	"errors"
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

// Team and solo applications share one set of review rules. ACCEPTED and
// REJECTED are spelled the same for both kinds; the "waiting for the
// organizer" status is not: LOCKED for teams, AWAITING_APPROVAL for solo
// players. The scenarios in tests/application_review cover both kinds, so a
// spelling drift would be caught there.
var (
	statusAccepted         = string(models.TournamentTeamStatusAccepted)
	statusRejected         = string(models.TournamentTeamStatusRejected)
	awaitingReviewStatuses = []string{
		string(models.TournamentTeamStatusLocked),
		string(models.SoloAwaitingApproval),
	}
)

// ReviewResult is what one decision looks like to callers, whichever kind of
// application it was made on.
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

// TeamReviewUseCase reviews team applications and solo applications alike.
// The name is historical: it was written for teams first.
type TeamReviewUseCase struct {
	teamRepo   repositories.TeamReviewRepository
	soloRepo   repositories.SoloReviewRepository
	accountSvc ports.AccountService
}

func NewTeamReviewUseCase(
	teamRepo repositories.TeamReviewRepository,
	soloRepo repositories.SoloReviewRepository,
	accountSvc ports.AccountService,
) *TeamReviewUseCase {
	return &TeamReviewUseCase{teamRepo: teamRepo, soloRepo: soloRepo, accountSvc: accountSvc}
}

// Review accepts or rejects an application (a team's or a solo player's),
// recording who decided and why.
//
// The entry is looked up as a team first and as a solo registration if no
// team has that ID. It deliberately does not depend on the tournament's
// registration mode, which an organizer can still change after solo players
// have registered.
//
// A decision can be switched later (ACCEPTED <-> REJECTED); every decision is
// kept in the audit trail. A paid applicant who gets rejected is NOT refunded
// here: refunds belong to US7-2, and switching such an applicant back to
// ACCEPTED must be revisited once that story automates them.
//
// TODO(sprint-next): notify the applicant. No notification system exists yet.
func (u *TeamReviewUseCase) Review(
	ctx context.Context,
	accountID, tournamentID, entryID uuid.UUID,
	rawDecision, rawReason string,
) (*ReviewResult, error) {
	var to string
	switch ReviewDecision(strings.ToUpper(strings.TrimSpace(rawDecision))) {
	case ReviewDecisionAccept:
		to = statusAccepted
	case ReviewDecisionReject:
		to = statusRejected
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

	teamReview, err := u.teamRepo.ApplyReview(ctx, accountID, reviewer.DisplayName, tournamentID, entryID, reason,
		func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error) {
			if err := checkReview(tournament, string(team.Status), to, acceptedEntries); err != nil {
				return "", err
			}
			return models.TournamentTeamStatus(to), nil
		})
	if err == nil {
		return &ReviewResult{
			ID:                  teamReview.ID,
			TournamentID:        teamReview.TournamentID,
			EntryID:             teamReview.TournamentTeamID,
			FromStatus:          string(teamReview.FromStatus),
			ToStatus:            string(teamReview.ToStatus),
			Reason:              teamReview.Reason,
			ReviewerAccountID:   teamReview.ReviewerAccountID,
			ReviewerDisplayName: teamReview.ReviewerDisplayName,
			CreatedAt:           teamReview.CreatedAt,
		}, nil
	}
	if !errors.Is(err, repositories.ErrTeamEntryNotFound) {
		return nil, translateReviewError(err)
	}

	// No team has this ID in the tournament; it may be a solo registration.
	soloReview, err := u.soloRepo.ApplyReview(ctx, accountID, reviewer.DisplayName, tournamentID, entryID, reason,
		func(tournament *models.Tournament, solo *models.SoloRegistration, acceptedEntries int) (models.SoloRegistrationStatus, error) {
			if err := checkReview(tournament, string(solo.Status), to, acceptedEntries); err != nil {
				return "", err
			}
			return models.SoloRegistrationStatus(to), nil
		})
	if err != nil {
		return nil, translateReviewError(err)
	}
	return &ReviewResult{
		ID:                  soloReview.ID,
		TournamentID:        soloReview.TournamentID,
		EntryID:             soloReview.SoloRegistrationID,
		FromStatus:          string(soloReview.FromStatus),
		ToStatus:            string(soloReview.ToStatus),
		Reason:              soloReview.Reason,
		ReviewerAccountID:   soloReview.ReviewerAccountID,
		ReviewerDisplayName: soloReview.ReviewerDisplayName,
		CreatedAt:           soloReview.CreatedAt,
	}, nil
}

// checkReview holds the decision rules shared by both kinds of application.
// It runs inside the repository's transaction, while the rows are locked.
func checkReview(tournament *models.Tournament, from, to string, acceptedEntries int) error {
	if !slices.Contains(reviewableStages, tournament.Status) {
		return ErrReviewNotAllowedInStage
	}
	if !slices.Contains(awaitingReviewStatuses, from) && from != statusAccepted && from != statusRejected {
		return ErrApplicationNotReviewable
	}
	if from == to {
		return ErrApplicationAlreadyDecided
	}
	// acceptedEntries never includes this application here (its status differs
	// from the target), so a full tournament has no spot left for it.
	if to == statusAccepted && acceptedEntries >= tournament.Capacity {
		return ErrTournamentFull
	}
	return nil
}

func translateReviewError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrTournamentNotFound):
		return ErrTournamentNotFound
	case errors.Is(err, repositories.ErrTeamEntryNotFound),
		errors.Is(err, repositories.ErrSoloNotFound):
		return ErrTeamEntryNotFound
	}
	return err
}
