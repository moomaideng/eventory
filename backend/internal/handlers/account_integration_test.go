package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/handlers"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/usecases"
)

// mockAccountRepo is an in-memory repository for integration testing.
type mockAccountRepo struct {
	accounts []*models.Account
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

func (m *mockAccountRepo) FindByUsername(ctx context.Context, username string) (*models.Account, error) {
	for _, a := range m.accounts {
		if a.Username == username {
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

// TestAccountDomain_Integration tests the complete integration of the account domain
// across HTTP routing, Huma API endpoints, authentication context, and use case logic.
func TestAccountDomain_Integration(t *testing.T) {
	repo := &mockAccountRepo{}
	useCase := usecases.NewAccountUseCase(repo)

	_, api := humatest.New(t)
	handlers.RegisterAccountRoutes(api, useCase)

	user1Ctx := context.WithValue(context.Background(), middlewares.UserEmailContextKey, "player1@eventory.gg")
	user2Ctx := context.WithValue(context.Background(), middlewares.UserEmailContextKey, "player2@eventory.gg")

	var createdAccountID uuid.UUID

	// 1. Initial State: Unregistered user calls GET /me -> Returns 404 (prompts frontend onboarding)
	t.Run("GET /me returns 404 before onboarding", func(t *testing.T) {
		resp := api.GetCtx(user1Ctx, "/me")
		if resp.Code != http.StatusNotFound {
			t.Errorf("expected status 404 Not Found, got %d", resp.Code)
		}
	})

	// 2. Registration: User completes onboarding via POST /onboard -> Returns 201 Created
	t.Run("POST /onboard registers new account", func(t *testing.T) {
		resp := api.PostCtx(user1Ctx, "/onboard", handlers.OnboardAccountRequest{
			Username: "PlayerOne",
		})

		if resp.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d", resp.Code)
		}

		var created handlers.AccountResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &created); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if created.Username != "PlayerOne" || created.Email != "player1@eventory.gg" {
			t.Errorf("unexpected account data: %+v", created)
		}
		createdAccountID = created.ID
	})

	// 3. Post-Registration: User calls GET /me -> Returns 200 OK with active account
	t.Run("GET /me returns 200 after onboarding", func(t *testing.T) {
		resp := api.GetCtx(user1Ctx, "/me")
		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", resp.Code)
		}

		var account handlers.AccountResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &account); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if account.Username != "PlayerOne" {
			t.Errorf("expected username 'PlayerOne', got %s", account.Username)
		}
	})

	// 4. Duplicate Check: Another user tries taking the same username -> Returns 409 Conflict
	t.Run("POST /onboard with duplicate username returns 409 conflict", func(t *testing.T) {
		resp := api.PostCtx(user2Ctx, "/onboard", handlers.OnboardAccountRequest{
			Username: "PlayerOne",
		})

		if resp.Code != http.StatusConflict {
			t.Errorf("expected status 409 Conflict, got %d", resp.Code)
		}
	})

	// 5. Update Profile: User changes username via PATCH /me -> Returns 200 OK
	t.Run("PATCH /me updates username successfully", func(t *testing.T) {
		resp := api.PatchCtx(user1Ctx, "/me", handlers.UpdateUsernameRequest{
			Username: "PlayerOneUpdated",
		})

		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", resp.Code)
		}

		var updated handlers.AccountResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &updated); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if updated.Username != "PlayerOneUpdated" {
			t.Errorf("expected username 'PlayerOneUpdated', got %s", updated.Username)
		}
	})

	// 6. Public Lookup: Lookup account by ID via GET /{id} -> Returns 200 OK
	t.Run("GET /{id} retrieves public account details", func(t *testing.T) {
		resp := api.Get("/" + createdAccountID.String())
		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", resp.Code)
		}
	})

	// 7. Security: Request without authentication -> Returns 401 Unauthorized
	t.Run("unauthenticated request returns 401 unauthorized", func(t *testing.T) {
		resp := api.Post("/onboard", handlers.OnboardAccountRequest{
			Username: "Stranger",
		})

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401 Unauthorized, got %d", resp.Code)
		}
	})
}
