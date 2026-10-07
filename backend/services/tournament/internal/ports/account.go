package ports

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrOrganizerNotFound  = errors.New("organizer profile not found")
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
}

type OrganizerInfo struct {
	ID            uuid.UUID
	AccountID     uuid.UUID
	OrganizerName string
}

type AccountService interface {
	GetAccount(ctx context.Context, id uuid.UUID) (*AccountInfo, error)
	BatchGetAccounts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]AccountInfo, error)
	BatchGetOrganizerProfiles(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]OrganizerInfo, error)
	GetOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*OrganizerInfo, error)
}
