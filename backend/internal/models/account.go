package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Account struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Username  string    `gorm:"type:varchar(32);not null;uniqueIndex"`
	Status    string    `gorm:"type:varchar(16);not null;default:'ACTIVE'"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// * -> optional/nullable
	OrganizerProfile *OrganizerProfile `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE;"`
	SponsorProfile   *SponsorProfile   `gorm:"foreignKey:AccountID;constraint:OnDelete:CASCADE;"`
}

func (account *Account) BeforeSave(_ *gorm.DB) error {
	account.Email = strings.ToLower(strings.TrimSpace(account.Email))
	account.Username = strings.TrimSpace(account.Username)
	return nil
}

type OrganizerProfile struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AccountID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	OrganizerName  string    `gorm:"type:varchar(150);not null"`
	OrganizerEmail string    `gorm:"type:varchar(255)"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SponsorProfile struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AccountID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	SponsorName  string    `gorm:"type:varchar(150);not null"`
	SponsorEmail string    `gorm:"type:varchar(255)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
