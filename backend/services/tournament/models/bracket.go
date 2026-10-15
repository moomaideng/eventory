package models

import (
	"time"

	"github.com/google/uuid"
)

// Match represents a stage node in the bracket graph.
type Match struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID   uuid.UUID    `gorm:"type:uuid;not null;index"`
	Tournament     Tournament   `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	Name           string       `gorm:"type:varchar(160);not null"`
	Venue          string       `gorm:"type:varchar(160);not null"`
	StageLevel     int          `gorm:"not null;index"`
	AdvancingCount int          `gorm:"not null"`
	TargetMatchID  *uuid.UUID   `gorm:"type:uuid;index"`
	TargetMatch    *Match       `gorm:"foreignKey:TargetMatchID;constraint:OnDelete:SET NULL;"`
	IsCompleted    bool         `gorm:"not null;default:false"`
	Entries        []MatchEntry `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE;"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Match) TableName() string {
	return "matches"
}

// MatchEntry represents a team assigned to a specific match.
type MatchEntry struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MatchID   uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_match_entry_team,priority:1"`
	Match     Match          `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE;"`
	TeamID    uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_match_entry_team,priority:2"`
	Team      TournamentTeam `gorm:"foreignKey:TeamID;constraint:OnDelete:CASCADE;"`
	FinalRank *int           `gorm:"type:int"`
	Advances  bool           `gorm:"not null;default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (MatchEntry) TableName() string {
	return "match_entries"
}
