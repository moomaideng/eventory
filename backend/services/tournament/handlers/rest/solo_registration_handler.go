package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type SoloRegistrationInput struct {
	TournamentID uuid.UUID `path:"tournamentId"`
}
type StartSoloRegistrationInput struct {
	TournamentID uuid.UUID `path:"tournamentId"`
	Body         struct {
		Answers []models.FieldAnswer `json:"answers"`
	}
}
type SoloRegistrationResponse struct {
	Status    models.SoloRegistrationStatus `json:"status"`
	PaymentID *uuid.UUID                    `json:"paymentId,omitempty"`
}
type SoloRegistrationOutput struct{ Body SoloRegistrationResponse }
type SoloEligibilityOutput struct{ Body usecases.SoloEligibility }

func soloRegistrationHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, repositories.ErrTournamentNotFound), errors.Is(err, repositories.ErrSoloNotFound):
		return huma.Error404NotFound(err.Error(), err)
	case errors.Is(err, repositories.ErrSoloWrongMode), errors.Is(err, repositories.ErrSoloClosed), errors.Is(err, repositories.ErrSoloFull), errors.Is(err, repositories.ErrInvalidRegistrationFee), errors.Is(err, usecases.ErrSoloAnswersLocked):
		return huma.Error409Conflict(err.Error(), err)
	case errors.Is(err, usecases.ErrPaymentUnavailable):
		return huma.Error503ServiceUnavailable("Payment verification unavailable; retry without creating another registration", err)
	case errors.Is(err, usecases.ErrUnrecognizedField), errors.Is(err, usecases.ErrDuplicateFieldAnswer), errors.Is(err, forms.ErrFieldRequired), errors.Is(err, forms.ErrInvalidFieldValue), errors.Is(err, forms.ErrInvalidFieldConfig):
		return huma.Error400BadRequest(err.Error(), err)
	default:
		return registrationFormHTTPError(err)
	}
}

func RegisterSoloRegistrationRoutes(authGroup huma.API, useCase *usecases.SoloRegistrationUseCase) {
	huma.Register(authGroup, huma.Operation{OperationID: "get-solo-registration-eligibility", Method: http.MethodGet, Path: "/tournaments/{tournamentId}/solo-registration/eligibility", Summary: "Check personal solo registration eligibility", Tags: []string{"Solo Registration"}, Security: []map[string][]string{{"bearer": {}}}}, func(ctx context.Context, input *SoloRegistrationInput) (*SoloEligibilityOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		result, err := useCase.Eligibility(ctx, input.TournamentID, accountID)
		if err != nil {
			return nil, soloRegistrationHTTPError(err)
		}
		return &SoloEligibilityOutput{Body: result}, nil
	})
	huma.Register(authGroup, huma.Operation{OperationID: "start-solo-registration", Method: http.MethodPost, Path: "/tournaments/{tournamentId}/solo-registration", Summary: "Save solo answers and start registration payment", Tags: []string{"Solo Registration"}, Security: []map[string][]string{{"bearer": {}}}}, func(ctx context.Context, input *StartSoloRegistrationInput) (*SoloRegistrationOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		result, err := useCase.Start(ctx, input.TournamentID, accountID, input.Body.Answers)
		if err != nil {
			return nil, soloRegistrationHTTPError(err)
		}
		return &SoloRegistrationOutput{Body: SoloRegistrationResponse{Status: result.Status, PaymentID: result.PaymentID}}, nil
	})
	huma.Register(authGroup, huma.Operation{OperationID: "refresh-solo-registration", Method: http.MethodPost, Path: "/tournaments/{tournamentId}/solo-registration/refresh", Summary: "Verify solo registration payment", Tags: []string{"Solo Registration"}, Security: []map[string][]string{{"bearer": {}}}}, func(ctx context.Context, input *SoloRegistrationInput) (*SoloRegistrationOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		result, err := useCase.Refresh(ctx, input.TournamentID, accountID)
		if err != nil {
			return nil, soloRegistrationHTTPError(err)
		}
		return &SoloRegistrationOutput{Body: SoloRegistrationResponse{Status: result.Status, PaymentID: result.PaymentID}}, nil
	})
}
