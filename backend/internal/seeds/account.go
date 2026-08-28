package seeds

import (
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedAccounts(db *gorm.DB) error {
	accounts := []models.Account{
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Email:    "alice@example.com",
			Username: "alice_events",
			Status:   "ACTIVE",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Email:    "somchai@example.com",
			Username: "brightfuture",
			Status:   "ACTIVE",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Email:    "maya@example.com",
			Username: "maya_works",
			Status:   "ACTIVE",
		},
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&accounts).Error; err != nil {
		return err
	}

	profiles := []models.Profile{
		// Alice: Organizer Profile Only
		{
			ID:          uuid.MustParse("11111111-0000-0000-0000-000000000001"),
			AccountID:   accounts[0].ID,
			Type:        models.ProfileTypeOrganizer,
			DisplayName: "Alice Events",
			ContactInfo: "alice.events@example.com | +66 81 000 0001 | https://example.com/alice-events",
		},
		// Somchai: Sponsor Profile Only
		{
			ID:          uuid.MustParse("11111111-0000-0000-0000-000000000002"),
			AccountID:   accounts[1].ID,
			Type:        models.ProfileTypeSponsor,
			DisplayName: "Bright Future Co.",
			ContactInfo: "partnerships@example.com | +66 81 000 0002",
		},
		// Maya: Both Organizer and Sponsor Profiles
		{
			ID:          uuid.MustParse("11111111-0000-0000-0000-000000000003"),
			AccountID:   accounts[2].ID,
			Type:        models.ProfileTypeOrganizer,
			DisplayName: "Maya Works",
			ContactInfo: "maya.events@example.com | +66 81 000 0003 | Local gaming communities",
		},
		{
			ID:          uuid.MustParse("11111111-0000-0000-0000-000000000004"),
			AccountID:   accounts[2].ID,
			Type:        models.ProfileTypeSponsor,
			DisplayName: "Maya Works Studio",
			ContactInfo: "studio@example.com | https://example.com/maya-works-studio",
		},
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&profiles).Error
}
