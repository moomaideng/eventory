package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/registration"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

type RegistrationFormInput struct {
	TournamentID uuid.UUID `path:"tournamentId"`
}
type RegistrationFormOutput struct{ Body *models.RegistrationForm }
type SaveRegistrationFormInput struct {
	TournamentID uuid.UUID `path:"tournamentId"`
	Body         models.RegistrationFormConfig
}
type MemberAnswersInput struct {
	MemberID uuid.UUID `path:"memberId"`
}
type MemberAnswersOutput struct {
	Body struct {
		FormVersion   int                           `json:"formVersion"`
		Answers       []models.RegistrationAnswer   `json:"answers"`
		Questions     []models.RegistrationQuestion `json:"questions"`
		ConsentNotice string                        `json:"consentNotice"`
	}
}

func RegisterRegistrationRoutes(api huma.API, uc *usecases.RegistrationUseCase) {
	huma.Register(api, huma.Operation{OperationID: "get-registration-form", Method: http.MethodGet, Path: "/tournaments/{tournamentId}/registration-form", Tags: []string{"Registration"}}, func(ctx context.Context, input *RegistrationFormInput) (*RegistrationFormOutput, error) {
		viewer, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		form, err := uc.GetForm(ctx, input.TournamentID, viewer)
		if err != nil {
			return nil, registrationHTTPError(err)
		}
		return &RegistrationFormOutput{Body: form}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "save-registration-form", Method: http.MethodPut, Path: "/tournaments/{tournamentId}/registration-form", Tags: []string{"Registration"}}, func(ctx context.Context, input *SaveRegistrationFormInput) (*RegistrationFormOutput, error) {
		viewer, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		form := &models.RegistrationForm{Questions: input.Body.Questions, ConsentNotice: input.Body.ConsentNotice}
		if err := uc.SaveForm(ctx, input.TournamentID, viewer, form); err != nil {
			return nil, registrationHTTPError(err)
		}
		return &RegistrationFormOutput{Body: form}, nil
	})
	huma.Register(api, huma.Operation{OperationID: "get-member-registration", Method: http.MethodGet, Path: "/lobbies/members/{memberId}/registration", Tags: []string{"Registration"}}, func(ctx context.Context, input *MemberAnswersInput) (*MemberAnswersOutput, error) {
		viewer, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		member, err := uc.GetAnswers(ctx, input.MemberID, viewer)
		if err != nil {
			return nil, registrationHTTPError(err)
		}
		out := &MemberAnswersOutput{}
		out.Body.FormVersion, out.Body.Answers = member.FormVersion, member.RegistrationAnswers
		out.Body.Questions, out.Body.ConsentNotice = member.RegistrationQuestions, member.ConsentNotice
		return out, nil
	})
}

func registrationHTTPError(err error) error {
	switch {
	case errors.Is(err, registration.ErrInvalid):
		return huma.Error422UnprocessableEntity(err.Error())
	case errors.Is(err, registration.ErrFormChanged):
		return huma.Error409Conflict(err.Error())
	case errors.Is(err, repositories.ErrFormForbidden):
		return huma.Error403Forbidden("You cannot access this registration data")
	case errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, repositories.ErrTournamentNotFound):
		return huma.Error404NotFound("Registration resource not found")
	default:
		return teamLobbyHTTPError(err)
	}
}
