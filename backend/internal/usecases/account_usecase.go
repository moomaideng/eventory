package usecases

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
)

var (
	ErrAccountNotFound          = errors.New("account not found")
	ErrHandleAlreadyExists       = errors.New("handle is already taken")
	ErrInvalidHandle            = errors.New("handle must be between 3 and 32 lowercase alphanumeric characters or underscores")
	ErrInvalidDisplayName       = errors.New("display name cannot be empty")
	ErrInvalidEmail             = errors.New("email cannot be empty")
	ErrInvalidAccountID         = errors.New("valid account id is required")
	ErrOrganizerProfileNotFound = errors.New("organizer profile not found")
	ErrSponsorProfileNotFound   = errors.New("sponsor profile not found")
	ErrInvalidOrganizerName     = errors.New("organizer name cannot be empty")
	ErrInvalidSponsorName       = errors.New("sponsor name cannot be empty")
)

// CreateAccountInput specifies input parameters for creating or ensuring an account.
type CreateAccountInput struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	Handle      string
	AvatarURL   *string
}

// UpdateAccountInput specifies fields that can be updated on an account.
type UpdateAccountInput struct {
	DisplayName *string
	Handle      *string
	Phone       *string
	AvatarURL   *string
}

// UpsertOrganizerProfileInput specifies input parameters for creating or updating an organizer profile.
type UpsertOrganizerProfileInput struct {
	OrganizerName  string
	OrganizerEmail *string
}

// UpsertSponsorProfileInput specifies input parameters for creating or updating a sponsor profile.
type UpsertSponsorProfileInput struct {
	SponsorName  string
	SponsorEmail *string
}


// AccountUseCase handles core business logic for user accounts.
type AccountUseCase struct {
	accountRepo repositories.AccountRepository
}

// NewAccountUseCase creates a new instance of AccountUseCase.
func NewAccountUseCase(accountRepo repositories.AccountRepository) *AccountUseCase {
	return &AccountUseCase{
		accountRepo: accountRepo,
	}
}

// SanitizeHandle cleans an input string into a lowercase alphanumeric handle with underscores.
func SanitizeHandle(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	var sb strings.Builder
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			sb.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '.' || r == '@' {
			sb.WriteRune('_')
		}
	}
	res := strings.Trim(sb.String(), "_")
	if len(res) > 24 {
		res = res[:24]
	}
	if len(res) < 3 {
		res = "player"
	}
	return res
}

// GetAccountByEmail retrieves an account by its unique email address.
func (u *AccountUseCase) GetAccountByEmail(ctx context.Context, email string) (*models.Account, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, ErrInvalidEmail
	}

	account, err := u.accountRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

// GetAccountByID retrieves an account by its internal UUID.
func (u *AccountUseCase) GetAccountByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	account, err := u.accountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

// GetAccountByHandle retrieves an account by its unique public handle.
func (u *AccountUseCase) GetAccountByHandle(ctx context.Context, handle string) (*models.Account, error) {
	handle = strings.ToLower(strings.TrimSpace(handle))
	if handle == "" {
		return nil, ErrInvalidHandle
	}

	account, err := u.accountRepo.FindByHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

// CreateAccount creates or ensures an account exists for the authenticated user (Idempotent JIT provisioning).
func (u *AccountUseCase) CreateAccount(ctx context.Context, input CreateAccountInput) (*models.Account, error) {
	if input.ID == uuid.Nil {
		return nil, ErrInvalidAccountID
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return nil, ErrInvalidEmail
	}

	// 1. Idempotency: return existing account if already created
	if existing, err := u.accountRepo.FindByID(ctx, input.ID); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	if existing, err := u.accountRepo.FindByEmail(ctx, email); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	// 2. Default display name to email prefix if not provided
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = strings.Split(email, "@")[0]
	}
	if len(displayName) > 64 {
		displayName = displayName[:64]
	}

	// 3. Resolve handle (use requested if available, otherwise generate unique)
	handle := u.resolveUniqueHandle(ctx, input.Handle, displayName)

	account := &models.Account{
		ID:          input.ID,
		Email:       email,
		DisplayName: displayName,
		Handle:      handle,
		AvatarURL:   input.AvatarURL,
		Status:      models.AccountStatusOnboarding,
	}

	if err := u.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// resolveUniqueHandle ensures a non-colliding handle using requested or fallback base name.
func (u *AccountUseCase) resolveUniqueHandle(ctx context.Context, requested string, fallbackBase string) string {
	if requested != "" {
		candidate := SanitizeHandle(requested)
		if existing, _ := u.accountRepo.FindByHandle(ctx, candidate); existing == nil {
			return candidate
		}
	}

	base := SanitizeHandle(fallbackBase)
	candidate := base
	for i := 0; i < 5; i++ {
		if existing, _ := u.accountRepo.FindByHandle(ctx, candidate); existing == nil {
			return candidate
		}
		n, _ := rand.Int(rand.Reader, big.NewInt(9000))
		candidate = fmt.Sprintf("%s_%04d", base, n.Int64()+1000)
		if len(candidate) > 32 {
			candidate = candidate[:32]
		}
	}
	return fmt.Sprintf("user_%s", uuid.New().String()[:8])
}

// UpdateAccount updates the profile fields of an existing account.
func (u *AccountUseCase) UpdateAccount(ctx context.Context, id uuid.UUID, input UpdateAccountInput) (*models.Account, error) {
	account, err := u.accountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	if input.DisplayName != nil {
		dName := strings.TrimSpace(*input.DisplayName)
		if dName == "" {
			return nil, ErrInvalidDisplayName
		}
		if len(dName) > 64 {
			dName = dName[:64]
		}
		account.DisplayName = dName
	}

	if input.Handle != nil {
		h := SanitizeHandle(*input.Handle)
		if len(h) < 3 || len(h) > 32 {
			return nil, ErrInvalidHandle
		}
		if h != account.Handle {
			existing, err := u.accountRepo.FindByHandle(ctx, h)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != id {
				return nil, ErrHandleAlreadyExists
			}
			account.Handle = h
		}
	}

	if input.Phone != nil {
		account.Phone = cleanOptionalString(*input.Phone)
	}

	if input.AvatarURL != nil {
		account.AvatarURL = cleanOptionalString(*input.AvatarURL)
	}

	if account.Status == models.AccountStatusOnboarding {
		account.Status = models.AccountStatusActive
	}

	if err := u.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

func cleanOptionalString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// GetOrganizerProfile retrieves the organizer profile linked to an account UUID.
func (u *AccountUseCase) GetOrganizerProfile(ctx context.Context, accountID uuid.UUID) (*models.OrganizerProfile, error) {
	profile, err := u.accountRepo.FindOrganizerProfileByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrOrganizerProfileNotFound
	}
	return profile, nil
}

// UpsertOrganizerProfile creates or updates the organizer profile linked to an account UUID.
func (u *AccountUseCase) UpsertOrganizerProfile(ctx context.Context, accountID uuid.UUID, input UpsertOrganizerProfileInput) (*models.OrganizerProfile, error) {
	account, err := u.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	name := strings.TrimSpace(input.OrganizerName)
	if name == "" {
		return nil, ErrInvalidOrganizerName
	}
	if len(name) > 150 {
		name = name[:150]
	}

	email := account.Email
	if input.OrganizerEmail != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*input.OrganizerEmail))
		if trimmed != "" {
			email = trimmed
		}
	}

	profile := &models.OrganizerProfile{
		AccountID:      accountID,
		OrganizerName:  name,
		OrganizerEmail: email,
	}

	return u.accountRepo.UpsertOrganizerProfile(ctx, profile)
}

// GetSponsorProfile retrieves the sponsor profile linked to an account UUID.
func (u *AccountUseCase) GetSponsorProfile(ctx context.Context, accountID uuid.UUID) (*models.SponsorProfile, error) {
	profile, err := u.accountRepo.FindSponsorProfileByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrSponsorProfileNotFound
	}
	return profile, nil
}

// UpsertSponsorProfile creates or updates the sponsor profile linked to an account UUID.
func (u *AccountUseCase) UpsertSponsorProfile(ctx context.Context, accountID uuid.UUID, input UpsertSponsorProfileInput) (*models.SponsorProfile, error) {
	account, err := u.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}

	name := strings.TrimSpace(input.SponsorName)
	if name == "" {
		return nil, ErrInvalidSponsorName
	}
	if len(name) > 150 {
		name = name[:150]
	}

	email := account.Email
	if input.SponsorEmail != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*input.SponsorEmail))
		if trimmed != "" {
			email = trimmed
		}
	}

	profile := &models.SponsorProfile{
		AccountID:    accountID,
		SponsorName:  name,
		SponsorEmail: email,
	}

	return u.accountRepo.UpsertSponsorProfile(ctx, profile)
}

