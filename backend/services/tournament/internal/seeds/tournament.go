package seeds

import (
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedTournaments(db *gorm.DB) error {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	lockedAt := now.Add(-24 * time.Hour)
	aliceAccountID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	devAccountID := uuid.MustParse("99999999-0000-4000-8000-000000000001")
	mayaAccountID := uuid.MustParse("00000000-0000-4000-8000-000000000003")

	// Alice keeps the original sample tournaments while Alex has two dedicated
	// dev-owned tournaments for exercising the organizer dashboard.
	tournaments := []models.Tournament{
		{
			ID:          uuid.MustParse("33333333-0000-4000-8000-000000000001"),
			Name:        "Bangkok Valorant Open",
			Description: "A free community tournament for new and experienced Valorant teams.",
			Game:        "Valorant", Location: "Online", StartsAt: now.AddDate(0, 0, 3).Add(12 * time.Hour),
			EndsAt: now.AddDate(0, 0, 3).Add(20 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 2),
			RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 5, MaxTeamSize: 5,
			EntryFee: 0, Currency: "THB", Capacity: 32,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-4000-8000-000000000002"),
			Name:        "Chula Mobile Legends Cup",
			Description: "A weekend Mobile Legends competition for university squads.",
			Game:        "Mobile Legends", Location: "Chulalongkorn University", StartsAt: now.AddDate(0, 0, 10).Add(9 * time.Hour),
			EndsAt: now.AddDate(0, 0, 11).Add(18 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 7),
			RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 5, MaxTeamSize: 5,
			EntryFee: 250, Currency: "THB", Capacity: 24,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-4000-8000-000000000003"),
			Name:        "SEA Tekken Challenger",
			Description: "An offline fighting-game bracket for challengers across Southeast Asia.",
			Game:        "Tekken 8", Location: "Siam Paragon, Bangkok", StartsAt: now.AddDate(0, 0, 21).Add(6 * time.Hour),
			EndsAt: now.AddDate(0, 0, 21).Add(14 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 17),
			RegistrationMode: models.TournamentRegistrationModeSolo, MinTeamSize: 1, MaxTeamSize: 1,
			EntryFee: 500, Currency: "THB", Capacity: 64,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-4000-8000-000000000004"),
			Name:        "Weekend Chess Blitz",
			Description: "Fast-paced Swiss-system chess for players of every rating.",
			Game:        "Chess", Location: "Maya Works Studio", StartsAt: now.AddDate(0, 0, 6).Add(7 * time.Hour),
			EndsAt: now.AddDate(0, 0, 6).Add(12 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, -1),
			RegistrationMode: models.TournamentRegistrationModeSolo, MinTeamSize: 1, MaxTeamSize: 1,
			EntryFee: 100, Currency: "THB", Capacity: 40,
			Status: models.TournamentStatusRegistrationClosed, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-4000-8000-000000000005"),
			Name:        "Eventory Invitational",
			Description: "A premium multi-day invitational featuring Thailand's top esports teams.",
			Game:        "Counter-Strike 2", Location: "Queen Sirikit Convention Center", StartsAt: now.AddDate(0, 0, 45).Add(4 * time.Hour),
			EndsAt: now.AddDate(0, 0, 47).Add(14 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 35),
			RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 5, MaxTeamSize: 5,
			EntryFee: 1200, Currency: "THB", Capacity: 16,
			Status: models.TournamentStatusRegistrationOpen, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-4000-8000-000000000006"),
			Name:        "Dev Arena Clash",
			Description: "A local development tournament for testing team registration and organizer workflows.",
			Game:        "Apex Legends", Location: "Online", StartsAt: now.AddDate(0, 0, 14).Add(11 * time.Hour),
			EndsAt: now.AddDate(0, 0, 14).Add(19 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 11),
			RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 1, MaxTeamSize: 3,
			EntryFee: 300, Currency: "THB", Capacity: 24,
			Status: models.TournamentStatusRegistrationClosed, Published: true,
		},
		{
			ID:          uuid.MustParse("33333333-0000-4000-8000-000000000007"),
			Name:        "Alex's Strategy Cup",
			Description: "A development sample for an upcoming solo or duo strategy tournament.",
			Game:        "Teamfight Tactics", Location: "Bangkok", StartsAt: now.AddDate(0, 0, 28).Add(10 * time.Hour),
			EndsAt: now.AddDate(0, 0, 28).Add(17 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 24),
			RegistrationMode: models.TournamentRegistrationModeSolo, MinTeamSize: 1, MaxTeamSize: 2,
			EntryFee: 150, Currency: "THB", Capacity: 48,
			Status: models.TournamentStatusRegistrationOpen, Published: false,
		},
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&tournaments).Error; err != nil {
		return err
	}

	staff := []models.TournamentStaff{
		{TournamentID: tournaments[0].ID, AccountID: aliceAccountID, Role: models.TournamentStaffRoleOwner},
		{TournamentID: tournaments[1].ID, AccountID: mayaAccountID, Role: models.TournamentStaffRoleOwner},
		{TournamentID: tournaments[2].ID, AccountID: aliceAccountID, Role: models.TournamentStaffRoleOwner},
		{TournamentID: tournaments[3].ID, AccountID: mayaAccountID, Role: models.TournamentStaffRoleOwner},
		{TournamentID: tournaments[4].ID, AccountID: aliceAccountID, Role: models.TournamentStaffRoleOwner},
		{TournamentID: tournaments[5].ID, AccountID: devAccountID, Role: models.TournamentStaffRoleOwner},
		{TournamentID: tournaments[6].ID, AccountID: devAccountID, Role: models.TournamentStaffRoleOwner},
	}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tournament_id"}, {Name: "account_id"}},
		UpdateAll: true,
	}).Create(&staff).Error; err != nil {
		return err
	}

	teams := []models.TournamentTeam{
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000001"), TournamentID: tournaments[0].ID, Name: "Neon Tigers", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000002"), TournamentID: tournaments[0].ID, Name: "Bangkok Byte", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000003"), TournamentID: tournaments[0].ID, Name: "Siam Sentinels", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000004"), TournamentID: tournaments[1].ID, Name: "Chula Phoenix", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000005"), TournamentID: tournaments[1].ID, Name: "River Guardians", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000006"), TournamentID: tournaments[1].ID, Name: "Lotus Legends", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000007"), TournamentID: tournaments[2].ID, Name: "Iron Fist BKK", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000008"), TournamentID: tournaments[2].ID, Name: "Manila Punishers", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000009"), TournamentID: tournaments[2].ID, Name: "Jakarta Kings", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000010"), TournamentID: tournaments[3].ID, Name: "Knight Owls", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000011"), TournamentID: tournaments[3].ID, Name: "Queen's Gambit Club", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000012"), TournamentID: tournaments[4].ID, Name: "Eventory Elite", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000013"), TournamentID: tournaments[4].ID, Name: "Northern Stars", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000014"), TournamentID: tournaments[4].ID, Name: "Crimson Circuit", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},

		// Dev Arena Clash (tournaments[5]) teams - 16 Accepted, 2 Rejected
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000021"), TournamentID: tournaments[5].ID, Name: "Apex Titans", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000022"), TournamentID: tournaments[5].ID, Name: "Neon Dragons", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000023"), TournamentID: tournaments[5].ID, Name: "Cyber Valkyrie", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000024"), TournamentID: tournaments[5].ID, Name: "Shadow Wolves", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000025"), TournamentID: tournaments[5].ID, Name: "Solar Flare", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000026"), TournamentID: tournaments[5].ID, Name: "Phantom Strike", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000027"), TournamentID: tournaments[5].ID, Name: "Echo Esports", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000028"), TournamentID: tournaments[5].ID, Name: "Vortex Gaming", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000029"), TournamentID: tournaments[5].ID, Name: "Quantum Pulse", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000030"), TournamentID: tournaments[5].ID, Name: "Ironclad Legion", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000031"), TournamentID: tournaments[5].ID, Name: "Crimson Vipers", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000032"), TournamentID: tournaments[5].ID, Name: "Astral Horizon", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000033"), TournamentID: tournaments[5].ID, Name: "Thunderbolt Gaming", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000034"), TournamentID: tournaments[5].ID, Name: "Frostbite Clan", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000035"), TournamentID: tournaments[5].ID, Name: "Midnight Stalkers", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000036"), TournamentID: tournaments[5].ID, Name: "Hyperion Rising", InviteCode: nil, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000037"), TournamentID: tournaments[5].ID, Name: "Disqualified Squad", InviteCode: nil, Status: models.TournamentTeamStatusRejected, LockedAt: &lockedAt},
		{ID: uuid.MustParse("44444444-0000-4000-8000-000000000038"), TournamentID: tournaments[5].ID, Name: "Late Entry Crew", InviteCode: nil, Status: models.TournamentTeamStatusRejected, LockedAt: &lockedAt},
	}
	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, UpdateAll: true,
	}).Create(&teams).Error; err != nil {
		return err
	}

	playerIDs := []uuid.UUID{
		uuid.MustParse("00000000-0000-4000-8000-000000000010"),
		uuid.MustParse("00000000-0000-4000-8000-000000000011"),
		uuid.MustParse("00000000-0000-4000-8000-000000000012"),
		uuid.MustParse("00000000-0000-4000-8000-000000000013"),
		uuid.MustParse("00000000-0000-4000-8000-000000000014"),
		uuid.MustParse("00000000-0000-4000-8000-000000000015"),
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
		{ID: uuid.MustParse("55555555-0000-4000-8000-000000000001"), TournamentID: tournaments[0].ID, GoalAmount: 50000, RaisedAmount: 32500, SupporterCount: 24},
		{ID: uuid.MustParse("55555555-0000-4000-8000-000000000002"), TournamentID: tournaments[1].ID, GoalAmount: 75000, RaisedAmount: 51000, SupporterCount: 31},
		{ID: uuid.MustParse("55555555-0000-4000-8000-000000000003"), TournamentID: tournaments[2].ID, GoalAmount: 120000, RaisedAmount: 84000, SupporterCount: 47},
		{ID: uuid.MustParse("55555555-0000-4000-8000-000000000004"), TournamentID: tournaments[3].ID, GoalAmount: 25000, RaisedAmount: 25000, SupporterCount: 19},
		{ID: uuid.MustParse("55555555-0000-4000-8000-000000000005"), TournamentID: tournaments[4].ID, GoalAmount: 500000, RaisedAmount: 287500, SupporterCount: 83},
		{ID: uuid.MustParse("55555555-0000-4000-8000-000000000006"), TournamentID: tournaments[5].ID, GoalAmount: 60000, RaisedAmount: 18000, SupporterCount: 12},
		{ID: uuid.MustParse("55555555-0000-4000-8000-000000000007"), TournamentID: tournaments[6].ID, GoalAmount: 30000, RaisedAmount: 0, SupporterCount: 0},
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, UpdateAll: true,
	}).Create(&funding).Error; err != nil {
		return err
	}

	// Seed realistic tournament bracket DAG & live match progression for Dev Arena Clash (tournaments[5])
	devTournamentID := tournaments[5].ID
	finalID := uuid.MustParse("77777777-0000-4000-8000-000000000001")
	semi1ID := uuid.MustParse("77777777-0000-4000-8000-000000000002")
	semi2ID := uuid.MustParse("77777777-0000-4000-8000-000000000003")
	qf1ID := uuid.MustParse("77777777-0000-4000-8000-000000000004")
	qf2ID := uuid.MustParse("77777777-0000-4000-8000-000000000005")
	qf3ID := uuid.MustParse("77777777-0000-4000-8000-000000000006")
	qf4ID := uuid.MustParse("77777777-0000-4000-8000-000000000007")
	r16_1 := uuid.MustParse("77777777-0000-4000-8000-000000000008")
	r16_2 := uuid.MustParse("77777777-0000-4000-8000-000000000009")
	r16_3 := uuid.MustParse("77777777-0000-4000-8000-000000000010")
	r16_4 := uuid.MustParse("77777777-0000-4000-8000-000000000011")
	r16_5 := uuid.MustParse("77777777-0000-4000-8000-000000000012")
	r16_6 := uuid.MustParse("77777777-0000-4000-8000-000000000013")
	r16_7 := uuid.MustParse("77777777-0000-4000-8000-000000000014")
	r16_8 := uuid.MustParse("77777777-0000-4000-8000-000000000015")

	matches := []models.Match{
		// Stage 4: Championship
		{ID: finalID, TournamentID: devTournamentID, Name: "Championship Final", Venue: "Main Arena Stage", StageLevel: 4, TeamCount: 2, AdvancingCount: 0, TargetMatchID: nil, IsCompleted: false},
		// Stage 3: Semifinals
		{ID: semi1ID, TournamentID: devTournamentID, Name: "Semifinal A", Venue: "Main Arena A", StageLevel: 3, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &finalID, IsCompleted: false},
		{ID: semi2ID, TournamentID: devTournamentID, Name: "Semifinal B", Venue: "Main Arena B", StageLevel: 3, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &finalID, IsCompleted: false},
		// Stage 2: Quarterfinals
		{ID: qf1ID, TournamentID: devTournamentID, Name: "Quarterfinal 1", Venue: "Station 1", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &semi1ID, IsCompleted: false},
		{ID: qf2ID, TournamentID: devTournamentID, Name: "Quarterfinal 2", Venue: "Station 2", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &semi1ID, IsCompleted: false},
		{ID: qf3ID, TournamentID: devTournamentID, Name: "Quarterfinal 3", Venue: "Station 3", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &semi2ID, IsCompleted: false},
		{ID: qf4ID, TournamentID: devTournamentID, Name: "Quarterfinal 4", Venue: "Station 4", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &semi2ID, IsCompleted: false},
		// Stage 1: Round of 16 (8 matches x 2 teams = 16 total capacity matching the 16 accepted teams)
		{ID: r16_1, TournamentID: devTournamentID, Name: "Round 1 - Heat 1", Venue: "Station 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &qf1ID, IsCompleted: true},
		{ID: r16_2, TournamentID: devTournamentID, Name: "Round 1 - Heat 2", Venue: "Station 2", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &qf1ID, IsCompleted: true},
		{ID: r16_3, TournamentID: devTournamentID, Name: "Round 1 - Heat 3", Venue: "Station 3", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &qf2ID, IsCompleted: false},
		{ID: r16_4, TournamentID: devTournamentID, Name: "Round 1 - Heat 4", Venue: "Station 4", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &qf2ID, IsCompleted: false},
		{ID: r16_5, TournamentID: devTournamentID, Name: "Round 1 - Heat 5", Venue: "Station 5", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &qf3ID, IsCompleted: false},
		{ID: r16_6, TournamentID: devTournamentID, Name: "Round 1 - Heat 6", Venue: "Station 6", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &qf3ID, IsCompleted: false},
		{ID: r16_7, TournamentID: devTournamentID, Name: "Round 1 - Heat 7", Venue: "Station 7", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &qf4ID, IsCompleted: false},
		{ID: r16_8, TournamentID: devTournamentID, Name: "Round 1 - Heat 8", Venue: "Station 8", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &qf4ID, IsCompleted: false},
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, UpdateAll: true,
	}).Create(&matches).Error; err != nil {
		return err
	}

	rank1 := 1
	rank2 := 2
	score21 := "2-1"
	score12 := "1-2"
	score20 := "2-0"
	score02 := "0-2"

	matchEntries := []models.MatchEntry{
		// Match 1.1 (completed)
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000001"), MatchID: r16_1, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000021"), FinalRank: &rank1, Score: &score21, Advances: true},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000002"), MatchID: r16_1, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000022"), FinalRank: &rank2, Score: &score12, Advances: false},
		// Match 1.2 (completed)
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000003"), MatchID: r16_2, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000023"), FinalRank: &rank1, Score: &score20, Advances: true},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000004"), MatchID: r16_2, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000024"), FinalRank: &rank2, Score: &score02, Advances: false},
		// Advanced into Quarterfinal 1
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000005"), MatchID: qf1ID, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000021"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000006"), MatchID: qf1ID, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000023"), Advances: false},
		// Remaining Stage 1 matches (pending)
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000007"), MatchID: r16_3, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000025"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000008"), MatchID: r16_3, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000026"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000009"), MatchID: r16_4, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000027"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000010"), MatchID: r16_4, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000028"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000011"), MatchID: r16_5, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000029"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000012"), MatchID: r16_5, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000030"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000013"), MatchID: r16_6, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000031"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000014"), MatchID: r16_6, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000032"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000015"), MatchID: r16_7, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000033"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000016"), MatchID: r16_7, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000034"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000017"), MatchID: r16_8, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000035"), Advances: false},
		{ID: uuid.MustParse("88888888-0000-4000-8000-000000000018"), MatchID: r16_8, TeamID: uuid.MustParse("44444444-0000-4000-8000-000000000036"), Advances: false},
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}}, UpdateAll: true,
	}).Create(&matchEntries).Error
}
