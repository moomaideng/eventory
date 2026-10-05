package rest

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	sharedmiddlewares "github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
)

func authenticatedAccountID(ctx context.Context) (uuid.UUID, error) {
	accountID, err := sharedmiddlewares.GetAuthUserID(ctx)
	if err != nil {
		return uuid.Nil, huma.Error401Unauthorized("Authentication required", err)
	}
	return accountID, nil
}

func mapAccountPortError(err error) error {
	switch {
	case errors.Is(err, ports.ErrAccountNotFound):
		return huma.Error404NotFound("Account not found.", err)
	case errors.Is(err, ports.ErrAccountOnboarding):
		return huma.Error403Forbidden("Onboarding required: please complete your profile first")
	case errors.Is(err, ports.ErrOrganizerNotFound):
		return huma.Error403Forbidden("Organizer profile required to perform this action")
	case errors.Is(err, ports.ErrAccountUnavailable):
		return huma.Error502BadGateway("Account service unavailable", err)
	default:
		return nil
	}
}
