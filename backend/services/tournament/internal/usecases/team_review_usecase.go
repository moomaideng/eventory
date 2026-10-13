package usecases

import (
	"context"
	"errors"
	"slices"
	"strings"
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

type TeamReviewUseCase struct {
	repo       repositories.TeamReviewRepository
	accountSvc ports.AccountService
}

func NewTeamReviewUseCase(repo repositories.TeamReviewRepository, accountSvc ports.AccountService) *TeamReviewUseCase {
	return &TeamReviewUseCase{repo: repo, accountSvc: accountSvc}
}

// Review accepts or rejects a team's application, recording who decided and
// why.
//
// A decision can be switched later (ACCEPTED <-> REJECTED); every decision is
// kept in the audit trail. A paid team that gets rejected is NOT refunded here:
// refunds belong to US7-2, and switching such a team back to ACCEPTED must be
// revisited once that story automates them.
//
// TODO(sprint-next): notify the applicant. No notification system exists yet.
func (u *TeamReviewUseCase) Review(
	ctx context.Context,
	accountID, tournamentID, teamID uuid.UUID,
	rawDecision, rawReason string,
) (*models.TournamentTeamReview, error) {
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

	review, err := u.repo.ApplyReview(ctx, accountID, reviewer.DisplayName, tournamentID, teamID, reason,
		func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error) {
			if !slices.Contains(reviewableStages, tournament.Status) {
				return "", ErrReviewNotAllowedInStage
			}
			switch team.Status {
			case models.TournamentTeamStatusLocked,
				models.TournamentTeamStatusAccepted,
				models.TournamentTeamStatusRejected:
			default:
				return "", ErrApplicationNotReviewable
			}
			if team.Status == to {
				return "", ErrApplicationAlreadyDecided
			}
			// acceptedEntries never includes this team here (its status differs
			// from the target), so a full tournament has no spot left for it.
			if to == models.TournamentTeamStatusAccepted && acceptedEntries >= tournament.Capacity {
				return "", ErrTournamentFull
			}
			return to, nil
		})
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrTournamentNotFound):
			return nil, ErrTournamentNotFound
		case errors.Is(err, repositories.ErrTeamEntryNotFound):
			return nil, ErrTeamEntryNotFound
		}
		return nil, err
	}
	return review, nil
}
