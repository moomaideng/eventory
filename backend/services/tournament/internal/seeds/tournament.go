package seeds

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedTournaments(db *gorm.DB) error {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	lockedAt := now.Add(-24 * time.Hour)
	sponsorDeadline := now.AddDate(0, 0, 3)
	id := func(value string) uuid.UUID { return uuid.MustParse(value) }
	fixtureID := func(prefix, suffix int) uuid.UUID {
		return id(fmt.Sprintf("%08d-0000-4000-8000-%012d", prefix, suffix))
	}

	const notice = "Local seed fixture only — not a real tournament. "
	alexID := id("99999999-0000-4000-8000-000000000001")
	aliceID := id("00000000-0000-4000-8000-000000000001")
	mayaID := id("00000000-0000-4000-8000-000000000003")
	sponsorID := fixtureID(33333333, 1)
	setupID := fixtureID(33333333, 6)
	liveID := fixtureID(33333333, 7)
	completedID := fixtureID(33333333, 8)
	aliceTournamentID := fixtureID(33333333, 2)
	mayaTournamentID := fixtureID(33333333, 3)

	tournaments := []models.Tournament{
		{ID: sponsorID, Name: "[DEMO · Sponsor] Checkout Playground", Description: notice + "Use this tournament to test sponsorship checkout, payments, and payment events.", Detail: "## Local demo\n\nUse **Dev Quick Login** to sponsor this Maya-owned tournament.", Game: "Sponsor demo", Location: "Local Dev Environment", StartsAt: now.AddDate(0, 0, 14).Add(10 * time.Hour), EndsAt: now.AddDate(0, 0, 15).Add(18 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 7), RegistrationMode: models.TournamentRegistrationModeSolo, MinTeamSize: 1, MaxTeamSize: 1, Capacity: 32, Currency: "THB", Status: models.TournamentStatusCrowdfunding, Crowdfunding: models.CrowdfundingInfo{GoalAmountMinor: 5000000, Deadline: &sponsorDeadline}},
		{ID: setupID, Name: "[DEMO · Bracket] Setup Playground", Description: notice + "Use this tournament to define and seed a 16-team bracket before it starts.", Detail: "## Local demo\n\nSixteen accepted teams are ready. Define and seed a Stage 1 bracket as Alex (Dev).", Game: "Bracket demo", Location: "Local Dev Environment", StartsAt: now.AddDate(0, 0, 10).Add(10 * time.Hour), EndsAt: now.AddDate(0, 0, 10).Add(18 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 7), RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 1, MaxTeamSize: 3, Capacity: 16, Currency: "THB", Status: models.TournamentStatusRegistration},
		{ID: liveID, Name: "[DEMO · Bracket] Live Results", Description: notice + "Use this tournament to inspect a 16-team live bracket and record the remaining results.", Detail: "## Local demo\n\nRound of 16 is underway; two completed matches have already advanced to the quarterfinal.", Game: "Bracket demo", Location: "Local Dev Environment", StartsAt: now.AddDate(0, 0, -1).Add(10 * time.Hour), EndsAt: now.AddDate(0, 0, 1).Add(18 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, -3), RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 1, MaxTeamSize: 3, Capacity: 16, Currency: "THB", Status: models.TournamentStatusOngoing},
		{ID: completedID, Name: "[DEMO · Bracket] Completed Archive", Description: notice + "Use this tournament to inspect a fully completed bracket and final standings.", Detail: "## Local demo\n\nEvery match is complete; the final has a winner.", Game: "Bracket demo", Location: "Local Dev Environment", StartsAt: now.AddDate(0, 0, -7).Add(10 * time.Hour), EndsAt: now.AddDate(0, 0, -6).Add(18 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, -9), RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 1, MaxTeamSize: 3, Capacity: 4, Currency: "THB", Status: models.TournamentStatusCompleted},
		{ID: aliceTournamentID, Name: "[SAMPLE · Alice] Valorant Community Cup", Description: notice + "A small public listing owned by Alice Events.", Game: "Valorant", Location: "Local Dev Environment", StartsAt: now.AddDate(0, 0, 18).Add(12 * time.Hour), EndsAt: now.AddDate(0, 0, 18).Add(20 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 14), RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 5, MaxTeamSize: 5, Capacity: 32, Currency: "THB", Status: models.TournamentStatusRegistration},
		{ID: mayaTournamentID, Name: "[SAMPLE · Maya] Mobile Legends Weekend", Description: notice + "A small public listing owned by Maya Works.", Game: "Mobile Legends", Location: "Local Dev Environment", StartsAt: now.AddDate(0, 0, 24).Add(9 * time.Hour), EndsAt: now.AddDate(0, 0, 25).Add(18 * time.Hour), RegistrationDeadline: now.AddDate(0, 0, 20), RegistrationMode: models.TournamentRegistrationModeTeam, MinTeamSize: 5, MaxTeamSize: 5, Capacity: 24, Currency: "THB", Status: models.TournamentStatusRegistration},
	}
	if err := upsert(db, "id", &tournaments); err != nil {
		return err
	}

	staff := []models.TournamentStaff{{TournamentID: sponsorID, AccountID: mayaID, Role: models.TournamentStaffRoleOwner}, {TournamentID: setupID, AccountID: alexID, Role: models.TournamentStaffRoleOwner}, {TournamentID: liveID, AccountID: alexID, Role: models.TournamentStaffRoleOwner}, {TournamentID: completedID, AccountID: alexID, Role: models.TournamentStaffRoleOwner}, {TournamentID: aliceTournamentID, AccountID: aliceID, Role: models.TournamentStaffRoleOwner}, {TournamentID: mayaTournamentID, AccountID: mayaID, Role: models.TournamentStaffRoleOwner}}
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tournament_id"}, {Name: "account_id"}}, UpdateAll: true}).Create(&staff).Error; err != nil {
		return err
	}

	tiers := []models.SponsorshipTier{{ID: fixtureID(55555555, 1), TournamentID: sponsorID, Name: "Bronze", PriceMinor: 100000, Slots: 10, Perks: []string{"Logo on tournament page"}}, {ID: fixtureID(55555555, 2), TournamentID: sponsorID, Name: "Silver", PriceMinor: 500000, Slots: 3, Perks: []string{"Logo on tournament page", "Social media mention"}}, {ID: fixtureID(55555555, 3), TournamentID: sponsorID, Name: "Gold", PriceMinor: 1500000, Slots: 1, Perks: []string{"Featured logo", "Social media mention", "Opening announcement"}}}
	if err := upsert(db, "id", &tiers); err != nil {
		return err
	}

	setupNames := []string{"Setup Tigers", "Setup Dragons", "Setup Wolves", "Setup Falcons", "Setup Comets", "Setup Ravens", "Setup Phoenix", "Setup Panthers", "Setup Sharks", "Setup Cobras", "Setup Bears", "Setup Owls", "Setup Lions", "Setup Foxes", "Setup Knights", "Setup Giants"}
	liveNames := []string{"Apex Titans", "Neon Dragons", "Cyber Valkyrie", "Shadow Wolves", "Solar Flare", "Phantom Strike", "Echo Esports", "Vortex Gaming", "Quantum Pulse", "Ironclad Legion", "Crimson Vipers", "Astral Horizon", "Thunderbolt Gaming", "Frostbite Clan", "Midnight Stalkers", "Hyperion Rising"}
	archiveNames := []string{"Archive Tigers", "Archive Dragons", "Archive Wolves", "Archive Falcons"}
	teams := make([]models.TournamentTeam, 0, 36)
	addTeams := func(tournamentID uuid.UUID, names []string, firstID int) {
		for index, name := range names {
			teams = append(teams, models.TournamentTeam{ID: fixtureID(44444444, firstID+index), TournamentID: tournamentID, Name: name, Status: models.TournamentTeamStatusAccepted, LockedAt: &lockedAt})
		}
	}
	addTeams(setupID, setupNames, 61)
	liveTeamStart := len(teams)
	addTeams(liveID, liveNames, 21)
	archiveTeamStart := len(teams)
	addTeams(completedID, archiveNames, 45)
	if err := upsert(db, "id", &teams); err != nil {
		return err
	}

	playerIDs := []uuid.UUID{id("00000000-0000-4000-8000-000000000010"), id("00000000-0000-4000-8000-000000000011"), id("00000000-0000-4000-8000-000000000012"), id("00000000-0000-4000-8000-000000000013"), id("00000000-0000-4000-8000-000000000014"), id("00000000-0000-4000-8000-000000000015")}
	members := make([]models.TournamentTeamMember, 0, len(teams)*2)
	for index, team := range teams {
		for offset := 0; offset < 2; offset++ {
			role := models.TournamentTeamMemberRoleMember
			if offset == 0 {
				role = models.TournamentTeamMemberRoleCaptain
			}
			accountID := playerIDs[(index*2+offset)%len(playerIDs)]
			members = append(members, models.TournamentTeamMember{ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte(team.ID.String()+":"+accountID.String())), TournamentTeamID: team.ID, AccountID: accountID, Role: role})
		}
	}
	if err := upsert(db, "id", &members); err != nil {
		return err
	}

	finalID := fixtureID(77777777, 1)
	semis := []uuid.UUID{fixtureID(77777777, 2), fixtureID(77777777, 3)}
	quarters := []uuid.UUID{fixtureID(77777777, 4), fixtureID(77777777, 5), fixtureID(77777777, 6), fixtureID(77777777, 7)}
	roundOf16 := []uuid.UUID{fixtureID(77777777, 8), fixtureID(77777777, 9), fixtureID(77777777, 10), fixtureID(77777777, 11), fixtureID(77777777, 12), fixtureID(77777777, 13), fixtureID(77777777, 14), fixtureID(77777777, 15)}
	matches := []models.Match{{ID: finalID, TournamentID: liveID, Name: "Championship Final", Venue: "Main Arena", StageLevel: 4, TeamCount: 2}}
	for index, matchID := range semis {
		matches = append(matches, models.Match{ID: matchID, TournamentID: liveID, Name: fmt.Sprintf("Semifinal %d", index+1), Venue: "Main Arena", StageLevel: 3, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &finalID})
	}
	for index, matchID := range quarters {
		target := semis[index/2]
		matches = append(matches, models.Match{ID: matchID, TournamentID: liveID, Name: fmt.Sprintf("Quarterfinal %d", index+1), Venue: fmt.Sprintf("Station %d", index+1), StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &target})
	}
	for index, matchID := range roundOf16 {
		target := quarters[index/2]
		matches = append(matches, models.Match{ID: matchID, TournamentID: liveID, Name: fmt.Sprintf("Round of 16 · Match %d", index+1), Venue: fmt.Sprintf("Station %d", index+1), StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &target, IsCompleted: index < 2})
	}

	archiveFinalID := fixtureID(77777777, 31)
	archiveSemi1ID := fixtureID(77777777, 32)
	archiveSemi2ID := fixtureID(77777777, 33)
	matches = append(matches, models.Match{ID: archiveFinalID, TournamentID: completedID, Name: "Archive Final", Venue: "Local Dev Arena", StageLevel: 2, TeamCount: 2, IsCompleted: true}, models.Match{ID: archiveSemi1ID, TournamentID: completedID, Name: "Archive Semifinal 1", Venue: "Local Dev Arena", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &archiveFinalID, IsCompleted: true}, models.Match{ID: archiveSemi2ID, TournamentID: completedID, Name: "Archive Semifinal 2", Venue: "Local Dev Arena", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: &archiveFinalID, IsCompleted: true})
	if err := upsert(db, "id", &matches); err != nil {
		return err
	}

	rank1, rank2 := 1, 2
	score21, score12, score20, score02 := "2-1", "1-2", "2-0", "0-2"
	entries := make([]models.MatchEntry, 0, 26)
	for index, matchID := range roundOf16 {
		first, second := teams[liveTeamStart+index*2].ID, teams[liveTeamStart+index*2+1].ID
		entryID := 1 + index*2
		if index < 2 {
			entries = append(entries, models.MatchEntry{ID: fixtureID(88888888, entryID), MatchID: matchID, TeamID: first, FinalRank: &rank1, Score: &score21, Advances: true}, models.MatchEntry{ID: fixtureID(88888888, entryID+1), MatchID: matchID, TeamID: second, FinalRank: &rank2, Score: &score12})
		} else {
			entries = append(entries, models.MatchEntry{ID: fixtureID(88888888, entryID), MatchID: matchID, TeamID: first}, models.MatchEntry{ID: fixtureID(88888888, entryID+1), MatchID: matchID, TeamID: second})
		}
	}
	entries = append(entries, models.MatchEntry{ID: fixtureID(88888888, 17), MatchID: quarters[0], TeamID: teams[liveTeamStart].ID}, models.MatchEntry{ID: fixtureID(88888888, 18), MatchID: quarters[0], TeamID: teams[liveTeamStart+2].ID})
	archiveTeams := teams[archiveTeamStart:]
	entries = append(entries, models.MatchEntry{ID: fixtureID(88888888, 31), MatchID: archiveSemi1ID, TeamID: archiveTeams[0].ID, FinalRank: &rank1, Score: &score20, Advances: true}, models.MatchEntry{ID: fixtureID(88888888, 32), MatchID: archiveSemi1ID, TeamID: archiveTeams[1].ID, FinalRank: &rank2, Score: &score02}, models.MatchEntry{ID: fixtureID(88888888, 33), MatchID: archiveSemi2ID, TeamID: archiveTeams[2].ID, FinalRank: &rank1, Score: &score21, Advances: true}, models.MatchEntry{ID: fixtureID(88888888, 34), MatchID: archiveSemi2ID, TeamID: archiveTeams[3].ID, FinalRank: &rank2, Score: &score12}, models.MatchEntry{ID: fixtureID(88888888, 35), MatchID: archiveFinalID, TeamID: archiveTeams[0].ID, FinalRank: &rank1, Score: &score21}, models.MatchEntry{ID: fixtureID(88888888, 36), MatchID: archiveFinalID, TeamID: archiveTeams[2].ID, FinalRank: &rank2, Score: &score12})
	return upsert(db, "id", &entries)
}

func upsert(db *gorm.DB, column string, value any) error {
	return db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: column}}, UpdateAll: true}).Create(value).Error
}
