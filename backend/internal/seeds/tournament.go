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
	lockedAt := now.Add(-24 * time.Hour)
	aliceOrganizerID := uuid.MustParse("11111111-0000-0000-0000-000000000001")
	mayaOrganizerID := uuid.MustParse("11111111-0000-0000-0000-000000000003")

	// Mock data, I only add 5 cuz i want it to be simple not too much
	tournaments := []models.Tournament{
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000001"),
			OrganizerID: aliceOrganizerID, Name: "Bangkok Valorant Open",
			Description: "A free community tournament for new and experienced Valorant teams.",
			Game:        "Valorant", Location: "Online", StartsAt: now.AddDate(0, 0, 3).Add(12 * time.Hour),
			EndsAt: now.AddDate(0, 0, 3).Add(20 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 2),
			EntryFee: 0, Currency: "THB", Capacity: 32,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000002"),
			OrganizerID: mayaOrganizerID, Name: "Chula Mobile Legends Cup",
			Description: "A weekend Mobile Legends competition for university squads.",
			Game:        "Mobile Legends", Location: "Chulalongkorn University", StartsAt: now.AddDate(0, 0, 10).Add(9 * time.Hour),
			EndsAt: now.AddDate(0, 0, 11).Add(18 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 7),
			EntryFee: 250, Currency: "THB", Capacity: 24,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000003"),
			OrganizerID: aliceOrganizerID, Name: "SEA Tekken Challenger",
			Description: "An offline fighting-game bracket for challengers across Southeast Asia.",
			Game:        "Tekken 8", Location: "Siam Paragon, Bangkok", StartsAt: now.AddDate(0, 0, 21).Add(6 * time.Hour),
			EndsAt: now.AddDate(0, 0, 21).Add(14 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 17),
			EntryFee: 500, Currency: "THB", Capacity: 64,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000004"),
			OrganizerID: mayaOrganizerID, Name: "Weekend Chess Blitz",
			Description: "Fast-paced Swiss-system chess for players of every rating.",
			Game:        "Chess", Location: "Maya Works Studio", StartsAt: now.AddDate(0, 0, 6).Add(7 * time.Hour),
			EndsAt: now.AddDate(0, 0, 6).Add(12 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, -1),
			EntryFee: 100, Currency: "THB", Capacity: 40,
			Status: models.TournamentStatusRegistrationClosed, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-0000-0000-000000000005"),
			OrganizerID: aliceOrganizerID, Name: "Eventory Invitational",
			Description: "A premium multi-day invitational featuring Thailand's top esports teams.",
			Game:        "Counter-Strike 2", Location: "Queen Sirikit Convention Center", StartsAt: now.AddDate(0, 0, 45).Add(4 * time.Hour),
			EndsAt: now.AddDate(0, 0, 47).Add(14 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 35),
			EntryFee: 1200, Currency: "THB", Capacity: 16,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&tournaments).Error; err != nil {
		return err
	}

	teams := []models.TournamentTeam{
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000001"), TournamentID: tournaments[0].ID, Name: "Neon Tigers", InviteCode: "VAL-NEON-001", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000002"), TournamentID: tournaments[0].ID, Name: "Bangkok Byte", InviteCode: "VAL-BYTE-002", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000003"), TournamentID: tournaments[0].ID, Name: "Siam Sentinels", InviteCode: "VAL-SIAM-003", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000004"), TournamentID: tournaments[1].ID, Name: "Chula Phoenix", InviteCode: "MLBB-CHULA-004", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000005"), TournamentID: tournaments[1].ID, Name: "River Guardians", InviteCode: "MLBB-RIVER-005", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000006"), TournamentID: tournaments[1].ID, Name: "Lotus Legends", InviteCode: "MLBB-LOTUS-006", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000007"), TournamentID: tournaments[2].ID, Name: "Iron Fist BKK", InviteCode: "TEK-IRON-007", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000008"), TournamentID: tournaments[2].ID, Name: "Manila Punishers", InviteCode: "TEK-MANILA-008", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000009"), TournamentID: tournaments[2].ID, Name: "Jakarta Kings", InviteCode: "TEK-JAKARTA-009", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000010"), TournamentID: tournaments[3].ID, Name: "Knight Owls", InviteCode: "CHESS-KNIGHT-010", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000011"), TournamentID: tournaments[3].ID, Name: "Queen's Gambit Club", InviteCode: "CHESS-QUEEN-011", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000012"), TournamentID: tournaments[4].ID, Name: "Eventory Elite", InviteCode: "CS2-ELITE-012", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000013"), TournamentID: tournaments[4].ID, Name: "Northern Stars", InviteCode: "CS2-NORTH-013", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-0000-0000-000000000014"), TournamentID: tournaments[4].ID, Name: "Crimson Circuit", InviteCode: "CS2-CRIMSON-014", Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
	}
	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, UpdateAll: true,
	}).Create(&teams).Error; err != nil {
		return err
	}

	playerIDs := []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000000010"),
		uuid.MustParse("00000000-0000-0000-0000-000000000011"),
		uuid.MustParse("00000000-0000-0000-0000-000000000012"),
		uuid.MustParse("00000000-0000-0000-0000-000000000013"),
		uuid.MustParse("00000000-0000-0000-0000-000000000014"),
		uuid.MustParse("00000000-0000-0000-0000-000000000015"),
	}
	members := make([]models.TournamentTeamMember, 0, len(teams)*2)
	for index, team := range teams {
		pairStart := (index % 3) * 2
		for memberOffset := 0; memberOffset < 2; memberOffset++ {
			accountID := playerIDs[pairStart+memberOffset]
			role := models.TournamentTeamMemberRoleMember
			if memberOffset == 0 {
				role = models.TournamentTeamMemberRoleCaptain
			}
			members = append(members, models.TournamentTeamMember{
				ID:               uuid.NewSHA1(uuid.NameSpaceOID, []byte(team.ID.String()+":"+accountID.String())),
				TournamentTeamID: team.ID,
				AccountID:        accountID,
				Role:             role,
			})
		}
	}
	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, UpdateAll: true,
	}).Create(&members).Error; err != nil {
		return err
	}

	funding := []models.TournamentFunding{
		{ID: uuid.MustParse("55555555-0000-0000-0000-000000000001"), TournamentID: tournaments[0].ID, GoalAmount: 50000, RaisedAmount: 32500, SupporterCount: 24},
		{ID: uuid.MustParse("55555555-0000-0000-0000-000000000002"), TournamentID: tournaments[1].ID, GoalAmount: 75000, RaisedAmount: 51000, SupporterCount: 31},
		{ID: uuid.MustParse("55555555-0000-0000-0000-000000000003"), TournamentID: tournaments[2].ID, GoalAmount: 120000, RaisedAmount: 84000, SupporterCount: 47},
		{ID: uuid.MustParse("55555555-0000-0000-0000-000000000004"), TournamentID: tournaments[3].ID, GoalAmount: 25000, RaisedAmount: 25000, SupporterCount: 19},
		{ID: uuid.MustParse("55555555-0000-0000-0000-000000000005"), TournamentID: tournaments[4].ID, GoalAmount: 500000, RaisedAmount: 287500, SupporterCount: 83},
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, UpdateAll: true,
	}).Create(&funding).Error
}
