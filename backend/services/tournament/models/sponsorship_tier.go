package models

import (
	"time"

	"github.com/google/uuid"
)

// SponsorshipTier is a package (e.g. Gold, Silver) that sponsors can buy for a
// tournament (US5-4). Tiers contribute toward the tournament's single
// TournamentFunding goal; they do not carry a goal of their own.
//
// Sold slots and pledges are owned by sponsor checkout (US5-5), not stored here.
type SponsorshipTier struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:idx_sponsorship_tier_name"`
	Tournament   Tournament `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	Name         string     `gorm:"type:varchar(80);not null;uniqueIndex:idx_sponsorship_tier_name"`
	// Price is in whole currency units, matching Tournament.EntryFee and Currency.
	Price     int64    `gorm:"not null;check:chk_sponsorship_tiers_price,price > 0"`
	Slots     int      `gorm:"not null;check:chk_sponsorship_tiers_slots,slots > 0"`
	Perks     []string `gorm:"type:jsonb;serializer:json;not null"`
	SortOrder int      `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
