package models

import (
	"time"

	"github.com/google/uuid"
)

// TournamentStatusChange is the append-only audit trail for manual status
// overrides (US3-3). One row is written per successful override; rows are
// never updated or deleted, so the history stays trustworthy.
//
// Statuses are stored as plain strings rather than a database enum, matching
// Tournament.Status. Adding a new TournamentStatus therefore needs no
// migration here, and historical rows keep the value that was written at the
// time even if that status is later retired.
type TournamentStatusChange struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID uuid.UUID  `gorm:"type:uuid;not null;index:idx_tournament_status_changes_history,priority:1"`
	Tournament   Tournament `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`

	// FromStatus is the status the tournament held before the override.
	FromStatus TournamentStatus `gorm:"type:varchar(32);not null"`
	// ToStatus is the status the organizer moved it to.
	ToStatus TournamentStatus `gorm:"type:varchar(32);not null"`
	// Reason is the organizer's justification. Required for overrides that the
	// use case marks as reason-mandatory (e.g. cancelling a tournament).
	Reason string `gorm:"type:text;not null"`

	// ActorAccountID is the account that performed the override. Deliberately an
	// Account rather than an OrganizerProfile so the trail still resolves once
	// delegated staff can override on an organizer's behalf (US3-7).
	ActorAccountID uuid.UUID `gorm:"type:uuid;not null;index"`
	Actor          Account   `gorm:"foreignKey:ActorAccountID;constraint:OnDelete:RESTRICT;"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime;index:idx_tournament_status_changes_history,priority:2,sort:desc"`
}
