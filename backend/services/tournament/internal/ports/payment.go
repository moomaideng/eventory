package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
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
	ID             uuid.UUID            `json:"id"`
	PayerAccountID uuid.UUID            `json:"payerAccountId"`
	Source         string               `json:"source"`
	Reference      string               `json:"reference"`
	AmountMinor    int64                `json:"amountMinor"`
	Currency       string               `json:"currency"`
	ExpiresAt      time.Time            `json:"expiresAt"`
	Status         models.PaymentStatus `json:"status"`
}

type PaymentService interface {
	EnsureInvoice(context.Context, RegistrationPaymentRequest) (*RegistrationPayment, error)
}

type RegistrationRefundService interface {
	RefundRegistration(context.Context, uuid.UUID, uuid.UUID) (*RegistrationPayment, error)
}
