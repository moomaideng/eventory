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
	// ApplyReview moves a team to the status returned by decide and writes the
	// audit row, all in one transaction. decide receives the locked tournament
	// and team plus the number of teams currently accepted; returning an error
	// aborts the whole review, so the decision rules stay in the use case while
	// the rows stay locked.
	ApplyReview(
		ctx context.Context,
		accountID uuid.UUID,
		reviewerDisplayName string,
		tournamentID, teamID uuid.UUID,
		reason string,
		decide func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error),
	) (*models.TournamentTeamReview, *uuid.UUID, error)
	PendingRefunds(ctx context.Context) ([]models.TournamentTeam, error)
	MarkRefunded(ctx context.Context, teamID, paymentID uuid.UUID) error
}

type teamReviewRepository struct {
	db *gorm.DB
}

func NewTeamReviewRepository(db *gorm.DB) TeamReviewRepository {
	return &teamReviewRepository{db: db}
}

func (r *teamReviewRepository) ApplyReview(
	ctx context.Context,
	accountID uuid.UUID,
	reviewerDisplayName string,
	tournamentID, teamID uuid.UUID,
	reason string,
	decide func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error),
) (*models.TournamentTeamReview, *uuid.UUID, error) {
	var review models.TournamentTeamReview
	var paymentID *uuid.UUID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tournament models.Tournament

		// Lock the tournament row for the rest of the transaction so two
		// concurrent approvals cannot both see a free spot and overfill it.
		// Scoping through ownedTournaments also enforces who may review.
		err := ownedTournaments(tx, accountID).
			Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "tournaments"}}).
			Where("id = ?", tournamentID).
			First(&tournament).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTournamentNotFound
		}
		if err != nil {
			return err
		}

		// Lock the team too: payment confirmation changes its status without
		// taking the tournament lock.
		var team models.TournamentTeam
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND tournament_id = ?", teamID, tournamentID).
			First(&team).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTeamEntryNotFound
		}
		if err != nil {
			return err
		}

		var accepted int64
		if err := tx.Model(&models.TournamentTeam{}).
			Where("tournament_id = ? AND status = ?", tournamentID, models.TournamentTeamStatusAccepted).
			Count(&accepted).Error; err != nil {
			return err
		}

		to, err := decide(&tournament, &team, int(accepted))
		if err != nil {
			return err
		}

		from := team.Status
		paymentID = team.PaymentID
		if err := tx.Model(&team).Update("status", to).Error; err != nil {
			return err
		}

		review = models.TournamentTeamReview{
			TournamentID:        tournament.ID,
			TournamentTeamID:    team.ID,
			FromStatus:          from,
			ToStatus:            to,
			Reason:              reason,
			ReviewerAccountID:   accountID,
			ReviewerDisplayName: reviewerDisplayName,
		}
		return tx.Create(&review).Error
	})
	if err != nil {
		return nil, nil, err
	}
	return &review, paymentID, nil
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
