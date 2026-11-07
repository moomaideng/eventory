package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/moomaideng/eventory/services/account/handlers/rest"
	"github.com/moomaideng/eventory/services/account/internal/usecases"
	"github.com/moomaideng/eventory/services/account/models"
)

type mockAccountRepo struct {
	organizerProfiles map[uuid.UUID]*models.OrganizerProfile
}

func (m *mockAccountRepo) FindByID(_ context.Context, _ uuid.UUID) (*models.Account, error) {
	return nil, nil
}
func (m *mockAccountRepo) FindByIDs(_ context.Context, _ []uuid.UUID) ([]models.Account, error) {
	return nil, nil
}
func (m *mockAccountRepo) FindByEmail(_ context.Context, _ string) (*models.Account, error) {
	return nil, nil
}
func (m *mockAccountRepo) FindByHandle(_ context.Context, _ string) (*models.Account, error) {
	return nil, nil
}
func (m *mockAccountRepo) Create(_ context.Context, _ *models.Account) error {
	return nil
}
func (m *mockAccountRepo) Update(_ context.Context, _ *models.Account) error {
	return nil
}
func (m *mockAccountRepo) FindOrganizerProfileByAccountID(_ context.Context, accountID uuid.UUID) (*models.OrganizerProfile, error) {
	return m.organizerProfiles[accountID], nil
}
func (m *mockAccountRepo) FindOrganizerProfilesByIDs(_ context.Context, _ []uuid.UUID) ([]models.OrganizerProfile, error) {
	return nil, nil
}
func (m *mockAccountRepo) UpsertOrganizerProfile(_ context.Context, profile *models.OrganizerProfile) (*models.OrganizerProfile, error) {
	m.organizerProfiles[profile.AccountID] = profile
	return profile, nil
}
func (m *mockAccountRepo) FindSponsorProfileByAccountID(_ context.Context, _ uuid.UUID) (*models.SponsorProfile, error) {
	return nil, nil
}
func (m *mockAccountRepo) UpsertSponsorProfile(_ context.Context, profile *models.SponsorProfile) (*models.SponsorProfile, error) {
	return profile, nil
}

func TestGetOrganizerProfileByID(t *testing.T) {
	repo := &mockAccountRepo{
		organizerProfiles: make(map[uuid.UUID]*models.OrganizerProfile),
	}
	useCase := usecases.NewAccountUseCase(repo)

	organizerID := uuid.New()
	repo.organizerProfiles[organizerID] = &models.OrganizerProfile{
		AccountID:      organizerID,
		OrganizerName:  "Epic Tournaments Ltd",
		OrganizerEmail: "contact@epic.gg",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, rootAPI := humatest.New(t)
	accountGroup := huma.NewGroup(rootAPI, "/api/v1/accounts")
	publicGroup := huma.NewGroup(rootAPI, "/api/v1/accounts")
	rest.RegisterAccountRoutes(accountGroup, publicGroup, useCase)

	t.Run("returns 200 OK for existing organizer", func(t *testing.T) {
		resp := rootAPI.Get("/api/v1/accounts/" + organizerID.String() + "/organizer-profile")
		assert.Equal(t, http.StatusOK, resp.Code)

		var out rest.OrganizerProfileResponse
		err := json.Unmarshal(resp.Body.Bytes(), &out)
		assert.NoError(t, err)
		assert.Equal(t, organizerID, out.AccountID)
		assert.Equal(t, "Epic Tournaments Ltd", out.OrganizerName)
		assert.Equal(t, "contact@epic.gg", out.OrganizerEmail)
	})

	t.Run("returns 404 for non-existent organizer", func(t *testing.T) {
		unknownID := uuid.New()
		resp := rootAPI.Get("/api/v1/accounts/" + unknownID.String() + "/organizer-profile")
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})
}
