package models

import (
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	DeliveryStatusPending        DeliveryStatus = "PENDING"
	DeliveryStatusSending        DeliveryStatus = "SENDING"
	DeliveryStatusRetryScheduled DeliveryStatus = "RETRY_SCHEDULED"
	DeliveryStatusSent           DeliveryStatus = "SENT"
	DeliveryStatusFailed         DeliveryStatus = "FAILED"
)

type Delivery struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EventID         uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_notification_delivery_handler_v2"`
	EventType       string         `gorm:"type:varchar(120);not null;index"`
	RecipientEmail  string         `gorm:"type:varchar(255);not null"`
	Channel         string         `gorm:"type:varchar(24);not null;uniqueIndex:idx_notification_delivery_handler_v2"`
	TemplateKey     string         `gorm:"type:varchar(120);not null;uniqueIndex:idx_notification_delivery_handler_v2"`
	PayloadHash     string         `gorm:"type:char(64);not null;default:''"`
	Status          DeliveryStatus `gorm:"type:varchar(32);not null;index"`
	AttemptCount    int            `gorm:"not null;default:0"`
	LastError       string         `gorm:"type:text;not null;default:''"`
	LockedUntil     *time.Time     `gorm:"index"`
	EventOccurredAt time.Time      `gorm:"not null"`
	SentAt          *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Delivery) TableName() string { return "notification_deliveries" }

func All() []any { return []any{&Delivery{}} }
