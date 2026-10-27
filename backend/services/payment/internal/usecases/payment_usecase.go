package usecases

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/payment/internal/repositories"
	"github.com/moomaideng/eventory/services/payment/models"
	"gorm.io/gorm"
)

var ErrInvalid = errors.New("invalid payment input")

type CreatePayment struct {
	PayerAccountID uuid.UUID `json:"payerAccountId"`
	Source         string    `json:"source" minLength:"1" maxLength:"64"`
	Reference      string    `json:"reference" minLength:"1" maxLength:"128"`
	IdempotencyKey string    `json:"idempotencyKey" minLength:"1" maxLength:"128"`
	Description    string    `json:"description" minLength:"1" maxLength:"255"`
	AmountMinor    int64     `json:"amountMinor" minimum:"1" maximum:"9007199254740991" doc:"Amount in satang; 100 satang = 1 THB"`
	Currency       string    `json:"currency" enum:"THB"`
	ExpiresAt      time.Time `json:"expiresAt" doc:"Caller-selected deadline; must be in the future for a new invoice"`
}

type PaymentUseCase struct {
	repo repositories.PaymentRepository
}

func NewPaymentUseCase(repo repositories.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo}
}

func (uc *PaymentUseCase) Create(ctx context.Context, input CreatePayment) (*models.Invoice, error) {
	if input.PayerAccountID == uuid.Nil || input.AmountMinor <= 0 || input.AmountMinor > 9007199254740991 || input.Currency != "THB" {
		return nil, ErrInvalid
	}
	for _, field := range []struct {
		value string
		max   int
	}{{input.Source, 64}, {input.Reference, 128}, {input.IdempotencyKey, 128}, {input.Description, 255}} {
		if strings.TrimSpace(field.value) == "" || utf8.RuneCountInString(field.value) > field.max {
			return nil, ErrInvalid
		}
	}
	invoice := &models.Invoice{ID: uuid.New(), PayerAccountID: input.PayerAccountID,
		Source: input.Source, Reference: input.Reference, IdempotencyKey: input.IdempotencyKey, Description: input.Description,
		AmountMinor: input.AmountMinor, Currency: input.Currency, ExpiresAt: input.ExpiresAt.UTC().Truncate(time.Microsecond), Provider: "MOCK"}
	// Exact retries still return the original invoice after its deadline passes.
	existing, err := uc.repo.FindByKey(ctx, input.Source, input.IdempotencyKey)
	if err == nil {
		if !repositories.SameRequest(existing, invoice) {
			return nil, repositories.ErrConflict
		}
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// A durable team attempt can be recovered after an outage. Creating its
	// already-expired invoice is safe: CompleteMock cannot pay it.
	if !invoice.ExpiresAt.After(time.Now()) && input.Source != "TEAM_REGISTRATION" && input.Source != "SOLO_REGISTRATION" && input.Source != "SPONSORSHIP" {
		return nil, ErrInvalid
	}
	return uc.repo.Create(ctx, invoice)
}

func (uc *PaymentUseCase) GetOwned(ctx context.Context, payer, id uuid.UUID) (*models.Invoice, error) {
	return uc.repo.GetOwned(ctx, payer, id)
}

func (uc *PaymentUseCase) GetForService(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	return uc.repo.Get(ctx, id)
}

func (uc *PaymentUseCase) RefundRegistration(ctx context.Context, invoiceID, entryID uuid.UUID) (*models.Invoice, error) {
	if invoiceID == uuid.Nil || entryID == uuid.Nil {
		return nil, ErrInvalid
	}
	return uc.repo.Refund(ctx, invoiceID, func(invoice *models.Invoice) error {
		if invoice.Source != "TEAM_REGISTRATION" || !strings.HasPrefix(invoice.Reference, entryID.String()+":") || invoice.PaidAt == nil {
			return repositories.ErrConflict
		}
		if invoice.RefundedAt != nil {
			return nil
		}
		now := time.Now().UTC().Truncate(time.Microsecond)
		invoice.RefundedAt = &now
		invoice.RefundReason = "APPLICATION_REJECTED"
		return nil
	})
}

// Date filters include both selected days in UTC, matching history display.
func HistoryFilter(from, to string, page, limit int) (repositories.Filter, error) {
	filter := repositories.Filter{Page: page, Limit: limit}
	if page < 1 || page > 1000000 || limit < 1 || limit > 100 {
		return filter, ErrInvalid
	}
	if from != "" {
		date, err := time.Parse("2006-01-02", from)
		if err != nil {
			return filter, ErrInvalid
		}
		filter.From = &date
	}
	if to != "" {
		date, err := time.Parse("2006-01-02", to)
		if err != nil {
			return filter, ErrInvalid
		}
		if filter.From != nil && date.Before(*filter.From) {
			return filter, ErrInvalid
		}
		until := date.AddDate(0, 0, 1)
		filter.Until = &until
	}
	return filter, nil
}

func (uc *PaymentUseCase) ListOwned(ctx context.Context, payer uuid.UUID, from, to string, page, limit int) ([]models.Invoice, int64, error) {
	filter, err := HistoryFilter(from, to, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return uc.repo.ListOwned(ctx, payer, filter)
}

func (uc *PaymentUseCase) CompleteMock(ctx context.Context, payer, id uuid.UUID, result models.MockResult) (*models.Invoice, error) {
	if result != models.Succeeded && result != models.Failed && result != models.Cancelled {
		return nil, ErrInvalid
	}
	return uc.repo.Complete(ctx, payer, id, func(payment *models.Invoice) error {
		if payment.Provider != "MOCK" {
			return repositories.ErrConflict
		}
		if payment.PaidAt != nil && result == models.Succeeded {
			return nil
		}
		now := time.Now().UTC().Truncate(time.Microsecond)
		if payment.Status(now) != models.Unpaid {
			return repositories.ErrConflict
		}
		if result == models.Succeeded {
			payment.PaidAt = &now
		} else if payment.Source == "TEAM_REGISTRATION" || payment.Source == "SOLO_REGISTRATION" || payment.Source == "SPONSORSHIP" {
			// Close this attempt before the tournament service unlocks the roster.
			payment.CancelledAt = &now
		}
		return nil
	})
}

func (uc *PaymentUseCase) GetReceipt(ctx context.Context, payer, id uuid.UUID) (*models.Invoice, error) {
	invoice, err := uc.repo.GetOwned(ctx, payer, id)
	if err != nil {
		return nil, err
	}
	if invoice.PaidAt == nil || invoice.Receipt == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return invoice, nil
}
