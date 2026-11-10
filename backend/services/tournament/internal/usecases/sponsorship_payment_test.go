package usecases

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type sponsorshipPaymentStub struct{ payment *ports.RegistrationPayment }

func (s sponsorshipPaymentStub) EnsureInvoice(context.Context, ports.RegistrationPaymentRequest) (*ports.RegistrationPayment, error) {
	return s.payment, nil
}

func TestSponsorshipRepositoryErrorsBecomeDomainErrors(t *testing.T) {
	unexpected := errors.New("database unavailable")
	for _, tc := range []struct{ repository, domain error }{
		{repositories.ErrPledgeNotFound, ErrSponsorshipNotFound},
		{repositories.ErrTournamentNotFound, ErrTournamentNotFound},
		{repositories.ErrSponsorshipTierNotFound, ErrSponsorshipTierNotFound},
		{repositories.ErrSponsorshipTierNameDuplicate, ErrSponsorshipTierNameDuplicate},
		{ErrSponsorRequired, ErrSponsorRequired},
		{unexpected, unexpected},
	} {
		if got := mapSponsorshipError(fmt.Errorf("lookup: %w", tc.repository)); !errors.Is(got, tc.domain) {
			t.Errorf("want %v, got %v", tc.domain, got)
		}
	}
	if mapSponsorshipError(nil) != nil {
		t.Fatal("successful operation returned an error")
	}
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
			p := &ports.RegistrationPayment{ID: paymentID, PayerAccountID: payer, Source: models.PaymentSourceSponsorship, Reference: pledgeID.String(), AmountMinor: 10000, Currency: "THB", ExpiresAt: expires, Status: "PAID"}
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

type pledgeLookupStub struct {
	repositories.SponsorshipRepository // Begin is unset: reaching it fails the test
	pledge                             *models.SponsorshipPledge
}

func (s pledgeLookupStub) Get(context.Context, uuid.UUID) (*models.SponsorshipPledge, error) {
	if s.pledge == nil {
		return nil, repositories.ErrPledgeNotFound
	}
	return s.pledge, nil
}

func TestSponsorshipPaymentOutcomeChecksThePledgeBeforeSettling(t *testing.T) {
	payer, pledgeID, paymentID := uuid.New(), uuid.New(), uuid.New()
	expires := time.Now().UTC().Add(time.Minute).Truncate(time.Microsecond)
	pledge := func(status models.PledgeStatus) *models.SponsorshipPledge {
		return &models.SponsorshipPledge{ID: pledgeID, PayerAccountID: payer, AmountMinor: 10000, Currency: "THB", ExpiresAt: expires, Status: status}
	}
	for _, tc := range []struct {
		name    string
		pledge  *models.SponsorshipPledge
		change  func(*PaymentOutcome)
		wantErr error
	}{
		{"deleted pledge is ignored", nil, nil, nil},
		{"settled pledge is left alone", pledge(models.PledgeConfirmed), nil, nil},
		{"malformed reference", pledge(models.PledgePending), func(o *PaymentOutcome) { o.Reference = "nope" }, ErrPaymentMismatch},
		{"different payer", pledge(models.PledgePending), func(o *PaymentOutcome) { o.PayerAccountID = uuid.New() }, ErrPaymentMismatch},
		{"different price", pledge(models.PledgePending), func(o *PaymentOutcome) { o.AmountMinor++ }, ErrPaymentMismatch},
		{"different currency", pledge(models.PledgePending), func(o *PaymentOutcome) { o.Currency = "USD" }, ErrPaymentMismatch},
		{"different deadline", pledge(models.PledgePending), func(o *PaymentOutcome) { o.ExpiresAt = o.ExpiresAt.Add(time.Second) }, ErrPaymentMismatch},
		{"different invoice", func() *models.SponsorshipPledge {
			p := pledge(models.PledgePending)
			other := uuid.New()
			p.PaymentID = &other
			return p
		}(), nil, ErrPaymentMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			outcome := PaymentOutcome{Status: models.PaymentPaid, PaymentID: paymentID, PayerAccountID: payer, Reference: pledgeID.String(), AmountMinor: 10000, Currency: "THB", ExpiresAt: expires}
			if tc.change != nil {
				tc.change(&outcome)
			}
			u := NewSponsorshipUseCase(pledgeLookupStub{pledge: tc.pledge}, nil, nil, nil)
			if err := u.ApplyPaymentOutcome(context.Background(), outcome); !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}
