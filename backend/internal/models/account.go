package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProfileType string

const (
	ProfileTypeOrganizer ProfileType = "ORGANIZER"
	ProfileTypeSponsor   ProfileType = "SPONSOR"
)

type Account struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Username  string    `gorm:"type:varchar(32);not null;uniqueIndex"`
	Status    string    `gorm:"type:varchar(16);not null;default:'ACTIVE'"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Profiles []Profile `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE;"`
}

func (account *Account) BeforeSave(_ *gorm.DB) error {
	account.Email = strings.ToLower(strings.TrimSpace(account.Email))
	account.Username = strings.ToLower(strings.TrimSpace(account.Username))
	return nil
}

type Profile struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AccountID   uuid.UUID   `gorm:"type:uuid;not null;uniqueIndex:idx_account_type"`
	Type        ProfileType `gorm:"type:varchar(16);not null;check:type IN ('ORGANIZER', 'SPONSOR');uniqueIndex:idx_account_type"`
	DisplayName string      `gorm:"type:varchar(150);not null"`
	ContactInfo string      `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
