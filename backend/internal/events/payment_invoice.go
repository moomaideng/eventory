package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	PaymentCancelledType = "payment.cancelled"
	PaymentExpiredType   = "payment.expired"
)

// PaymentInvoice is the invoice a payment event is about. Every payment event
// carries it, so the service that requested the invoice can find its own
// record and check the terms without calling Payment back.
type PaymentInvoice struct {
	PaymentID      uuid.UUID `json:"payment_id"`
	PayerAccountID uuid.UUID `json:"payer_account_id"`
	// InvoiceSource and Reference say what the invoice pays for, for example
	// TEAM_REGISTRATION and "<team id>:<attempt id>".
	InvoiceSource string    `json:"invoice_source"`
	Reference     string    `json:"reference"`
	AmountMinor   int64     `json:"amount_minor"`
	Currency      string    `json:"currency"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// PaymentCancelled is published by Payment Service when a payer fails or
// cancels checkout and the invoice is closed for good.
type PaymentCancelled struct {
	SpecVersion   string         `json:"spec_version"`
	ID            uuid.UUID      `json:"id"`
	Type          string         `json:"type"`
	Source        string         `json:"source"`
	OccurredAt    time.Time      `json:"occurred_at"`
	CorrelationID uuid.UUID      `json:"correlation_id"`
	Data          PaymentInvoice `json:"data"`
}

// PaymentExpired is published by Payment Service once an unpaid invoice passes
// its deadline. No request causes expiry, so this event is the only signal.
type PaymentExpired struct {
	SpecVersion   string         `json:"spec_version"`
	ID            uuid.UUID      `json:"id"`
	Type          string         `json:"type"`
	Source        string         `json:"source"`
	OccurredAt    time.Time      `json:"occurred_at"`
	CorrelationID uuid.UUID      `json:"correlation_id"`
	Data          PaymentInvoice `json:"data"`
}
