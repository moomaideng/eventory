package seeds

import (
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const mockPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

func SeedAccounts(db *gorm.DB) error {
	roles := []models.Role{
		{Code: "organizer", DisplayName: "Organizer"},
		{Code: "sponsor", DisplayName: "Sponsor"},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&roles).Error; err != nil {
		return err
	}

	accounts := []models.Account{
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Email:        "alice@example.com",
			PasswordHash: mockPasswordHash,
			PublicHandle: "alice_events",
			Status:       "active",
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Email:        "somchai@example.com",
			PasswordHash: mockPasswordHash,
			PublicHandle: "brightfuture",
			Status:       "active",
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Email:        "maya@example.com",
			PasswordHash: mockPasswordHash,
			PublicHandle: "maya_works",
			Status:       "active",
		},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&accounts).Error; err != nil {
		return err
	}

	accountRoles := []models.AccountRole{
		{AccountID: accounts[0].ID, RoleCode: "organizer"},
		{AccountID: accounts[1].ID, RoleCode: "sponsor"},
		{AccountID: accounts[2].ID, RoleCode: "organizer"},
		{AccountID: accounts[2].ID, RoleCode: "sponsor"},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&accountRoles).Error; err != nil {
		return err
	}

	organizerProfiles := []models.OrganizerProfile{
		{
			AccountID:   accounts[0].ID,
			DisplayName: "Alice Events",
			Bio:         "Community tournament organizer.",
			PublicEmail: "alice.events@example.com",
			PublicPhone: "+66 81 000 0001",
			WebsiteURL:  "https://example.com/alice-events",
		},
		{
			AccountID:   accounts[2].ID,
			DisplayName: "Maya Works",
			Bio:         "Organizer for local gaming communities.",
			PublicEmail: "maya.events@example.com",
			PublicPhone: "+66 81 000 0003",
			WebsiteURL:  "https://example.com/maya-works",
		},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&organizerProfiles).Error; err != nil {
		return err
	}

	sponsorProfiles := []models.SponsorProfile{
		{
			AccountID:        accounts[1].ID,
			OrganizationName: "Bright Future Co.",
			Description:      "Mock sponsor profile for local development.",
			PublicEmail:      "partnerships@example.com",
			PublicPhone:      "+66 81 000 0002",
			WebsiteURL:       "https://example.com/bright-future",
		},
		{
			AccountID:        accounts[2].ID,
			OrganizationName: "Maya Works Studio",
			Description:      "Mock organization with both organizer and sponsor roles.",
			PublicEmail:      "studio@example.com",
			PublicPhone:      "+66 81 000 0003",
			WebsiteURL:       "https://example.com/maya-works-studio",
		},
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sponsorProfiles).Error
}
