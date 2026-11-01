package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/payment/internal/repositories"
	"github.com/moomaideng/eventory/services/payment/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryFilter(t *testing.T) {
	for _, tc := range []struct {
		from, to    string
		page, limit int
		valid       bool
	}{
		{"", "", 1, 20, true}, {"2024-02-29", "2024-02-29", 1, 20, true},
		{"2025-02-29", "", 1, 20, false}, {"2026-01-02", "2026-01-01", 1, 20, false},
		{"2026-1-01", "", 1, 20, false}, {"", "bad", 1, 20, false},
		{"", "", 0, 20, false}, {"", "", 1, 101, false}, {"", "", 1000001, 20, false},
	} {
		filter, err := HistoryFilter(tc.from, tc.to, tc.page, tc.limit)
		if (err == nil) != tc.valid {
			t.Errorf("%+v: %v", tc, err)
		}
		if tc.from != "" && tc.from == tc.to && tc.valid && filter.Until.Sub(*filter.From).Hours() != 24 {
			t.Error("End date must include the entire day")
		}
	}
}

type mockPaymentRepo struct {
	repositories.PaymentRepository
	invoice *models.Invoice
}

func (m *mockPaymentRepo) Complete(ctx context.Context, payer, id uuid.UUID, fn func(*models.Invoice) error) (*models.Invoice, error) {
	if err := fn(m.invoice); err != nil {
		return nil, err
	}
	return m.invoice, nil
}

type recordedEvent struct {
	routingKey string
	event      any
}

type mockPublisher struct {
	events []recordedEvent
}

func (m *mockPublisher) Publish(ctx context.Context, routingKey string, event any) error {
	m.events = append(m.events, recordedEvent{routingKey: routingKey, event: event})
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

func TestPaymentUseCase_CompleteMockPublishesEvent(t *testing.T) {
	payerID := uuid.New()
	invoiceID := uuid.New()

	invoice := &models.Invoice{
		ID:             invoiceID,
		PayerAccountID: payerID,
		Provider:       "MOCK",
		AmountMinor:    50000,
		Currency:       "THB",
		Description:    "Tournament Entry Fee",
		ExpiresAt:      time.Now().Add(1 * time.Hour),
	}

	pub := &mockPublisher{}
	repo := &mockPaymentRepo{invoice: invoice}
	uc := NewPaymentUseCase(repo, pub)

	res, err := uc.CompleteMock(context.Background(), payerID, invoiceID, models.Succeeded)
	require.NoError(t, err)
	assert.NotNil(t, res.PaidAt)

	require.Len(t, pub.events, 1)
	assert.Equal(t, events.PaymentConfirmedType, pub.events[0].routingKey)
	evt, ok := pub.events[0].event.(events.PaymentConfirmed)
	require.True(t, ok)
	assert.Equal(t, events.PaymentConfirmedType, evt.Type)
	assert.Equal(t, invoiceID, evt.Data.PaymentID)
	assert.Equal(t, payerID, evt.Data.PayerAccountID)
	assert.Equal(t, int64(50000), evt.Data.AmountMinor)
	assert.Equal(t, "THB", evt.Data.Currency)
	assert.Equal(t, "Tournament Entry Fee", evt.Data.Description)
}
