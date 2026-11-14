package models

import "github.com/google/uuid"

type PledgeStatus string

const (
	PledgePending   PledgeStatus = "PENDING"
	PledgeConfirmed PledgeStatus = "CONFIRMED"
	PledgeCancelled PledgeStatus = "CANCELLED"
	PledgeExpired   PledgeStatus = "EXPIRED"
)

// SponsorshipTier is an organizer-defined sponsorship package on a tournament.
type SponsorshipTier struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TournamentID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_sponsorship_tiers_tournament_name,priority:1"`
	Name         string    `gorm:"size:80;not null;uniqueIndex:idx_sponsorship_tiers_tournament_name,priority:2"`
	PriceMinor   int64     `gorm:"not null"`
	Slots        int       `gorm:"not null"`
	Perks        []string  `gorm:"serializer:json;type:jsonb;not null"`
}

// SponsorshipPledge represents a sponsor checkout for a tournament tier.
// Only CONFIRMED pledges count toward crowdfunding totals and public sponsor listings.
type SponsorshipPledge struct {
	ID             uuid.UUID        `gorm:"type:uuid;primaryKey"`
	TournamentID   uuid.UUID        `gorm:"type:uuid;not null;index;uniqueIndex:idx_active_sponsorship,priority:1,where:(status = 'PENDING' OR status = 'CONFIRMED')"`
	TierID         uuid.UUID        `gorm:"type:uuid;not null;index"`
	Tier           *SponsorshipTier `gorm:"foreignKey:TierID;references:ID"`
	PayerAccountID uuid.UUID        `gorm:"type:uuid;not null;index;uniqueIndex:idx_active_sponsorship,priority:2,where:(status = 'PENDING' OR status = 'CONFIRMED')"`
	AmountMinor    int64            `gorm:"not null"`
	Currency       string           `gorm:"size:3;not null"`
	Status         PledgeStatus     `gorm:"size:16;not null;index"`
	PaymentID      *uuid.UUID       `gorm:"type:uuid;index"`
}
