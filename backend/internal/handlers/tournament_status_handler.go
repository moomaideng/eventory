package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/usecases"
)

type TournamentStatusChangeResponse struct {
	ID           uuid.UUID `json:"id"`
	TournamentID uuid.UUID `json:"tournamentId"`
	FromStatus   string    `json:"fromStatus" enum:"DRAFT,CROWDFUNDING,REGISTRATION_OPEN,REGISTRATION_CLOSED,ONGOING,COMPLETED,CANCELLED"`
	ToStatus     string    `json:"toStatus" enum:"DRAFT,CROWDFUNDING,REGISTRATION_OPEN,REGISTRATION_CLOSED,ONGOING,COMPLETED,CANCELLED"`
	Reason       string    `json:"reason"`
	ActorID      uuid.UUID `json:"actorId"`
	ActorName    string    `json:"actorName,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type OverrideTournamentStatusInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
	Body         struct {
		Status string `json:"status" enum:"DRAFT,CROWDFUNDING,REGISTRATION_OPEN,REGISTRATION_CLOSED,ONGOING,COMPLETED,CANCELLED" doc:"Target status"`
		Reason string `json:"reason" minLength:"10" maxLength:"500" doc:"Why the status is being overridden; recorded in the audit trail"`
	}
}

type OverrideTournamentStatusOutput struct {
	Body struct {
		TournamentID uuid.UUID                      `json:"tournamentId"`
		Status       string                         `json:"status"`
		Change       TournamentStatusChangeResponse `json:"change"`
	}
}

type TournamentStatusHistoryInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
}

type TournamentStatusHistoryOutput struct {
	Body struct {
		CurrentStatus string `json:"currentStatus" enum:"DRAFT,CROWDFUNDING,REGISTRATION_OPEN,REGISTRATION_CLOSED,ONGOING,COMPLETED,CANCELLED"`
		// AllowedTransitions is empty for terminal statuses, which is how the
		// client knows to disable the override control entirely.
		AllowedTransitions []string                         `json:"allowedTransitions"`
		Items              []TournamentStatusChangeResponse `json:"items"`
	}
}

func RegisterTournamentStatusRoutes(
	api huma.API,
	statuses *usecases.TournamentStatusUseCase,
	accounts *usecases.AccountUseCase,
) {
	huma.Register(api, huma.Operation{
		OperationID: "override-tournament-status",
		Method:      http.MethodPatch,
		Path:        "/tournaments/{id}/status",
		Summary:     "Override a tournament's status",
		Description: "Manually moves an owned tournament to another lifecycle status, " +
			"recording the previous status, the reason, the acting account, and a timestamp.",
		Tags:     []string{"Organizer dashboard"},
		Security: []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *OverrideTournamentStatusInput) (*OverrideTournamentStatusOutput, error) {
		account, err := authenticatedAccount(ctx, accounts)
		if err != nil {
			return nil, err
		}

		result, err := statuses.Override(ctx, account.ID, input.TournamentID, input.Body.Status, input.Body.Reason)
		if err != nil {
			return nil, tournamentStatusHTTPError(err)
		}

		output := &OverrideTournamentStatusOutput{}
		output.Body.TournamentID = result.TournamentID
		output.Body.Status = string(result.ToStatus)
		output.Body.Change = toStatusChangeResponse(result.Change)
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-tournament-status-history",
		Method:      http.MethodGet,
		Path:        "/tournaments/{id}/status/history",
		Summary:     "List a tournament's status override history",
		Description: "Returns the audit trail of manual status overrides for an owned tournament, newest first.",
		Tags:        []string{"Organizer dashboard"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *TournamentStatusHistoryInput) (*TournamentStatusHistoryOutput, error) {
		account, err := authenticatedAccount(ctx, accounts)
		if err != nil {
			return nil, err
		}

		detail, err := statuses.History(ctx, account.ID, input.TournamentID)
		if err != nil {
			return nil, tournamentStatusHTTPError(err)
		}

		output := &TournamentStatusHistoryOutput{}
		output.Body.CurrentStatus = string(detail.CurrentStatus)
		output.Body.AllowedTransitions = make([]string, 0, len(detail.AllowedTransitions))
		for _, status := range detail.AllowedTransitions {
			output.Body.AllowedTransitions = append(output.Body.AllowedTransitions, string(status))
		}
		output.Body.Items = make([]TournamentStatusChangeResponse, 0, len(detail.Changes))
		for _, change := range detail.Changes {
			output.Body.Items = append(output.Body.Items, toStatusChangeResponse(change))
		}
		return output, nil
	})
}

func toStatusChangeResponse(change models.TournamentStatusChange) TournamentStatusChangeResponse {
	return TournamentStatusChangeResponse{
		ID:           change.ID,
		TournamentID: change.TournamentID,
		FromStatus:   string(change.FromStatus),
		ToStatus:     string(change.ToStatus),
		Reason:       change.Reason,
		ActorID:      change.ActorAccountID,
		ActorName:    change.Actor.DisplayName,
		CreatedAt:    change.CreatedAt,
	}
}

func tournamentStatusHTTPError(err error) error {
	switch {
	case errors.Is(err, usecases.ErrTournamentNotFound):
		// Ownership failures are reported as 404, never 403: confirming a
		// tournament exists would leak it to accounts that do not own it.
		return huma.Error404NotFound("Tournament not found or not owned by your account")
	case errors.Is(err, usecases.ErrInvalidStatusTransition):
		return huma.Error422UnprocessableEntity("That status change is not allowed from the tournament's current status")
	case errors.Is(err, usecases.ErrStatusUnchanged):
		return huma.Error422UnprocessableEntity("Tournament is already in that status")
	case errors.Is(err, usecases.ErrOverrideReasonRequired):
		return huma.Error422UnprocessableEntity("A reason between 10 and 500 characters is required")
	case errors.Is(err, usecases.ErrUnknownStatus):
		return huma.Error400BadRequest("Unknown tournament status")
	default:
		return huma.Error500InternalServerError("Unable to override tournament status")
	}
}
