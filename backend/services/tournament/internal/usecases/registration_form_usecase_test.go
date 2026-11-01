package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type tournamentRepoStub struct {
	mockTournamentRepository
	tournaments map[uuid.UUID]*models.Tournament
}

func (s *tournamentRepoStub) Search(_ context.Context, _ repositories.TournamentFilters) ([]repositories.TournamentSearchItem, int64, error) {
	return nil, 0, nil
}
func (s *tournamentRepoStub) GetPublishedByID(_ context.Context, id uuid.UUID) (*models.Tournament, error) {
	return s.GetByID(context.Background(), id)
}
func (s *tournamentRepoStub) GetByID(_ context.Context, id uuid.UUID) (*models.Tournament, error) {
	t, ok := s.tournaments[id]
	if !ok {
		return nil, repositories.ErrTournamentNotFound
	}
	return t, nil
}
func (s *tournamentRepoStub) Create(_ context.Context, _ *models.Tournament, _ *models.TournamentFunding) error {
	return nil
}
func (s *tournamentRepoStub) Update(_ context.Context, _ *models.Tournament) error {
	return nil
}
func (s *tournamentRepoStub) GetActiveTeamCount(_ context.Context, _ uuid.UUID) (int64, int64, error) {
	return 0, 0, nil
}

type formRepoStub struct {
	forms     map[uuid.UUID]*models.RegistrationForm
	responses map[string]*models.FormResponse // key: tournamentID:accountID
}

func newFormRepoStub() *formRepoStub {
	return &formRepoStub{
		forms:     make(map[uuid.UUID]*models.RegistrationForm),
		responses: make(map[string]*models.FormResponse),
	}
}

func (s *formRepoStub) EnsureIndexes(_ context.Context) error {
	return nil
}

func (s *formRepoStub) GetFormByTournamentID(_ context.Context, tournamentID uuid.UUID) (*models.RegistrationForm, error) {
	f, ok := s.forms[tournamentID]
	if !ok {
		return nil, repositories.ErrFormNotFound
	}
	return f, nil
}

func (s *formRepoStub) UpsertForm(_ context.Context, form *models.RegistrationForm) (*models.RegistrationForm, error) {
	s.forms[form.TournamentID] = form
	return form, nil
}

func (s *formRepoStub) CreateResponse(_ context.Context, resp *models.FormResponse) (*models.FormResponse, error) {
	key := resp.TournamentID.String() + ":" + resp.AccountID.String()
	if _, exists := s.responses[key]; exists {
		return nil, repositories.ErrResponseAlreadyExists
	}
	s.responses[key] = resp
	return resp, nil
}

func (s *formRepoStub) UpsertResponse(_ context.Context, resp *models.FormResponse) (*models.FormResponse, error) {
	key := resp.TournamentID.String() + ":" + resp.AccountID.String()
	s.responses[key] = resp
	return resp, nil
}

func (s *formRepoStub) GetResponseByAccount(_ context.Context, tournamentID uuid.UUID, accountID uuid.UUID) (*models.FormResponse, error) {
	key := tournamentID.String() + ":" + accountID.String()
	r, ok := s.responses[key]
	if !ok {
		return nil, repositories.ErrResponseNotFound
	}
	return r, nil
}

func (s *formRepoStub) ListResponsesByTournamentID(_ context.Context, tournamentID uuid.UUID) ([]models.FormResponse, error) {
	var list []models.FormResponse
	for _, r := range s.responses {
		if r.TournamentID == tournamentID {
			list = append(list, *r)
		}
	}
	return list, nil
}

func TestRegistrationFormUseCase_SaveForm(t *testing.T) {
	ctx := context.Background()
	organizerID := uuid.New()
	otherOrganizerID := uuid.New()
	tournamentID := uuid.New()

	tournaments := map[uuid.UUID]*models.Tournament{
		tournamentID: {
			ID: tournamentID,
			Staff: []models.TournamentStaff{
				{AccountID: organizerID, Role: models.TournamentStaffRoleOwner},
			},
			Status: models.TournamentStatusRegistrationOpen,
		},
	}
	tRepo := &tournamentRepoStub{tournaments: tournaments}
	fRepo := newFormRepoStub()
	registry := forms.NewDefaultRegistry()
	uc := usecases.NewRegistrationFormUseCase(tRepo, fRepo, registry)

	t.Run("tournament not found", func(t *testing.T) {
		_, err := uc.SaveForm(ctx, organizerID, uuid.New(), "Title", "Desc", nil)
		if !errors.Is(err, usecases.ErrTournamentNotFound) {
			t.Fatalf("expected ErrTournamentNotFound, got %v", err)
		}
	})

	t.Run("forbidden not owner", func(t *testing.T) {
		_, err := uc.SaveForm(ctx, otherOrganizerID, tournamentID, "Title", "Desc", nil)
		if !errors.Is(err, usecases.ErrNotTournamentOwner) {
			t.Fatalf("expected ErrNotTournamentOwner, got %v", err)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		_, err := uc.SaveForm(ctx, organizerID, tournamentID, "   ", "Desc", nil)
		if !errors.Is(err, usecases.ErrInvalidFormTitle) {
			t.Fatalf("expected ErrInvalidFormTitle, got %v", err)
		}
	})

	t.Run("duplicate field IDs", func(t *testing.T) {
		fields := []models.FormField{
			{ID: "f1", Label: "Field 1", Type: models.FieldTypeShortText},
			{ID: "f1", Label: "Field 2", Type: models.FieldTypeShortText},
		}
		_, err := uc.SaveForm(ctx, organizerID, tournamentID, "Title", "Desc", fields)
		if !errors.Is(err, usecases.ErrDuplicateFieldID) {
			t.Fatalf("expected ErrDuplicateFieldID, got %v", err)
		}
	})

	t.Run("invalid field config", func(t *testing.T) {
		fields := []models.FormField{
			{ID: "f1", Label: "Dropdown", Type: models.FieldTypeSelectOne}, // missing options
		}
		_, err := uc.SaveForm(ctx, organizerID, tournamentID, "Title", "Desc", fields)
		if !errors.Is(err, forms.ErrInvalidFieldConfig) {
			t.Fatalf("expected ErrInvalidFieldConfig, got %v", err)
		}
	})

	t.Run("success save", func(t *testing.T) {
		fields := []models.FormField{
			{
				ID:       "ign",
				Label:    "In-game Name",
				Type:     models.FieldTypeShortText,
				Required: true,
			},
			{
				ID:    "size",
				Label: "T-Shirt Size",
				Type:  models.FieldTypeSelectOne,
				Config: forms.SelectOneConfig{
					Options: []models.FormFieldOption{
						{ID: "s", Label: "Small"},
						{ID: "m", Label: "Medium"},
					},
				},
			},
		}
		form, err := uc.SaveForm(ctx, organizerID, tournamentID, "Player Reg Form", "Rules", fields)
		if err != nil {
			t.Fatalf("SaveForm() error = %v", err)
		}
		if form.Title != "Player Reg Form" {
			t.Fatalf("form title = %q, want %q", form.Title, "Player Reg Form")
		}
		if len(form.Fields) != 2 {
			t.Fatalf("form fields count = %d, want 2", len(form.Fields))
		}
	})
}

func TestRegistrationFormUseCase_SubmitResponse(t *testing.T) {
	ctx := context.Background()
	organizerID := uuid.New()
	tournamentID := uuid.New()
	accountID := uuid.New()

	tournaments := map[uuid.UUID]*models.Tournament{
		tournamentID: {
			ID: tournamentID,
			Staff: []models.TournamentStaff{
				{AccountID: organizerID, Role: models.TournamentStaffRoleOwner},
			},
			Status: models.TournamentStatusRegistrationOpen,
		},
	}
	tRepo := &tournamentRepoStub{tournaments: tournaments}
	fRepo := newFormRepoStub()
	registry := forms.NewDefaultRegistry()
	uc := usecases.NewRegistrationFormUseCase(tRepo, fRepo, registry)

	// Setup form
	_, err := uc.SaveForm(ctx, organizerID, tournamentID, "Reg Form", "", []models.FormField{
		{ID: "ign", Label: "In-game Name", Type: models.FieldTypeShortText, Required: true},
		{ID: "age", Label: "Age", Type: models.FieldTypeNumber, Required: false},
	})
	if err != nil {
		t.Fatalf("SaveForm() error = %v", err)
	}

	t.Run("unrecognized field", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "unknown", Value: "something"},
		}
		_, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		if !errors.Is(err, usecases.ErrUnrecognizedField) {
			t.Fatalf("expected ErrUnrecognizedField, got %v", err)
		}
	})

	t.Run("duplicate answer for field", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "ign", Value: "Player1"},
			{FieldID: "ign", Value: "Player2"},
		}
		_, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		if !errors.Is(err, usecases.ErrDuplicateFieldAnswer) {
			t.Fatalf("expected ErrDuplicateFieldAnswer, got %v", err)
		}
	})

	t.Run("missing required field", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "age", Value: 20},
		}
		_, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		if !errors.Is(err, forms.ErrFieldRequired) {
			t.Fatalf("expected ErrFieldRequired, got %v", err)
		}
	})

	t.Run("successful submission", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "ign", Value: "Faker"},
			{FieldID: "age", Value: 28},
		}
		resp, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		if err != nil {
			t.Fatalf("SubmitResponse() error = %v", err)
		}
		if resp.TournamentID != tournamentID {
			t.Fatalf("tournamentID = %v, want %v", resp.TournamentID, tournamentID)
		}
		if resp.AccountID != accountID {
			t.Fatalf("accountID = %v, want %v", resp.AccountID, accountID)
		}
		if len(resp.Answers) != 2 {
			t.Fatalf("answers count = %d, want 2", len(resp.Answers))
		}
	})

	t.Run("duplicate submission", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "ign", Value: "Faker"},
		}
		_, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		if !errors.Is(err, usecases.ErrResponseAlreadyExists) {
			t.Fatalf("expected ErrResponseAlreadyExists, got %v", err)
		}
	})
}

func (s *formRepoStub) GetResponsesByAccounts(ctx context.Context, tournamentID uuid.UUID, accountIDs []uuid.UUID) ([]models.FormResponse, error) {
	responses := make([]models.FormResponse, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		response, err := s.GetResponseByAccount(ctx, tournamentID, accountID)
		if err == nil {
			responses = append(responses, *response)
		}
	}
	return responses, nil
}
