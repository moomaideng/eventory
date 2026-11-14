package seeds

import (
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/payment/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func SeedPayments(db *gorm.DB) error {
	payer := uuid.MustParse("99999999-0000-4000-8000-000000000001")
	for _, status := range []models.Status{models.Unpaid, models.Paid, models.Expired} {
		now := time.Now().UTC().Truncate(time.Microsecond)
		key := string(status)
		payment := models.Invoice{ID: uuid.NewSHA1(uuid.NameSpaceURL, []byte("eventory/mock-invoice/"+key)),
			PayerAccountID: payer, Source: "seed", Reference: "sample-" + key, IdempotencyKey: key,
			Description: "Sample mock invoice — " + key, AmountMinor: 15000, Currency: "THB", Provider: "MOCK", ExpiresAt: now.Add(24 * time.Hour)}
		if status == models.Expired {
			payment.ExpiresAt = now.Add(-time.Hour)
		}
		if status == models.Paid {
			payment.PaidAt = &now
			payment.Receipt = &models.Receipt{ID: uuid.NewSHA1(uuid.NameSpaceURL, []byte("eventory/mock-receipt/"+key)), InvoiceID: payment.ID, CreatedAt: now}
		}
		// Seeding again must not reopen an already completed checkout.
		if err := db.Transaction(func(tx *gorm.DB) error {
			result := tx.Omit("Receipt").Clauses(clause.OnConflict{DoNothing: true}).Create(&payment)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 1 && payment.Receipt != nil {
				return tx.Create(payment.Receipt).Error
			}
			return nil
		}); err != nil {
			return err
		}
	}

	// Matches the confirmed Bronze pledge in the tournament demo seed.
	// Keeping both sides of the cross-service payment record makes the funding
	// total, sponsor list, and payment history describe the same checkout.
	now := time.Now().UTC().Truncate(time.Microsecond)
	pledgeID := uuid.MustParse("66666666-0000-4000-8000-000000000001")
	sponsorship := models.Invoice{
		ID:             uuid.MustParse("66666666-0000-4000-8000-000000000002"),
		PayerAccountID: uuid.MustParse("00000000-0000-4000-8000-000000000002"),
		Source:         models.SourceSponsorship,
		Reference:      pledgeID.String(),
		IdempotencyKey: "sponsorship:" + pledgeID.String(),
		Description:    "Sponsorship: [DEMO · Sponsor] Checkout Playground (Bronze)",
		AmountMinor:    100000,
		Currency:       "THB",
		Provider:       "MOCK",
		ExpiresAt:      now.Add(24 * time.Hour),
		PaidAt:         &now,
		Receipt: &models.Receipt{
			ID:        uuid.MustParse("66666666-0000-4000-8000-000000000003"),
			InvoiceID: uuid.MustParse("66666666-0000-4000-8000-000000000002"),
			CreatedAt: now,
		},
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.Omit("Receipt").Clauses(clause.OnConflict{DoNothing: true}).Create(&sponsorship)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 1 {
			return tx.Create(sponsorship.Receipt).Error
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
