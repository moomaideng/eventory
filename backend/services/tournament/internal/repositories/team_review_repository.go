package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrTeamEntryNotFound is returned when the team does not exist in the
// tournament. A tournament the caller cannot manage reports
// ErrTournamentNotFound instead, so the two look identical to outsiders.
var ErrTeamEntryNotFound = errors.New("team entry not found")

// TeamReviewRepository persists organizer decisions on team applications
// (US3-6) together with their audit trail.
type TeamReviewRepository interface {
	Begin(ctx context.Context) (TeamReviewTransaction, error)
	PendingRefunds(ctx context.Context) ([]models.TournamentTeam, error)
	MarkRefunded(ctx context.Context, teamID, paymentID uuid.UUID) error
}

// TeamReviewTransaction locks the tournament and team rows while the use case
// evaluates review permissions and capacity rules.
type TeamReviewTransaction interface {
	LockTournament(tournamentID uuid.UUID) (*models.Tournament, error)
	LockTeamAndAcceptedCount(tournamentID, teamID uuid.UUID) (*models.TournamentTeam, int, error)
	SaveReview(team *models.TournamentTeam, toStatus models.TournamentTeamStatus, review *models.TournamentTeamReview) error
	Commit() error
	Rollback()
}

type teamReviewRepository struct {
	db *gorm.DB
}

type teamReviewTransaction struct {
	db *gorm.DB
}

func NewTeamReviewRepository(db *gorm.DB) TeamReviewRepository {
	return &teamReviewRepository{db: db}
}

func (r *teamReviewRepository) Begin(ctx context.Context) (TeamReviewTransaction, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &teamReviewTransaction{db: tx}, nil
}

func (tx *teamReviewTransaction) LockTournament(tournamentID uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament

	// Lock the tournament row for the rest of the transaction so two
	// concurrent approvals cannot both see a free spot and overfill it.
	err := tx.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Staff", orderedStaff).
		Where("id = ?", tournamentID).
		First(&tournament).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tournament, nil
}

func (tx *teamReviewTransaction) LockTeamAndAcceptedCount(
	tournamentID, teamID uuid.UUID,
) (*models.TournamentTeam, int, error) {
	// Lock the team too: payment confirmation changes its status without
	// taking the tournament lock.
	var team models.TournamentTeam
	err := tx.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Members").
		Where("id = ? AND tournament_id = ?", teamID, tournamentID).
		First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, ErrTeamEntryNotFound
	}
	if err != nil {
		return nil, 0, err
	}

	var accepted int64
	if err := tx.db.Model(&models.TournamentTeam{}).
		Where("tournament_id = ? AND status = ?", tournamentID, models.TournamentTeamStatusAccepted).
		Count(&accepted).Error; err != nil {
		return nil, 0, err
	}

	return &team, int(accepted), nil
}

func (tx *teamReviewTransaction) SaveReview(
	team *models.TournamentTeam,
	toStatus models.TournamentTeamStatus,
	review *models.TournamentTeamReview,
) error {
	if err := tx.db.Model(team).Update("status", toStatus).Error; err != nil {
		return err
	}
	return tx.db.Create(review).Error
}

func (tx *teamReviewTransaction) Commit() error {
	return tx.db.Commit().Error
}

func (tx *teamReviewTransaction) Rollback() {
	tx.db.Rollback()
}

func (r *teamReviewRepository) PendingRefunds(ctx context.Context) ([]models.TournamentTeam, error) {
	var teams []models.TournamentTeam
	err := r.db.WithContext(ctx).Where("status = ? AND payment_id IS NOT NULL AND refunded_at IS NULL", models.TournamentTeamStatusRejected).Find(&teams).Error
	return teams, err
}

func (r *teamReviewRepository) MarkRefunded(ctx context.Context, teamID, paymentID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.TournamentTeam{}).
		Where("id = ? AND status = ? AND payment_id = ? AND refunded_at IS NULL", teamID, models.TournamentTeamStatusRejected, paymentID).
		Update("refunded_at", time.Now().UTC()).Error
}
