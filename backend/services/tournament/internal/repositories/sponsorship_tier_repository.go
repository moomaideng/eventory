package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrSponsorshipTierNotFound      = errors.New("sponsorship tier not found")
	ErrSponsorshipTierNameDuplicate = errors.New("sponsorship tier name already used in this tournament")
	ErrSponsorshipTierInUse         = errors.New("this package has a pending or confirmed sponsorship and cannot be changed")
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
		tiers[i].AvailableSlots = max(tiers[i].Slots-byTier[tiers[i].ID], 0)
		tiers[i].Locked = byTier[tiers[i].ID] > 0
	}
	return tiers, nil
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
	return translateTierError(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUnusedTier(tx, tier.TournamentID, tier.ID); err != nil {
			return err
		}
		return tx.Omit("Tournament").Save(tier).Error
	}))
}

func (r *sponsorshipTierRepository) Delete(ctx context.Context, tournamentID, tierID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockUnusedTier(tx, tournamentID, tierID); err != nil {
			return err
		}
		// Closed, unpaid attempts no longer reserve capacity or need their draft logos.
		if err := tx.Where("tier_id = ?", tierID).Delete(&models.SponsorshipPledge{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.SponsorshipTier{}, "id = ?", tierID).Error
	})
}

func lockUnusedTier(tx *gorm.DB, tournamentID, tierID uuid.UUID) error {
	var tier models.SponsorshipTier
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&tier, "id = ? AND tournament_id = ?", tierID, tournamentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSponsorshipTierNotFound
		}
		return err
	}
	var count int64
	if err := tx.Model(&models.SponsorshipPledge{}).Where("tier_id = ? AND status IN ?", tierID, []models.PledgeStatus{models.PledgePending, models.PledgeConfirmed}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrSponsorshipTierInUse
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
