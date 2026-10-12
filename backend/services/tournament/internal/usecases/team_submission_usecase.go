package usecases

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/models"
	"log"
	"time"
)

var ErrPaymentUnavailable = errors.New("payment status could not be verified; roster remains locked, please retry")

type SubmissionRepository interface {
	Pending(context.Context) ([]models.TournamentTeam, error)
	ApplyPayment(context.Context, *models.TournamentTeam, uuid.UUID, string) error
}

type TeamSubmissionUseCase struct {
	repo     SubmissionRepository
	payments ports.PaymentService
}

func NewTeamSubmissionUseCase(repo SubmissionRepository, payments ports.PaymentService) *TeamSubmissionUseCase {
	return &TeamSubmissionUseCase{repo, payments}
}

func (u *TeamSubmissionUseCase) Reconcile(ctx context.Context, team *models.TournamentTeam) error {
	if team.Status != models.TournamentTeamStatusPaymentPending {
		return nil
	}
	if team.PaymentAttemptID == nil || team.PaymentPayerID == nil || team.PaymentExpiresAt == nil {
		return ErrPaymentUnavailable
	}
	request := ports.RegistrationPaymentRequest{
		PayerAccountID: *team.PaymentPayerID, Source: "TEAM_REGISTRATION",
		Reference: team.ID.String() + ":" + team.PaymentAttemptID.String(), IdempotencyKey: team.PaymentAttemptID.String(),
		Description: "Team registration " + team.ID.String(), AmountMinor: team.PaymentAmountMinor, Currency: "THB", ExpiresAt: *team.PaymentExpiresAt,
	}
	payment, err := u.payments.EnsureInvoice(ctx, request)
	if err != nil {
		return ErrPaymentUnavailable
	}
	if payment.ID == uuid.Nil || payment.PayerAccountID != request.PayerAccountID || payment.Source != request.Source || payment.Reference != request.Reference || payment.AmountMinor != request.AmountMinor || payment.Currency != request.Currency || !payment.ExpiresAt.Equal(request.ExpiresAt) {
		return ErrPaymentUnavailable
	}
	if team.PaymentID != nil && *team.PaymentID != payment.ID {
		return ErrPaymentUnavailable
	}
	switch payment.Status {
	case "UNPAID", "PAID", "EXPIRED", "CANCELLED":
		return u.repo.ApplyPayment(ctx, team, payment.ID, payment.Status)
	default:
		return ErrPaymentUnavailable
	}
}

// Durable attempts are recovered even if the captain closes the checkout or a
// service restarts between payment confirmation and application submission.
func (u *TeamSubmissionUseCase) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			teams, err := u.repo.Pending(ctx)
			if err != nil {
				log.Printf("list pending team submissions: %v", err)
				continue
			}
			for i := range teams {
				if ctx.Err() != nil {
					return
				}
				if err := u.Reconcile(ctx, &teams[i]); err != nil {
					log.Printf("reconcile team %s: %v", teams[i].ID, err)
				}
			}
		}
	}
}
