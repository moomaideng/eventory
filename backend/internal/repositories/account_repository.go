package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"gorm.io/gorm"
)

// AccountRepository defines database operations for the Account entity and associated profiles.
type AccountRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	FindByEmail(ctx context.Context, email string) (*models.Account, error)
	FindByHandle(ctx context.Context, handle string) (*models.Account, error)
	Create(ctx context.Context, account *models.Account) error
	Update(ctx context.Context, account *models.Account) error

	FindOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.OrganizerProfile, error)
	UpsertOrganizerProfile(ctx context.Context, profile *models.OrganizerProfile) (*models.OrganizerProfile, error)
	FindSponsorProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.SponsorProfile, error)
	UpsertSponsorProfile(ctx context.Context, profile *models.SponsorProfile) (*models.SponsorProfile, error)
}


// accountRepositoryImpl is the concrete implementation of AccountRepository using GORM.
type accountRepositoryImpl struct {
	db *gorm.DB
}

// NewAccountRepository creates a new instance of AccountRepository implementation.
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepositoryImpl{db: db}
}

// FindByID retrieves an account by its unique UUID.
func (r *accountRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	var account models.Account
	err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// FindByEmail retrieves an account by its unique email address.
func (r *accountRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.Account, error) {
	var account models.Account
	err := r.db.WithContext(ctx).First(&account, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// FindByHandle retrieves an account by its unique handle.
func (r *accountRepositoryImpl) FindByHandle(ctx context.Context, handle string) (*models.Account, error) {
	var account models.Account
	err := r.db.WithContext(ctx).First(&account, "handle = ?", handle).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// Create inserts a new account record into the database.
func (r *accountRepositoryImpl) Create(ctx context.Context, account *models.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// Update saves modifications to an existing account record.
func (r *accountRepositoryImpl) Update(ctx context.Context, account *models.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// FindOrganizerProfileByAccountID retrieves the linked organizer profile for a given account UUID.
func (r *accountRepositoryImpl) FindOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.OrganizerProfile, error) {
	var profile models.OrganizerProfile
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// UpsertOrganizerProfile creates or updates the single organizer profile linked to an account.
func (r *accountRepositoryImpl) UpsertOrganizerProfile(ctx context.Context, profile *models.OrganizerProfile) (*models.OrganizerProfile, error) {
	var existing models.OrganizerProfile
	err := r.db.WithContext(ctx).Where("account_id = ?", profile.AccountID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if profile.ID == uuid.Nil {
				profile.ID = uuid.New()
			}
			if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
				// Retry in case of concurrent insert
				var fallback models.OrganizerProfile
				if rerr := r.db.WithContext(ctx).Where("account_id = ?", profile.AccountID).First(&fallback).Error; rerr == nil {
					fallback.OrganizerName = profile.OrganizerName
					fallback.OrganizerEmail = profile.OrganizerEmail
					if err := r.db.WithContext(ctx).Save(&fallback).Error; err != nil {
						return nil, err
					}
					return &fallback, nil
				}
				return nil, err
			}
			return profile, nil
		}
		return nil, err
	}

	existing.OrganizerName = profile.OrganizerName
	existing.OrganizerEmail = profile.OrganizerEmail
	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

// FindSponsorProfileByAccountID retrieves the linked sponsor profile for a given account UUID.
func (r *accountRepositoryImpl) FindSponsorProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.SponsorProfile, error) {
	var profile models.SponsorProfile
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// UpsertSponsorProfile creates or updates the single sponsor profile linked to an account.
func (r *accountRepositoryImpl) UpsertSponsorProfile(ctx context.Context, profile *models.SponsorProfile) (*models.SponsorProfile, error) {
	var existing models.SponsorProfile
	err := r.db.WithContext(ctx).Where("account_id = ?", profile.AccountID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if profile.ID == uuid.Nil {
				profile.ID = uuid.New()
			}
			if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
				// Retry in case of concurrent insert
				var fallback models.SponsorProfile
				if rerr := r.db.WithContext(ctx).Where("account_id = ?", profile.AccountID).First(&fallback).Error; rerr == nil {
					fallback.SponsorName = profile.SponsorName
					fallback.SponsorEmail = profile.SponsorEmail
					if err := r.db.WithContext(ctx).Save(&fallback).Error; err != nil {
						return nil, err
					}
					return &fallback, nil
				}
				return nil, err
			}
			return profile, nil
		}
		return nil, err
	}

	existing.SponsorName = profile.SponsorName
	existing.SponsorEmail = profile.SponsorEmail
	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

