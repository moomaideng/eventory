package usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.ErrorIs(t, err, usecases.ErrTournamentNotFound)
	})

	t.Run("forbidden not owner", func(t *testing.T) {
		_, err := uc.SaveForm(ctx, otherOrganizerID, tournamentID, "Title", "Desc", nil)
		assert.ErrorIs(t, err, usecases.ErrNotTournamentOwner)
	})

	t.Run("empty title", func(t *testing.T) {
		_, err := uc.SaveForm(ctx, organizerID, tournamentID, "   ", "Desc", nil)
		assert.ErrorIs(t, err, usecases.ErrInvalidFormTitle)
	})

	t.Run("duplicate field IDs", func(t *testing.T) {
		fields := []models.FormField{
			{ID: "f1", Label: "Field 1", Type: models.FieldTypeShortText},
			{ID: "f1", Label: "Field 2", Type: models.FieldTypeShortText},
		}
		_, err := uc.SaveForm(ctx, organizerID, tournamentID, "Title", "Desc", fields)
		assert.ErrorIs(t, err, usecases.ErrDuplicateFieldID)
	})

	t.Run("invalid field config", func(t *testing.T) {
		fields := []models.FormField{
			{ID: "f1", Label: "Dropdown", Type: models.FieldTypeSelectOne}, // missing options
		}
		_, err := uc.SaveForm(ctx, organizerID, tournamentID, "Title", "Desc", fields)
		assert.ErrorIs(t, err, forms.ErrInvalidFieldConfig)
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
		require.NoError(t, err)
		assert.Equal(t, "Player Reg Form", form.Title)
		assert.Len(t, form.Fields, 2)
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
	require.NoError(t, err)

	t.Run("unrecognized field", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "unknown", Value: "something"},
		}
		_, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		assert.ErrorIs(t, err, usecases.ErrUnrecognizedField)
	})

	t.Run("duplicate answer for field", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "ign", Value: "Player1"},
			{FieldID: "ign", Value: "Player2"},
		}
		_, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		assert.ErrorIs(t, err, usecases.ErrDuplicateFieldAnswer)
	})

	t.Run("missing required field", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "age", Value: 20},
		}
		_, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		assert.ErrorIs(t, err, forms.ErrFieldRequired)
	})

	t.Run("successful submission", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "ign", Value: "Faker"},
			{FieldID: "age", Value: 28},
		}
		resp, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		require.NoError(t, err)
		assert.Equal(t, tournamentID, resp.TournamentID)
		assert.Equal(t, accountID, resp.AccountID)
		assert.Len(t, resp.Answers, 2)
	})

	t.Run("duplicate submission", func(t *testing.T) {
		answers := []models.FieldAnswer{
			{FieldID: "ign", Value: "Faker"},
		}
		_, err := uc.SubmitResponse(ctx, accountID, tournamentID, answers)
		assert.ErrorIs(t, err, usecases.ErrResponseAlreadyExists)
	})
}
