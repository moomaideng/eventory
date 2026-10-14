package models

import (
	"time"

	"github.com/google/uuid"
)

// SoloRegistrationReview is the append-only audit trail for organizer
// decisions on solo applications (US3-6). It mirrors TournamentTeamReview:
// one row per successful decision, never updated or deleted.
//
// It is a separate table because solo statuses are longer than team ones
// (AWAITING_APPROVAL is 17 characters), so they do not fit the varchar(16)
// columns of tournament_team_reviews.
type SoloRegistrationReview struct {
	ID                 uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID       uuid.UUID        `gorm:"type:uuid;not null;index"`
	Tournament         Tournament       `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	SoloRegistrationID uuid.UUID        `gorm:"type:uuid;not null;index:idx_solo_registration_reviews_history,priority:1"`
	SoloRegistration   SoloRegistration `gorm:"foreignKey:SoloRegistrationID;constraint:OnDelete:CASCADE;"`

	// FromStatus is the status the registration held before the decision.
	FromStatus SoloRegistrationStatus `gorm:"type:varchar(24);not null"`
	// ToStatus is the decision's outcome (ACCEPTED or REJECTED).
	ToStatus SoloRegistrationStatus `gorm:"type:varchar(24);not null"`
	// Reason is the organizer's optional justification. Empty when none given.
	Reason string `gorm:"type:text;not null;default:''"`

	// ReviewerAccountID is the account that made the decision.
	ReviewerAccountID uuid.UUID `gorm:"type:uuid;not null;index"`
	// ReviewerDisplayName is captured at decision time so history keeps showing
	// who acted then, even if the account later renames.
	ReviewerDisplayName string `gorm:"type:varchar(64);not null"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime;index:idx_solo_registration_reviews_history,priority:2,sort:desc"`
}
