package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/payment/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrConflict = errors.New("invoice request conflicts with an existing invoice")

type Filter struct {
	From, Until *time.Time
	Page, Limit int
}

type PaymentRepository interface {
	Create(context.Context, *models.Invoice) (*models.Invoice, error)
	FindByKey(context.Context, string, string) (*models.Invoice, error)
	Get(context.Context, uuid.UUID) (*models.Invoice, error)
	GetOwned(context.Context, uuid.UUID, uuid.UUID) (*models.Invoice, error)
	ListOwned(context.Context, uuid.UUID, Filter) ([]models.Invoice, int64, error)
	Complete(context.Context, uuid.UUID, uuid.UUID, func(*models.Invoice) error) (*models.Invoice, error)
}

type paymentRepository struct{ db *gorm.DB }

func NewPaymentRepository(db *gorm.DB) PaymentRepository { return &paymentRepository{db} }

func (r *paymentRepository) FindByKey(ctx context.Context, source, key string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.WithContext(ctx).Preload("Receipt").Where("source = ? AND idempotency_key = ?", source, key).First(&invoice).Error
	return &invoice, err
}

func SameRequest(existing, requested *models.Invoice) bool {
	return existing.PayerAccountID == requested.PayerAccountID && existing.Reference == requested.Reference &&
		existing.AmountMinor == requested.AmountMinor && existing.Currency == requested.Currency && existing.Description == requested.Description &&
		existing.ExpiresAt.Equal(requested.ExpiresAt)
}

func (r *paymentRepository) Create(ctx context.Context, payment *models.Invoice) (*models.Invoice, error) {
	// ON CONFLICT waits for concurrent inserts before checking the winning key.
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}, clause.Returning{}).Create(payment)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 1 {
		return payment, nil
	}
	var existing models.Invoice
	err := r.db.WithContext(ctx).Preload("Receipt").Where("source = ? AND idempotency_key = ?", payment.Source, payment.IdempotencyKey).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrConflict
	}
	if err != nil {
		return nil, err
	}
	if !SameRequest(&existing, payment) {
		return nil, ErrConflict
	}
	return &existing, nil
}

func (r *paymentRepository) GetOwned(ctx context.Context, payer, id uuid.UUID) (*models.Invoice, error) {
	var payment models.Invoice
	err := r.db.WithContext(ctx).Preload("Receipt").Where("id = ? AND payer_account_id = ?", id, payer).First(&payment).Error
	return &payment, err
}

func (r *paymentRepository) Get(ctx context.Context, id uuid.UUID) (*models.Invoice, error) {
	var payment models.Invoice
	err := r.db.WithContext(ctx).Preload("Receipt").Where("id = ?", id).First(&payment).Error
	return &payment, err
}

func (r *paymentRepository) ListOwned(ctx context.Context, payer uuid.UUID, filter Filter) ([]models.Invoice, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("payer_account_id = ?", payer)
	if filter.From != nil {
		query = query.Where("created_at >= ?", *filter.From)
	}
	if filter.Until != nil {
		query = query.Where("created_at < ?", *filter.Until)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	payments := make([]models.Invoice, 0)
	err := query.Preload("Receipt").Order("created_at DESC, id DESC").Limit(filter.Limit).Offset((filter.Page - 1) * filter.Limit).Find(&payments).Error
	return payments, total, err
}

func (r *paymentRepository) Complete(ctx context.Context, payer, id uuid.UUID, mutate func(*models.Invoice) error) (*models.Invoice, error) {
	var payment models.Invoice
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Receipt").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND payer_account_id = ?", id, payer).First(&payment).Error; err != nil {
			return err
		}
		previouslyPaid := payment.PaidAt != nil
		if err := mutate(&payment); err != nil {
			return err
		}
		if previouslyPaid || payment.PaidAt == nil {
			return nil
		}
		payment.Receipt = &models.Receipt{ID: uuid.New(), InvoiceID: payment.ID, CreatedAt: *payment.PaidAt}
		if err := tx.Create(payment.Receipt).Error; err != nil {
			return err
		}
		return tx.Model(&payment).Omit("Receipt").Update("paid_at", payment.PaidAt).Error
	})
	return &payment, err
}
