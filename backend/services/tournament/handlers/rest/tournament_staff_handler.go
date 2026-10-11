package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type TournamentStaffResponse struct {
	TournamentID uuid.UUID `json:"tournamentId"`
	AccountID    uuid.UUID `json:"accountId"`
	Handle       string    `json:"handle"`
	DisplayName  string    `json:"displayName"`
	Role         string    `json:"role" enum:"OWNER,ADMIN,REFEREE"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CreateTournamentStaffOutput struct {
	Status int `default:"201"`
	Body   TournamentStaffResponse
}

type TournamentStaffOutput struct {
	Body TournamentStaffResponse
}

type TournamentStaffListBody struct {
	Items []TournamentStaffResponse `json:"items"`
}

type TournamentStaffListOutput struct {
	Body TournamentStaffListBody
}

type EmptyTournamentStaffOutput struct {
	Status int
}

type InviteStaffRequest struct {
	Handle string `json:"handle" minLength:"3" maxLength:"32" doc:"Public handle of the user to invite"`
	Role   string `json:"role" enum:"ADMIN,REFEREE" doc:"Staff role to assign (ADMIN or REFEREE)"`
}

type InviteStaffInput struct {
	TournamentID uuid.UUID          `path:"id" doc:"Tournament ID"`
	Body         InviteStaffRequest
}

type UpdateStaffRoleRequest struct {
	Role string `json:"role" enum:"ADMIN,REFEREE" doc:"New staff role (ADMIN or REFEREE)"`
}

type UpdateStaffRoleInput struct {
	TournamentID uuid.UUID              `path:"id" doc:"Tournament ID"`
	AccountID    uuid.UUID              `path:"accountId" doc:"Staff member account ID"`
	Body         UpdateStaffRoleRequest
}

type ListStaffInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
}

type RemoveStaffInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
	AccountID    uuid.UUID `path:"accountId" doc:"Staff member account ID to remove"`
}

func RegisterTournamentStaffRoutes(api huma.API, staffUseCase *usecases.TournamentStaffUseCase) {
	huma.Register(api, huma.Operation{
		OperationID: "invite-tournament-staff",
		Method:      http.MethodPost,
		Path:        "/tournaments/{id}/staff",
		Summary:     "Invite a tournament staff member",
		Description: "Invites an active user by their handle and assigns them an ADMIN or REFEREE role.",
		Tags:        []string{"Tournament Staff"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *InviteStaffInput) (*CreateTournamentStaffOutput, error) {
		callerID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		view, err := staffUseCase.InviteStaff(
			ctx,
			callerID,
			input.TournamentID,
			input.Body.Handle,
			models.TournamentStaffRole(input.Body.Role),
		)
		if err != nil {
			return nil, tournamentStaffHTTPError(err)
		}

		return &CreateTournamentStaffOutput{
			Status: http.StatusCreated,
			Body:   toTournamentStaffResponse(*view),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-tournament-staff",
		Method:      http.MethodGet,
		Path:        "/tournaments/{id}/staff",
		Summary:     "List tournament staff",
		Description: "Lists all staff members (owner, admins, referees) assigned to a tournament.",
		Tags:        []string{"Tournament Staff"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *ListStaffInput) (*TournamentStaffListOutput, error) {
		callerID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		views, err := staffUseCase.ListStaff(ctx, callerID, input.TournamentID)
		if err != nil {
			return nil, tournamentStaffHTTPError(err)
		}

		items := make([]TournamentStaffResponse, 0, len(views))
		for _, v := range views {
			items = append(items, toTournamentStaffResponse(v))
		}

		return &TournamentStaffListOutput{
			Body: TournamentStaffListBody{Items: items},
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-tournament-staff-role",
		Method:      http.MethodPatch,
		Path:        "/tournaments/{id}/staff/{accountId}",
		Summary:     "Update staff member role",
		Description: "Updates an existing staff member's role between ADMIN and REFEREE.",
		Tags:        []string{"Tournament Staff"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *UpdateStaffRoleInput) (*TournamentStaffOutput, error) {
		callerID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		view, err := staffUseCase.UpdateStaffRole(
			ctx,
			callerID,
			input.TournamentID,
			input.AccountID,
			models.TournamentStaffRole(input.Body.Role),
		)
		if err != nil {
			return nil, tournamentStaffHTTPError(err)
		}

		return &TournamentStaffOutput{
			Body: toTournamentStaffResponse(*view),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "remove-tournament-staff",
		Method:      http.MethodDelete,
		Path:        "/tournaments/{id}/staff/{accountId}",
		Summary:     "Remove tournament staff member",
		Description: "Removes an assigned staff member from the tournament.",
		Tags:        []string{"Tournament Staff"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *RemoveStaffInput) (*EmptyTournamentStaffOutput, error) {
		callerID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		if err := staffUseCase.RemoveStaff(ctx, callerID, input.TournamentID, input.AccountID); err != nil {
			return nil, tournamentStaffHTTPError(err)
		}

		return &EmptyTournamentStaffOutput{Status: http.StatusNoContent}, nil
	})
}

func toTournamentStaffResponse(v usecases.TournamentStaffView) TournamentStaffResponse {
	return TournamentStaffResponse{
		TournamentID: v.TournamentID,
		AccountID:    v.AccountID,
		Handle:       v.Handle,
		DisplayName:  v.DisplayName,
		Role:         string(v.Role),
		CreatedAt:    v.CreatedAt,
	}
}

func tournamentStaffHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, usecases.ErrTournamentNotFound):
		return huma.Error404NotFound("Tournament not found")
	case errors.Is(err, usecases.ErrCallerNotAuthorized):
		return huma.Error403Forbidden("You are not authorized to manage staff for this tournament", err)
	case errors.Is(err, usecases.ErrInvalidStaffRole):
		return huma.Error400BadRequest("Invalid staff role: must be ADMIN or REFEREE", err)
	case errors.Is(err, usecases.ErrAccountNotActive):
		return huma.Error422UnprocessableEntity("The invited user is not active and cannot be assigned as staff", err)
	case errors.Is(err, usecases.ErrStaffAlreadyExists):
		return huma.Error409Conflict("This user is already a staff member for this tournament", err)
	case errors.Is(err, usecases.ErrStaffNotFound):
		return huma.Error404NotFound("Staff member not found")
	case errors.Is(err, usecases.ErrCannotModifyOwner):
		return huma.Error403Forbidden("Tournament owner role cannot be modified or removed", err)
	default:
		return huma.Error500InternalServerError("Tournament staff operation failed", err)
	}
}
