package usecases

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/models"
)

// ErrPaymentMismatch means a payment event contradicts the record it points at,
// so retrying it cannot help.
var ErrPaymentMismatch = errors.New("payment event does not match the record it references")

// PaymentOutcome is the state of one invoice as Payment reported it, either in
// a reply to Reconcile or in a payment event. It carries the terms the invoice
// was issued under so they can be checked against Tournament's own record.
type PaymentOutcome struct {
	Status         models.PaymentStatus
	PaymentID      uuid.UUID
	PayerAccountID uuid.UUID
	Reference      string
	AmountMinor    int64
	Currency       string
	ExpiresAt      time.Time
}

func outcomeOf(p *ports.RegistrationPayment) PaymentOutcome {
	return PaymentOutcome{
		Status: p.Status, PaymentID: p.ID, PayerAccountID: p.PayerAccountID, Reference: p.Reference,
		AmountMinor: p.AmountMinor, Currency: p.Currency, ExpiresAt: p.ExpiresAt,
	}
}

// settleable reports whether status is one Tournament knows how to record.
func settleable(status models.PaymentStatus) bool {
	switch status {
	case models.PaymentUnpaid, models.PaymentPaid, models.PaymentExpired, models.PaymentCancelled:
		return true
	}
	return false
}
