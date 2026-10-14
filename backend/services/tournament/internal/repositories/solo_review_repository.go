package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SoloReviewRepository persists organizer decisions on solo applications
// (US3-6) together with their audit trail. It is the solo twin of
// TeamReviewRepository and follows the same transaction shape.
//
// A solo entry that does not exist in the tournament reports
// ErrSoloNotFound; a tournament the caller cannot manage reports
// ErrTournamentNotFound instead, so the two look identical to outsiders.
type SoloReviewRepository interface {
	// ApplyReview moves a solo registration to the status returned by decide
	// and writes the audit row, all in one transaction. decide receives the
	// locked tournament and registration plus the number of solo players
	// currently accepted; returning an error aborts the whole review.
	ApplyReview(
		ctx context.Context,
		accountID uuid.UUID,
		reviewerDisplayName string,
		tournamentID, soloID uuid.UUID,
		reason string,
		decide func(tournament *models.Tournament, solo *models.SoloRegistration, acceptedEntries int) (models.SoloRegistrationStatus, error),
	) (*models.SoloRegistrationReview, error)
}

type soloReviewRepository struct {
	db *gorm.DB
}

func NewSoloReviewRepository(db *gorm.DB) SoloReviewRepository {
	return &soloReviewRepository{db: db}
}

func (r *soloReviewRepository) ApplyReview(
	ctx context.Context,
	accountID uuid.UUID,
	reviewerDisplayName string,
	tournamentID, soloID uuid.UUID,
	reason string,
	decide func(tournament *models.Tournament, solo *models.SoloRegistration, acceptedEntries int) (models.SoloRegistrationStatus, error),
) (*models.SoloRegistrationReview, error) {
	var review models.SoloRegistrationReview

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

		// Lock the registration too, so two reviews of the same player
		// serialise and the second one sees the first one's result.
		var solo models.SoloRegistration
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND tournament_id = ?", soloID, tournamentID).
			First(&solo).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSoloNotFound
		}
		if err != nil {
			return err
		}

		var accepted int64
		if err := tx.Model(&models.SoloRegistration{}).
			Where("tournament_id = ? AND status = ?", tournamentID, models.SoloAccepted).
			Count(&accepted).Error; err != nil {
			return err
		}

		to, err := decide(&tournament, &solo, int(accepted))
		if err != nil {
			return err
		}

		from := solo.Status
		if err := tx.Model(&solo).Update("status", to).Error; err != nil {
			return err
		}

		review = models.SoloRegistrationReview{
			TournamentID:        tournament.ID,
			SoloRegistrationID:  solo.ID,
			FromStatus:          from,
			ToStatus:            to,
			Reason:              reason,
			ReviewerAccountID:   accountID,
			ReviewerDisplayName: reviewerDisplayName,
		}
		return tx.Create(&review).Error
	})
	if err != nil {
		return nil, err
	}
	return &review, nil
}
