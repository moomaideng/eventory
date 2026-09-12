package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/models"
)

type mockAccountRepoForOrganizer struct {
	profiles map[uuid.UUID]*models.OrganizerProfile
}

func (m *mockAccountRepoForOrganizer) FindByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	return nil, nil
}

func (m *mockAccountRepoForOrganizer) FindByEmail(ctx context.Context, email string) (*models.Account, error) {
	return nil, nil
}

func (m *mockAccountRepoForOrganizer) FindByHandle(ctx context.Context, handle string) (*models.Account, error) {
	return nil, nil
}

func (m *mockAccountRepoForOrganizer) Create(ctx context.Context, account *models.Account) error {
	return nil
}

func (m *mockAccountRepoForOrganizer) Update(ctx context.Context, account *models.Account) error {
	return nil
}

func (m *mockAccountRepoForOrganizer) FindOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.OrganizerProfile, error) {
	return m.profiles[accountID], nil
}

func (m *mockAccountRepoForOrganizer) UpsertOrganizerProfile(ctx context.Context, profile *models.OrganizerProfile) (*models.OrganizerProfile, error) {
	return nil, nil
}

func (m *mockAccountRepoForOrganizer) FindSponsorProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.SponsorProfile, error) {
	return nil, nil
}

func (m *mockAccountRepoForOrganizer) UpsertSponsorProfile(ctx context.Context, profile *models.SponsorProfile) (*models.SponsorProfile, error) {
	return nil, nil
}

func TestGetOrganizerProfile_Missing(t *testing.T) {
	_, err := middlewares.GetOrganizerProfile(context.Background())
	if err == nil {
		t.Fatal("expected error when organizer profile is missing from context")
	}
}

func TestGetOrganizerProfile_Present(t *testing.T) {
	profile := &models.OrganizerProfile{
		ID:            uuid.New(),
		OrganizerName: "Alice",
	}
	ctx := context.WithValue(context.Background(), middlewares.OrganizerProfileContextKey, profile)

	extracted, err := middlewares.GetOrganizerProfile(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if extracted.ID != profile.ID || extracted.OrganizerName != "Alice" {
		t.Fatalf("unexpected profile extracted: %+v", extracted)
	}
}

func TestOrganizerMiddleware_HTTPGuard(t *testing.T) {
	organizerAccountID := uuid.New()
	competitorAccountID := uuid.New()

	repo := &mockAccountRepoForOrganizer{
		profiles: map[uuid.UUID]*models.OrganizerProfile{
			organizerAccountID: {
				ID:            uuid.New(),
				AccountID:     organizerAccountID,
				OrganizerName: "Organizer One",
			},
		},
	}

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "1.0.0"))
	orgMiddleware := middlewares.NewOrganizerMiddleware(api, repo)

	group := huma.NewGroup(api, "/organizer-only")
	group.UseMiddleware(orgMiddleware.HumaMiddleware())

	huma.Register(group, huma.Operation{
		OperationID: "test-endpoint",
		Method:      http.MethodGet,
		Path:        "",
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		return &struct{}{}, nil
	})

	// Case 1: Competitor without organizer profile -> 403 Forbidden
	req1 := httptest.NewRequest(http.MethodGet, "/organizer-only", nil)
	ctx1 := context.WithValue(req1.Context(), middlewares.UserSubContextKey, competitorAccountID.String())
	ctx1 = context.WithValue(ctx1, middlewares.UserEmailContextKey, "competitor@example.com")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1.WithContext(ctx1))
	if rec1.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-organizer, got %d", rec1.Code)
	}

	// Case 2: Verified organizer -> 204 or 200 OK
	req2 := httptest.NewRequest(http.MethodGet, "/organizer-only", nil)
	ctx2 := context.WithValue(req2.Context(), middlewares.UserSubContextKey, organizerAccountID.String())
	ctx2 = context.WithValue(ctx2, middlewares.UserEmailContextKey, "organizer@example.com")
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2.WithContext(ctx2))
	if rec2.Code != http.StatusOK && rec2.Code != http.StatusNoContent {
		t.Fatalf("expected 200/204 for organizer, got %d", rec2.Code)
	}
}
