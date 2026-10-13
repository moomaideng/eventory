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

var (
	ErrSoloWrongMode = errors.New("tournament is not for solo registration")
	ErrSoloFull      = errors.New("tournament is full")
	ErrSoloClosed    = errors.New("registration is closed")
	ErrSoloNotFound  = errors.New("solo registration not found")
)

type SoloRegistrationRepository struct{ db *gorm.DB }

func NewSoloRegistrationRepository(db *gorm.DB) *SoloRegistrationRepository {
	return &SoloRegistrationRepository{db: db}
}

func (r *SoloRegistrationRepository) Get(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.SoloRegistration, error) {
	var row models.SoloRegistration
	err := r.db.WithContext(ctx).First(&row, "tournament_id = ? AND account_id = ?", tournamentID, accountID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSoloNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SoloRegistrationRepository) Count(ctx context.Context, tournamentID uuid.UUID, includePending bool) (int64, error) {
	statuses := []models.SoloRegistrationStatus{models.SoloRegistered}
	if includePending {
		statuses = append(statuses, models.SoloPaymentPending)
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&models.SoloRegistration{}).Where("tournament_id = ? AND status IN ?", tournamentID, statuses).Count(&count).Error
	return count, err
}

func (r *SoloRegistrationRepository) Begin(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.SoloRegistration, error) {
	var result models.SoloRegistration
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tournament models.Tournament
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&tournament, "id = ?", tournamentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTournamentNotFound
			}
			return err
		}
		if tournament.RegistrationMode != models.TournamentRegistrationModeSolo {
			return ErrSoloWrongMode
		}
		var existing models.SoloRegistration
		err := tx.First(&existing, "tournament_id = ? AND account_id = ?", tournamentID, accountID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && (existing.Status == models.SoloRegistered || existing.Status == models.SoloPaymentPending) {
			result = existing
			return nil
		}
		if !tournament.Published || tournament.Status != models.TournamentStatusRegistrationOpen || !time.Now().UTC().Before(tournament.RegistrationDeadline) {
			return ErrSoloClosed
		}
		if tournament.EntryFee < 0 || tournament.EntryFee > 9007199254740991/100 || tournament.Currency != "THB" {
			return ErrInvalidRegistrationFee
		}
		var occupied int64
		if err := tx.Model(&models.SoloRegistration{}).Where("tournament_id = ? AND status IN ?", tournamentID, []models.SoloRegistrationStatus{models.SoloRegistered, models.SoloPaymentPending}).Count(&occupied).Error; err != nil {
			return err
		}
		if occupied >= int64(tournament.Capacity) {
			return ErrSoloFull
		}
		now := time.Now().UTC().Truncate(time.Microsecond)
		status := models.SoloRegistered
		var attempt *uuid.UUID
		var expires *time.Time
		var registered *time.Time
		if tournament.EntryFee > 0 {
			status = models.SoloPaymentPending
			id := uuid.New()
			attempt = &id
			expiry := now.Add(15 * time.Minute)
			if tournament.RegistrationDeadline.Before(expiry) {
				expiry = tournament.RegistrationDeadline
			}
			expiry = expiry.Truncate(time.Microsecond)
			expires = &expiry
		} else {
			registered = &now
		}
		if err == nil {
			if err := tx.Model(&existing).Updates(map[string]any{"status": status, "payment_attempt_id": attempt, "payment_id": nil, "payment_amount_minor": tournament.EntryFee * 100, "payment_expires_at": expires, "registered_at": registered}).Error; err != nil {
				return err
			}
			return tx.First(&result, "id = ?", existing.ID).Error
		}
		result = models.SoloRegistration{ID: uuid.New(), TournamentID: tournamentID, AccountID: accountID, Status: status, PaymentAttemptID: attempt, PaymentAmountMinor: tournament.EntryFee * 100, PaymentExpiresAt: expires, RegisteredAt: registered}
		return tx.Create(&result).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *SoloRegistrationRepository) Pending(ctx context.Context) ([]models.SoloRegistration, error) {
	var rows []models.SoloRegistration
	err := r.db.WithContext(ctx).Where("status = ?", models.SoloPaymentPending).Find(&rows).Error
	return rows, err
}

func (r *SoloRegistrationRepository) ApplyPayment(ctx context.Context, row *models.SoloRegistration, paymentID uuid.UUID, status string) error {
	updates := map[string]any{"payment_id": paymentID}
	switch status {
	case "UNPAID":
		// Store the invoice ID so the player can resume the same checkout.
	case "PAID":
		updates["status"] = models.SoloRegistered
		updates["registered_at"] = time.Now().UTC()
	case "EXPIRED", "CANCELLED":
		updates["status"] = models.SoloPaymentFailed
	default:
		return nil
	}
	return r.db.WithContext(ctx).Model(&models.SoloRegistration{}).Where("id = ? AND status = ? AND payment_attempt_id = ?", row.ID, models.SoloPaymentPending, row.PaymentAttemptID).Updates(updates).Error
}
