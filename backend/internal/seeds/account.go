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
			ID:          uuid.MustParse("99999999-0000-4000-8000-000000000001"),
			Email:       "dev@eventory.gg",
			Handle:      "dev_alex",
			DisplayName: "Alex (Dev)",
			Status:      "ACTIVE",
		},
		{
			ID:          uuid.MustParse("00000000-0000-4000-8000-000000000001"),
			Email:       "alice@example.com",
			Handle:      "alice_events",
			DisplayName: "Alice Events",
			Status:      "ACTIVE",
		},
		{
			ID:          uuid.MustParse("00000000-0000-4000-8000-000000000002"),
			Email:       "somchai@example.com",
			Handle:      "brightfuture",
			DisplayName: "Bright Future",
			Status:      "ACTIVE",
		},
		{
			ID:          uuid.MustParse("00000000-0000-4000-8000-000000000003"),
			Email:       "maya@example.com",
			Handle:      "maya_works",
			DisplayName: "Maya Works",
			Status:      "ACTIVE",
		},
		{
			ID:          uuid.MustParse("00000000-0000-4000-8000-000000000010"),
			Email:       "player1@example.com",
			Handle:      "player_one",
			DisplayName: "Player One",
			Status:      "ACTIVE",
		},
		{
			ID:          uuid.MustParse("00000000-0000-4000-8000-000000000011"),
			Email:       "player2@example.com",
			Handle:      "player_two",
			DisplayName: "Player Two",
			Status:      "ACTIVE",
		},
		{
			ID:          uuid.MustParse("00000000-0000-4000-8000-000000000012"),
			Email:       "player3@example.com",
			Handle:      "player_three",
			DisplayName: "Player Three",
			Status:      "ACTIVE",
		},
		{
			ID:          uuid.MustParse("00000000-0000-4000-8000-000000000013"),
			Email:       "player4@example.com",
			Handle:      "player_four",
			DisplayName: "Player Four",
			Status:      "ACTIVE",
		},
		{
			ID:          uuid.MustParse("00000000-0000-4000-8000-000000000014"),
			Email:       "player5@example.com",
			Handle:      "player_five",
			DisplayName: "Player Five",
			Status:      "ACTIVE",
		},
		{
			ID:          uuid.MustParse("00000000-0000-4000-8000-000000000015"),
			Email:       "player6@example.com",
			Handle:      "player_six",
			DisplayName: "Player Six",
			Status:      "ACTIVE",
		},
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&accounts).Error; err != nil {
		return err
	}

	organizerProfiles := []models.OrganizerProfile{
		// Alex (Dev): Full Organizer access in dev mode
		{
			ID:             uuid.MustParse("99999999-0000-4000-8000-000000000002"),
			AccountID:      accounts[0].ID,
			OrganizerName:  "Alex Events (Dev)",
			OrganizerEmail: "dev@eventory.gg",
		},
		// Alice: Organizer Profile Only
		{
			ID:             uuid.MustParse("11111111-0000-4000-8000-000000000001"),
			AccountID:      accounts[1].ID,
			OrganizerName:  "Alice Events",
			OrganizerEmail: "alice.events@example.com",
		},
		// Maya: Both Organizer and Sponsor Profiles
		{
			ID:             uuid.MustParse("11111111-0000-4000-8000-000000000003"),
			AccountID:      accounts[3].ID,
			OrganizerName:  "Maya Works",
			OrganizerEmail: "maya.events@example.com",
		},
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&organizerProfiles).Error; err != nil {
		return err
	}

	sponsorProfiles := []models.SponsorProfile{
		// Alex (Dev): Sponsor access in dev mode
		{
			ID:           uuid.MustParse("99999999-0000-4000-8000-000000000003"),
			AccountID:    accounts[0].ID,
			SponsorName:  "Alex Ventures (Dev)",
			SponsorEmail: "dev@eventory.gg",
		},
		// Somchai: Sponsor Profile Only
		{
			ID:           uuid.MustParse("22222222-0000-4000-8000-000000000002"),
			AccountID:    accounts[2].ID,
			SponsorName:  "Bright Future Co.",
			SponsorEmail: "partnerships@example.com",
		},
		// Maya: Both Organizer and Sponsor Profiles
		{
			ID:           uuid.MustParse("22222222-0000-4000-8000-000000000004"),
			AccountID:    accounts[3].ID,
			SponsorName:  "Maya Works Studio",
			SponsorEmail: "studio@example.com",
		},
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&sponsorProfiles).Error
}
