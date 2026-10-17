package models

import (
	"time"

	"github.com/google/uuid"
)

type PledgeStatus string

const (
	PledgePending   PledgeStatus = "PENDING"
	PledgeConfirmed PledgeStatus = "CONFIRMED"
	PledgeCancelled PledgeStatus = "CANCELLED"
	PledgeExpired   PledgeStatus = "EXPIRED"
)

// Terms and branding are snapshots of this purchase, independent of later profile edits.
type SponsorshipPledge struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey"`
	TournamentID   uuid.UUID       `gorm:"type:uuid;not null;index"`
	TierID         uuid.UUID       `gorm:"type:uuid;not null;index"`
	Tier           SponsorshipTier `gorm:"foreignKey:TierID;constraint:OnDelete:RESTRICT;"`
	PayerAccountID uuid.UUID       `gorm:"type:uuid;not null;index"`
	SponsorName    string          `gorm:"size:150;not null"`
	WebsiteURL     string          `gorm:"size:2048;not null"`
	TierName       string          `gorm:"size:80;not null"`
	AmountMinor    int64           `gorm:"not null;check:pledge_amount,amount_minor > 0"`
	Currency       string          `gorm:"size:3;not null"`
	Status         PledgeStatus    `gorm:"size:16;not null;index"`
	PaymentID      *uuid.UUID      `gorm:"type:uuid;uniqueIndex"`
	ExpiresAt      time.Time       `gorm:"not null"`
	CreatedAt      time.Time
	ConfirmedAt    *time.Time
}

// Keep binary data out of pledge/list queries. The database reset removes it too.
type SponsorshipLogo struct {
	PledgeID    uuid.UUID         `gorm:"type:uuid;primaryKey"`
	Pledge      SponsorshipPledge `gorm:"foreignKey:PledgeID;constraint:OnDelete:CASCADE;"`
	ContentType string            `gorm:"size:32;not null"`
	Data        []byte            `gorm:"type:bytea;not null"`
}
