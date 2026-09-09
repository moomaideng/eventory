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
	accounts          map[uuid.UUID]*models.Account
	organizerProfiles map[uuid.UUID]*models.OrganizerProfile
	sponsorProfiles   map[uuid.UUID]*models.SponsorProfile
}

func newMockAccountRepository() *mockAccountRepository {
	return &mockAccountRepository{
		accounts:          make(map[uuid.UUID]*models.Account),
		organizerProfiles: make(map[uuid.UUID]*models.OrganizerProfile),
		sponsorProfiles:   make(map[uuid.UUID]*models.SponsorProfile),
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

func (m *mockAccountRepository) FindByHandle(ctx context.Context, handle string) (*models.Account, error) {
	for _, acc := range m.accounts {
		if acc.Handle == handle {
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

func (m *mockAccountRepository) FindOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.OrganizerProfile, error) {
	if p, exists := m.organizerProfiles[accountID]; exists {
		return p, nil
	}
	return nil, nil
}

func (m *mockAccountRepository) UpsertOrganizerProfile(ctx context.Context, profile *models.OrganizerProfile) (*models.OrganizerProfile, error) {
	if existing, exists := m.organizerProfiles[profile.AccountID]; exists {
		existing.OrganizerName = profile.OrganizerName
		existing.OrganizerEmail = profile.OrganizerEmail
		return existing, nil
	}
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	m.organizerProfiles[profile.AccountID] = profile
	return profile, nil
}

func (m *mockAccountRepository) FindSponsorProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.SponsorProfile, error) {
	if p, exists := m.sponsorProfiles[accountID]; exists {
		return p, nil
	}
	return nil, nil
}

func (m *mockAccountRepository) UpsertSponsorProfile(ctx context.Context, profile *models.SponsorProfile) (*models.SponsorProfile, error) {
	if existing, exists := m.sponsorProfiles[profile.AccountID]; exists {
		existing.SponsorName = profile.SponsorName
		existing.SponsorEmail = profile.SponsorEmail
		return existing, nil
	}
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	m.sponsorProfiles[profile.AccountID] = profile
	return profile, nil
}

func TestGetAccountByEmail_Found(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	email := "player1@eventory.gg"

	mockRepo.accounts[userID] = &models.Account{
		ID:          userID,
		Email:       email,
		Handle:      "player_one",
		DisplayName: "Player One",
		Status:      "ACTIVE",
	}

	account, err := useCase.GetAccountByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if account.DisplayName != "Player One" {
		t.Errorf("expected displayName 'Player One', got %v", account.DisplayName)
	}
	if account.Handle != "player_one" {
		t.Errorf("expected handle 'player_one', got %v", account.Handle)
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

func TestCreateAccount_Success(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	email := "newuser@eventory.gg"
	displayName := "MooMai"
	requestedHandle := "moomai"
	avatar := "https://example.com/avatar.png"

	account, err := useCase.CreateAccount(context.Background(), usecases.CreateAccountInput{
		ID:          userID,
		Email:       email,
		DisplayName: displayName,
		Handle:      requestedHandle,
		AvatarURL:   &avatar,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if account.ID != userID {
		t.Errorf("expected ID %v, got %v", userID, account.ID)
	}
	if account.Email != email {
		t.Errorf("expected Email %v, got %v", email, account.Email)
	}
	if account.DisplayName != "MooMai" {
		t.Errorf("expected DisplayName 'MooMai', got %v", account.DisplayName)
	}
	if account.Handle != "moomai" {
		t.Errorf("expected Handle 'moomai', got %v", account.Handle)
	}
	if account.AvatarURL == nil || *account.AvatarURL != avatar {
		t.Errorf("expected AvatarURL %v, got %v", avatar, account.AvatarURL)
	}
}

func TestCreateAccount_Idempotent(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	email := "idempotent@eventory.gg"

	acc1, err := useCase.CreateAccount(context.Background(), usecases.CreateAccountInput{
		ID:          userID,
		Email:       email,
		DisplayName: "First Name",
		Handle:      "first_handle",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Second call with same ID or Email should return existing account
	acc2, err := useCase.CreateAccount(context.Background(), usecases.CreateAccountInput{
		ID:          userID,
		Email:       email,
		DisplayName: "Different Name",
		Handle:      "diff_handle",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if acc1.ID != acc2.ID {
		t.Errorf("expected same account ID, got %v and %v", acc1.ID, acc2.ID)
	}
	if acc2.DisplayName != "First Name" {
		t.Errorf("expected original DisplayName 'First Name', got %v", acc2.DisplayName)
	}
}

func TestCreateAccount_HandleCollisionFallback(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	existingID := uuid.New()
	mockRepo.accounts[existingID] = &models.Account{
		ID:          existingID,
		Email:       "existing@eventory.gg",
		Handle:      "alex",
		DisplayName: "Alex Original",
		Status:      "ACTIVE",
	}

	newID := uuid.New()
	account, err := useCase.CreateAccount(context.Background(), usecases.CreateAccountInput{
		ID:          newID,
		Email:       "alex2@eventory.gg",
		DisplayName: "Alex",
		Handle:      "alex",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should not fail! Should auto-generate a fallback handle like alex_XXXX
	if account.Handle == "alex" {
		t.Error("expected different handle due to collision with 'alex'")
	}
	if account.DisplayName != "Alex" {
		t.Errorf("expected DisplayName 'Alex', got %v", account.DisplayName)
	}
}

func TestUpdateAccount_Success(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.accounts[userID] = &models.Account{
		ID:          userID,
		Email:       "ninja@eventory.gg",
		Handle:      "old_handle",
		DisplayName: "Old Name",
		Status:      "ACTIVE",
	}

	newName := "New Shadow Ninja"
	newHandle := "shadow_ninja"
	phone := "+66812345678"

	updated, err := useCase.UpdateAccount(context.Background(), userID, usecases.UpdateAccountInput{
		DisplayName: &newName,
		Handle:      &newHandle,
		Phone:       &phone,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if updated.DisplayName != "New Shadow Ninja" {
		t.Errorf("expected updated DisplayName 'New Shadow Ninja', got %v", updated.DisplayName)
	}
	if updated.Handle != "shadow_ninja" {
		t.Errorf("expected updated Handle 'shadow_ninja', got %v", updated.Handle)
	}
	if updated.Phone == nil || *updated.Phone != phone {
		t.Errorf("expected updated Phone %v, got %v", phone, updated.Phone)
	}
}

func TestUpdateAccount_HandleAlreadyExists(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	user1ID := uuid.New()
	user2ID := uuid.New()

	mockRepo.accounts[user1ID] = &models.Account{
		ID:          user1ID,
		Email:       "player1@eventory.gg",
		Handle:      "player_one",
		DisplayName: "Player One",
		Status:      "ACTIVE",
	}
	mockRepo.accounts[user2ID] = &models.Account{
		ID:          user2ID,
		Email:       "player2@eventory.gg",
		Handle:      "player_two",
		DisplayName: "Player Two",
		Status:      "ACTIVE",
	}

	takenHandle := "player_two"
	_, err := useCase.UpdateAccount(context.Background(), user1ID, usecases.UpdateAccountInput{
		Handle: &takenHandle,
	})
	if err != usecases.ErrHandleAlreadyExists {
		t.Errorf("expected ErrHandleAlreadyExists, got: %v", err)
	}
}

func TestCreateAccount_NilIDFails(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	_, err := useCase.CreateAccount(context.Background(), usecases.CreateAccountInput{
		ID:          uuid.Nil,
		Email:       "test@eventory.gg",
		DisplayName: "Test",
	})
	if err != usecases.ErrInvalidAccountID {
		t.Errorf("expected ErrInvalidAccountID, got: %v", err)
	}
}
