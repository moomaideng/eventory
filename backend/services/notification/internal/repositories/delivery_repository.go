package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/notification/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ClaimResult int

const (
	Claimed ClaimResult = iota
	AlreadyHandled
	AlreadyProcessing
	PayloadConflict
)

type DeliveryRepository struct{ db *gorm.DB }

func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository { return &DeliveryRepository{db: db} }

func (repo *DeliveryRepository) Claim(
	ctx context.Context,
	eventID uuid.UUID,
	eventType, recipient, template, payloadHash string,
	eventOccurredAt, now time.Time,
	lease time.Duration,
) (*models.Delivery, ClaimResult, error) {
	var delivery models.Delivery
	result := Claimed
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("event_id = ? AND channel = ? AND template_key = ?", eventID, "EMAIL", template).
			First(&delivery).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			candidate := models.Delivery{
				ID: uuid.New(), EventID: eventID, EventType: eventType,
				RecipientEmail: strings.ToLower(recipient), Channel: "EMAIL",
				TemplateKey: template, PayloadHash: payloadHash, Status: models.DeliveryStatusPending,
				EventOccurredAt: eventOccurredAt,
			}
			created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&candidate)
			if created.Error != nil {
				return created.Error
			}
			if created.RowsAffected == 1 {
				delivery = candidate
			} else if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("event_id = ? AND channel = ? AND template_key = ?", eventID, "EMAIL", template).
				First(&delivery).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// Rows created before payload hashing was introduced are safely adopted
		// on their first delivery after migration.
		if delivery.PayloadHash == "" {
			delivery.PayloadHash = payloadHash
		} else if delivery.PayloadHash != payloadHash {
			result = PayloadConflict
			return nil
		}

		if delivery.Status == models.DeliveryStatusSent || delivery.Status == models.DeliveryStatusFailed {
			result = AlreadyHandled
			return nil
		}
		if delivery.Status == models.DeliveryStatusSending && delivery.LockedUntil != nil && delivery.LockedUntil.After(now) {
			result = AlreadyProcessing
			return nil
		}

		lockedUntil := now.Add(lease)
		delivery.Status = models.DeliveryStatusSending
		delivery.AttemptCount++
		delivery.LockedUntil = &lockedUntil
		delivery.LastError = ""
		return tx.Save(&delivery).Error
	})
	return &delivery, result, err
}

func (repo *DeliveryRepository) ResetFailed(ctx context.Context, eventID uuid.UUID) (bool, error) {
	result := repo.db.WithContext(ctx).Model(&models.Delivery{}).
		Where("event_id = ? AND status = ?", eventID, models.DeliveryStatusFailed).
		Updates(map[string]any{
			"status": models.DeliveryStatusPending, "attempt_count": 0,
			"last_error": "", "locked_until": nil, "sent_at": nil,
		})
	return result.RowsAffected > 0, result.Error
}

func (repo *DeliveryRepository) MarkSent(ctx context.Context, id uuid.UUID, now time.Time) error {
	return repo.db.WithContext(ctx).Model(&models.Delivery{}).Where("id = ?", id).Updates(map[string]any{
		"status": models.DeliveryStatusSent, "sent_at": now, "locked_until": nil, "last_error": "",
	}).Error
}

func (repo *DeliveryRepository) MarkRetry(ctx context.Context, id uuid.UUID, message string) error {
	return repo.updateFailure(ctx, id, models.DeliveryStatusRetryScheduled, message)
}

func (repo *DeliveryRepository) MarkFailed(ctx context.Context, id uuid.UUID, message string) error {
	return repo.updateFailure(ctx, id, models.DeliveryStatusFailed, message)
}

func (repo *DeliveryRepository) updateFailure(ctx context.Context, id uuid.UUID, status models.DeliveryStatus, message string) error {
	if len(message) > 1000 {
		message = message[:1000]
	}
	return repo.db.WithContext(ctx).Model(&models.Delivery{}).Where("id = ?", id).Updates(map[string]any{
		"status": status, "last_error": message, "locked_until": nil,
	}).Error
}
