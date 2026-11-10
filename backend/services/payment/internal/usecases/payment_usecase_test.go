package usecases

import (
	"context"
	"encoding/json"
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

// mockPaymentRepo runs the use case's mutation on one invoice and records the
// outbox events it asks to store with the change.
type mockPaymentRepo struct {
	repositories.PaymentRepository
	invoice *models.Invoice
	queued  []models.OutboxEvent
}

func (m *mockPaymentRepo) Complete(_ context.Context, _, _ uuid.UUID, fn func(*models.Invoice) ([]models.OutboxEvent, error)) (*models.Invoice, error) {
	queued, err := fn(m.invoice)
	if err != nil {
		return nil, err
	}
	m.queued = append(m.queued, queued...)
	return m.invoice, nil
}

func TestCompleteMockQueuesTheOutcomeEvent(t *testing.T) {
	for _, tc := range []struct {
		name     string
		source   string
		result   models.MockResult
		wantKey  string // empty: no event
		wantDone func(*models.Invoice) bool
	}{
		{"success", "TEAM_REGISTRATION", models.Succeeded, events.PaymentConfirmedType, func(i *models.Invoice) bool { return i.PaidAt != nil }},
		{"failed registration payment closes the attempt", "TEAM_REGISTRATION", models.Failed, events.PaymentCancelledType, func(i *models.Invoice) bool { return i.CancelledAt != nil }},
		{"cancelled sponsorship closes the attempt", "SPONSORSHIP", models.Cancelled, events.PaymentCancelledType, func(i *models.Invoice) bool { return i.CancelledAt != nil }},
		{"failed payment for another source stays retryable", "OTHER", models.Failed, "", func(i *models.Invoice) bool { return i.CancelledAt == nil && i.PaidAt == nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			invoice := &models.Invoice{
				ID: uuid.New(), PayerAccountID: uuid.New(), Provider: "MOCK", Source: tc.source, Reference: "team:attempt",
				AmountMinor: 50000, Currency: "THB", ExpiresAt: time.Now().Add(time.Hour),
			}
			repo := &mockPaymentRepo{invoice: invoice}

			_, err := NewPaymentUseCase(repo).CompleteMock(context.Background(), invoice.PayerAccountID, invoice.ID, tc.result)
			require.NoError(t, err)

			assert.True(t, tc.wantDone(invoice))
			if tc.wantKey == "" {
				require.Empty(t, repo.queued)
				return
			}
			require.Len(t, repo.queued, 1)
			queued := repo.queued[0]
			assert.Equal(t, tc.wantKey, queued.RoutingKey)
			assert.Equal(t, invoice.ID.String()+":"+tc.wantKey, queued.DedupKey)
			var body struct {
				Data events.PaymentInvoice `json:"data"`
			}
			require.NoError(t, json.Unmarshal(queued.Payload, &body))
			assert.Equal(t, invoice.ID, body.Data.PaymentID)
			assert.Equal(t, tc.source, body.Data.InvoiceSource)
			assert.Equal(t, "team:attempt", body.Data.Reference)
			assert.Equal(t, int64(50000), body.Data.AmountMinor)
			assert.True(t, invoice.ExpiresAt.Equal(body.Data.ExpiresAt))
		})
	}
}

func TestCompleteMockReplayQueuesNothing(t *testing.T) {
	paidAt := time.Now().UTC()
	invoice := &models.Invoice{ID: uuid.New(), PayerAccountID: uuid.New(), Provider: "MOCK", Source: "TEAM_REGISTRATION", PaidAt: &paidAt, ExpiresAt: time.Now().Add(time.Hour)}
	repo := &mockPaymentRepo{invoice: invoice}

	_, err := NewPaymentUseCase(repo).CompleteMock(context.Background(), invoice.PayerAccountID, invoice.ID, models.Succeeded)
	require.NoError(t, err)
	require.Empty(t, repo.queued)
}
