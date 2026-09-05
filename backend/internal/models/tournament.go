package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	TournamentStatusRegistrationOpen   = "REGISTRATION_OPEN"
	TournamentStatusRegistrationClosed = "REGISTRATION_CLOSED"
	TournamentStatusOngoing            = "ONGOING"
	TournamentStatusCompleted          = "COMPLETED"
)

// Add any fields you want if you think this is not enough
type Tournament struct {
	ID                   uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizerID          uuid.UUID          `gorm:"type:uuid;not null;index"`
	Organizer            OrganizerProfile   `gorm:"foreignKey:OrganizerID;constraint:OnDelete:RESTRICT;"`
	Name                 string             `gorm:"type:varchar(160);not null"`
	Description          string             `gorm:"type:text;not null"`
	Game                 string             `gorm:"type:varchar(80);not null"`
	Location             string             `gorm:"type:varchar(160);not null"`
	StartAt              time.Time          `gorm:"not null;index:idx_tournaments_public_start,priority:3"`
	EndAt                time.Time          `gorm:"not null"`
	RegistrationDeadline time.Time          `gorm:"not null"`
	EntryFee             int64              `gorm:"not null;default:0;index"`
	Currency             string             `gorm:"type:char(3);not null;default:'THB'"`
	Capacity             int                `gorm:"not null"`
	RegisteredCount      int                `gorm:"not null;default:0"`
	Status               string             `gorm:"type:varchar(32);not null;index:idx_tournaments_public_start,priority:2"`
	Published            bool               `gorm:"not null;default:false;index:idx_tournaments_public_start,priority:1"`
	Teams                []TournamentTeam   `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	Funding              *TournamentFunding `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Add team & funding (Hoof pls review)
type TournamentTeam struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_tournament_team_name"`
	Name         string    `gorm:"type:varchar(120);not null;uniqueIndex:idx_tournament_team_name"`
	MemberCount  int       `gorm:"not null;default:1"`
	Seed         int       `gorm:"not null;default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TournamentFunding struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	GoalAmount     int64     `gorm:"not null;default:0"`
	RaisedAmount   int64     `gorm:"not null;default:0"`
	SupporterCount int       `gorm:"not null;default:0"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
