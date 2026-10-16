package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

var (
	ErrSponsorshipTierNotFound      = errors.New("sponsorship tier not found")
	ErrSponsorshipTierNameDuplicate = errors.New("sponsorship tier name already used in this tournament")
)

// SponsorshipTierRepository persists sponsorship tier packages (US5-4).
type SponsorshipTierRepository interface {
	// List returns a tournament's tiers in display order.
	List(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipTier, error)
	Get(ctx context.Context, tournamentID, tierID uuid.UUID) (*models.SponsorshipTier, error)
	Create(ctx context.Context, tier *models.SponsorshipTier) error
	Update(ctx context.Context, tier *models.SponsorshipTier) error
	Delete(ctx context.Context, tournamentID, tierID uuid.UUID) error
}

type sponsorshipTierRepository struct {
	db *gorm.DB
}

func NewSponsorshipTierRepository(db *gorm.DB) SponsorshipTierRepository {
	return &sponsorshipTierRepository{db: db}
}

func (r *sponsorshipTierRepository) List(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipTier, error) {
	var tiers []models.SponsorshipTier
	err := r.db.WithContext(ctx).
		Where("tournament_id = ?", tournamentID).
		Order("sort_order ASC, price DESC, created_at ASC").
		Find(&tiers).Error
	return tiers, err
}

func (r *sponsorshipTierRepository) Get(ctx context.Context, tournamentID, tierID uuid.UUID) (*models.SponsorshipTier, error) {
	var tier models.SponsorshipTier
	err := r.db.WithContext(ctx).
		Where("id = ? AND tournament_id = ?", tierID, tournamentID).
		First(&tier).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSponsorshipTierNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *sponsorshipTierRepository) Create(ctx context.Context, tier *models.SponsorshipTier) error {
	return translateTierError(r.db.WithContext(ctx).Omit("Tournament").Create(tier).Error)
}

func (r *sponsorshipTierRepository) Update(ctx context.Context, tier *models.SponsorshipTier) error {
	return translateTierError(r.db.WithContext(ctx).Omit("Tournament").Save(tier).Error)
}

func (r *sponsorshipTierRepository) Delete(ctx context.Context, tournamentID, tierID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tournament_id = ?", tierID, tournamentID).
		Delete(&models.SponsorshipTier{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSponsorshipTierNotFound
	}
	return nil
}

func translateTierError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "23505") {
		return ErrSponsorshipTierNameDuplicate
	}
	return err
}
