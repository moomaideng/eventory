package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
)

type TeamReviewInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
	TeamID       uuid.UUID `path:"teamId" doc:"Team (application) ID"`
	Body         struct {
		Decision string `json:"decision" enum:"ACCEPT,REJECT" doc:"Accept or reject the application"`
		Reason   string `json:"reason,omitempty" maxLength:"500" doc:"Optional justification, mainly for rejections; recorded in the audit trail"`
	}
}

type TeamReviewResponse struct {
	ID           uuid.UUID `json:"id"`
	FromStatus   string    `json:"fromStatus" enum:"LOCKED,ACCEPTED,REJECTED"`
	ToStatus     string    `json:"toStatus" enum:"ACCEPTED,REJECTED"`
	Reason       string    `json:"reason"`
	ReviewerID   uuid.UUID `json:"reviewerId"`
	ReviewerName string    `json:"reviewerName,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type TeamReviewOutput struct {
	Body struct {
		TournamentID uuid.UUID          `json:"tournamentId"`
		TeamID       uuid.UUID          `json:"teamId"`
		Status       string             `json:"status" enum:"ACCEPTED,REJECTED"`
		Review       TeamReviewResponse `json:"review"`
	}
}

func RegisterTeamReviewRoutes(api huma.API, reviews *usecases.TeamReviewUseCase) {
	huma.Register(api, huma.Operation{
		OperationID:   "review-team-application",
		Method:        http.MethodPost,
		Path:          "/tournaments/{id}/entries/{teamId}/review",
		DefaultStatus: http.StatusOK,
		Summary:       "Accept or reject a team's application",
		Description: "Records an organizer's decision on a locked team application, " +
			"enforcing the tournament's capacity. A decision may be switched later; " +
			"every decision is kept in the audit trail.",
		Tags:     []string{"Organizer dashboard"},
		Security: []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *TeamReviewInput) (*TeamReviewOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		review, err := reviews.Review(ctx, accountID, input.TournamentID, input.TeamID,
			input.Body.Decision, input.Body.Reason)
		if err != nil {
			return nil, teamReviewHTTPError(err)
		}

		output := &TeamReviewOutput{}
		output.Body.TournamentID = review.TournamentID
		output.Body.TeamID = review.TournamentTeamID
		output.Body.Status = string(review.ToStatus)
		output.Body.Review = TeamReviewResponse{
			ID:           review.ID,
			FromStatus:   string(review.FromStatus),
			ToStatus:     string(review.ToStatus),
			Reason:       review.Reason,
			ReviewerID:   review.ReviewerAccountID,
			ReviewerName: review.ReviewerDisplayName,
			CreatedAt:    review.CreatedAt,
		}
		return output, nil
	})
}

func teamReviewHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, usecases.ErrTournamentNotFound):
		// Ownership failures are reported as 404, never 403, so a tournament
		// the caller cannot manage looks identical to one that does not exist.
		return huma.Error404NotFound("Tournament not found or not managed by your account")
	case errors.Is(err, usecases.ErrTeamEntryNotFound):
		return huma.Error404NotFound("Application not found in this tournament")
	case errors.Is(err, usecases.ErrUnknownReviewDecision):
		return huma.Error400BadRequest("Decision must be ACCEPT or REJECT")
	case errors.Is(err, usecases.ErrReviewReasonTooLong):
		return huma.Error422UnprocessableEntity("Reason must be at most 500 characters")
	case errors.Is(err, usecases.ErrReviewNotAllowedInStage):
		return huma.Error409Conflict("Applications cannot be reviewed in the tournament's current stage")
	case errors.Is(err, usecases.ErrApplicationNotReviewable):
		return huma.Error409Conflict("Only locked applications can be reviewed")
	case errors.Is(err, usecases.ErrApplicationAlreadyDecided):
		return huma.Error409Conflict("The application already has that decision")
	case errors.Is(err, usecases.ErrTournamentFull):
		return huma.Error409Conflict("The tournament has reached its capacity")
	default:
		return huma.Error500InternalServerError("Unable to review the application")
	}
}
