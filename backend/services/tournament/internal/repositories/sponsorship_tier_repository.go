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
	Create(ctx context.Context, tier *models.SponsorshipTier) error
	Begin(ctx context.Context) (SponsorshipTierTransaction, error)
}

type SponsorshipTierTransaction interface {
	LockTier(tournamentID, tierID uuid.UUID) (*models.Tournament, *models.SponsorshipTier, error)
	Reserved(tierID uuid.UUID) (int64, error)
	UpdateTier(tier *models.SponsorshipTier) error
	DeleteTier(tierID uuid.UUID) error
	Commit() error
	Rollback()
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
	if err != nil {
		return nil, err
	}
	var counts []struct {
		TierID   uuid.UUID
		Reserved int
	}
	if err := r.db.WithContext(ctx).Model(&models.SponsorshipPledge{}).Select("tier_id, COUNT(*) AS reserved").
		Where("tournament_id = ? AND status IN ?", tournamentID, []models.PledgeStatus{models.PledgePending, models.PledgeConfirmed}).Group("tier_id").Scan(&counts).Error; err != nil {
		return nil, err
	}
	byTier := make(map[uuid.UUID]int, len(counts))
	for _, count := range counts {
		byTier[count.TierID] = count.Reserved
	}
	for i := range tiers {
		tiers[i].ReservedSlots = byTier[tiers[i].ID]
	}
	return tiers, nil
}

func (r *sponsorshipTierRepository) Create(ctx context.Context, tier *models.SponsorshipTier) error {
	return translateTierError(r.db.WithContext(ctx).Omit("Tournament").Create(tier).Error)
}

func (r *sponsorshipTierRepository) Begin(ctx context.Context) (SponsorshipTierTransaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &sponsorshipTransaction{db: tx}, nil
}

func (tx *sponsorshipTransaction) UpdateTier(tier *models.SponsorshipTier) error {
	return translateTierError(tx.db.Omit("Tournament").Save(tier).Error)
}

func (tx *sponsorshipTransaction) DeleteTier(tierID uuid.UUID) error {
	if err := tx.db.Where("tier_id = ?", tierID).Delete(&models.SponsorshipPledge{}).Error; err != nil {
		return err
	}
	return tx.db.Delete(&models.SponsorshipTier{}, "id = ?", tierID).Error
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
