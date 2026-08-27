package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	Code        string `gorm:"type:varchar(32);primaryKey"`
	DisplayName string `gorm:"type:varchar(100);not null;uniqueIndex"`
	CreatedAt   time.Time
}

type Account struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"type:text;not null;uniqueIndex"`
	PasswordHash string    `gorm:"type:text;not null"`
	PublicHandle string    `gorm:"type:varchar(32);not null;uniqueIndex"`
	Status       string    `gorm:"type:varchar(16);not null;default:active;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (account *Account) BeforeCreate(_ *gorm.DB) error {
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	return nil
}

func (account *Account) BeforeSave(_ *gorm.DB) error {
	account.Email = strings.ToLower(strings.TrimSpace(account.Email))
	account.PublicHandle = strings.ToLower(strings.TrimSpace(account.PublicHandle))
	return nil
}

type AccountRole struct {
	AccountID uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoleCode  string    `gorm:"type:varchar(32);primaryKey"`
	CreatedAt time.Time

	Account Account `gorm:"foreignKey:AccountID;references:ID;constraint:OnDelete:CASCADE"`
	Role    Role    `gorm:"foreignKey:RoleCode;references:Code;constraint:OnDelete:RESTRICT"`
}

type OrganizerProfile struct {
	AccountID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	DisplayName string    `gorm:"type:varchar(150);not null"`
	Bio         string    `gorm:"type:text"`
	PublicEmail string    `gorm:"type:text"`
	PublicPhone string    `gorm:"type:varchar(32)"`
	WebsiteURL  string    `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Account Account `gorm:"foreignKey:AccountID;references:ID;constraint:OnDelete:CASCADE"`
}

type SponsorProfile struct {
	AccountID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	OrganizationName string    `gorm:"type:varchar(150);not null"`
	Description      string    `gorm:"type:text"`
	PublicEmail      string    `gorm:"type:text"`
	PublicPhone      string    `gorm:"type:varchar(32)"`
	WebsiteURL       string    `gorm:"type:text"`
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Account Account `gorm:"foreignKey:AccountID;references:ID;constraint:OnDelete:CASCADE"`
}
