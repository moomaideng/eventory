package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	sharedmiddlewares "github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/handlers/rest"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	tournamentmiddlewares "github.com/moomaideng/eventory/services/tournament/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockTournamentRepo struct {
	tournament *models.Tournament
}

func (m *mockTournamentRepo) Search(_ context.Context, _ repositories.TournamentFilters) ([]repositories.TournamentSearchItem, int64, error) {
	return nil, 0, nil
}
func (m *mockTournamentRepo) GetPublishedByID(_ context.Context, id uuid.UUID) (*models.Tournament, error) {
	return m.GetByID(context.Background(), id)
}
func (m *mockTournamentRepo) GetByID(_ context.Context, id uuid.UUID) (*models.Tournament, error) {
	if m.tournament != nil && m.tournament.ID == id {
		return m.tournament, nil
	}
	return nil, repositories.ErrTournamentNotFound
}
func (m *mockTournamentRepo) Create(_ context.Context, _ *models.Tournament, _ *models.TournamentFunding) error {
	return nil
}
func (m *mockTournamentRepo) Update(_ context.Context, _ *models.Tournament) error {
	return nil
}
func (m *mockTournamentRepo) GetActiveTeamCount(_ context.Context, _ uuid.UUID) (int64, int64, error) {
	return 0, 0, nil
}
func (m *mockTournamentRepo) ListOwned(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Tournament, int64, error) {
	return nil, 0, nil
}
func (m *mockTournamentRepo) GetOwned(_ context.Context, _, _ uuid.UUID) (*models.Tournament, error) {
	return nil, nil
}
func (m *mockTournamentRepo) ApplyOverride(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID, _ string, _ func(*models.Tournament) (models.TournamentStatus, error)) (*models.TournamentStatusChange, error) {
	return nil, nil
}
func (m *mockTournamentRepo) ListHistory(_ context.Context, _, _ uuid.UUID) (*repositories.TournamentStatusHistory, error) {
	return nil, nil
}

type mockFormRepo struct {
	form      *models.RegistrationForm
	responses map[string]*models.FormResponse
}

func newMockFormRepo() *mockFormRepo {
	return &mockFormRepo{
		responses: make(map[string]*models.FormResponse),
	}
}

func (m *mockFormRepo) EnsureIndexes(_ context.Context) error {
	return nil
}

func (m *mockFormRepo) GetFormByTournamentID(_ context.Context, id uuid.UUID) (*models.RegistrationForm, error) {
	if m.form != nil && m.form.TournamentID == id {
		return m.form, nil
	}
	return nil, repositories.ErrFormNotFound
}

func (m *mockFormRepo) UpsertForm(_ context.Context, form *models.RegistrationForm) (*models.RegistrationForm, error) {
	m.form = form
	return form, nil
}

func (m *mockFormRepo) CreateResponse(_ context.Context, resp *models.FormResponse) (*models.FormResponse, error) {
	key := resp.TournamentID.String() + ":" + resp.AccountID.String()
	if _, ok := m.responses[key]; ok {
		return nil, repositories.ErrResponseAlreadyExists
	}
	resp.ID = bson.NewObjectID()
	m.responses[key] = resp
	return resp, nil
}

func (m *mockFormRepo) GetResponseByAccount(_ context.Context, tournamentID, accountID uuid.UUID) (*models.FormResponse, error) {
	key := tournamentID.String() + ":" + accountID.String()
	r, ok := m.responses[key]
	if !ok {
		return nil, repositories.ErrResponseNotFound
	}
	return r, nil
}

func (m *mockFormRepo) ListResponsesByTournamentID(_ context.Context, tournamentID uuid.UUID) ([]models.FormResponse, error) {
	var list []models.FormResponse
	for _, r := range m.responses {
		if r.TournamentID == tournamentID {
			list = append(list, *r)
		}
	}
	return list, nil
}

func TestRegistrationFormRoutes(t *testing.T) {
	organizerAccountID := uuid.New()
	participantAccountID := uuid.New()
	tournamentID := uuid.New()

	tourney := &models.Tournament{
		ID: tournamentID,
		Staff: []models.TournamentStaff{
			{AccountID: organizerAccountID, Role: models.TournamentStaffRoleOwner},
		},
		Status: models.TournamentStatusRegistrationOpen,
	}

	tRepo := &mockTournamentRepo{tournament: tourney}
	fRepo := newMockFormRepo()
	registry := forms.NewDefaultRegistry()
	useCase := usecases.NewRegistrationFormUseCase(tRepo, fRepo, registry)

	_, api := humatest.New(t)

	authMW := func(ctx huma.Context, next func(huma.Context)) {
		newCtx := context.WithValue(ctx.Context(), sharedmiddlewares.UserSubContextKey, participantAccountID.String())
		ctx = huma.WithContext(ctx, newCtx)
		next(ctx)
	}

	organizerMW := func(ctx huma.Context, next func(huma.Context)) {
		newCtx := context.WithValue(ctx.Context(), sharedmiddlewares.UserSubContextKey, organizerAccountID.String())
		newCtx = context.WithValue(newCtx, tournamentmiddlewares.OrganizerProfileContextKey, &ports.OrganizerInfo{
			AccountID: organizerAccountID,
		})
		ctx = huma.WithContext(ctx, newCtx)
		next(ctx)
	}

	rest.RegisterRegistrationFormRoutes(api, authMW, organizerMW, useCase)

	t.Run("GET form when not found", func(t *testing.T) {
		resp := api.Get("/api/v1/tournaments/" + tournamentID.String() + "/form")
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("PUT form by organizer", func(t *testing.T) {
		body := map[string]any{
			"title":       "Championship Registration",
			"description": "Please fill out required answers",
			"fields": []map[string]any{
				{
					"id":       "ign",
					"type":     "SHORT_TEXT",
					"label":    "In Game Name",
					"required": true,
					"order":    1,
				},
				{
					"id":       "tier",
					"type":     "SELECT_ONE",
					"label":    "Rank Tier",
					"required": true,
					"order":    2,
					"config": map[string]any{
						"options": []map[string]any{
							{"id": "gold", "label": "Gold"},
							{"id": "plat", "label": "Platinum"},
						},
					},
				},
			},
		}

		resp := api.Put("/api/v1/tournaments/"+tournamentID.String()+"/form", body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var out rest.RegistrationFormResponse
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
		assert.Equal(t, "Championship Registration", out.Title)
		assert.Len(t, out.Fields, 2)
	})

	t.Run("GET form by public", func(t *testing.T) {
		resp := api.Get("/api/v1/tournaments/" + tournamentID.String() + "/form")
		assert.Equal(t, http.StatusOK, resp.Code)

		var out rest.RegistrationFormResponse
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
		assert.Equal(t, "Championship Registration", out.Title)
	})

	t.Run("POST submit form response", func(t *testing.T) {
		body := map[string]any{
			"answers": []map[string]any{
				{"fieldId": "ign", "value": "Faker"},
				{"fieldId": "tier", "value": "gold"},
			},
		}

		resp := api.Post("/api/v1/tournaments/"+tournamentID.String()+"/form/responses", body)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var out rest.FormResponseItem
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
		assert.Equal(t, tournamentID, out.TournamentID)
		assert.Equal(t, participantAccountID, out.AccountID)
		assert.Len(t, out.Answers, 2)
	})

	t.Run("POST duplicate submit response gives 409", func(t *testing.T) {
		body := map[string]any{
			"answers": []map[string]any{
				{"fieldId": "ign", "value": "Faker"},
				{"fieldId": "tier", "value": "gold"},
			},
		}

		resp := api.Post("/api/v1/tournaments/"+tournamentID.String()+"/form/responses", body)
		assert.Equal(t, http.StatusConflict, resp.Code)
	})

	t.Run("GET my form response", func(t *testing.T) {
		resp := api.Get("/api/v1/tournaments/" + tournamentID.String() + "/form/responses/me")
		assert.Equal(t, http.StatusOK, resp.Code)

		var out rest.FormResponseItem
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
		assert.Equal(t, tournamentID, out.TournamentID)
		assert.Equal(t, participantAccountID, out.AccountID)
	})

	t.Run("GET organizer list responses", func(t *testing.T) {
		resp := api.Get("/api/v1/tournaments/" + tournamentID.String() + "/form/responses")
		assert.Equal(t, http.StatusOK, resp.Code)

		var out struct {
			Items []rest.FormResponseItem `json:"items"`
			Total int                     `json:"total"`
		}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &out))
		assert.Equal(t, 1, out.Total)
		assert.Len(t, out.Items, 1)
	})
}
