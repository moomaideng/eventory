package models

import (
	"time"

	"github.com/google/uuid"
)

// OutboxEvent is a domain event written in the same transaction as the invoice
// change it describes. A relay publishes it afterwards, so an event exists
// exactly when the change committed.
type OutboxEvent struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoutingKey string    `gorm:"size:64;not null"`
	// DedupKey makes writing the same event twice harmless, e.g. "<invoice id>:payment.confirmed".
	DedupKey string `gorm:"size:160;not null;uniqueIndex"`
	Payload  []byte `gorm:"type:jsonb;not null"`
	// The partial index keeps the relay's poll cheap however many events were published.
	CreatedAt   time.Time `gorm:"not null;index:idx_payment_outbox_unpublished,where:published_at IS NULL"`
	PublishedAt *time.Time
}

func (OutboxEvent) TableName() string { return "payment_outbox" }
