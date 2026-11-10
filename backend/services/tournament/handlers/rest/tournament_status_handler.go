package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
)

type OverrideTournamentStatusInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
	Body         struct {
		Status string `json:"status" enum:"DRAFT,CROWDFUNDING,REGISTRATION,ONGOING,COMPLETED" doc:"Target status"`
		Reason string `json:"reason,omitempty" maxLength:"500" doc:"Optional reason for the status change"`
		Force  bool   `json:"force,omitempty" doc:"Allow a crowdfunding owner to open registration before the funding goal is reached"`
	}
}

type OverrideTournamentStatusOutput struct {
	Body struct {
		TournamentID uuid.UUID `json:"tournamentId"`
		Status       string    `json:"status"`
	}
}

func RegisterTournamentStatusRoutes(
	api huma.API,
	auth, organizer func(huma.Context, func(huma.Context)),
	tournaments *usecases.TournamentUseCase,
) {
	huma.Register(api, huma.Operation{
		OperationID: "override-tournament-status",
		Method:      http.MethodPatch,
		Path:        "/api/v1/tournaments/{id}/status",
		Summary:     "Override a tournament's status",
		Description: "Moves an owned tournament to another lifecycle status.",
		Tags:        []string{"Organizer dashboard"},
		Middlewares: huma.Middlewares{auth, organizer},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *OverrideTournamentStatusInput) (*OverrideTournamentStatusOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		if _, err := middlewares.GetOrganizerProfile(ctx); err != nil {
			return nil, tournamentStatusHTTPError(err)
		}

		result, err := tournaments.OverrideWithForce(ctx, accountID, input.TournamentID, input.Body.Status, input.Body.Reason, input.Body.Force)
		if err != nil {
			return nil, tournamentStatusHTTPError(err)
		}

		output := &OverrideTournamentStatusOutput{}
		output.Body.TournamentID = result.TournamentID
		output.Body.Status = string(result.ToStatus)
		return output, nil
	})
}

func tournamentStatusHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, middlewares.ErrOrganizerProfileRequired):
		return huma.Error403Forbidden("Organizer profile required to perform this action", err)
	case errors.Is(err, usecases.ErrTournamentNotFound):
		return huma.Error404NotFound("Tournament not found or not owned by your account")
	case errors.Is(err, usecases.ErrNotTournamentOwner):
		return huma.Error403Forbidden("Only the tournament owner can change crowdfunding or registration status")
	case errors.Is(err, usecases.ErrCrowdfundingGoalNotReached):
		return huma.Error422UnprocessableEntity("Funding goal has not been reached. Use force start to open registration early")
	case errors.Is(err, usecases.ErrCrowdfundingSetupIncomplete):
		return huma.Error422UnprocessableEntity(err.Error())
	case errors.Is(err, usecases.ErrCrowdfundingPlanIncomplete),
		errors.Is(err, usecases.ErrRegistrationGateIncomplete),
		errors.Is(err, usecases.ErrRegistrationStillOpen),
		errors.Is(err, usecases.ErrUnresolvedApplicationsExist),
		errors.Is(err, usecases.ErrInsufficientAcceptedTeams),
		errors.Is(err, usecases.ErrBracketMissing),
		errors.Is(err, usecases.ErrBracketStage1Mismatch),
		errors.Is(err, usecases.ErrBracketNotSeeded),
		errors.Is(err, usecases.ErrUnfinishedMatchesExist),
		errors.Is(err, usecases.ErrNoChampionshipWinner):
		return huma.Error422UnprocessableEntity(err.Error())
	case errors.Is(err, usecases.ErrInvalidStatusTransition):
		return huma.Error422UnprocessableEntity("That status change is not allowed from the tournament's current status")
	case errors.Is(err, usecases.ErrStatusUnchanged):
		return huma.Error422UnprocessableEntity("Tournament is already in that status")
	case errors.Is(err, usecases.ErrUnknownStatus):
		return huma.Error400BadRequest("Unknown tournament status")
	default:
		return huma.Error500InternalServerError("Unable to override tournament status")
	}
}
