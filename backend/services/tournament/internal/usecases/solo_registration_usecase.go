package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var ErrSoloAnswersLocked = errors.New("submitted answers cannot be changed")

type SoloRegistrationUseCase struct {
	repo        *repositories.SoloRegistrationRepository
	tournaments repositories.TournamentRepository
	forms       *RegistrationFormUseCase
	accounts    ports.AccountService
	payments    ports.PaymentService
}

func NewSoloRegistrationUseCase(repo *repositories.SoloRegistrationRepository, tournaments repositories.TournamentRepository, forms *RegistrationFormUseCase, accounts ports.AccountService, payments ports.PaymentService) *SoloRegistrationUseCase {
	return &SoloRegistrationUseCase{repo, tournaments, forms, accounts, payments}
}

type SoloEligibility struct {
	CanRegister bool       `json:"canRegister"`
	Reason      string     `json:"reason"`
	Status      string     `json:"status"`
	PaymentID   *uuid.UUID `json:"paymentId,omitempty"`
}

func (u *SoloRegistrationUseCase) Eligibility(ctx context.Context, tournamentID, accountID uuid.UUID) (SoloEligibility, error) {
	if _, err := requireActiveAccount(ctx, u.accounts, accountID); err != nil {
		return SoloEligibility{}, err
	}
	tournament, err := u.tournaments.GetPublishedByID(ctx, tournamentID)
	if err != nil {
		return SoloEligibility{}, err
	}
	if tournament.RegistrationMode != models.TournamentRegistrationModeSolo {
		return SoloEligibility{Reason: "NOT_SOLO"}, nil
	}
	row, err := u.repo.Get(ctx, tournamentID, accountID)
	if err != nil && !errors.Is(err, repositories.ErrSoloNotFound) {
		return SoloEligibility{}, err
	}
	if err == nil {
		switch row.Status {
		case models.SoloRegistered:
			return SoloEligibility{Reason: "ALREADY_REGISTERED", Status: string(row.Status)}, nil
		case models.SoloPaymentPending:
			return SoloEligibility{Reason: "PAYMENT_PENDING", Status: string(row.Status), PaymentID: row.PaymentID}, nil
		}
	}
	if tournament.Status != models.TournamentStatusRegistrationOpen || !time.Now().UTC().Before(tournament.RegistrationDeadline) {
		return SoloEligibility{Reason: "REGISTRATION_CLOSED"}, nil
	}
	count, err := u.repo.Count(ctx, tournamentID, true)
	if err != nil {
		return SoloEligibility{}, err
	}
	if count >= int64(tournament.Capacity) {
		return SoloEligibility{Reason: "TOURNAMENT_FULL"}, nil
	}
	return SoloEligibility{CanRegister: true, Reason: "OPEN", Status: ""}, nil
}

func (u *SoloRegistrationUseCase) Start(ctx context.Context, tournamentID, accountID uuid.UUID, answers []models.FieldAnswer) (*models.SoloRegistration, error) {
	if _, err := requireActiveAccount(ctx, u.accounts, accountID); err != nil {
		return nil, err
	}
	// An existing attempt is idempotent. Reconcile first so a finished mock
	// payment is never replaced with a new invoice.
	if existing, err := u.repo.Get(ctx, tournamentID, accountID); err == nil {
		if existing.Status == models.SoloPaymentPending {
			if err := u.Reconcile(ctx, existing); err != nil {
				return nil, err
			}
			existing, err = u.repo.Get(ctx, tournamentID, accountID)
			if err != nil {
				return nil, err
			}
		}
		if existing.Status == models.SoloRegistered || existing.Status == models.SoloPaymentPending {
			return existing, nil
		}
	} else if !errors.Is(err, repositories.ErrSoloNotFound) {
		return nil, err
	}
	eligibility, err := u.Eligibility(ctx, tournamentID, accountID)
	if err != nil {
		return nil, err
	}
	switch eligibility.Reason {
	case "NOT_SOLO":
		return nil, repositories.ErrSoloWrongMode
	case "REGISTRATION_CLOSED":
		return nil, repositories.ErrSoloClosed
	case "TOURNAMENT_FULL":
		return nil, repositories.ErrSoloFull
	}
	if u.forms != nil {
		_, err := u.forms.SubmitResponse(ctx, accountID, tournamentID, answers)
		if errors.Is(err, ErrResponseAlreadyExists) {
			stored, getErr := u.forms.GetMyResponse(ctx, accountID, tournamentID)
			if getErr != nil {
				return nil, getErr
			}
			original, _ := json.Marshal(stored.Answers)
			proposed, _ := json.Marshal(answers)
			if string(original) != string(proposed) {
				return nil, ErrSoloAnswersLocked
			}
		} else if errors.Is(err, ErrFormNotFound) && len(answers) == 0 {
			// A tournament without custom questions can still accept solo players.
		} else if err != nil {
			return nil, err
		}
	} else if len(answers) != 0 {
		return nil, ErrUnrecognizedField
	}
	row, err := u.repo.Begin(ctx, tournamentID, accountID)
	if err != nil {
		return nil, err
	}
	if row.Status == models.SoloPaymentPending {
		if err := u.Reconcile(ctx, row); err != nil {
			return nil, err
		}
		return u.repo.Get(ctx, tournamentID, accountID)
	}
	return row, nil
}

func (u *SoloRegistrationUseCase) Refresh(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.SoloRegistration, error) {
	row, err := u.repo.Get(ctx, tournamentID, accountID)
	if err != nil {
		return nil, err
	}
	if err := u.Reconcile(ctx, row); err != nil {
		return nil, err
	}
	return u.repo.Get(ctx, tournamentID, accountID)
}

func (u *SoloRegistrationUseCase) Reconcile(ctx context.Context, row *models.SoloRegistration) error {
	if row.Status != models.SoloPaymentPending {
		return nil
	}
	if row.PaymentAttemptID == nil || row.PaymentExpiresAt == nil {
		return ErrPaymentUnavailable
	}
	request := ports.RegistrationPaymentRequest{PayerAccountID: row.AccountID, Source: "SOLO_REGISTRATION", Reference: row.TournamentID.String() + ":" + row.PaymentAttemptID.String(), IdempotencyKey: row.PaymentAttemptID.String(), Description: "Solo registration " + row.TournamentID.String(), AmountMinor: row.PaymentAmountMinor, Currency: "THB", ExpiresAt: *row.PaymentExpiresAt}
	payment, err := u.payments.EnsureInvoice(ctx, request)
	if err != nil {
		return ErrPaymentUnavailable
	}
	if payment.ID == uuid.Nil || payment.PayerAccountID != request.PayerAccountID || payment.Source != request.Source || payment.Reference != request.Reference || payment.AmountMinor != request.AmountMinor || payment.Currency != request.Currency || !payment.ExpiresAt.Equal(request.ExpiresAt) {
		return ErrPaymentUnavailable
	}
	if row.PaymentID != nil && *row.PaymentID != payment.ID {
		return ErrPaymentUnavailable
	}
	switch payment.Status {
	case "UNPAID", "PAID", "EXPIRED", "CANCELLED":
		return u.repo.ApplyPayment(ctx, row, payment.ID, payment.Status)
	default:
		return ErrPaymentUnavailable
	}
}

func (u *SoloRegistrationUseCase) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rows, err := u.repo.Pending(ctx)
			if err != nil {
				log.Printf("list pending solo registrations: %v", err)
				continue
			}
			for i := range rows {
				if ctx.Err() != nil {
					return
				}
				if err := u.Reconcile(ctx, &rows[i]); err != nil {
					log.Printf("reconcile solo registration %s: %v", rows[i].ID, err)
				}
			}
		}
	}
}
