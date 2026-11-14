package seeds

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/seedids"
	"github.com/moomaideng/eventory/services/payment/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedPayments populates deterministic sponsorship and tournament registration
// invoices (with receipts for PAID and REFUNDED entries) linked to tournament fixtures.
func SeedPayments(db *gorm.DB) error {
	now := time.Now().UTC().Truncate(time.Second)

	var invoices []models.Invoice
	var receipts []models.Receipt

	for i, s := range seedids.SponsorshipInvoices() {
		created := now.Add(-time.Duration(72-i*2) * time.Hour)
		inv := models.Invoice{
			ID:             s.InvoiceID,
			PayerAccountID: s.PayerAccountID,
			Source:         models.SourceSponsorship,
			Reference:      s.PledgeID.String(),
			IdempotencyKey: "sponsorship:" + s.PledgeID.String(),
			Description:    fmt.Sprintf("Sponsorship: %s (%s)", s.TournamentName, s.TierName),
			AmountMinor:    s.AmountMinor,
			Currency:       "THB",
			Provider:       "MOCK",
			CreatedAt:      created,
			UpdatedAt:      created,
		}

		switch s.Status {
		case "PAID":
			paidAt := created.Add(5 * time.Minute)
			inv.ExpiresAt = created.Add(30 * time.Minute)
			inv.PaidAt = &paidAt
			inv.UpdatedAt = paidAt
			receipts = append(receipts, models.Receipt{
				ID:        s.ReceiptID,
				InvoiceID: s.InvoiceID,
				CreatedAt: paidAt,
			})
		case "UNPAID":
			inv.ExpiresAt = now.Add(2 * time.Hour)
		case "EXPIRED":
			expiredAt := created.Add(30 * time.Minute)
			inv.ExpiresAt = expiredAt
			inv.ExpiryAnnouncedAt = &expiredAt
			inv.UpdatedAt = expiredAt
		default: // CANCELLED
			cancelledAt := created.Add(10 * time.Minute)
			inv.ExpiresAt = created.Add(30 * time.Minute)
			inv.CancelledAt = &cancelledAt
			inv.UpdatedAt = cancelledAt
		}

		invoices = append(invoices, inv)
	}

	for i, r := range seedids.RegistrationInvoices() {
		created := now.Add(-time.Duration(96-i*2) * time.Hour)
		inv := models.Invoice{
			ID:             r.InvoiceID,
			PayerAccountID: r.PayerAccountID,
			Source:         models.SourceTeamRegistration,
			Reference:      fmt.Sprintf("%s:%s:%s", r.TournamentID, r.TeamID, r.AttemptID),
			IdempotencyKey: r.AttemptID.String(),
			Description:    fmt.Sprintf("Registration: %s (%s)", r.TournamentName, r.TeamID),
			AmountMinor:    r.AmountMinor,
			Currency:       "THB",
			Provider:       "MOCK",
			CreatedAt:      created,
			UpdatedAt:      created,
		}

		switch r.Status {
		case "PAID":
			paidAt := created.Add(4 * time.Minute)
			inv.ExpiresAt = created.Add(45 * time.Minute)
			inv.PaidAt = &paidAt
			inv.UpdatedAt = paidAt
			receipts = append(receipts, models.Receipt{
				ID:        r.ReceiptID,
				InvoiceID: r.InvoiceID,
				CreatedAt: paidAt,
			})
		case "REFUNDED":
			paidAt := created.Add(4 * time.Minute)
			refundedAt := created.Add(6 * time.Hour)
			inv.ExpiresAt = created.Add(45 * time.Minute)
			inv.PaidAt = &paidAt
			inv.RefundedAt = &refundedAt
			inv.RefundReason = r.RefundReason
			inv.UpdatedAt = refundedAt
			receipts = append(receipts, models.Receipt{
				ID:        r.ReceiptID,
				InvoiceID: r.InvoiceID,
				CreatedAt: paidAt,
			})
		default: // UNPAID
			inv.ExpiresAt = now.Add(45 * time.Minute)
		}

		invoices = append(invoices, inv)
	}

	for i := range invoices {
		// Preserve state transitions if a developer pays or refunds a seeded invoice locally.
		if err := db.Omit("Receipt").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&invoices[i]).Error; err != nil {
			return err
		}
	}

	for i := range receipts {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&receipts[i]).Error; err != nil {
			return err
		}
	}

	_ = uuid.Nil
	return nil
}
