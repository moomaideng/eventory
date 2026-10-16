package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type RegistrationFormResponse struct {
	ID           string             `json:"id"`
	TournamentID uuid.UUID          `json:"tournamentId"`
	Title        string             `json:"title"`
	Description  string             `json:"description,omitempty"`
	Fields       []models.FormField `json:"fields"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

type GetRegistrationFormInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
}

type GetRegistrationFormOutput struct {
	Body RegistrationFormResponse
}

type SaveRegistrationFormInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	Body         struct {
		Title       string             `json:"title" minLength:"1" maxLength:"160" doc:"Form title"`
		Description string             `json:"description,omitempty" maxLength:"2000" doc:"Form description"`
		Fields      []models.FormField `json:"fields" doc:"Form field definitions"`
	}
}

type SaveRegistrationFormOutput struct {
	Body RegistrationFormResponse
}

type FormResponseItem struct {
	ID           string               `json:"id"`
	TournamentID uuid.UUID            `json:"tournamentId"`
	AccountID    uuid.UUID            `json:"accountId"`
	Answers      []models.FieldAnswer `json:"answers"`
}

type SubmitFormResponseInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	Body         struct {
		Answers []models.FieldAnswer `json:"answers" doc:"Submitted answers"`
	}
}

type SubmitFormResponseOutput struct {
	Status int `default:"201"`
	Body   FormResponseItem
}

type GetMyFormResponseInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
}

type GetMyFormResponseOutput struct {
	Body FormResponseItem
}

type ListFormResponsesInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
}

type ListFormResponsesOutput struct {
	Body struct {
		Items []FormResponseItem `json:"items"`
		Total int                `json:"total"`
	}
}

func toRegistrationFormResponse(form *models.RegistrationForm) RegistrationFormResponse {
	return RegistrationFormResponse{
		ID:           form.ID.Hex(),
		TournamentID: form.TournamentID,
		Title:        form.Title,
		Description:  form.Description,
		Fields:       form.Fields,
		CreatedAt:    form.CreatedAt,
		UpdatedAt:    form.UpdatedAt,
	}
}

func toFormResponseItem(resp *models.FormResponse) FormResponseItem {
	return FormResponseItem{
		ID:           resp.ID.Hex(),
		TournamentID: resp.TournamentID,
		AccountID:    resp.AccountID,
		Answers:      resp.Answers,
	}
}

func registrationFormHTTPError(err error) error {
	switch {
	case errors.Is(err, usecases.ErrTournamentNotFound):
		return huma.Error404NotFound("Tournament not found", err)
	case errors.Is(err, usecases.ErrFormNotFound):
		return huma.Error404NotFound("Registration form not found", err)
	case errors.Is(err, usecases.ErrResponseNotFound):
		return huma.Error404NotFound("Registration form response not found", err)
	case errors.Is(err, usecases.ErrNotTournamentOwner):
		return huma.Error403Forbidden("Only the tournament organizer can perform this action", err)
	case errors.Is(err, usecases.ErrResponseAlreadyExists):
		return huma.Error409Conflict("Registration form response already submitted", err)
	case errors.Is(err, usecases.ErrRegistrationClosed):
		return huma.Error400BadRequest("Tournament registration is not open", err)
	case errors.Is(err, usecases.ErrInvalidFormTitle),
		errors.Is(err, usecases.ErrDuplicateFieldID),
		errors.Is(err, usecases.ErrEmptyFieldID),
		errors.Is(err, usecases.ErrEmptyFieldLabel),
		errors.Is(err, usecases.ErrUnrecognizedField),
		errors.Is(err, usecases.ErrDuplicateFieldAnswer),
		errors.Is(err, forms.ErrFieldRequired),
		errors.Is(err, forms.ErrInvalidFieldValue),
		errors.Is(err, forms.ErrInvalidFieldConfig),
		errors.Is(err, forms.ErrUnknownFieldType):
		return huma.Error400BadRequest(err.Error(), err)
	default:
		return huma.Error500InternalServerError("An unexpected error occurred", err)
	}
}

func RegisterRegistrationFormRoutes(
	api huma.API,
	authGroup huma.API,
	organizerTournamentGroup huma.API,
	formUseCase *usecases.RegistrationFormUseCase,
) {
	huma.Register(api, huma.Operation{
		OperationID: "get-registration-form",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments/{tournamentId}/form",
		Summary:     "Get registration form for a tournament",
		Description: "Returns the custom registration form definition configured for this tournament.",
		Tags:        []string{"Registration Forms"},
	}, func(ctx context.Context, input *GetRegistrationFormInput) (*GetRegistrationFormOutput, error) {
		form, err := formUseCase.GetForm(ctx, input.TournamentID)
		if err != nil {
			return nil, registrationFormHTTPError(err)
		}
		return &GetRegistrationFormOutput{Body: toRegistrationFormResponse(form)}, nil
	})

	huma.Register(organizerTournamentGroup, huma.Operation{
		OperationID: "save-registration-form",
		Method:      http.MethodPut,
		Path:        "/{tournamentId}/form",
		Summary:     "Create or update tournament registration form",
		Description: "Configures or updates custom registration form fields. Only the organizing owner can perform this action.",
		Tags:        []string{"Registration Forms"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *SaveRegistrationFormInput) (*SaveRegistrationFormOutput, error) {
		profile, err := middlewares.GetOrganizerProfile(ctx)
		if err != nil {
			return nil, registrationFormHTTPError(err)
		}

		form, err := formUseCase.SaveForm(
			ctx,
			profile.AccountID,
			input.TournamentID,
			input.Body.Title,
			input.Body.Description,
			input.Body.Fields,
		)
		if err != nil {
			return nil, registrationFormHTTPError(err)
		}
		return &SaveRegistrationFormOutput{Body: toRegistrationFormResponse(form)}, nil
	})

	huma.Register(authGroup, huma.Operation{
		OperationID: "submit-form-response",
		Method:      http.MethodPost,
		Path:        "/tournaments/{tournamentId}/form/responses",
		Summary:     "Submit answers for tournament registration form",
		Description: "Submits answers for a tournament registration form. One submission per participant.",
		Tags:        []string{"Registration Forms"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *SubmitFormResponseInput) (*SubmitFormResponseOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := formUseCase.SubmitResponse(ctx, accountID, input.TournamentID, input.Body.Answers)
		if err != nil {
			return nil, registrationFormHTTPError(err)
		}
		return &SubmitFormResponseOutput{
			Status: http.StatusCreated,
			Body:   toFormResponseItem(resp),
		}, nil
	})

	huma.Register(authGroup, huma.Operation{
		OperationID: "get-my-form-response",
		Method:      http.MethodGet,
		Path:        "/tournaments/{tournamentId}/form/responses/me",
		Summary:     "Get current user's submitted answers for a tournament",
		Description: "Returns the authenticated participant's submitted answers for this tournament.",
		Tags:        []string{"Registration Forms"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *GetMyFormResponseInput) (*GetMyFormResponseOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		resp, err := formUseCase.GetMyResponse(ctx, accountID, input.TournamentID)
		if err != nil {
			return nil, registrationFormHTTPError(err)
		}
		return &GetMyFormResponseOutput{Body: toFormResponseItem(resp)}, nil
	})

	huma.Register(organizerTournamentGroup, huma.Operation{
		OperationID: "list-form-responses",
		Method:      http.MethodGet,
		Path:        "/{tournamentId}/form/responses",
		Summary:     "List all submitted responses for tournament",
		Description: "Lists all participant responses for the tournament. Only the organizing owner can perform this action.",
		Tags:        []string{"Registration Forms"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *ListFormResponsesInput) (*ListFormResponsesOutput, error) {
		profile, err := middlewares.GetOrganizerProfile(ctx)
		if err != nil {
			return nil, registrationFormHTTPError(err)
		}

		responses, err := formUseCase.ListResponses(ctx, profile.AccountID, input.TournamentID)
		if err != nil {
			return nil, registrationFormHTTPError(err)
		}

		items := make([]FormResponseItem, len(responses))
		for i, r := range responses {
			items[i] = toFormResponseItem(&r)
		}

		out := &ListFormResponsesOutput{}
		out.Body.Items = items
		out.Body.Total = len(items)
		return out, nil
	})
}
