package usecases_test

import (
	"context"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/account/internal/usecases"
	"github.com/moomaideng/eventory/services/account/models"
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

func (m *mockAccountRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Account, error) {
	out := make([]models.Account, 0, len(ids))
	for _, id := range ids {
		if acc, exists := m.accounts[id]; exists {
			out = append(out, *acc)
		}
	}
	return out, nil
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

func (m *mockAccountRepository) FindOrganizerProfilesByIDs(ctx context.Context, ids []uuid.UUID) ([]models.OrganizerProfile, error) {
	out := make([]models.OrganizerProfile, 0, len(ids))
	for _, profile := range m.organizerProfiles {
		for _, id := range ids {
			if profile.AccountID == id {
				out = append(out, *profile)
				break
			}
		}
	}
	return out, nil
}

func (m *mockAccountRepository) UpsertOrganizerProfile(ctx context.Context, profile *models.OrganizerProfile) (*models.OrganizerProfile, error) {
	if existing, exists := m.organizerProfiles[profile.AccountID]; exists {
		existing.OrganizerName = profile.OrganizerName
		existing.OrganizerEmail = profile.OrganizerEmail
		return existing, nil
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

func (m *mockAccountRepository) FindSponsorProfilesByIDs(ctx context.Context, ids []uuid.UUID) ([]models.SponsorProfile, error) {
	var out []models.SponsorProfile
	for _, id := range ids {
		if p, exists := m.sponsorProfiles[id]; exists {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (m *mockAccountRepository) UpsertSponsorProfile(ctx context.Context, profile *models.SponsorProfile) (*models.SponsorProfile, error) {
	if existing, exists := m.sponsorProfiles[profile.AccountID]; exists {
		existing.SponsorName = profile.SponsorName
		existing.SponsorEmail = profile.SponsorEmail
		existing.Phone = profile.Phone
		existing.WebsiteURL = profile.WebsiteURL
		existing.LogoURL = profile.LogoURL
		existing.Tags = profile.Tags
		return existing, nil
	}
	m.sponsorProfiles[profile.AccountID] = profile
	return profile, nil
}

func (m *mockAccountRepository) ListVisibleSponsorProfiles(_ context.Context, search, category string, page, pageSize int) ([]models.SponsorProfile, int64, error) {
	var items []models.SponsorProfile
	for _, profile := range m.sponsorProfiles {
		account := m.accounts[profile.AccountID]
		if account == nil || account.Status != models.AccountStatusActive || profile.SponsorEmail == "" {
			continue
		}
		matchesSearch := strings.Contains(strings.ToLower(profile.SponsorName), strings.ToLower(search))
		for _, tag := range profile.Tags {
			matchesSearch = matchesSearch || strings.Contains(tag, strings.ToLower(search))
		}
		if search != "" && !matchesSearch {
			continue
		}
		matchesTags := true
		for _, tag := range strings.Split(category, ",") {
			if tag != "" && !slices.Contains(profile.Tags, strings.ToLower(tag)) {
				matchesTags = false
			}
		}
		if !matchesTags {
			continue
		}
		items = append(items, *profile)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].SponsorName < items[j].SponsorName })
	total := int64(len(items))
	start := (page - 1) * pageSize
	if start >= len(items) {
		return nil, total, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func (m *mockAccountRepository) ListSponsorTags(_ context.Context) ([]string, error) {
	seen := map[string]bool{}
	for _, profile := range m.sponsorProfiles {
		account := m.accounts[profile.AccountID]
		if account != nil && account.Status == models.AccountStatusActive && profile.SponsorEmail != "" {
			for _, tag := range profile.Tags {
				seen[tag] = true
			}
		}
	}
	var categories []string
	for category := range seen {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	return categories, nil
}

func TestSponsorDirectoryListsProfilesAndRequiresOrganizer(t *testing.T) {
	ctx := context.Background()
	repo := newMockAccountRepository()
	uc := usecases.NewAccountUseCase(repo)
	organizerID, sponsorID, suspendedID := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{organizerID, sponsorID, suspendedID} {
		repo.accounts[id] = &models.Account{ID: id, Email: "test@example.com", Status: models.AccountStatusActive}
	}
	repo.organizerProfiles[organizerID] = &models.OrganizerProfile{AccountID: organizerID, OrganizerName: "Organizer"}
	for _, id := range []uuid.UUID{sponsorID, suspendedID} {
		_, err := uc.UpsertSponsorProfile(ctx, id, usecases.UpsertSponsorProfileInput{SponsorName: "Test Brand", Tags: []string{" Valorant ", "CHESS", "valorant"}})
		if err != nil {
			t.Fatal(err)
		}
	}
	repo.accounts[suspendedID].Status = models.AccountStatusSuspended
	page, err := uc.ListSponsorDirectory(ctx, organizerID, "test", "CHESS", 1, 12)
	if err != nil || page.Total != 1 || len(page.Items) != 1 || page.Items[0].AccountID != sponsorID {
		t.Fatalf("expected new profile immediately visible, got %+v %v", page, err)
	}
	if !slices.Equal(page.Items[0].Tags, []string{"valorant", "chess"}) {
		t.Fatalf("tags not normalized: %v", page.Items[0].Tags)
	}
	if len(page.Tags) != 2 {
		t.Fatalf("expected two tag filters: %v", page.Tags)
	}
	page, err = uc.ListSponsorDirectory(ctx, organizerID, "#VALO", "#CHESS, valorant,CHESS", 1, 12)
	if err != nil || page.Total != 1 {
		t.Fatalf("hashtag search and normalized AND filters should match: %+v %v", page, err)
	}
	page, err = uc.ListSponsorDirectory(ctx, organizerID, "", "chess,volleyball", 1, 12)
	if err != nil || page.Total != 0 || len(page.Tags) != 2 {
		t.Fatalf("no AND matches must still return the full tag list: %+v %v", page, err)
	}
	if _, err := uc.ListSponsorDirectory(ctx, organizerID, "", strings.Repeat("a", 41), 1, 12); err != usecases.ErrInvalidSponsorDirectory {
		t.Fatalf("invalid tag filter should be rejected: %v", err)
	}
	if _, err := uc.ListSponsorDirectory(ctx, sponsorID, "", "", 1, 12); err != usecases.ErrSponsorDirectoryForbidden {
		t.Fatalf("expected organizer guard: %v", err)
	}
	email := "updated@example.com"
	_, err = uc.UpsertSponsorProfile(ctx, sponsorID, usecases.UpsertSponsorProfileInput{SponsorName: "New Brand", SponsorEmail: &email, Tags: []string{"volleyball"}})
	if err != nil {
		t.Fatal(err)
	}
	page, err = uc.ListSponsorDirectory(ctx, organizerID, "new", "volleyball", 1, 12)
	if err != nil || page.Total != 1 || page.Items[0].SponsorEmail != email {
		t.Fatalf("updated profile should remain visible: %+v %v", page, err)
	}
	page, err = uc.ListSponsorDirectory(ctx, organizerID, "", "chess", 1, 12)
	if err != nil || page.Total != 0 {
		t.Fatalf("old tags should no longer match: %+v %v", page, err)
	}
}

func TestSponsorProfileRequiresValidEmailAndTags(t *testing.T) {
	for _, tc := range []struct {
		name, email string
		tags        []string
	}{
		{"empty tags", "brand@example.com", nil},
		{"blank tags", "brand@example.com", []string{"  "}},
		{"long tag", "brand@example.com", []string{strings.Repeat("a", 41)}},
		{"many tags", "brand@example.com", []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"}},
		{"bad email", "invalid", []string{"chess"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newMockAccountRepository()
			id := uuid.New()
			repo.accounts[id] = &models.Account{ID: id, Email: tc.email}
			_, err := usecases.NewAccountUseCase(repo).UpsertSponsorProfile(context.Background(), id, usecases.UpsertSponsorProfileInput{SponsorName: "Brand", SponsorEmail: &tc.email, Tags: tc.tags})
			if err != usecases.ErrInvalidSponsorDirectory {
				t.Fatalf("expected validation error, got %v", err)
			}
			if len(repo.sponsorProfiles) != 0 {
				t.Fatal("invalid profile was saved")
			}
		})
	}
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

func TestGetAccountByHandle_Success(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.accounts[userID] = &models.Account{
		ID:          userID,
		Email:       "player1@eventory.gg",
		Handle:      "player_one",
		DisplayName: "Player One",
		Status:      models.AccountStatusActive,
	}

	account, err := useCase.GetAccountByHandle(context.Background(), "player_one")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if account.ID != userID {
		t.Errorf("expected id %v, got %v", userID, account.ID)
	}
	if account.Handle != "player_one" {
		t.Errorf("expected handle 'player_one', got %v", account.Handle)
	}
}

func TestGetAccountByHandle_NotFound(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	_, err := useCase.GetAccountByHandle(context.Background(), "unknown_handle")
	if err != usecases.ErrAccountNotFound {
		t.Errorf("expected ErrAccountNotFound, got: %v", err)
	}
}

func TestGetAccountByHandle_InvalidHandle(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	_, err := useCase.GetAccountByHandle(context.Background(), "   ")
	if err != usecases.ErrInvalidHandle {
		t.Errorf("expected ErrInvalidHandle, got: %v", err)
	}
}

func TestCreateAccount_Success(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	email := "newuser@eventory.gg"

	account, err := useCase.CreateAccount(context.Background(), usecases.CreateAccountInput{
		ID:    userID,
		Email: email,
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
	if account.DisplayName != "newuser" {
		t.Errorf("expected DisplayName 'newuser', got %v", account.DisplayName)
	}
	if account.Handle == "" {
		t.Errorf("expected non-empty handle, got %v", account.Handle)
	}
	if account.Status != models.AccountStatusOnboarding {
		t.Errorf("expected Status %v, got %v", models.AccountStatusOnboarding, account.Status)
	}
}

func TestCreateAccount_Idempotent(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	email := "idempotent@eventory.gg"

	acc1, err := useCase.CreateAccount(context.Background(), usecases.CreateAccountInput{
		ID:    userID,
		Email: email,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Second call with same ID or Email should return existing account
	acc2, err := useCase.CreateAccount(context.Background(), usecases.CreateAccountInput{
		ID:    userID,
		Email: email,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if acc1.ID != acc2.ID {
		t.Errorf("expected same account ID, got %v and %v", acc1.ID, acc2.ID)
	}
	if acc2.DisplayName != "idempotent" {
		t.Errorf("expected original DisplayName 'idempotent', got %v", acc2.DisplayName)
	}
}

func TestCreateAccount_HandleCollisionFallback(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	existingID := uuid.New()
	mockRepo.accounts[existingID] = &models.Account{
		ID:          existingID,
		Email:       "alex@eventory.gg",
		Handle:      "alex",
		DisplayName: "alex",
		Status:      "ACTIVE",
	}

	newID := uuid.New()
	account, err := useCase.CreateAccount(context.Background(), usecases.CreateAccountInput{
		ID:    newID,
		Email: "alex@different.gg",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should not fail! Should auto-generate a fallback handle like alex_XXXX
	if account.Handle == "alex" {
		t.Error("expected different handle due to collision with 'alex'")
	}
	if account.DisplayName != "alex" {
		t.Errorf("expected DisplayName 'alex', got %v", account.DisplayName)
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

	updated, err := useCase.UpdateAccount(context.Background(), userID, usecases.UpdateAccountInput{
		DisplayName: &newName,
		Handle:      &newHandle,
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
}

func TestUpdateAccount_OnboardingToActive(t *testing.T) {
	mockRepo := newMockAccountRepository()
	useCase := usecases.NewAccountUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.accounts[userID] = &models.Account{
		ID:          userID,
		Email:       "newplayer@eventory.gg",
		Handle:      "user_1024",
		DisplayName: "New Player",
		Status:      models.AccountStatusOnboarding,
	}

	newHandle := "pro_player"
	updated, err := useCase.UpdateAccount(context.Background(), userID, usecases.UpdateAccountInput{
		Handle: &newHandle,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if updated.Status != models.AccountStatusActive {
		t.Errorf("expected status %v, got %v", models.AccountStatusActive, updated.Status)
	}
	if updated.Handle != "pro_player" {
		t.Errorf("expected handle 'pro_player', got %v", updated.Handle)
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
		ID:    uuid.Nil,
		Email: "test@eventory.gg",
	})
	if err != usecases.ErrInvalidAccountID {
		t.Errorf("expected ErrInvalidAccountID, got: %v", err)
	}
}

func TestBatchGetSponsorProfiles(t *testing.T) {
	repo := newMockAccountRepository()
	uc := usecases.NewAccountUseCase(repo)
	foundID := uuid.New()
	missingID := uuid.New()
	repo.sponsorProfiles[foundID] = &models.SponsorProfile{
		AccountID:   foundID,
		SponsorName: "Apex Sponsors",
		WebsiteURL:  "https://apex.example",
		LogoURL:     "https://cdn.example/apex.png",
	}

	found, missing, err := uc.BatchGetSponsorProfiles(context.Background(), []uuid.UUID{foundID, foundID, missingID, uuid.Nil})
	if err != nil {
		t.Fatalf("BatchGetSponsorProfiles() error = %v", err)
	}
	if len(found) != 1 || found[0].AccountID != foundID || found[0].SponsorName != "Apex Sponsors" {
		t.Fatalf("unexpected found profiles: %+v", found)
	}
	if len(missing) != 1 || missing[0] != missingID {
		t.Fatalf("unexpected missing IDs: %+v", missing)
	}
}
