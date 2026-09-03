package seeds

import (
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedTournaments(db *gorm.DB) error {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	aliceOrganizerID := uuid.MustParse("11111111-0000-0000-0000-000000000001")
	mayaOrganizerID := uuid.MustParse("11111111-0000-0000-0000-000000000003")

	// Mock data, I only add 5 cuz i want it to be simple not too much
	tournaments := []models.Tournament{
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000001"),
			OrganizerID: aliceOrganizerID, Name: "Bangkok Valorant Open",
			Description: "A free community tournament for new and experienced Valorant teams.",
			Game:        "Valorant", Location: "Online", StartAt: now.AddDate(0, 0, 3).Add(12 * time.Hour),
			EndAt: now.AddDate(0, 0, 3).Add(20 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 2),
			EntryFee: 0, Currency: "THB", Capacity: 32, RegisteredCount: 18,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000002"),
			OrganizerID: mayaOrganizerID, Name: "Chula Mobile Legends Cup",
			Description: "A weekend Mobile Legends competition for university squads.",
			Game:        "Mobile Legends", Location: "Chulalongkorn University", StartAt: now.AddDate(0, 0, 10).Add(9 * time.Hour),
			EndAt: now.AddDate(0, 0, 11).Add(18 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 7),
			EntryFee: 250, Currency: "THB", Capacity: 24, RegisteredCount: 16,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000003"),
			OrganizerID: aliceOrganizerID, Name: "SEA Tekken Challenger",
			Description: "An offline fighting-game bracket for challengers across Southeast Asia.",
			Game:        "Tekken 8", Location: "Siam Paragon, Bangkok", StartAt: now.AddDate(0, 0, 21).Add(6 * time.Hour),
			EndAt: now.AddDate(0, 0, 21).Add(14 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 17),
			EntryFee: 500, Currency: "THB", Capacity: 64, RegisteredCount: 41,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000004"),
			OrganizerID: mayaOrganizerID, Name: "Weekend Chess Blitz",
			Description: "Fast-paced Swiss-system chess for players of every rating.",
			Game:        "Chess", Location: "Maya Works Studio", StartAt: now.AddDate(0, 0, 6).Add(7 * time.Hour),
			EndAt: now.AddDate(0, 0, 6).Add(12 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, -1),
			EntryFee: 100, Currency: "THB", Capacity: 40, RegisteredCount: 40,
			Status: models.TournamentStatusRegistrationClosed, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000005"),
			OrganizerID: aliceOrganizerID, Name: "Eventory Invitational",
			Description: "A premium multi-day invitational featuring Thailand's top esports teams.",
			Game:        "Counter-Strike 2", Location: "Queen Sirikit Convention Center", StartAt: now.AddDate(0, 0, 45).Add(4 * time.Hour),
			EndAt: now.AddDate(0, 0, 47).Add(14 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 35),
			EntryFee: 1200, Currency: "THB", Capacity: 16, RegisteredCount: 8,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&tournaments).Error
}
