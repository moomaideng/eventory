package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"gorm.io/gorm"
)

// AccountRepository defines database operations for the Account entity.
type AccountRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	FindByEmail(ctx context.Context, email string) (*models.Account, error)
	FindByUsername(ctx context.Context, username string) (*models.Account, error)
	Create(ctx context.Context, account *models.Account) error
	Update(ctx context.Context, account *models.Account) error
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

// FindByUsername retrieves an account by its unique display username.
func (r *accountRepositoryImpl) FindByUsername(ctx context.Context, username string) (*models.Account, error) {
	var account models.Account
	err := r.db.WithContext(ctx).First(&account, "username = ?", username).Error
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
