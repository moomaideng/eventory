package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type sponsorshipPaymentStub struct{ payment *ports.RegistrationPayment }

func (s sponsorshipPaymentStub) EnsureInvoice(context.Context, ports.RegistrationPaymentRequest) (*ports.RegistrationPayment, error) {
	return s.payment, nil
}

func TestSponsorshipRejectsMismatchedPayment(t *testing.T) {
	payer, pledgeID, paymentID := uuid.New(), uuid.New(), uuid.New()
	expires := time.Now().UTC().Add(time.Minute)
	for _, tc := range []struct {
		name   string
		change func(*ports.RegistrationPayment)
	}{
		{"missing ID", func(p *ports.RegistrationPayment) { p.ID = uuid.Nil }},
		{"different payer", func(p *ports.RegistrationPayment) { p.PayerAccountID = uuid.New() }},
		{"different source", func(p *ports.RegistrationPayment) { p.Source = "TEAM_REGISTRATION" }},
		{"different pledge", func(p *ports.RegistrationPayment) { p.Reference = uuid.NewString() }},
		{"different price", func(p *ports.RegistrationPayment) { p.AmountMinor++ }},
		{"different currency", func(p *ports.RegistrationPayment) { p.Currency = "USD" }},
		{"different deadline", func(p *ports.RegistrationPayment) { p.ExpiresAt = p.ExpiresAt.Add(time.Second) }},
		{"different invoice", func(p *ports.RegistrationPayment) { p.ID = uuid.New() }},
		{"unknown status", func(p *ports.RegistrationPayment) { p.Status = "INVALID" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &ports.RegistrationPayment{ID: paymentID, PayerAccountID: payer, Source: SponsorshipPaymentSource, Reference: pledgeID.String(), AmountMinor: 10000, Currency: "THB", ExpiresAt: expires, Status: "PAID"}
			tc.change(p)
			pledge := &models.SponsorshipPledge{ID: pledgeID, PayerAccountID: payer, AmountMinor: 10000, Currency: "THB", ExpiresAt: expires, Status: models.PledgePending, PaymentID: &paymentID}
			// Invalid payment data must be rejected before any repository transaction starts.
			u := NewSponsorshipUseCase(nil, nil, nil, sponsorshipPaymentStub{p})
			if err := u.Reconcile(context.Background(), pledge); !errors.Is(err, ErrPaymentUnavailable) {
				t.Fatalf("mismatched payment was accepted: %v", err)
			}
		})
	}
}
