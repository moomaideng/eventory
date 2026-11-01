package models

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeRegistrationApproved NotificationType = "REGISTRATION_APPROVED"
	NotificationTypeRegistrationRejected NotificationType = "REGISTRATION_REJECTED"
	NotificationTypePaymentConfirmed     NotificationType = "PAYMENT_CONFIRMED"
)

// Notification represents an in-app user notification scoped to an account.
type Notification struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null;index:idx_notifications_user_created,priority:1;index:idx_notifications_user_unread,priority:1"`
	Type      NotificationType `gorm:"type:varchar(64);not null"`
	Title     string           `gorm:"type:varchar(255);not null"`
	Message   string           `gorm:"type:text;not null"`
	Link      string           `gorm:"type:varchar(512);not null;default:''"`
	IsRead    bool             `gorm:"not null;default:false;index:idx_notifications_user_unread,priority:2"`
	ReadAt    *time.Time
	CreatedAt time.Time `gorm:"not null;index:idx_notifications_user_created,priority:2,sort:desc"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Notification) TableName() string { return "user_notifications" }
