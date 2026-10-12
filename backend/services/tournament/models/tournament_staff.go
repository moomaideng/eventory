package models

import (
	"time"

	"github.com/google/uuid"
)

type TournamentStaffRole string

const (
	TournamentStaffRoleOwner   TournamentStaffRole = "OWNER"
	TournamentStaffRoleAdmin   TournamentStaffRole = "ADMIN"
	TournamentStaffRoleReferee TournamentStaffRole = "REFEREE"
)

// IsAssignableRole checks whether a role can be assigned or invited by staff.
func IsAssignableRole(role TournamentStaffRole) bool {
	return role == TournamentStaffRoleAdmin || role == TournamentStaffRoleReferee
}

// TournamentStaff represents an assigned staff member for a tournament.
// The composite primary key is (tournament_id, account_id).
type TournamentStaff struct {
	TournamentID uuid.UUID           `gorm:"type:uuid;primaryKey;index"`
	Tournament   Tournament          `gorm:"foreignKey:TournamentID;constraint:OnDelete:CASCADE;"`
	AccountID    uuid.UUID           `gorm:"type:uuid;primaryKey;index"`
	Role         TournamentStaffRole `gorm:"type:varchar(32);not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (TournamentStaff) TableName() string {
	return "tournament_staff"
}

func (s *TournamentStaff) IsOwner() bool {
	return s != nil && s.Role == TournamentStaffRoleOwner
}

func (s *TournamentStaff) IsAdmin() bool {
	return s != nil && s.Role == TournamentStaffRoleAdmin
}

func (s *TournamentStaff) IsReferee() bool {
	return s != nil && s.Role == TournamentStaffRoleReferee
}
