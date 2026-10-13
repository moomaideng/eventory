package repositories

import (
	"context"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"time"
)

type TeamSubmissionRepository struct{ db *gorm.DB }

func NewTeamSubmissionRepository(db *gorm.DB) *TeamSubmissionRepository {
	return &TeamSubmissionRepository{db}
}

func (r *TeamSubmissionRepository) Pending(ctx context.Context) ([]models.TournamentTeam, error) {
	var teams []models.TournamentTeam
	err := r.db.WithContext(ctx).Where("status = ?", models.TournamentTeamStatusPaymentPending).Find(&teams).Error
	return teams, err
}

// Compare-and-set prevents stale invoice responses from changing a newer attempt,
// or an UNPAID response from undoing another request's completed submission.
func (r *TeamSubmissionRepository) ApplyPayment(ctx context.Context, team *models.TournamentTeam, paymentID uuid.UUID, status string) error {
	updates := map[string]any{"payment_id": paymentID}
	switch status {
	case "PAID":
		updates["status"] = models.TournamentTeamStatusLocked
		updates["locked_at"] = time.Now().UTC()
	case "EXPIRED", "CANCELLED":
		updates["status"] = models.TournamentTeamStatusForming
		updates["payment_attempt_id"] = nil
		updates["payment_id"] = nil
		updates["payment_payer_id"] = nil
		updates["payment_expires_at"] = nil
		updates["payment_amount_minor"] = 0
	}
	return r.db.WithContext(ctx).Model(&models.TournamentTeam{}).
		Where("id = ? AND status = ? AND payment_attempt_id = ?", team.ID, models.TournamentTeamStatusPaymentPending, team.PaymentAttemptID).
		Updates(updates).Error
}
