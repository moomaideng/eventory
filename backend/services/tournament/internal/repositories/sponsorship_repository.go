package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrPledgeNotFound = errors.New("sponsorship not found")

type OwnedSponsorship struct {
	ID             uuid.UUID
	TournamentID   uuid.UUID
	TournamentName string
	TierID         uuid.UUID
	TierName       string
	AmountMinor    int64
	Currency       string
	Status         models.PledgeStatus
	PaymentID      *uuid.UUID
}

type SponsorshipRepository interface {
	Begin(ctx context.Context) (SponsorshipTransaction, error)
	Get(ctx context.Context, id uuid.UUID) (*models.SponsorshipPledge, error)
	ListConfirmed(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipPledge, error)
	ListOwned(ctx context.Context, payerID uuid.UUID, page, limit int) ([]OwnedSponsorship, int64, error)
}

type SponsorshipTransaction interface {
	LockTier(tournamentID, tierID uuid.UUID) (*models.Tournament, *models.SponsorshipTier, error)
	FindByID(id uuid.UUID) (*models.SponsorshipPledge, error)
	FindActive(tournamentID, payerID uuid.UUID) (*models.SponsorshipPledge, error)
	Reserved(tierID uuid.UUID) (int64, error)
	Create(pledge *models.SponsorshipPledge) error
	SaveResult(pledge *models.SponsorshipPledge) (bool, error)
	Commit() error
	Rollback()
}

type sponsorshipRepository struct{ db *gorm.DB }
type sponsorshipTransaction struct{ db *gorm.DB }

func NewSponsorshipRepository(db *gorm.DB) SponsorshipRepository {
	return &sponsorshipRepository{db: db}
}

func (r *sponsorshipRepository) Begin(ctx context.Context) (SponsorshipTransaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &sponsorshipTransaction{db: tx}, nil
}

func (r *sponsorshipRepository) Get(ctx context.Context, id uuid.UUID) (*models.SponsorshipPledge, error) {
	var pledge models.SponsorshipPledge
	if err := r.db.WithContext(ctx).First(&pledge, "id = ?", id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPledgeNotFound
	} else if err != nil {
		return nil, err
	}
	return &pledge, nil
}

func (r *sponsorshipRepository) ListConfirmed(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipPledge, error) {
	var pledges []models.SponsorshipPledge
	err := r.db.WithContext(ctx).Preload("Tier").
		Where("tournament_id = ? AND status = ?", tournamentID, models.PledgeConfirmed).
		Order("id ASC").Find(&pledges).Error
	return pledges, err
}

func (r *sponsorshipRepository) ListOwned(ctx context.Context, payerID uuid.UUID, page, limit int) ([]OwnedSponsorship, int64, error) {
	var total int64
	base := r.db.WithContext(ctx).Table("sponsorship_pledges AS p").Where("p.payer_account_id = ?", payerID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []OwnedSponsorship
	err := base.Select("p.id, p.tournament_id, t.name AS tournament_name, p.tier_id, st.name AS tier_name, p.amount_minor, p.currency, p.status, p.payment_id").
		Joins("JOIN tournaments AS t ON t.id = p.tournament_id").
		Joins("JOIN sponsorship_tiers AS st ON st.id = p.tier_id").
		Order("p.id DESC").Offset((page - 1) * limit).Limit(limit).Scan(&items).Error
	return items, total, err
}

func (tx *sponsorshipTransaction) LockTier(tournamentID, tierID uuid.UUID) (*models.Tournament, *models.SponsorshipTier, error) {
	tournament, err := lockTournamentRow(tx.db, tournamentID)
	if err != nil {
		return nil, nil, err
	}
	var tier models.SponsorshipTier
	err = tx.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&tier, "id = ? AND tournament_id = ?", tierID, tournamentID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrSponsorshipTierNotFound
	}
	return tournament, &tier, err
}

func (tx *sponsorshipTransaction) FindByID(id uuid.UUID) (*models.SponsorshipPledge, error) {
	var pledge models.SponsorshipPledge
	err := tx.db.First(&pledge, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &pledge, err
}

func (tx *sponsorshipTransaction) FindActive(tournamentID, payerID uuid.UUID) (*models.SponsorshipPledge, error) {
	var pledge models.SponsorshipPledge
	err := tx.db.Where("tournament_id = ? AND payer_account_id = ? AND status IN ?",
		tournamentID, payerID, []models.PledgeStatus{models.PledgePending, models.PledgeConfirmed}).First(&pledge).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &pledge, err
}

func (tx *sponsorshipTransaction) Reserved(tierID uuid.UUID) (int64, error) {
	var count int64
	err := tx.db.Model(&models.SponsorshipPledge{}).
		Where("tier_id = ? AND status IN ?", tierID, []models.PledgeStatus{models.PledgePending, models.PledgeConfirmed}).Count(&count).Error
	return count, err
}

func (tx *sponsorshipTransaction) Create(pledge *models.SponsorshipPledge) error {
	return tx.db.Create(pledge).Error
}

func (tx *sponsorshipTransaction) SaveResult(pledge *models.SponsorshipPledge) (bool, error) {
	result := tx.db.Model(&models.SponsorshipPledge{}).Where("id = ? AND status = ?", pledge.ID, models.PledgePending).
		Select("status", "payment_id").Updates(pledge)
	return result.RowsAffected == 1, result.Error
}

func (tx *sponsorshipTransaction) Commit() error { return tx.db.Commit().Error }
func (tx *sponsorshipTransaction) Rollback()     { tx.db.Rollback() }
