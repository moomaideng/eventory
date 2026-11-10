package usecases

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/payment/models"
)

const eventSource = "payment-service"

func confirmedEvent(invoice *models.Invoice, payerEmail string) (models.OutboxEvent, error) {
	return outboxEvent(events.PaymentConfirmedType, invoice.ID, events.PaymentConfirmed{
		SpecVersion:   events.EventSpecVersion,
		ID:            uuid.New(),
		Type:          events.PaymentConfirmedType,
		Source:        eventSource,
		OccurredAt:    *invoice.PaidAt,
		CorrelationID: uuid.New(),
		Data: events.PaymentConfirmedData{
			PaymentInvoice: eventInvoice(invoice),
			Description:    invoice.Description,
			Recipient:      events.EventRecipient{AccountID: invoice.PayerAccountID, Email: payerEmail},
		},
	})
}

func cancelledEvent(invoice *models.Invoice) (models.OutboxEvent, error) {
	return outboxEvent(events.PaymentCancelledType, invoice.ID, events.PaymentCancelled{
		SpecVersion:   events.EventSpecVersion,
		ID:            uuid.New(),
		Type:          events.PaymentCancelledType,
		Source:        eventSource,
		OccurredAt:    *invoice.CancelledAt,
		CorrelationID: uuid.New(),
		Data:          eventInvoice(invoice),
	})
}

func expiredEvent(invoice *models.Invoice) (models.OutboxEvent, error) {
	return outboxEvent(events.PaymentExpiredType, invoice.ID, events.PaymentExpired{
		SpecVersion:   events.EventSpecVersion,
		ID:            uuid.New(),
		Type:          events.PaymentExpiredType,
		Source:        eventSource,
		OccurredAt:    invoice.ExpiresAt,
		CorrelationID: uuid.New(),
		Data:          eventInvoice(invoice),
	})
}

func eventInvoice(invoice *models.Invoice) events.PaymentInvoice {
	return events.PaymentInvoice{
		PaymentID:      invoice.ID,
		PayerAccountID: invoice.PayerAccountID,
		InvoiceSource:  invoice.Source,
		Reference:      invoice.Reference,
		AmountMinor:    invoice.AmountMinor,
		Currency:       invoice.Currency,
		ExpiresAt:      invoice.ExpiresAt,
	}
}

// outboxEvent keys the event by invoice and type: an invoice reaches each of
// those outcomes once, so a repeat write is a no-op.
func outboxEvent(routingKey string, invoiceID uuid.UUID, payload any) (models.OutboxEvent, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return models.OutboxEvent{}, err
	}
	return models.OutboxEvent{RoutingKey: routingKey, DedupKey: invoiceID.String() + ":" + routingKey, Payload: body}, nil
}
