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
)

type SponsorshipTierRow struct {
	models.SponsorshipTier
	ReservedSlots int
}

// SponsorshipTierRepository persists sponsorship tier packages (US5-4).
type SponsorshipTierRepository interface {
	List(ctx context.Context, tournamentID uuid.UUID) ([]SponsorshipTierRow, error)
	Create(ctx context.Context, tier *models.SponsorshipTier) error
	Begin(ctx context.Context) (SponsorshipTierTransaction, error)
}

type SponsorshipTierTransaction interface {
	LockTier(tournamentID, tierID uuid.UUID) (*models.Tournament, *models.SponsorshipTier, error)
	Reserved(tierID uuid.UUID) (int64, error)
	HasPledges(tierID uuid.UUID) (bool, error)
	UpdateTier(tier *models.SponsorshipTier) error
	DeleteTier(tierID uuid.UUID) error
	Commit() error
	Rollback()
}

type sponsorshipTierRepository struct{ db *gorm.DB }

func NewSponsorshipTierRepository(db *gorm.DB) SponsorshipTierRepository {
	return &sponsorshipTierRepository{db: db}
}

func (r *sponsorshipTierRepository) List(ctx context.Context, tournamentID uuid.UUID) ([]SponsorshipTierRow, error) {
	var tiers []models.SponsorshipTier
	if err := r.db.WithContext(ctx).
		Where("tournament_id = ?", tournamentID).
		Order("price_minor DESC, name ASC, id ASC").
		Find(&tiers).Error; err != nil {
		return nil, err
	}
	var counts []struct {
		TierID uuid.UUID
		Count  int
	}
	if err := r.db.WithContext(ctx).Model(&models.SponsorshipPledge{}).
		Select("tier_id, COUNT(*) AS count").
		Where("tournament_id = ? AND status IN ?", tournamentID, []models.PledgeStatus{models.PledgePending, models.PledgeConfirmed}).
		Group("tier_id").Scan(&counts).Error; err != nil {
		return nil, err
	}
	reserved := make(map[uuid.UUID]int, len(counts))
	for _, c := range counts {
		reserved[c.TierID] = c.Count
	}
	rows := make([]SponsorshipTierRow, len(tiers))
	for i, tier := range tiers {
		rows[i] = SponsorshipTierRow{SponsorshipTier: tier, ReservedSlots: reserved[tier.ID]}
	}
	return rows, nil
}

func (r *sponsorshipTierRepository) Create(ctx context.Context, tier *models.SponsorshipTier) error {
	return mapTierWriteError(r.db.WithContext(ctx).Create(tier).Error)
}

func (r *sponsorshipTierRepository) Begin(ctx context.Context) (SponsorshipTierTransaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &sponsorshipTransaction{db: tx}, nil
}

func (tx *sponsorshipTransaction) HasPledges(tierID uuid.UUID) (bool, error) {
	var count int64
	err := tx.db.Model(&models.SponsorshipPledge{}).Where("tier_id = ?", tierID).Count(&count).Error
	return count > 0, err
}

func (tx *sponsorshipTransaction) UpdateTier(tier *models.SponsorshipTier) error {
	return mapTierWriteError(tx.db.Model(&models.SponsorshipTier{}).Where("id = ?", tier.ID).
		Select("name", "price_minor", "slots", "perks").Updates(tier).Error)
}

func (tx *sponsorshipTransaction) DeleteTier(tierID uuid.UUID) error {
	return tx.db.Delete(&models.SponsorshipTier{}, "id = ?", tierID).Error
}

func mapTierWriteError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "23505") {
		return ErrSponsorshipTierNameDuplicate
	}
	return err
}

// lockTournament loads the tournament with staff under row lock (also used by sponsorship checkouts).
func lockTournamentRow(db *gorm.DB, tournamentID uuid.UUID) (*models.Tournament, error) {
	var t models.Tournament
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Staff").First(&t, "id = ?", tournamentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}
	return &t, nil
}
