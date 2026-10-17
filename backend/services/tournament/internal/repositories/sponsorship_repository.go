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

type SponsorshipRepository interface {
	Begin(ctx context.Context) (SponsorshipTransaction, error)
	Get(ctx context.Context, id uuid.UUID) (*models.SponsorshipPledge, error)
	Pending(ctx context.Context) ([]models.SponsorshipPledge, error)
	ListConfirmed(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipPledge, error)
	ListOwned(ctx context.Context, payerID uuid.UUID, page, limit int) ([]OwnedSponsorship, int64, error)
	Logo(ctx context.Context, pledgeID uuid.UUID) (*models.SponsorshipLogo, error)
}

type OwnedSponsorship struct {
	models.SponsorshipPledge
	TournamentName string
}

func (r *sponsorshipRepository) ListOwned(ctx context.Context, payerID uuid.UUID, page, limit int) ([]OwnedSponsorship, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.SponsorshipPledge{}).Where("payer_account_id = ?", payerID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]OwnedSponsorship, 0)
	err := query.Select("sponsorship_pledges.*, tournaments.name AS tournament_name").
		Joins("JOIN tournaments ON tournaments.id = sponsorship_pledges.tournament_id").
		Order("sponsorship_pledges.created_at DESC, sponsorship_pledges.id DESC").
		Offset((page - 1) * limit).Limit(limit).Scan(&items).Error
	return items, total, err
}

// The use case evaluates rules while these rows are locked; persistence owns the transaction.
type SponsorshipTransaction interface {
	LockTier(tournamentID, tierID uuid.UUID) (*models.Tournament, *models.SponsorshipTier, error)
	FindActive(tierID, payerID uuid.UUID) (*models.SponsorshipPledge, error)
	Reserved(tierID uuid.UUID) (int64, error)
	Create(pledge *models.SponsorshipPledge, logo *models.SponsorshipLogo) error
	SaveResult(pledge *models.SponsorshipPledge) (bool, error)
	RefreshFunding(tournamentID uuid.UUID) error
	Commit() error
	Rollback()
}

type sponsorshipRepository struct{ db *gorm.DB }
type sponsorshipTransaction struct{ db *gorm.DB }

func NewSponsorshipRepository(db *gorm.DB) SponsorshipRepository { return &sponsorshipRepository{db} }

func (r *sponsorshipRepository) Begin(ctx context.Context) (SponsorshipTransaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &sponsorshipTransaction{tx}, nil
}

func (r *sponsorshipRepository) Get(ctx context.Context, id uuid.UUID) (*models.SponsorshipPledge, error) {
	var pledge models.SponsorshipPledge
	err := r.db.WithContext(ctx).First(&pledge, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPledgeNotFound
	}
	return &pledge, err
}

func (r *sponsorshipRepository) Pending(ctx context.Context) ([]models.SponsorshipPledge, error) {
	var pledges []models.SponsorshipPledge
	err := r.db.WithContext(ctx).Where("status = ?", models.PledgePending).Order("created_at").Find(&pledges).Error
	return pledges, err
}

func (r *sponsorshipRepository) ListConfirmed(ctx context.Context, tournamentID uuid.UUID) ([]models.SponsorshipPledge, error) {
	var pledges []models.SponsorshipPledge
	err := r.db.WithContext(ctx).Where("tournament_id = ? AND status = ?", tournamentID, models.PledgeConfirmed).Order("created_at").Find(&pledges).Error
	return pledges, err
}

func (r *sponsorshipRepository) Logo(ctx context.Context, pledgeID uuid.UUID) (*models.SponsorshipLogo, error) {
	var logo models.SponsorshipLogo
	err := r.db.WithContext(ctx).Joins("JOIN sponsorship_pledges p ON p.id = sponsorship_logos.pledge_id").
		Joins("JOIN tournaments t ON t.id = p.tournament_id").
		Where("pledge_id = ? AND p.status = ? AND t.published = true", pledgeID, models.PledgeConfirmed).First(&logo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPledgeNotFound
	}
	return &logo, err
}

func (tx *sponsorshipTransaction) LockTier(tournamentID, tierID uuid.UUID) (*models.Tournament, *models.SponsorshipTier, error) {
	var tournament models.Tournament
	if err := tx.db.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Staff").First(&tournament, "id = ?", tournamentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = ErrTournamentNotFound
		}
		return nil, nil, err
	}
	var tier models.SponsorshipTier
	err := tx.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&tier, "id = ? AND tournament_id = ?", tierID, tournamentID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrSponsorshipTierNotFound
	}
	return &tournament, &tier, err
}

func (tx *sponsorshipTransaction) FindActive(tierID, payerID uuid.UUID) (*models.SponsorshipPledge, error) {
	var pledge models.SponsorshipPledge
	err := tx.db.Where("tier_id = ? AND payer_account_id = ? AND status IN ?", tierID, payerID, []models.PledgeStatus{models.PledgePending, models.PledgeConfirmed}).First(&pledge).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &pledge, err
}

func (tx *sponsorshipTransaction) Reserved(tierID uuid.UUID) (int64, error) {
	var count int64
	err := tx.db.Model(&models.SponsorshipPledge{}).Where("tier_id = ? AND status IN ?", tierID, []models.PledgeStatus{models.PledgePending, models.PledgeConfirmed}).Count(&count).Error
	return count, err
}

func (tx *sponsorshipTransaction) Create(pledge *models.SponsorshipPledge, logo *models.SponsorshipLogo) error {
	if err := tx.db.Omit("Tier").Create(pledge).Error; err != nil {
		return err
	}
	return tx.db.Omit("Pledge").Create(logo).Error
}

func (tx *sponsorshipTransaction) SaveResult(pledge *models.SponsorshipPledge) (bool, error) {
	result := tx.db.Model(&models.SponsorshipPledge{}).Where("id = ? AND status = ?", pledge.ID, models.PledgePending).
		Updates(map[string]any{"status": pledge.Status, "payment_id": pledge.PaymentID, "confirmed_at": pledge.ConfirmedAt})
	return result.RowsAffected == 1, result.Error
}

func (tx *sponsorshipTransaction) RefreshFunding(tournamentID uuid.UUID) error {
	var totals struct {
		RaisedAmount   int64
		SupporterCount int
	}
	if err := tx.db.Model(&models.SponsorshipPledge{}).Select("(COALESCE(SUM(amount_minor), 0) / 100)::bigint AS raised_amount, COUNT(DISTINCT payer_account_id) AS supporter_count").
		Where("tournament_id = ? AND status = ?", tournamentID, models.PledgeConfirmed).Scan(&totals).Error; err != nil {
		return err
	}
	return tx.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tournament_id"}}, DoUpdates: clause.AssignmentColumns([]string{"raised_amount", "supporter_count"})}).
		Create(&models.TournamentFunding{TournamentID: tournamentID, RaisedAmount: totals.RaisedAmount, SupporterCount: totals.SupporterCount}).Error
}

func (tx *sponsorshipTransaction) Commit() error { return tx.db.Commit().Error }
func (tx *sponsorshipTransaction) Rollback()     { tx.db.Rollback() }
