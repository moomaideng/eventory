package ports

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type RegistrationPaymentRequest struct {
	PayerAccountID uuid.UUID `json:"payerAccountId"`
	Source         string    `json:"source"`
	Reference      string    `json:"reference"`
	IdempotencyKey string    `json:"idempotencyKey"`
	Description    string    `json:"description"`
	AmountMinor    int64     `json:"amountMinor"`
	Currency       string    `json:"currency"`
	ExpiresAt      time.Time `json:"expiresAt"`
}

type RegistrationPayment struct {
	ID             uuid.UUID `json:"id"`
	PayerAccountID uuid.UUID `json:"payerAccountId"`
	Source         string    `json:"source"`
	Reference      string    `json:"reference"`
	AmountMinor    int64     `json:"amountMinor"`
	Currency       string    `json:"currency"`
	ExpiresAt      time.Time `json:"expiresAt"`
	Status         string    `json:"status"`
}

type PaymentService interface {
	EnsureInvoice(context.Context, RegistrationPaymentRequest) (*RegistrationPayment, error)
}
