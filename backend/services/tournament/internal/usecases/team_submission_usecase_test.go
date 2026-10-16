package usecases_test

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"testing"
	"time"
)

type submissionRepoStub struct{ applied string }

func (r *submissionRepoStub) Pending(context.Context) ([]models.TournamentTeam, error) {
	return nil, nil
}
func (r *submissionRepoStub) ApplyPayment(_ context.Context, _ *models.TournamentTeam, _ uuid.UUID, status string) error {
	r.applied = status
	return nil
}

type registrationPaymentStub struct {
	mutate func(*ports.RegistrationPayment)
	err    error
}

func (p registrationPaymentStub) EnsureInvoice(_ context.Context, r ports.RegistrationPaymentRequest) (*ports.RegistrationPayment, error) {
	payment := &ports.RegistrationPayment{ID: uuid.New(), PayerAccountID: r.PayerAccountID, Source: r.Source, Reference: r.Reference, AmountMinor: r.AmountMinor, Currency: r.Currency, ExpiresAt: r.ExpiresAt, Status: "PAID"}
	if p.mutate != nil {
		p.mutate(payment)
	}
	return payment, p.err
}

func TestSubmissionRequiresAuthoritativeMatchingInvoice(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*ports.RegistrationPayment)
		err    error
	}{
		{name: "unavailable", err: errors.New("timeout after creating invoice")},
		{name: "wrong payer", mutate: func(p *ports.RegistrationPayment) { p.PayerAccountID = uuid.New() }},
		{name: "wrong amount", mutate: func(p *ports.RegistrationPayment) { p.AmountMinor++ }},
		{name: "wrong team or attempt", mutate: func(p *ports.RegistrationPayment) { p.Reference = uuid.NewString() }},
		{name: "wrong currency", mutate: func(p *ports.RegistrationPayment) { p.Currency = "USD" }},
		{name: "wrong source", mutate: func(p *ports.RegistrationPayment) { p.Source = "OTHER" }},
		{name: "wrong deadline", mutate: func(p *ports.RegistrationPayment) { p.ExpiresAt = p.ExpiresAt.Add(time.Second) }},
		{name: "unknown status", mutate: func(p *ports.RegistrationPayment) { p.Status = "UNKNOWN" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			attempt, payer := uuid.New(), uuid.New()
			expires := time.Now().Add(time.Minute)
			team := &models.TournamentTeam{ID: uuid.New(), Status: models.TournamentTeamStatusPaymentPending, PaymentAttemptID: &attempt, PaymentPayerID: &payer, PaymentExpiresAt: &expires, PaymentAmountMinor: 10000}
			repo := &submissionRepoStub{}
			uc := usecases.NewTeamSubmissionUseCase(repo, registrationPaymentStub{tc.mutate, tc.err})
			if err := uc.Reconcile(context.Background(), team); !errors.Is(err, usecases.ErrPaymentUnavailable) {
				t.Fatalf("error = %v", err)
			}
			if repo.applied != "" {
				t.Fatal("unverified payment changed roster")
			}
			// A later retry of the same durable attempt can complete after an outage.
			uc = usecases.NewTeamSubmissionUseCase(repo, registrationPaymentStub{})
			if err := uc.Reconcile(context.Background(), team); err != nil {
				t.Fatal(err)
			}
			if repo.applied != "PAID" {
				t.Fatal("paid attempt was not submitted")
			}
		})
	}
}
