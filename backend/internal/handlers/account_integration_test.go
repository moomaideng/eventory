package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/handlers"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/usecases"
)

// mockAccountRepo is an in-memory repository for integration testing.
type mockAccountRepo struct {
	accounts          []*models.Account
	organizerProfiles []*models.OrganizerProfile
	sponsorProfiles   []*models.SponsorProfile
}

func (m *mockAccountRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	for _, a := range m.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAccountRepo) FindByEmail(ctx context.Context, email string) (*models.Account, error) {
	for _, a := range m.accounts {
		if a.Email == email {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAccountRepo) FindByHandle(ctx context.Context, handle string) (*models.Account, error) {
	for _, a := range m.accounts {
		if a.Handle == handle {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAccountRepo) Create(ctx context.Context, account *models.Account) error {
	m.accounts = append(m.accounts, account)
	return nil
}

func (m *mockAccountRepo) Update(ctx context.Context, account *models.Account) error {
	for i, a := range m.accounts {
		if a.ID == account.ID {
			m.accounts[i] = account
			return nil
		}
	}
	return nil
}

func (m *mockAccountRepo) FindOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.OrganizerProfile, error) {
	for _, p := range m.organizerProfiles {
		if p.AccountID == accountID {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockAccountRepo) UpsertOrganizerProfile(ctx context.Context, profile *models.OrganizerProfile) (*models.OrganizerProfile, error) {
	for i, p := range m.organizerProfiles {
		if p.AccountID == profile.AccountID {
			m.organizerProfiles[i] = profile
			return profile, nil
		}
	}
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	m.organizerProfiles = append(m.organizerProfiles, profile)
	return profile, nil
}

func (m *mockAccountRepo) FindSponsorProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.SponsorProfile, error) {
	for _, p := range m.sponsorProfiles {
		if p.AccountID == accountID {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockAccountRepo) UpsertSponsorProfile(ctx context.Context, profile *models.SponsorProfile) (*models.SponsorProfile, error) {
	for i, p := range m.sponsorProfiles {
		if p.AccountID == profile.AccountID {
			m.sponsorProfiles[i] = profile
			return profile, nil
		}
	}
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	m.sponsorProfiles = append(m.sponsorProfiles, profile)
	return profile, nil
}

func makeAuthContext(userID uuid.UUID, email string) context.Context {
	ctx := context.WithValue(context.Background(), middlewares.UserEmailContextKey, email)
	return context.WithValue(ctx, middlewares.UserSubContextKey, userID.String())
}

// TestAccountDomain_Integration tests the complete integration of the account domain
// across HTTP routing, Huma API endpoints, authentication context, and use case logic.
func TestAccountDomain_Integration(t *testing.T) {
	repo := &mockAccountRepo{}
	useCase := usecases.NewAccountUseCase(repo)

	_, api := humatest.New(t)
	accountGroup := huma.NewGroup(api, "/api/v1/accounts")
	handlers.RegisterAccountRoutes(accountGroup, useCase)

	user1ID := uuid.New()
	user1Email := "player1@eventory.gg"
	user1Ctx := makeAuthContext(user1ID, user1Email)

	user2ID := uuid.New()
	user2Email := "player2@eventory.gg"
	user2Ctx := makeAuthContext(user2ID, user2Email)

	// 1. Initial State: Unregistered user calls GET /me -> Returns 404 (account does not exist yet)
	t.Run("GET /me returns 404 before account creation", func(t *testing.T) {
		resp := api.GetCtx(user1Ctx, "/api/v1/accounts/me")
		if resp.Code != http.StatusNotFound {
			t.Errorf("expected status 404 Not Found, got %d", resp.Code)
		}
	})

	// 2. Provisioning: Authenticated user creates account via POST /api/v1/accounts -> Returns 201 Created
	t.Run("POST / creates new account", func(t *testing.T) {
		requestedHandle := "player_one"
		resp := api.PostCtx(user1Ctx, "/api/v1/accounts", handlers.CreateAccountRequest{
			DisplayName: "Player One",
			Handle:      &requestedHandle,
		})

		if resp.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d: %s", resp.Code, resp.Body.String())
		}

		var created handlers.AccountResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &created); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if created.ID != user1ID || created.Email != user1Email {
			t.Errorf("unexpected account ID or email: %+v", created)
		}
		if created.DisplayName != "Player One" || created.Handle != "player_one" {
			t.Errorf("unexpected displayName or handle: %+v", created)
		}
		if created.Status != "ACTIVE" {
			t.Errorf("expected status ACTIVE, got %s", created.Status)
		}
	})

	// 3. Idempotency: Re-calling POST /api/v1/accounts returns existing account (201 Created)
	t.Run("POST / is idempotent for existing user", func(t *testing.T) {
		resp := api.PostCtx(user1Ctx, "/api/v1/accounts", handlers.CreateAccountRequest{
			DisplayName: "Player One Duplicate Attempt",
		})

		if resp.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d", resp.Code)
		}

		var account handlers.AccountResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &account); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if account.ID != user1ID || account.DisplayName != "Player One" {
			t.Errorf("expected original account data preserved: %+v", account)
		}
	})

	// 4. Post-Creation: User calls GET /me -> Returns 200 OK with created account details
	t.Run("GET /me returns 200 after account creation", func(t *testing.T) {
		resp := api.GetCtx(user1Ctx, "/api/v1/accounts/me")
		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", resp.Code)
		}

		var account handlers.AccountResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &account); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if account.DisplayName != "Player One" || account.Handle != "player_one" {
			t.Errorf("unexpected account details: %+v", account)
		}
	})

	// 5. Update Profile: User updates profile fields via PATCH /me -> Returns 200 OK
	t.Run("PATCH /me updates profile successfully", func(t *testing.T) {
		newDisplayName := "Player One Updated"
		phone := "+66812345678"
		resp := api.PatchCtx(user1Ctx, "/api/v1/accounts/me", handlers.UpdateAccountRequest{
			DisplayName: &newDisplayName,
			Phone:       &phone,
		})

		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", resp.Code)
		}

		var updated handlers.AccountResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &updated); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if updated.DisplayName != "Player One Updated" {
			t.Errorf("expected updated displayName, got %s", updated.DisplayName)
		}
		if updated.Phone == nil || *updated.Phone != phone {
			t.Errorf("expected updated phone, got %v", updated.Phone)
		}
	})

	// 6. Conflict Check: Second user attempts to take existing handle -> Returns 409 Conflict
	t.Run("PATCH /me with taken handle returns 409 conflict", func(t *testing.T) {
		// First, provision second user
		resp2 := api.PostCtx(user2Ctx, "/api/v1/accounts", handlers.CreateAccountRequest{
			DisplayName: "Player Two",
		})
		if resp2.Code != http.StatusCreated {
			t.Fatalf("failed to provision second user: %d", resp2.Code)
		}

		// Try updating user 2's handle to user 1's handle ("player_one")
		takenHandle := "player_one"
		patchResp := api.PatchCtx(user2Ctx, "/api/v1/accounts/me", handlers.UpdateAccountRequest{
			Handle: &takenHandle,
		})

		if patchResp.Code != http.StatusConflict {
			t.Errorf("expected status 409 Conflict, got %d: %s", patchResp.Code, patchResp.Body.String())
		}
	})

	// 7. Public Lookup: Lookup account by UUID via GET /{id} -> Returns 200 OK
	t.Run("GET /{id} retrieves public account details", func(t *testing.T) {
		resp := api.Get("/api/v1/accounts/" + user1ID.String())
		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", resp.Code)
		}
	})

	// 8. Context Profile: Upsert organizer profile via PUT /me/organizer-profile -> Returns 200 OK
	t.Run("PUT /me/organizer-profile upserts organizer profile", func(t *testing.T) {
		orgEmail := "org@eventory.gg"
		resp := api.PutCtx(user1Ctx, "/api/v1/accounts/me/organizer-profile", handlers.UpsertOrganizerProfileRequest{
			OrganizerName:  "Chula Esports Club",
			OrganizerEmail: &orgEmail,
		})

		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d: %s", resp.Code, resp.Body.String())
		}

		var out handlers.OrganizerProfileResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if out.OrganizerName != "Chula Esports Club" {
			t.Errorf("expected organizer name 'Chula Esports Club', got %s", out.OrganizerName)
		}
	})

	// 9. Context Profile: Upsert sponsor profile via PUT /me/sponsor-profile -> Returns 200 OK
	t.Run("PUT /me/sponsor-profile upserts sponsor profile", func(t *testing.T) {
		spEmail := "sponsor@eventory.gg"
		resp := api.PutCtx(user1Ctx, "/api/v1/accounts/me/sponsor-profile", handlers.UpsertSponsorProfileRequest{
			SponsorName:  "Red Bull Gaming",
			SponsorEmail: &spEmail,
		})

		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d: %s", resp.Code, resp.Body.String())
		}

		var out handlers.SponsorProfileResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if out.SponsorName != "Red Bull Gaming" {
			t.Errorf("expected sponsor name 'Red Bull Gaming', got %s", out.SponsorName)
		}
	})

	// 10. Security Guard: Request without authentication -> Returns 401 Unauthorized
	t.Run("unauthenticated request returns 401 unauthorized", func(t *testing.T) {
		resp := api.Post("/api/v1/accounts", handlers.CreateAccountRequest{
			DisplayName: "Anonymous",
		})

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401 Unauthorized, got %d", resp.Code)
		}
	})
}
