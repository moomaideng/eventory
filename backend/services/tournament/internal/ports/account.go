package ports

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrOrganizerNotFound  = errors.New("organizer profile not found")
	ErrSponsorNotFound    = errors.New("sponsor profile not found")
	ErrAccountOnboarding  = errors.New("onboarding required")
	ErrAccountUnavailable = errors.New("account service unavailable")
)

type AccountStatus string

const (
	AccountStatusOnboarding AccountStatus = "ONBOARDING"
	AccountStatusActive     AccountStatus = "ACTIVE"
	AccountStatusSuspended  AccountStatus = "SUSPENDED"
)

type AccountInfo struct {
	ID          uuid.UUID
	Handle      string
	DisplayName string
	Status      AccountStatus
	Email       string
}

type OrganizerInfo struct {
	AccountID     uuid.UUID
	OrganizerName string
}

type SponsorInfo struct {
	AccountID   uuid.UUID
	SponsorName string
	WebsiteURL  string
	LogoURL     string
}

type AccountService interface {
	GetAccount(ctx context.Context, id uuid.UUID) (*AccountInfo, error)
	BatchGetAccounts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]AccountInfo, error)
	BatchGetOrganizerProfiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]OrganizerInfo, error)
	GetOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*OrganizerInfo, error)
	BatchGetSponsorProfiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]SponsorInfo, error)
	GetSponsorProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*SponsorInfo, error)
	GetAccountByHandle(ctx context.Context, handle string) (*AccountInfo, error)
}
