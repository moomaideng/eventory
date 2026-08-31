package usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/usecases"
)

// mockAccountRepository is an in-memory mock implementation of AccountRepository for unit tests.
type mockAccountRepository struct {
	accounts map[uuid.UUID]*models.Account
}

func newMockAccountRepository() *mockAccountRepository {
	return &mockAccountRepository{
		accounts: make(map[uuid.UUID]*models.Account),
	}
}

func (m *mockAccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	if acc, exists := m.accounts[id]; exists {
		return acc, nil
	}
	return nil, nil
}

func (m *mockAccountRepository) FindByEmail(ctx context.Context, email string) (*models.Account, error) {
	for _, acc := range m.accounts {
		if acc.Email == email {
			return acc, nil
		}
	}
	return nil, nil
}

func (m *mockAccountRepository) Create(ctx context.Context, account *models.Account) error {
	m.accounts[account.ID] = account
	return nil
}

func (m *mockAccountRepository) Update(ctx context.Context, account *models.Account) error {
	m.accounts[account.ID] = account
	return nil
}

func TestGetAccountByEmail_Found(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	email := "player1@eventory.gg"

	mockRepo.accounts[userID] = &models.Account{
		ID:       userID,
		Email:    email,
		Username: "PlayerOne",
		Status:   "ACTIVE",
	}

	account, err := useCase.GetAccountByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if account.Username != "PlayerOne" {
		t.Errorf("expected username 'PlayerOne', got %v", account.Username)
	}
}

func TestGetAccountByEmail_NotFound(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	_, err := useCase.GetAccountByEmail(context.Background(), "unknown@eventory.gg")
	if err != usecases.ErrAccountNotFound {
		t.Errorf("expected ErrAccountNotFound, got: %v", err)
	}
}

func TestOnboardAccount_Success(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	email := "newuser@eventory.gg"
	username := "MooMai"

	account, err := useCase.OnboardAccount(context.Background(), email, username)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if account.Email != email {
		t.Errorf("expected Email %v, got %v", email, account.Email)
	}
	if account.Username != "MooMai" {
		t.Errorf("expected Username 'MooMai', got %v", account.Username)
	}
	if account.ID == uuid.Nil {
		t.Error("expected non-nil internal UUID generated for new account")
	}
}

func TestOnboardAccount_AlreadyExists(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	email := "existing@eventory.gg"
	mockRepo.accounts[uuid.New()] = &models.Account{
		ID:       uuid.New(),
		Email:    email,
		Username: "ExistingUser",
		Status:   "ACTIVE",
	}

	_, err := useCase.OnboardAccount(context.Background(), email, "AnotherName")
	if err != usecases.ErrAccountAlreadyExists {
		t.Errorf("expected ErrAccountAlreadyExists, got: %v", err)
	}
}

func TestUpdateUsername_Success(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.accounts[userID] = &models.Account{
		ID:       userID,
		Email:    "ninja@eventory.gg",
		Username: "OldName",
		Status:   "ACTIVE",
	}

	updated, err := useCase.UpdateUsername(context.Background(), userID, "NewShadowNinja")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if updated.Username != "NewShadowNinja" {
		t.Errorf("expected updated username 'NewShadowNinja', got %v", updated.Username)
	}
}
