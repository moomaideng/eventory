package models

import (
	"time"

	"github.com/google/uuid"
)

type TournamentStatus string

const (
	TournamentStatusRegistrationOpen   TournamentStatus = "REGISTRATION_OPEN"
	TournamentStatusRegistrationClosed TournamentStatus = "REGISTRATION_CLOSED"
	TournamentStatusOngoing            TournamentStatus = "ONGOING"
	TournamentStatusCompleted          TournamentStatus = "COMPLETED"
)

type TournamentRegistrationMode string

const (
	TournamentRegistrationModeSolo TournamentRegistrationMode = "SOLO"
	TournamentRegistrationModeTeam TournamentRegistrationMode = "TEAM"
	TournamentRegistrationModeBoth TournamentRegistrationMode = "BOTH"
)

type TournamentTeamStatus string

const (
	TournamentTeamStatusForming  TournamentTeamStatus = "FORMING"
	TournamentTeamStatusLocked   TournamentTeamStatus = "LOCKED"
	TournamentTeamStatusAccepted TournamentTeamStatus = "ACCEPTED"
	TournamentTeamStatusRejected TournamentTeamStatus = "REJECTED"
)

type TournamentTeamMemberRole string

const (
	TournamentTeamMemberRoleCaptain TournamentTeamMemberRole = "CAPTAIN"
	TournamentTeamMemberRoleMember  TournamentTeamMemberRole = "MEMBER"
)

type Tournament struct {
	ID                   uuid.UUID                  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizerID          uuid.UUID                  `gorm:"type:uuid;not null;index"`
	Organizer            OrganizerProfile           `gorm:"foreignKey:OrganizerID;constraint:OnDelete:RESTRICT;"`
	Name                 string                     `gorm:"type:varchar(160);not null"`
	Description          string                     `gorm:"type:text;not null"`
	Game                 string                     `gorm:"type:varchar(80);not null"`
	Location             string                     `gorm:"type:varchar(160);not null"`
	StartsAt             time.Time                  `gorm:"column:start_at;not null;index:idx_tournaments_public_start,priority:3"`
	EndsAt               time.Time                  `gorm:"column:end_at;not null"`
	RegistrationDeadline time.Time                  `gorm:"not null"`
	EntryFee             int64                      `gorm:"not null;default:0;index"`
	Currency             string                     `gorm:"type:char(3);not null;default:'THB'"`
	RegistrationMode     TournamentRegistrationMode `gorm:"type:varchar(16);not null;default:'SOLO'"`
	MinTeamSize          int                        `gorm:"not null;default:1"`
	MaxTeamSize          int                        `gorm:"not null;default:1"`
	Capacity             int                        `gorm:"not null"`
	Status               TournamentStatus           `gorm:"type:varchar(32);not null;index:idx_tournaments_public_start,priority:2"`
	Published            bool                       `gorm:"not null;default:false;index:idx_tournaments_public_start,priority:1"`
	Teams                []TournamentTeam           `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	Funding              *TournamentFunding         `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// TournamentTeam is the single persisted record for a team throughout
// formation, locking, organizer review, and confirmed registration.
type TournamentTeam struct {
	ID           uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID uuid.UUID              `gorm:"type:uuid;not null;index;uniqueIndex:idx_tournament_team_name;index:idx_tournament_teams_status,priority:1"`
	Tournament   Tournament             `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	Name         string                 `gorm:"type:varchar(120);not null;uniqueIndex:idx_tournament_team_name"`
	InviteCode   string                 `gorm:"type:varchar(32);uniqueIndex"`
	Status       TournamentTeamStatus   `gorm:"type:varchar(16);not null;default:'FORMING';index:idx_tournament_teams_status,priority:2"`
	Members      []TournamentTeamMember `gorm:"foreignKey:TournamentTeamID;constraint:OnDelete:CASCADE;"`
	CreatedAt    time.Time
	LockedAt     *time.Time
}

type TournamentTeamMember struct {
	ID               uuid.UUID                `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentTeamID uuid.UUID                `gorm:"type:uuid;not null;index;uniqueIndex:idx_tournament_team_member"`
	AccountID        uuid.UUID                `gorm:"type:uuid;not null;index;uniqueIndex:idx_tournament_team_member"`
	Role             TournamentTeamMemberRole `gorm:"type:varchar(16);not null"`
	Account          Account                  `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE;"`
	JoinedAt         time.Time                `gorm:"not null;autoCreateTime"`
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
