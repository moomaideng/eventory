package events

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/money"
)

const (
	PaymentConfirmedType = "payment.confirmed"
)

// PaymentConfirmed is published by Payment Service when an invoice transitions
// to PAID / Succeeded.
type PaymentConfirmed struct {
	SpecVersion   string               `json:"spec_version"`
	ID            uuid.UUID            `json:"id"`
	Type          string               `json:"type"`
	Source        string               `json:"source"`
	OccurredAt    time.Time            `json:"occurred_at"`
	CorrelationID uuid.UUID            `json:"correlation_id"`
	Data          PaymentConfirmedData `json:"data"`
}

type PaymentConfirmedData struct {
	PaymentInvoice
	Description string         `json:"description"`
	Recipient   EventRecipient `json:"recipient"`
}

func (event *PaymentConfirmed) Normalize() {
	event.Source = strings.TrimSpace(event.Source)
	event.Data.Currency = strings.ToUpper(strings.TrimSpace(event.Data.Currency))
	event.Data.Description = strings.TrimSpace(event.Data.Description)
	event.Data.Recipient.Email = strings.ToLower(strings.TrimSpace(event.Data.Recipient.Email))
	event.Data.Recipient.DisplayName = strings.TrimSpace(event.Data.Recipient.DisplayName)
}

func (event PaymentConfirmed) Validate() error {
	if event.SpecVersion != EventSpecVersion {
		return fmt.Errorf("unsupported spec_version %q", event.SpecVersion)
	}
	if event.Type != PaymentConfirmedType {
		return fmt.Errorf("unexpected event type %q", event.Type)
	}
	if event.ID == uuid.Nil || event.CorrelationID == uuid.Nil {
		return errors.New("event id and correlation_id are required")
	}
	if event.Source == "" || event.OccurredAt.IsZero() {
		return errors.New("source and occurred_at are required")
	}
	if event.Data.PaymentID == uuid.Nil || event.Data.PayerAccountID == uuid.Nil {
		return errors.New("payment_id and payer_account_id are required")
	}
	if amt, err := money.FromMinor(event.Data.AmountMinor, money.CurrencyOrDefault(event.Data.Currency)); err != nil || amt.AmountMinor == 0 {
		return errors.New("amount_minor must be positive and currency must be supported")
	}
	if event.Data.Recipient.AccountID == uuid.Nil {
		return errors.New("recipient account_id is required")
	}
	if event.Data.Recipient.Email != "" {
		address, err := mail.ParseAddress(event.Data.Recipient.Email)
		if err != nil || !strings.EqualFold(address.Address, event.Data.Recipient.Email) {
			return errors.New("recipient email is invalid")
		}
	}
	return nil
}
