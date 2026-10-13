package models

import (
	"time"

	"github.com/google/uuid"
)

type SoloRegistrationStatus string

const (
	SoloPaymentPending SoloRegistrationStatus = "PAYMENT_PENDING"
	SoloRegistered     SoloRegistrationStatus = "REGISTERED"
	SoloPaymentFailed  SoloRegistrationStatus = "PAYMENT_FAILED"
)

// A row is unique per player and tournament. Pending rows reserve a seat;
// failed rows can be reused for a fresh payment attempt.
type SoloRegistration struct {
	ID                 uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID       uuid.UUID              `gorm:"type:uuid;not null;uniqueIndex:idx_solo_player_tournament,priority:1;index:idx_solo_capacity"`
	Tournament         Tournament             `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	AccountID          uuid.UUID              `gorm:"type:uuid;not null;uniqueIndex:idx_solo_player_tournament,priority:2"`
	Status             SoloRegistrationStatus `gorm:"type:varchar(24);not null;index:idx_solo_capacity"`
	PaymentAttemptID   *uuid.UUID             `gorm:"type:uuid"`
	PaymentID          *uuid.UUID             `gorm:"type:uuid"`
	PaymentAmountMinor int64                  `gorm:"not null;default:0"`
	PaymentExpiresAt   *time.Time
	RegisteredAt       *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
