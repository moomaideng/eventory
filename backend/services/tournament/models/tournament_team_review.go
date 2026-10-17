package models

import (
	"time"

	"github.com/google/uuid"
)

// TournamentTeamReview is the append-only audit trail for organizer decisions
// on team applications (US3-6). One row is written per successful decision;
// rows are never updated or deleted, so the history stays trustworthy even
// when an organizer later switches a decision.
//
// Statuses are stored as plain strings rather than a database enum, matching
// TournamentTeam.Status.
type TournamentTeamReview struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	Tournament       Tournament     `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	TournamentTeamID uuid.UUID      `gorm:"type:uuid;not null;index:idx_tournament_team_reviews_history,priority:1"`
	Team             TournamentTeam `gorm:"foreignKey:TournamentTeamID;constraint:OnDelete:CASCADE;"`

	// FromStatus is the status the team held before the decision.
	FromStatus TournamentTeamStatus `gorm:"type:varchar(16);not null"`
	// ToStatus is the decision's outcome (ACCEPTED or REJECTED).
	ToStatus TournamentTeamStatus `gorm:"type:varchar(16);not null"`
	// Reason is the organizer's optional justification, mainly for rejections.
	// Empty when none was given.
	Reason string `gorm:"type:text;not null;default:''"`

	// ReviewerAccountID is the account that made the decision. An Account rather
	// than an OrganizerProfile, so the trail still resolves for delegated staff.
	ReviewerAccountID uuid.UUID `gorm:"type:uuid;not null;index"`
	// ReviewerDisplayName is captured at decision time so history keeps showing
	// who acted then, even if the account later renames.
	ReviewerDisplayName string `gorm:"type:varchar(64);not null"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime;index:idx_tournament_team_reviews_history,priority:2,sort:desc"`
}
