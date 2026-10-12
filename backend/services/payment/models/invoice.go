package models

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	Unpaid  Status = "UNPAID"
	Paid    Status = "PAID"
	Expired Status = "EXPIRED"
	Voided  Status = "CANCELLED"
)

type MockResult string

const (
	Succeeded MockResult = "SUCCEEDED"
	Failed    MockResult = "FAILED"
	Cancelled MockResult = "CANCELLED"
)

// Invoice terms are immutable; PaidAt and CancelledAt record terminal results.
type Invoice struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	PayerAccountID uuid.UUID `gorm:"type:uuid;not null;index:idx_invoice_history,priority:1"`
	Source         string    `gorm:"size:64;not null;uniqueIndex:idx_invoice_idempotency,priority:1;uniqueIndex:idx_invoice_reference,priority:1"`
	Reference      string    `gorm:"size:128;not null;uniqueIndex:idx_invoice_reference,priority:2"`
	IdempotencyKey string    `gorm:"size:128;not null;uniqueIndex:idx_invoice_idempotency,priority:2"`
	Description    string    `gorm:"size:255;not null"`
	AmountMinor    int64     `gorm:"not null;check:payment_amount,amount_minor > 0 AND amount_minor <= 9007199254740991"`
	Currency       string    `gorm:"size:3;not null;check:payment_currency,currency = 'THB'"`
	Provider       string    `gorm:"size:16;not null;check:payment_provider,provider = 'MOCK'"`
	CreatedAt      time.Time `gorm:"not null;index:idx_invoice_history,priority:2"`
	UpdatedAt      time.Time `gorm:"not null"`
	ExpiresAt      time.Time `gorm:"not null"`
	PaidAt         *time.Time
	CancelledAt    *time.Time
	Receipt        *Receipt `gorm:"foreignKey:InvoiceID"`
}

func (Invoice) TableName() string { return "payment_invoices" }

func (p *Invoice) Status(now time.Time) Status {
	if p.PaidAt != nil {
		return Paid
	}
	if p.CancelledAt != nil {
		return Voided
	}
	if !now.Before(p.ExpiresAt) {
		return Expired
	}
	return Unpaid
}

type Receipt struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	InvoiceID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`
}
