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
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000010"),
			Email:    "player1@example.com",
			Username: "player_one",
			Status:   "ACTIVE",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000011"),
			Email:    "player2@example.com",
			Username: "player_two",
			Status:   "ACTIVE",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000012"),
			Email:    "player3@example.com",
			Username: "player_three",
			Status:   "ACTIVE",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000013"),
			Email:    "player4@example.com",
			Username: "player_four",
			Status:   "ACTIVE",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000014"),
			Email:    "player5@example.com",
			Username: "player_five",
			Status:   "ACTIVE",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000015"),
			Email:    "player6@example.com",
			Username: "player_six",
			Status:   "ACTIVE",
		},
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&accounts).Error; err != nil {
		return err
	}

	organizerProfiles := []models.OrganizerProfile{
		// Alice: Organizer Profile Only
		{
			ID:             uuid.MustParse("11111111-0000-0000-0000-000000000001"),
			AccountID:      accounts[0].ID,
			OrganizerName:  "Alice Events",
			OrganizerEmail: "alice.events@example.com",
		},
		// Maya: Both Organizer and Sponsor Profiles
		{
			ID:             uuid.MustParse("11111111-0000-0000-0000-000000000003"),
			AccountID:      accounts[2].ID,
			OrganizerName:  "Maya Works",
			OrganizerEmail: "maya.events@example.com",
		},
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&organizerProfiles).Error; err != nil {
		return err
	}

	sponsorProfiles := []models.SponsorProfile{
		// Somchai: Sponsor Profile Only
		{
			ID:           uuid.MustParse("22222222-0000-0000-0000-000000000002"),
			AccountID:    accounts[1].ID,
			SponsorName:  "Bright Future Co.",
			SponsorEmail: "partnerships@example.com",
		},
		// Maya: Both Organizer and Sponsor Profiles
		{
			ID:           uuid.MustParse("22222222-0000-0000-0000-000000000004"),
			AccountID:    accounts[2].ID,
			SponsorName:  "Maya Works Studio",
			SponsorEmail: "studio@example.com",
		},
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sponsorProfiles).Error
}
