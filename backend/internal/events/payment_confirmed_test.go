package events

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func validPaymentConfirmedEvent() PaymentConfirmed {
	return PaymentConfirmed{
		SpecVersion:   EventSpecVersion,
		ID:            uuid.New(),
		Type:          PaymentConfirmedType,
		Source:        "payment-service",
		OccurredAt:    time.Now().UTC(),
		CorrelationID: uuid.New(),
		Data: PaymentConfirmedData{
			PaymentID:      uuid.New(),
			PayerAccountID: uuid.New(),
			AmountMinor:    50000,
			Currency:       "THB",
			Description:    "Cyber Championship entry fee",
			Recipient: EventRecipient{
				AccountID:   uuid.New(),
				Email:       "payer@example.com",
				DisplayName: "Payer User",
			},
		},
	}
}

func TestPaymentConfirmedValidate(t *testing.T) {
	t.Run("valid payment confirmed", func(t *testing.T) {
		e := validPaymentConfirmedEvent()
		if err := e.Validate(); err != nil {
			t.Fatalf("expected valid event, got %v", err)
		}
	})

	t.Run("invalid spec version", func(t *testing.T) {
		e := validPaymentConfirmedEvent()
		e.SpecVersion = "2.0"
		if err := e.Validate(); err == nil {
			t.Fatal("expected error for invalid spec version")
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		e := validPaymentConfirmedEvent()
		e.Type = "payment.failed"
		if err := e.Validate(); err == nil {
			t.Fatal("expected error for invalid type")
		}
	})

	t.Run("missing payment id", func(t *testing.T) {
		e := validPaymentConfirmedEvent()
		e.Data.PaymentID = uuid.Nil
		if err := e.Validate(); err == nil {
			t.Fatal("expected error for missing payment id")
		}
	})

	t.Run("non-positive amount", func(t *testing.T) {
		e := validPaymentConfirmedEvent()
		e.Data.AmountMinor = 0
		if err := e.Validate(); err == nil {
			t.Fatal("expected error for non-positive amount")
		}
	})

	t.Run("missing recipient account id", func(t *testing.T) {
		e := validPaymentConfirmedEvent()
		e.Data.Recipient.AccountID = uuid.Nil
		if err := e.Validate(); err == nil {
			t.Fatal("expected error for missing recipient account id")
		}
	})
}
