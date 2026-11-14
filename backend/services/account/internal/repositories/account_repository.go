package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/account/models"
	"gorm.io/gorm"
)

// AccountRepository defines database operations for the Account entity and associated profiles.
type AccountRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Account, error)
	FindByEmail(ctx context.Context, email string) (*models.Account, error)
	FindByHandle(ctx context.Context, handle string) (*models.Account, error)
	Create(ctx context.Context, account *models.Account) error
	Update(ctx context.Context, account *models.Account) error

	FindOrganizerProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.OrganizerProfile, error)
	FindOrganizerProfilesByIDs(ctx context.Context, ids []uuid.UUID) ([]models.OrganizerProfile, error)
	UpsertOrganizerProfile(ctx context.Context, profile *models.OrganizerProfile) (*models.OrganizerProfile, error)
	FindSponsorProfileByAccountID(ctx context.Context, accountID uuid.UUID) (*models.SponsorProfile, error)
	FindSponsorProfilesByIDs(ctx context.Context, ids []uuid.UUID) ([]models.SponsorProfile, error)
	UpsertSponsorProfile(ctx context.Context, profile *models.SponsorProfile) (*models.SponsorProfile, error)
	ListVisibleSponsorProfiles(ctx context.Context, search, tag string, page, pageSize int) ([]models.SponsorProfile, int64, error)
	ListSponsorTags(ctx context.Context) ([]string, error)
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

// FindByIDs retrieves accounts matching the given UUIDs. Missing IDs are omitted.
func (r *accountRepositoryImpl) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Account, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var accounts []models.Account
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
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

// FindOrganizerProfilesByIDs retrieves organizer profiles matching the given account IDs.
func (r *accountRepositoryImpl) FindOrganizerProfilesByIDs(ctx context.Context, ids []uuid.UUID) ([]models.OrganizerProfile, error) {
	return findByAccountIDs[models.OrganizerProfile](ctx, r.db, ids)
}

// FindSponsorProfilesByIDs retrieves sponsor profiles matching the given account IDs.
func (r *accountRepositoryImpl) FindSponsorProfilesByIDs(ctx context.Context, ids []uuid.UUID) ([]models.SponsorProfile, error) {
	return findByAccountIDs[models.SponsorProfile](ctx, r.db, ids)
}

func findByAccountIDs[T any](ctx context.Context, db *gorm.DB, ids []uuid.UUID) ([]T, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var profiles []T
	if err := db.WithContext(ctx).Where("account_id IN ?", ids).Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// UpsertOrganizerProfile creates or updates the single organizer profile linked to an account.
func (r *accountRepositoryImpl) UpsertOrganizerProfile(ctx context.Context, profile *models.OrganizerProfile) (*models.OrganizerProfile, error) {
	var existing models.OrganizerProfile
	err := r.db.WithContext(ctx).Where("account_id = ?", profile.AccountID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
				// Retry in case of concurrent insert
				var fallback models.OrganizerProfile
				if rerr := r.db.WithContext(ctx).Where("account_id = ?", profile.AccountID).First(&fallback).Error; rerr == nil {
					fallback.OrganizerName = profile.OrganizerName
					fallback.OrganizerEmail = profile.OrganizerEmail
					fallback.Phone = profile.Phone
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
	existing.Phone = profile.Phone
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
			if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
				// Retry in case of concurrent insert
				var fallback models.SponsorProfile
				if rerr := r.db.WithContext(ctx).Where("account_id = ?", profile.AccountID).First(&fallback).Error; rerr == nil {
					fallback.SponsorName = profile.SponsorName
					fallback.SponsorEmail = profile.SponsorEmail
					fallback.Phone = profile.Phone
					fallback.WebsiteURL = profile.WebsiteURL
					fallback.LogoURL = profile.LogoURL
					fallback.Tags = profile.Tags
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
	existing.Phone = profile.Phone
	existing.WebsiteURL = profile.WebsiteURL
	existing.LogoURL = profile.LogoURL
	existing.Tags = profile.Tags
	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

// ListVisibleSponsorProfiles returns only active sponsor profiles with a contact address.
func (r *accountRepositoryImpl) ListVisibleSponsorProfiles(ctx context.Context, search, tag string, page, pageSize int) ([]models.SponsorProfile, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.SponsorProfile{}).
		Joins("JOIN accounts ON accounts.id = sponsor_profiles.account_id").
		Where("sponsor_profiles.sponsor_email <> ? AND accounts.status = ?", "", models.AccountStatusActive)
	if search != "" {
		// Escape LIKE wildcards so the search box treats input as plain text.
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.ToLower(search))
		query = query.Where(`(LOWER(sponsor_profiles.sponsor_name) LIKE ? ESCAPE '\' OR EXISTS (SELECT 1 FROM jsonb_array_elements_text(sponsor_profiles.tags) AS t(value) WHERE LOWER(t.value) LIKE ? ESCAPE '\'))`, "%"+escaped+"%", "%"+escaped+"%")
	}
	for _, selected := range strings.Split(tag, ",") {
		if selected = strings.TrimSpace(selected); selected != "" {
			query = query.Where("jsonb_exists(sponsor_profiles.tags, ?)", strings.ToLower(selected))
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var profiles []models.SponsorProfile
	if err := query.Order("sponsor_profiles.sponsor_name ASC, sponsor_profiles.account_id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&profiles).Error; err != nil {
		return nil, 0, err
	}
	return profiles, total, nil
}

func (r *accountRepositoryImpl) ListSponsorTags(ctx context.Context) ([]string, error) {
	var tags []string
	err := r.db.WithContext(ctx).Raw(`SELECT DISTINCT jsonb_array_elements_text(s.tags) AS tag
 FROM sponsor_profiles s JOIN accounts a ON a.id = s.account_id
 WHERE s.sponsor_email <> '' AND a.status = ? ORDER BY tag`, models.AccountStatusActive).Scan(&tags).Error
	return tags, err
}
