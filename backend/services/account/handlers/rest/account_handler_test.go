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
	"github.com/moomaideng/eventory/services/account/handlers/rest"
	"github.com/moomaideng/eventory/services/account/internal/usecases"
	"github.com/moomaideng/eventory/services/account/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAccountRepo struct {
	profiles map[uuid.UUID]*models.OrganizerProfile
}

func (s *stubAccountRepo) FindByID(context.Context, uuid.UUID) (*models.Account, error) {
	return nil, nil
}
func (s *stubAccountRepo) FindByIDs(context.Context, []uuid.UUID) ([]models.Account, error) {
	return nil, nil
}
func (s *stubAccountRepo) FindByEmail(context.Context, string) (*models.Account, error) {
	return nil, nil
}
func (s *stubAccountRepo) FindByHandle(context.Context, string) (*models.Account, error) {
	return nil, nil
}
func (s *stubAccountRepo) Create(context.Context, *models.Account) error { return nil }
func (s *stubAccountRepo) Update(context.Context, *models.Account) error { return nil }
func (s *stubAccountRepo) FindOrganizerProfileByAccountID(_ context.Context, accountID uuid.UUID) (*models.OrganizerProfile, error) {
	return s.profiles[accountID], nil
}
func (s *stubAccountRepo) FindOrganizerProfilesByIDs(context.Context, []uuid.UUID) ([]models.OrganizerProfile, error) {
	return nil, nil
}
func (s *stubAccountRepo) UpsertOrganizerProfile(context.Context, *models.OrganizerProfile) (*models.OrganizerProfile, error) {
	return nil, nil
}
func (s *stubAccountRepo) FindSponsorProfileByAccountID(context.Context, uuid.UUID) (*models.SponsorProfile, error) {
	return nil, nil
}
func (s *stubAccountRepo) FindSponsorProfilesByIDs(context.Context, []uuid.UUID) ([]models.SponsorProfile, error) {
	return nil, nil
}
func (s *stubAccountRepo) UpsertSponsorProfile(context.Context, *models.SponsorProfile) (*models.SponsorProfile, error) {
	return nil, nil
}
func (s *stubAccountRepo) ListVisibleSponsorProfiles(context.Context, string, string, int, int) ([]models.SponsorProfile, int64, error) {
	return nil, 0, nil
}
func (s *stubAccountRepo) ListSponsorTags(context.Context) ([]string, error) {
	return nil, nil
}

func TestGetPublicOrganizerProfile(t *testing.T) {
	organizerID := uuid.New()
	phone := "+66 81-234-5678"
	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	repo := &stubAccountRepo{
		profiles: map[uuid.UUID]*models.OrganizerProfile{
			organizerID: {
				AccountID:      organizerID,
				OrganizerName:  "Alex Events",
				OrganizerEmail: "alex@eventory.gg",
				Phone:          &phone,
				CreatedAt:      createdAt,
			},
		},
	}

	_, api := humatest.New(t)
	group := huma.NewGroup(api, "/api/v1/accounts")
	rest.RegisterPublicAccountRoutes(group, usecases.NewAccountUseCase(repo))

	t.Run("returns public organizer profile without auth", func(t *testing.T) {
		resp := api.Get("/api/v1/accounts/organizers/" + organizerID.String())
		require.Equal(t, http.StatusOK, resp.Code)

		var body rest.PublicOrganizerProfileResponse
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
		assert.Equal(t, organizerID, body.AccountID)
		assert.Equal(t, "Alex Events", body.OrganizerName)
		assert.Equal(t, "alex@eventory.gg", body.OrganizerEmail)
		require.NotNil(t, body.Phone)
		assert.Equal(t, phone, *body.Phone)
	})

	t.Run("returns 404 when organizer does not exist", func(t *testing.T) {
		resp := api.Get("/api/v1/accounts/organizers/" + uuid.New().String())
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("returns 422 on invalid UUID", func(t *testing.T) {
		resp := api.Get("/api/v1/accounts/organizers/not-a-uuid")
		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	})
}
