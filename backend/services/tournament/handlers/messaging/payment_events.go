// Package messaging is the Tournament transport for broker messages, the
// counterpart of handlers/rest for HTTP.
package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

// PaymentEventsQueue receives Payment's outcome events. All Tournament
// instances share it, so each event is handled once.
const PaymentEventsQueue = "tournament.payment-events"

// PaymentEventConsumer retries a failed event after 30 s, 5 min and 30 min,
// then parks it in "tournament.payment-events.dead".
func PaymentEventConsumer(brokerURL string) events.Consumer {
	return events.Consumer{
		URL:            brokerURL,
		Queue:          PaymentEventsQueue,
		RoutingKeys:    []string{events.PaymentConfirmedType, events.PaymentCancelledType, events.PaymentExpiredType},
		RetryDelays:    []time.Duration{30 * time.Second, 5 * time.Minute, 30 * time.Minute},
		Prefetch:       10,
		ReconnectDelay: 5 * time.Second,
	}
}

// PaymentEventHandler applies each payment outcome to the team registration or
// sponsorship pledge its invoice was issued for. Events for invoices of other
// sources are acknowledged and skipped.
func PaymentEventHandler(teams *usecases.TeamLobbyUseCase, sponsorships *usecases.SponsorshipUseCase) events.Handler {
	return func(ctx context.Context, routingKey string, body []byte) error {
		source, outcome, err := decodePaymentEvent(routingKey, body)
		if err != nil {
			return err
		}
		switch source {
		case models.PaymentSourceTeamRegistration:
			err = teams.ApplyPaymentOutcome(ctx, outcome)
		case models.PaymentSourceSponsorship:
			err = sponsorships.ApplyPaymentOutcome(ctx, outcome)
		}
		if errors.Is(err, usecases.ErrPaymentMismatch) {
			return fmt.Errorf("%w: %w", events.ErrPermanent, err)
		}
		return err
	}
}

// paymentStatusByEvent is the invoice status each payment event reports.
var paymentStatusByEvent = map[string]models.PaymentStatus{
	events.PaymentConfirmedType: models.PaymentPaid,
	events.PaymentCancelledType: models.PaymentCancelled,
	events.PaymentExpiredType:   models.PaymentExpired,
}

// decodePaymentEvent returns the invoice source and the outcome an event
// reports. Every payment event carries a PaymentInvoice as its data, so one
// decode serves them all. Failures are permanent: retrying will not fix them.
func decodePaymentEvent(routingKey string, body []byte) (string, usecases.PaymentOutcome, error) {
	status, ok := paymentStatusByEvent[routingKey]
	if !ok {
		return "", usecases.PaymentOutcome{}, fmt.Errorf("%w: unexpected routing key %q", events.ErrPermanent, routingKey)
	}
	var event struct {
		Data events.PaymentInvoice `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return "", usecases.PaymentOutcome{}, fmt.Errorf("%w: decode %s: %w", events.ErrPermanent, routingKey, err)
	}
	invoice := event.Data
	return invoice.InvoiceSource, usecases.PaymentOutcome{
		Status: status, PaymentID: invoice.PaymentID, PayerAccountID: invoice.PayerAccountID, Reference: invoice.Reference,
		AmountMinor: invoice.AmountMinor, Currency: invoice.Currency, ExpiresAt: invoice.ExpiresAt,
	}, nil
}
