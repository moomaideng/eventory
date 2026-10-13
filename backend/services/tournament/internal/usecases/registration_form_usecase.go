package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrFormNotFound          = repositories.ErrFormNotFound
	ErrResponseNotFound      = repositories.ErrResponseNotFound
	ErrResponseAlreadyExists = repositories.ErrResponseAlreadyExists
	ErrInvalidFormTitle      = errors.New("form title is required")
	ErrDuplicateFieldID      = errors.New("duplicate field id in form")
	ErrEmptyFieldID          = errors.New("field id cannot be empty")
	ErrEmptyFieldLabel       = errors.New("field label cannot be empty")
	ErrRegistrationClosed    = errors.New("tournament registration is not open")
	ErrUnrecognizedField     = errors.New("unrecognized field submitted")
	ErrDuplicateFieldAnswer  = errors.New("duplicate answer for field submitted")
)

type RegistrationFormUseCase struct {
	tournamentRepo repositories.TournamentRepository
	formRepo       repositories.RegistrationFormRepository
	registry       *forms.FieldRegistry
}

func NewRegistrationFormUseCase(
	tournamentRepo repositories.TournamentRepository,
	formRepo repositories.RegistrationFormRepository,
	registry *forms.FieldRegistry,
) *RegistrationFormUseCase {
	return &RegistrationFormUseCase{
		tournamentRepo: tournamentRepo,
		formRepo:       formRepo,
		registry:       registry,
	}
}

func (u *RegistrationFormUseCase) GetForm(ctx context.Context, tournamentID uuid.UUID) (*models.RegistrationForm, error) {
	_, err := u.tournamentRepo.GetByID(ctx, tournamentID)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	form, err := u.formRepo.GetFormByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (u *RegistrationFormUseCase) SaveForm(
	ctx context.Context,
	accountID uuid.UUID,
	tournamentID uuid.UUID,
	title string,
	description string,
	fields []models.FormField,
) (*models.RegistrationForm, error) {
	tournament, err := u.tournamentRepo.GetByID(ctx, tournamentID)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	if !isTournamentOrganizer(tournament, accountID) {
		return nil, ErrNotTournamentOwner
	}

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return nil, ErrInvalidFormTitle
	}

	seenFieldIDs := make(map[string]struct{}, len(fields))
	for i, f := range fields {
		fieldID := strings.TrimSpace(f.ID)
		if fieldID == "" {
			return nil, ErrEmptyFieldID
		}
		if _, exists := seenFieldIDs[fieldID]; exists {
			return nil, fmt.Errorf("%w: '%s'", ErrDuplicateFieldID, fieldID)
		}
		seenFieldIDs[fieldID] = struct{}{}

		if strings.TrimSpace(f.Label) == "" {
			return nil, ErrEmptyFieldLabel
		}

		if err := u.registry.ValidateFieldConfig(f); err != nil {
			return nil, fmt.Errorf("field '%s': %w", fieldID, err)
		}

		fields[i].ID = fieldID
		if fields[i].Order == 0 {
			fields[i].Order = i + 1
		}
	}

	form := &models.RegistrationForm{
		TournamentID: tournamentID,
		Title:        trimmedTitle,
		Description:  strings.TrimSpace(description),
		Fields:       fields,
	}

	return u.formRepo.UpsertForm(ctx, form)
}

func (u *RegistrationFormUseCase) SubmitResponse(
	ctx context.Context,
	accountID uuid.UUID,
	tournamentID uuid.UUID,
	answers []models.FieldAnswer,
) (*models.FormResponse, error) {
	tournament, err := u.tournamentRepo.GetByID(ctx, tournamentID)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	if tournament.Status != models.TournamentStatusRegistrationOpen {
		return nil, ErrRegistrationClosed
	}

	form, err := u.formRepo.GetFormByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	formFieldsMap := make(map[string]models.FormField, len(form.Fields))
	for _, f := range form.Fields {
		formFieldsMap[f.ID] = f
	}

	submittedMap := make(map[string]any, len(answers))
	for _, a := range answers {
		fieldID := strings.TrimSpace(a.FieldID)
		if _, exists := formFieldsMap[fieldID]; !exists {
			return nil, fmt.Errorf("%w: '%s'", ErrUnrecognizedField, fieldID)
		}
		if _, duplicate := submittedMap[fieldID]; duplicate {
			return nil, fmt.Errorf("%w: '%s'", ErrDuplicateFieldAnswer, fieldID)
		}
		submittedMap[fieldID] = a.Value
	}

	for _, field := range form.Fields {
		val, answered := submittedMap[field.ID]
		if !answered {
			val = nil
		}
		if err := u.registry.ValidateFieldAnswer(field, val); err != nil {
			return nil, fmt.Errorf("field '%s': %w", field.ID, err)
		}
	}

	resp := &models.FormResponse{
		TournamentID: tournamentID,
		AccountID:    accountID,
		Answers:      answers,
	}

	return u.formRepo.CreateResponse(ctx, resp)
}

func (u *RegistrationFormUseCase) GetMyResponse(
	ctx context.Context,
	accountID uuid.UUID,
	tournamentID uuid.UUID,
) (*models.FormResponse, error) {
	_, err := u.tournamentRepo.GetByID(ctx, tournamentID)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	return u.formRepo.GetResponseByAccount(ctx, tournamentID, accountID)
}

func (u *RegistrationFormUseCase) ListResponses(
	ctx context.Context,
	accountID uuid.UUID,
	tournamentID uuid.UUID,
) ([]models.FormResponse, error) {
	tournament, err := u.tournamentRepo.GetByID(ctx, tournamentID)
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	if !isTournamentOrganizer(tournament, accountID) {
		return nil, ErrNotTournamentOwner
	}

	responses, err := u.formRepo.ListResponsesByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if tournament.RegistrationMode == models.TournamentRegistrationModeSolo {
		if counter, ok := u.tournamentRepo.(interface {
			RegisteredSoloAccountIDs(context.Context, uuid.UUID) (map[uuid.UUID]bool, error)
		}); ok {
			registered, err := counter.RegisteredSoloAccountIDs(ctx, tournamentID)
			if err != nil {
				return nil, err
			}
			filtered := make([]models.FormResponse, 0, len(responses))
			for _, response := range responses {
				if registered[response.AccountID] {
					filtered = append(filtered, response)
				}
			}
			return filtered, nil
		}
	}
	return responses, nil
}

func isTournamentOrganizer(tournament *models.Tournament, accountID uuid.UUID) bool {
	if staff, ok := tournament.StaffMember(accountID); ok {
		return staff.IsOwner()
	}
	return false
}
