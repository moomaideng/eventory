package models

import (
	"time"

	"github.com/google/uuid"
)

type TournamentStatus string

const (
	TournamentStatusDraft        TournamentStatus = "DRAFT"
	TournamentStatusCrowdfunding TournamentStatus = "CROWDFUNDING"
	TournamentStatusRegistration TournamentStatus = "REGISTRATION"
	TournamentStatusOngoing      TournamentStatus = "ONGOING"
	TournamentStatusCompleted    TournamentStatus = "COMPLETED"
)

type TournamentRegistrationMode string

const (
	TournamentRegistrationModeSolo TournamentRegistrationMode = "SOLO"
	TournamentRegistrationModeTeam TournamentRegistrationMode = "TEAM"
)

type TournamentTeamStatus string

const (
	TournamentTeamStatusForming        TournamentTeamStatus = "FORMING"
	TournamentTeamStatusPaymentPending TournamentTeamStatus = "PAYMENT_PENDING"
	TournamentTeamStatusLocked         TournamentTeamStatus = "LOCKED"
	TournamentTeamStatusAccepted       TournamentTeamStatus = "ACCEPTED"
	TournamentTeamStatusRejected       TournamentTeamStatus = "REJECTED"
)

type TournamentTeamMemberRole string

const (
	TournamentTeamMemberRoleCaptain TournamentTeamMemberRole = "CAPTAIN"
	TournamentTeamMemberRoleMember  TournamentTeamMemberRole = "MEMBER"
)

type Tournament struct {
	ID                   uuid.UUID                  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name                 string                     `gorm:"type:varchar(160);not null"`
	Description          string                     `gorm:"type:text;not null"`
	Detail               string                     `gorm:"type:text;not null;default:''"`
	Game                 string                     `gorm:"type:varchar(80);not null"`
	Location             string                     `gorm:"type:varchar(160);not null"`
	StartsAt             time.Time                  `gorm:"column:start_at;not null;index:idx_tournaments_public_start,priority:2"`
	EndsAt               time.Time                  `gorm:"column:end_at;not null"`
	RegistrationDeadline time.Time                  `gorm:"not null"`
	EntryFeeMinor        int64                      `gorm:"not null;default:0;index"`
	Currency             string                     `gorm:"type:char(3);not null;default:'THB'"`
	RegistrationMode     TournamentRegistrationMode `gorm:"type:varchar(16);not null;default:'SOLO'"`
	MinTeamSize          int                        `gorm:"not null;default:1"`
	MaxTeamSize          int                        `gorm:"not null;default:1"`
	Capacity             int                        `gorm:"not null"`
	Status               TournamentStatus           `gorm:"type:varchar(32);not null;index:idx_tournaments_public_start,priority:1"`
	Staff                []TournamentStaff          `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	Teams                []TournamentTeam           `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	Matches              []Match                    `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	Crowdfunding         CrowdfundingInfo           `gorm:"embedded;embeddedPrefix:crowdfunding_"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// IsRegistrationClosed reports whether the registration deadline has passed.
func (t *Tournament) IsRegistrationClosed(now time.Time) bool {
	return !now.Before(t.RegistrationDeadline)
}

// IsSolo reports whether this tournament is intended for solo participants.
func (t *Tournament) IsSolo() bool {
	return t.MinTeamSize == 1 && t.MaxTeamSize == 1
}

// StaffMember looks up a staff member by account ID in the tournament's preloaded staff list.
func (t *Tournament) StaffMember(accountID uuid.UUID) (*TournamentStaff, bool) {
	for i := range t.Staff {
		if t.Staff[i].AccountID == accountID {
			return &t.Staff[i], true
		}
	}
	return nil, false
}

// OwnerID returns the account ID of the tournament owner, or uuid.Nil if not found.
func (t *Tournament) OwnerID() uuid.UUID {
	for i := range t.Staff {
		if t.Staff[i].IsOwner() {
			return t.Staff[i].AccountID
		}
	}
	return uuid.Nil
}

// ValidateTeamID reports whether id is an 8-character alphanumeric team identifier.
func ValidateTeamID(id string) bool {
	if len(id) != 8 {
		return false
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}

// TournamentTeam is the single persisted record for a team throughout
// formation, locking, organizer review, and confirmed registration.
type TournamentTeam struct {
	TournamentID       uuid.UUID              `gorm:"type:uuid;primaryKey;not null;uniqueIndex:idx_tournament_team_active_name,where:status <> 'REJECTED';index:idx_tournament_teams_status,priority:1"`
	ID                 string                 `gorm:"type:varchar(8);primaryKey;not null"`
	Tournament         Tournament             `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	Name               string                 `gorm:"type:varchar(120);not null;uniqueIndex:idx_tournament_team_active_name,where:status <> 'REJECTED'"`
	InviteCode         *string                `gorm:"type:varchar(6);uniqueIndex"`
	Status             TournamentTeamStatus   `gorm:"type:varchar(16);not null;default:'FORMING';index:idx_tournament_teams_status,priority:2"`
	Members            []TournamentTeamMember `gorm:"foreignKey:TournamentID,TournamentTeamID;references:TournamentID,ID;constraint:OnDelete:CASCADE;"`
	CreatedAt          time.Time
	LockedAt           *time.Time
	PaymentAttemptID   *uuid.UUID `gorm:"type:uuid"`
	PaymentID          *uuid.UUID `gorm:"type:uuid"`
	RefundedAt         *time.Time
	PaymentPayerID     *uuid.UUID `gorm:"type:uuid"`
	PaymentAmountMinor int64      `gorm:"not null;default:0"`
	PaymentExpiresAt   *time.Time
}

type TournamentTeamMember struct {
	ID               uuid.UUID                `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TournamentID     uuid.UUID                `gorm:"type:uuid;not null;index;uniqueIndex:idx_tournament_team_member"`
	TournamentTeamID string                   `gorm:"type:varchar(8);not null;index;uniqueIndex:idx_tournament_team_member"`
	AccountID        uuid.UUID                `gorm:"type:uuid;not null;index;uniqueIndex:idx_tournament_team_member"`
	Role             TournamentTeamMemberRole `gorm:"type:varchar(16);not null"`
	JoinedAt         time.Time                `gorm:"not null;autoCreateTime"`
}

type CrowdfundingInfo struct {
	GoalAmountMinor int64               `gorm:"not null;default:0"`
	Deadline        *time.Time
	Tiers           []SponsorshipTier   `gorm:"foreignKey:TournamentID;references:ID;constraint:OnDelete:CASCADE;"`
	Pledges         []SponsorshipPledge `gorm:"foreignKey:TournamentID;references:ID;constraint:OnDelete:CASCADE;"`
}
