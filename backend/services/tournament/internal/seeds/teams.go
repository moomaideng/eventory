package seeds

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/seedids"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedTeamsAndBrackets populates deterministic teams/solo entries, rosters,
// review audit trails, and tournament bracket matches & entries.
func SeedTeamsAndBrackets(db *gorm.DB) error {
	now := time.Now().UTC().Truncate(time.Second)

	teams, members := buildTeamsAndMembers(now)

	// Clear fixture child rows for seeded tournaments before re-inserting so rosters
	// and match slots remain consistent across seed updates without touching user-created tournaments.
	seededTournaments := []uuid.UUID{
		seedids.TournTestReviewQueue,
		seedids.TournTestSoloPaid,
		seedids.TournTestBracketSetup,
		seedids.TournTestPermAdmin,
		seedids.TournTestPermReferee,
		seedids.TournTestBracketLive,
		seedids.TournTestBracketCompleted,
		seedids.TournRealRegValorant,
		seedids.TournRealRegSF6Solo,
		seedids.TournRealOngoingDota2,
		seedids.TournRealCompletedMLBB,
	}

	if err := db.Where("tournament_id IN ?", seededTournaments).Delete(&models.MatchEntry{}).Error; err != nil {
		return err
	}
	if err := db.Where("tournament_id IN ?", seededTournaments).Delete(&models.Match{}).Error; err != nil {
		return err
	}

	// Upsert teams; preserve PAYMENT_PENDING state if a developer paid it during local testing?
	// Static structural teams can UpdateAll safely except TM050003 which links to a live testable invoice.
	for i := range teams {
		if teams[i].Status == models.TournamentTeamStatusPaymentPending {
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "tournament_id"}, {Name: "id"}},
				DoNothing: true,
			}).Create(&teams[i]).Error; err != nil {
				return err
			}
			continue
		}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tournament_id"}, {Name: "id"}},
			UpdateAll: true,
		}).Create(&teams[i]).Error; err != nil {
			return err
		}
	}

	for i := range members {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tournament_id"}, {Name: "tournament_team_id"}, {Name: "account_id"}},
			UpdateAll: true,
		}).Create(&members[i]).Error; err != nil {
			return err
		}
	}

	reviews := buildTeamReviews(now)
	for i := range reviews {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&reviews[i]).Error; err != nil {
			return err
		}
	}

	matches, entries := buildBrackets(now)
	for i := range matches {
		if err := db.Create(&matches[i]).Error; err != nil {
			return err
		}
	}
	for i := range entries {
		if err := db.Create(&entries[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

func strPtr(s string) *string    { return &s }
func intPtr(i int) *int          { return &i }
func uuidPtr(id uuid.UUID) *uuid.UUID { return &id }

func addTeam(
	teams *[]models.TournamentTeam,
	members *[]models.TournamentTeamMember,
	t models.TournamentTeam,
	captain uuid.UUID,
	teammates ...uuid.UUID,
) {
	*teams = append(*teams, t)
	*members = append(*members, models.TournamentTeamMember{
		ID:               uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s:%s:%s", t.TournamentID, t.ID, captain))),
		TournamentID:     t.TournamentID,
		TournamentTeamID: t.ID,
		AccountID:        captain,
		Role:             models.TournamentTeamMemberRoleCaptain,
		JoinedAt:         t.CreatedAt,
	})
	for _, m := range teammates {
		*members = append(*members, models.TournamentTeamMember{
			ID:               uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s:%s:%s", t.TournamentID, t.ID, m))),
			TournamentID:     t.TournamentID,
			TournamentTeamID: t.ID,
			AccountID:        m,
			Role:             models.TournamentTeamMemberRoleMember,
			JoinedAt:         t.CreatedAt.Add(time.Minute),
		})
	}
}

func buildTeamsAndMembers(now time.Time) ([]models.TournamentTeam, []models.TournamentTeamMember) {
	var teams []models.TournamentTeam
	var members []models.TournamentTeamMember

	regInvByTeam := map[string]seedids.RegistrationInvoiceFixture{}
	for _, inv := range seedids.RegistrationInvoices() {
		regInvByTeam[inv.TournamentID.String()+":"+inv.TeamID] = inv
	}
	attachPayment := func(t *models.TournamentTeam) {
		inv, ok := regInvByTeam[t.TournamentID.String()+":"+t.ID]
		if !ok {
			return
		}
		t.PaymentAttemptID = uuidPtr(inv.AttemptID)
		t.PaymentID = uuidPtr(inv.InvoiceID)
		t.PaymentPayerID = uuidPtr(inv.PayerAccountID)
		t.PaymentAmountMinor = inv.AmountMinor
		switch inv.Status {
		case "UNPAID":
			t.PaymentExpiresAt = ptrTime(now.Add(45 * time.Minute))
		case "REFUNDED":
			t.RefundedAt = ptrTime(now.Add(-24 * time.Hour))
		}
	}

	// -------------------------------------------------------------------------
	// A5: [TEST · Review] Queue, Forms & Refunds (TEAM 2-3, Fee 500 THB)
	// -------------------------------------------------------------------------
	addTeam(&teams, &members, models.TournamentTeam{
		TournamentID: seedids.TournTestReviewQueue,
		ID:           "TM050001",
		Name:         "Dev Alpha (Forming)",
		InviteCode:   strPtr("TEST01"),
		Status:       models.TournamentTeamStatusForming,
		CreatedAt:    now.Add(-72 * time.Hour),
	}, seedids.AlexID, seedids.PlayerID(16))

	// PlayerID(2) is intentionally missing MongoDB FormResponse to demonstrate incomplete form lock blocker.
	addTeam(&teams, &members, models.TournamentTeam{
		TournamentID: seedids.TournTestReviewQueue,
		ID:           "TM050002",
		Name:         "Missing Form Five",
		InviteCode:   strPtr("TEST02"),
		Status:       models.TournamentTeamStatusForming,
		CreatedAt:    now.Add(-68 * time.Hour),
	}, seedids.PlengID, seedids.PlayerID(2))

	tm0503 := models.TournamentTeam{
		TournamentID: seedids.TournTestReviewQueue,
		ID:           "TM050003",
		Name:         "Pending Checkout Squad",
		Status:       models.TournamentTeamStatusPaymentPending,
		CreatedAt:    now.Add(-12 * time.Hour),
	}
	attachPayment(&tm0503)
	addTeam(&teams, &members, tm0503, seedids.TonklaID, seedids.PlayerID(3))

	for idx, spec := range []struct {
		id      string
		name    string
		status  models.TournamentTeamStatus
		captain uuid.UUID
		mate    uuid.UUID
	}{
		{"TM050004", "Neon Pulse (Refund Test)", models.TournamentTeamStatusLocked, seedids.BeamID, seedids.PlayerID(4)},
		{"TM050005", "Krungthep Owls", models.TournamentTeamStatusLocked, seedids.PlayerID(5), seedids.PlayerID(6)},
		{"TM050006", "Asoke Phantoms", models.TournamentTeamStatusLocked, seedids.PlayerID(7), seedids.PlayerID(8)},
		{"TM050007", "Chao Phraya Kings", models.TournamentTeamStatusAccepted, seedids.PlayerID(9), seedids.PlayerID(10)},
		{"TM050008", "Siam Tigers", models.TournamentTeamStatusAccepted, seedids.PlayerID(11), seedids.PlayerID(12)},
		{"TM050009", "Silom Strikers (Rejected)", models.TournamentTeamStatusRejected, seedids.PlayerID(13), seedids.PlayerID(14)},
		{"TM050010", "Thonglor Rebels (Rejected)", models.TournamentTeamStatusRejected, seedids.PlayerID(15), seedids.PlayerID(17)},
	} {
		t := models.TournamentTeam{
			TournamentID: seedids.TournTestReviewQueue,
			ID:           spec.id,
			Name:         spec.name,
			Status:       spec.status,
			CreatedAt:    now.Add(-time.Duration(60-idx*4) * time.Hour),
			LockedAt:     ptrTime(now.Add(-time.Duration(48-idx*3) * time.Hour)),
		}
		attachPayment(&t)
		addTeam(&teams, &members, t, spec.captain, spec.mate)
	}

	// -------------------------------------------------------------------------
	// A6: [TEST · Solo] Paid Entry & Form Playground (Alex unregistered)
	// -------------------------------------------------------------------------
	for idx, spec := range []struct {
		id      string
		name    string
		status  models.TournamentTeamStatus
		account uuid.UUID
	}{
		{"SL060001", "Kaito", models.TournamentTeamStatusAccepted, seedids.KaitoID},
		{"SL060002", "Pleng", models.TournamentTeamStatusAccepted, seedids.PlengID},
		{"SL060003", "Tonkla", models.TournamentTeamStatusAccepted, seedids.TonklaID},
		{"SL060004", "Beam", models.TournamentTeamStatusAccepted, seedids.BeamID},
		{"SL060005", "Fah", models.TournamentTeamStatusLocked, seedids.FahID},
		{"SL060006", "Player 19", models.TournamentTeamStatusLocked, seedids.PlayerID(19)},
		{"SL060007", "Player 20", models.TournamentTeamStatusRejected, seedids.PlayerID(20)},
	} {
		t := models.TournamentTeam{
			TournamentID: seedids.TournTestSoloPaid,
			ID:           spec.id,
			Name:         spec.name,
			Status:       spec.status,
			CreatedAt:    now.Add(-time.Duration(72-idx*5) * time.Hour),
			LockedAt:     ptrTime(now.Add(-time.Duration(68-idx*5) * time.Hour)),
		}
		attachPayment(&t)
		addTeam(&teams, &members, t, spec.account)
	}

	// -------------------------------------------------------------------------
	// A7: [TEST · Bracket] Setup & Start Gate (16 ACCEPTED teams, 0 matches)
	// -------------------------------------------------------------------------
	bracketSetupNames := []string{
		"Phoenix Vanguard", "Bangkok Drift", "Siam Serpents", "Crimson Lotus",
		"Ayutthaya Knights", "Mekong Tempest", "Cyber Garuda", "Lanna Wolves",
		"Thonburi Guardians", "Andaman Sharks", "Rattanakosin Five", "Erawan Apex",
		"Sukhumvit Spartans", "Sathorn Ravens", "Chula Circuit", "Ladprao Legends",
	}
	for i := 0; i < 16; i++ {
		var extra []uuid.UUID
		if i < 4 {
			extra = []uuid.UUID{seedids.PlayerID(17 + i)}
		}
		addTeam(&teams, &members, models.TournamentTeam{
			TournamentID: seedids.TournTestBracketSetup,
			ID:           fmt.Sprintf("TM07%04d", i+1),
			Name:         bracketSetupNames[i],
			Status:       models.TournamentTeamStatusAccepted,
			CreatedAt:    now.Add(-time.Duration(240-i*6) * time.Hour),
			LockedAt:     ptrTime(now.Add(-time.Duration(220-i*6) * time.Hour)),
		}, seedids.PlayerID(i+1), extra...)
	}

	// -------------------------------------------------------------------------
	// A8: [TEST · Permissions] Admin Role View
	// -------------------------------------------------------------------------
	for idx, spec := range []struct {
		id      string
		name    string
		status  models.TournamentTeamStatus
		invite  *string
		captain uuid.UUID
		mate    uuid.UUID
	}{
		{"TM080001", "Valkyrie Zero", models.TournamentTeamStatusAccepted, nil, seedids.KaitoID, seedids.PlayerID(1)},
		{"TM080002", "Khao San Aces", models.TournamentTeamStatusAccepted, nil, seedids.PlengID, seedids.PlayerID(2)},
		{"TM080003", "Rama IX Syndicate", models.TournamentTeamStatusAccepted, nil, seedids.TonklaID, seedids.PlayerID(3)},
		{"TM080004", "Bangna Raiders", models.TournamentTeamStatusAccepted, nil, seedids.BeamID, seedids.PlayerID(4)},
		{"TM080005", "Phaya Thai Core", models.TournamentTeamStatusLocked, nil, seedids.FahID, seedids.PlayerID(5)},
		{"TM080006", "Ari Phantoms", models.TournamentTeamStatusLocked, nil, seedids.MekID, seedids.PlayerID(6)},
		{"TM080007", "Open Admin Lobby", models.TournamentTeamStatusForming, strPtr("ADM006"), seedids.PlayerID(7), seedids.PlayerID(8)},
	} {
		var lockedAt *time.Time
		if spec.status != models.TournamentTeamStatusForming {
			lockedAt = ptrTime(now.Add(-time.Duration(48-idx*3) * time.Hour))
		}
		addTeam(&teams, &members, models.TournamentTeam{
			TournamentID: seedids.TournTestPermAdmin,
			ID:           spec.id,
			Name:         spec.name,
			InviteCode:   spec.invite,
			Status:       spec.status,
			CreatedAt:    now.Add(-time.Duration(60-idx*4) * time.Hour),
			LockedAt:     lockedAt,
		}, spec.captain, spec.mate)
	}

	// -------------------------------------------------------------------------
	// A9: [TEST · Permissions] Referee Role View (8 ACCEPTED teams)
	// -------------------------------------------------------------------------
	refCaptains := []uuid.UUID{
		seedids.KaitoID, seedids.PlengID, seedids.TonklaID, seedids.BeamID,
		seedids.FahID, seedids.MekID, seedids.PlayerID(7), seedids.PlayerID(8),
	}
	for i := 0; i < 8; i++ {
		addTeam(&teams, &members, models.TournamentTeam{
			TournamentID: seedids.TournTestPermReferee,
			ID:           fmt.Sprintf("TM09%04d", i+1),
			Name:         fmt.Sprintf("Ref Squad %02d", i+1),
			Status:       models.TournamentTeamStatusAccepted,
			CreatedAt:    now.Add(-time.Duration(168-i*4) * time.Hour),
			LockedAt:     ptrTime(now.Add(-time.Duration(160-i*4) * time.Hour)),
		}, refCaptains[i], seedids.PlayerID(9+i))
	}

	// -------------------------------------------------------------------------
	// A10: [TEST · Bracket] Live Match & Swap (9 ACCEPTED teams)
	// -------------------------------------------------------------------------
	liveCaptains := []uuid.UUID{
		seedids.KaitoID, seedids.PlengID, seedids.TonklaID, seedids.BeamID,
		seedids.FahID, seedids.MekID, seedids.PlayerID(17), seedids.PlayerID(18),
		seedids.PlayerID(19),
	}
	liveNames := []string{
		"Paper Planes", "Mirage BKK", "Solaris Esports", "Omega Pulse",
		"Viper Strike", "Banyan Crew", "Delta Surge", "Zenith X",
		"Wildcard Fox",
	}
	for i := 0; i < 9; i++ {
		addTeam(&teams, &members, models.TournamentTeam{
			TournamentID: seedids.TournTestBracketLive,
			ID:           fmt.Sprintf("TM10%04d", i+1),
			Name:         liveNames[i],
			Status:       models.TournamentTeamStatusAccepted,
			CreatedAt:    now.Add(-time.Duration(180-i*4) * time.Hour),
			LockedAt:     ptrTime(now.Add(-time.Duration(170-i*4) * time.Hour)),
		}, liveCaptains[i])
	}

	// -------------------------------------------------------------------------
	// A11: [TEST · Bracket] Completed Archive (6 ACCEPTED teams)
	// -------------------------------------------------------------------------
	compNames := []string{
		"Champion Lions", "Silver Falcons", "Bronze Bears",
		"Fourth Foxes", "FastTrack Hawks", "Sixth Owls",
	}
	compCaptains := []uuid.UUID{
		seedids.KaitoID, seedids.PlengID, seedids.TonklaID,
		seedids.BeamID, seedids.FahID, seedids.MekID,
	}
	for i := 0; i < 6; i++ {
		addTeam(&teams, &members, models.TournamentTeam{
			TournamentID: seedids.TournTestBracketCompleted,
			ID:           fmt.Sprintf("TM11%04d", i+1),
			Name:         compNames[i],
			Status:       models.TournamentTeamStatusAccepted,
			CreatedAt:    now.Add(-25 * 24 * time.Hour),
			LockedAt:     ptrTime(now.Add(-23 * 24 * time.Hour)),
		}, compCaptains[i])
	}

	// -------------------------------------------------------------------------
	// B3: SEA Valorant Challengers Open (Paid 800 THB, TEAM 5-5)
	// PlayerID(2) is REFEREE so omitted from rosters.
	// -------------------------------------------------------------------------
	for idx, spec := range []struct {
		id     string
		name   string
		status models.TournamentTeamStatus
		invite *string
		roster []uuid.UUID
	}{
		{"TM130001", "Kaito Protocol", models.TournamentTeamStatusAccepted, nil, []uuid.UUID{seedids.KaitoID, seedids.PlayerID(3), seedids.PlayerID(4), seedids.PlayerID(5), seedids.PlayerID(6)}},
		{"TM130002", "Aurora Bangkok", models.TournamentTeamStatusAccepted, nil, []uuid.UUID{seedids.PlengID, seedids.PlayerID(7), seedids.PlayerID(8), seedids.PlayerID(9), seedids.PlayerID(10)}},
		{"TM130003", "Sentinel X", models.TournamentTeamStatusAccepted, nil, []uuid.UUID{seedids.MekID, seedids.PlayerID(11), seedids.PlayerID(12), seedids.PlayerID(14), seedids.PlayerID(15)}},
		{"TM130004", "Crimson Radiant", models.TournamentTeamStatusAccepted, nil, []uuid.UUID{seedids.PlayerID(13), seedids.PlayerID(16), seedids.PlayerID(17), seedids.PlayerID(18), seedids.PlayerID(19)}},
		{"TM130005", "Skyline 5", models.TournamentTeamStatusLocked, nil, []uuid.UUID{seedids.FahID, seedids.PlayerID(20), seedids.PlayerID(21), seedids.PlayerID(23), seedids.PlayerID(24)}},
		{"TM130006", "Obsidian VCT", models.TournamentTeamStatusLocked, nil, []uuid.UUID{seedids.PlayerID(22), seedids.PlayerID(25), seedids.PlayerID(26), seedids.PlayerID(27), seedids.PlayerID(28)}},
		{"TM130007", "Tonkla Open Lobby", models.TournamentTeamStatusForming, strPtr("VALO26"), []uuid.UUID{seedids.TonklaID, seedids.PlayerID(29), seedids.PlayerID(30), seedids.PlayerID(31)}},
		{"TM130008", "Duplicate ID Squad", models.TournamentTeamStatusRejected, nil, []uuid.UUID{seedids.BeamID, seedids.PlayerID(32)}},
	} {
		var lockedAt *time.Time
		if spec.status != models.TournamentTeamStatusForming {
			lockedAt = ptrTime(now.Add(-time.Duration(72-idx*4) * time.Hour))
		}
		t := models.TournamentTeam{
			TournamentID: seedids.TournRealRegValorant,
			ID:           spec.id,
			Name:         spec.name,
			InviteCode:   spec.invite,
			Status:       spec.status,
			CreatedAt:    now.Add(-time.Duration(96-idx*4) * time.Hour),
			LockedAt:     lockedAt,
		}
		attachPayment(&t)
		addTeam(&teams, &members, t, spec.roster[0], spec.roster[1:]...)
	}

	// -------------------------------------------------------------------------
	// B4: TH Street Fighter 6 Community Clash (Free SOLO; Alex not joined)
	// -------------------------------------------------------------------------
	sf6Players := []struct {
		id     string
		name   string
		acc    uuid.UUID
		status models.TournamentTeamStatus
	}{
		{"SL140001", "Kaito", seedids.KaitoID, models.TournamentTeamStatusAccepted},
		{"SL140002", "Pleng", seedids.PlengID, models.TournamentTeamStatusAccepted},
		{"SL140003", "Tonkla", seedids.TonklaID, models.TournamentTeamStatusAccepted},
		{"SL140004", "Beam", seedids.BeamID, models.TournamentTeamStatusAccepted},
		{"SL140005", "Fah", seedids.FahID, models.TournamentTeamStatusAccepted},
		{"SL140006", "Mek", seedids.MekID, models.TournamentTeamStatusAccepted},
		{"SL140007", "Nong FrameTrap", seedids.PlayerID(1), models.TournamentTeamStatusAccepted},
		{"SL140008", "ParryKing BKK", seedids.PlayerID(2), models.TournamentTeamStatusAccepted},
		{"SL140009", "MBK Hadoken", seedids.PlayerID(3), models.TournamentTeamStatusAccepted},
		{"SL140010", "Ratchada Drive", seedids.PlayerID(4), models.TournamentTeamStatusAccepted},
		{"SL140011", "Siam Luke", seedids.PlayerID(5), models.TournamentTeamStatusAccepted},
		{"SL140012", "ChiangMai Cammy", seedids.PlayerID(6), models.TournamentTeamStatusAccepted},
		{"SL140013", "Asoke ChunLi", seedids.PlayerID(7), models.TournamentTeamStatusLocked},
		{"SL140014", "Thonglor Ryu", seedids.PlayerID(8), models.TournamentTeamStatusLocked},
		{"SL140015", "Unverified Entry", seedids.PlayerID(9), models.TournamentTeamStatusRejected},
	}
	for idx, p := range sf6Players {
		addTeam(&teams, &members, models.TournamentTeam{
			TournamentID: seedids.TournRealRegSF6Solo,
			ID:           p.id,
			Name:         p.name,
			Status:       p.status,
			CreatedAt:    now.Add(-time.Duration(120-idx*3) * time.Hour),
			LockedAt:     ptrTime(now.Add(-time.Duration(118-idx*3) * time.Hour)),
		}, p.acc)
	}

	// -------------------------------------------------------------------------
	// B5: ASEAN Dota 2 Invitational 2026 (20 ACCEPTED teams)
	// -------------------------------------------------------------------------
	dotaNames := []string{
		"Bangkok Guardians", "Manila Roshan", "SG Radiance", "KL Immortal", "Jakarta Tigers",
		"Saigon Storm", "Mekong Dire", "Phnom Penh Five", "Lanna Ancients", "Pattaya Tide",
		"Cebu Invokers", "Vientiane Void", "Yangon Spirit", "BKK Highground", "SEA Phoenix",
		"TDP Carry", "Brunei Guardians", "Johor Tempest", "Chiang Mai Rune", "Korat Wards",
	}
	dotaCaptains := []uuid.UUID{
		seedids.KaitoID, seedids.PlengID, seedids.TonklaID, seedids.BeamID,
		seedids.FahID, seedids.MekID,
	}
	for i := 0; i < 20; i++ {
		var captain uuid.UUID
		if i < len(dotaCaptains) {
			captain = dotaCaptains[i]
		} else {
			captain = seedids.PlayerID(i - len(dotaCaptains) + 1)
		}
		mate := seedids.PlayerID(15 + i)
		addTeam(&teams, &members, models.TournamentTeam{
			TournamentID: seedids.TournRealOngoingDota2,
			ID:           fmt.Sprintf("TM15%04d", i+1),
			Name:         dotaNames[i],
			Status:       models.TournamentTeamStatusAccepted,
			CreatedAt:    now.Add(-20 * 24 * time.Hour),
			LockedAt:     ptrTime(now.Add(-18 * 24 * time.Hour)),
		}, captain, mate)
	}

	// -------------------------------------------------------------------------
	// B6: Thailand Mobile Legends Championship S4 (8 ACCEPTED + 1 REJECTED)
	// -------------------------------------------------------------------------
	mlbbNames := []string{
		"Siam Titans", "Patong Dawn", "Korat Mythics", "Hatyai Surge",
		"Chao Phraya Five", "Nonthaburi Apex", "Ubon Legends", "Phuket Savages",
	}
	mlbbCaptains := []uuid.UUID{
		seedids.KaitoID, seedids.PlengID, seedids.TonklaID, seedids.BeamID,
		seedids.FahID, seedids.MekID, seedids.PlayerID(1), seedids.PlayerID(2),
	}
	for i := 0; i < 8; i++ {
		addTeam(&teams, &members, models.TournamentTeam{
			TournamentID: seedids.TournRealCompletedMLBB,
			ID:           fmt.Sprintf("TM16%04d", i+1),
			Name:         mlbbNames[i],
			Status:       models.TournamentTeamStatusAccepted,
			CreatedAt:    now.Add(-38 * 24 * time.Hour),
			LockedAt:     ptrTime(now.Add(-35 * 24 * time.Hour)),
		}, mlbbCaptains[i], seedids.PlayerID(4+i))
	}
	addTeam(&teams, &members, models.TournamentTeam{
		TournamentID: seedids.TournRealCompletedMLBB,
		ID:           "TM160009",
		Name:         "Late Roster Entry",
		Status:       models.TournamentTeamStatusRejected,
		CreatedAt:    now.Add(-32 * 24 * time.Hour),
		LockedAt:     ptrTime(now.Add(-31 * 24 * time.Hour)),
	}, seedids.PlayerID(12), seedids.PlayerID(13))

	return teams, members
}

func buildTeamReviews(now time.Time) []models.TournamentTeamReview {
	return []models.TournamentTeamReview{
		// A5 Review audit history
		{
			ID:                  seedids.ReviewID(1),
			TournamentID:        seedids.TournTestReviewQueue,
			TournamentTeamID:    "TM050007",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusAccepted,
			Reason:              "Verified payment and complete team intake forms.",
			ReviewerAccountID:   seedids.AlexID,
			ReviewerDisplayName: "Alex Dev",
			CreatedAt:           now.Add(-30 * time.Hour),
		},
		{
			ID:                  seedids.ReviewID(2),
			TournamentID:        seedids.TournTestReviewQueue,
			TournamentTeamID:    "TM050008",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusAccepted,
			Reason:              "Roster confirmed.",
			ReviewerAccountID:   seedids.MayaID,
			ReviewerDisplayName: "Maya",
			CreatedAt:           now.Add(-28 * time.Hour),
		},
		{
			ID:                  seedids.ReviewID(3),
			TournamentID:        seedids.TournTestReviewQueue,
			TournamentTeamID:    "TM050009",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusRejected,
			Reason:              "Incomplete scrim history and duplicate account.",
			ReviewerAccountID:   seedids.AlexID,
			ReviewerDisplayName: "Alex Dev",
			CreatedAt:           now.Add(-24 * time.Hour),
		},
		{
			ID:                  seedids.ReviewID(4),
			TournamentID:        seedids.TournTestReviewQueue,
			TournamentTeamID:    "TM050010",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusRejected,
			Reason:              "Roster eligibility check failed.",
			ReviewerAccountID:   seedids.MayaID,
			ReviewerDisplayName: "Maya",
			CreatedAt:           now.Add(-22 * time.Hour),
		},

		// A6 Solo Paid Review audit history
		{
			ID:                  seedids.ReviewID(5),
			TournamentID:        seedids.TournTestSoloPaid,
			TournamentTeamID:    "SL060001",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusAccepted,
			Reason:              "",
			ReviewerAccountID:   seedids.MayaID,
			ReviewerDisplayName: "Maya",
			CreatedAt:           now.Add(-40 * time.Hour),
		},
		{
			ID:                  seedids.ReviewID(6),
			TournamentID:        seedids.TournTestSoloPaid,
			TournamentTeamID:    "SL060002",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusAccepted,
			Reason:              "",
			ReviewerAccountID:   seedids.MayaID,
			ReviewerDisplayName: "Maya",
			CreatedAt:           now.Add(-39 * time.Hour),
		},
		{
			ID:                  seedids.ReviewID(7),
			TournamentID:        seedids.TournTestSoloPaid,
			TournamentTeamID:    "SL060003",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusAccepted,
			Reason:              "",
			ReviewerAccountID:   seedids.MayaID,
			ReviewerDisplayName: "Maya",
			CreatedAt:           now.Add(-38 * time.Hour),
		},
		{
			ID:                  seedids.ReviewID(8),
			TournamentID:        seedids.TournTestSoloPaid,
			TournamentTeamID:    "SL060004",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusAccepted,
			Reason:              "",
			ReviewerAccountID:   seedids.MayaID,
			ReviewerDisplayName: "Maya",
			CreatedAt:           now.Add(-37 * time.Hour),
		},
		{
			ID:                  seedids.ReviewID(9),
			TournamentID:        seedids.TournTestSoloPaid,
			TournamentTeamID:    "SL060007",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusRejected,
			Reason:              "Invalid CFN profile region.",
			ReviewerAccountID:   seedids.MayaID,
			ReviewerDisplayName: "Maya",
			CreatedAt:           now.Add(-35 * time.Hour),
		},

		// B3 Valorant audit history
		{
			ID:                  seedids.ReviewID(10),
			TournamentID:        seedids.TournRealRegValorant,
			TournamentTeamID:    "TM130001",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusAccepted,
			Reason:              "All 5 Riot IDs verified.",
			ReviewerAccountID:   seedids.AlexID,
			ReviewerDisplayName: "Alex Dev",
			CreatedAt:           now.Add(-48 * time.Hour),
		},
		{
			ID:                  seedids.ReviewID(11),
			TournamentID:        seedids.TournRealRegValorant,
			TournamentTeamID:    "TM130002",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusAccepted,
			Reason:              "Approved by tournament admin.",
			ReviewerAccountID:   seedids.BobID,
			ReviewerDisplayName: "Bob Dev",
			CreatedAt:           now.Add(-44 * time.Hour),
		},
		{
			ID:                  seedids.ReviewID(12),
			TournamentID:        seedids.TournRealRegValorant,
			TournamentTeamID:    "TM130008",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusRejected,
			Reason:              "Duplicate Riot ID detected across submitted roster.",
			ReviewerAccountID:   seedids.AlexID,
			ReviewerDisplayName: "Alex Dev",
			CreatedAt:           now.Add(-30 * time.Hour),
		},

		// B4 SF6 audit history
		{
			ID:                  seedids.ReviewID(13),
			TournamentID:        seedids.TournRealRegSF6Solo,
			TournamentTeamID:    "SL140015",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusRejected,
			Reason:              "Unverified CFN ID.",
			ReviewerAccountID:   seedids.BobID,
			ReviewerDisplayName: "Bob Dev",
			CreatedAt:           now.Add(-20 * time.Hour),
		},

		// B6 MLBB audit history
		{
			ID:                  seedids.ReviewID(14),
			TournamentID:        seedids.TournRealCompletedMLBB,
			TournamentTeamID:    "TM160009",
			FromStatus:          models.TournamentTeamStatusLocked,
			ToStatus:            models.TournamentTeamStatusRejected,
			Reason:              "Submitted after check-in cutoff.",
			ReviewerAccountID:   seedids.MayaID,
			ReviewerDisplayName: "Maya",
			CreatedAt:           now.Add(-30 * 24 * time.Hour),
		},
	}
}

func buildBrackets(now time.Time) ([]models.Match, []models.MatchEntry) {
	var matches []models.Match
	var entries []models.MatchEntry
	entrySeq := 1
	newEntry := func(matchID, tournID uuid.UUID, teamID string, score *string, advances bool, finalRank *int) models.MatchEntry {
		e := models.MatchEntry{
			ID:           seedids.MatchEntryID(entrySeq),
			MatchID:      matchID,
			TournamentID: tournID,
			TeamID:       teamID,
			Score:        score,
			Advances:     advances,
			FinalRank:    finalRank,
		}
		entrySeq++
		return e
	}

	// -------------------------------------------------------------------------
	// A9: [TEST · Permissions] Referee Role View (8 teams, 3 stages)
	// -------------------------------------------------------------------------
	tA9 := seedids.TournTestPermReferee
	matches = append(matches,
		models.Match{ID: seedids.MatchID(1), TournamentID: tA9, Name: "Grand Final", Venue: "Stage A", StageLevel: 3, TeamCount: 2, AdvancingCount: 1, ScheduledAt: ptrTime(now.Add(24 * time.Hour))},
		models.Match{ID: seedids.MatchID(2), TournamentID: tA9, Name: "Semifinal 1", Venue: "Stage A", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(1)), ScheduledAt: ptrTime(now.Add(14 * time.Hour))},
		models.Match{ID: seedids.MatchID(3), TournamentID: tA9, Name: "Semifinal 2", Venue: "Stage B", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(1)), ScheduledAt: ptrTime(now.Add(16 * time.Hour))},
		models.Match{ID: seedids.MatchID(4), TournamentID: tA9, Name: "Quarterfinal 1", Venue: "Lobby 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(2)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-4 * time.Hour))},
		models.Match{ID: seedids.MatchID(5), TournamentID: tA9, Name: "Quarterfinal 2", Venue: "Lobby 2", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(2)), ScheduledAt: ptrTime(now.Add(4 * time.Hour))},
		models.Match{ID: seedids.MatchID(6), TournamentID: tA9, Name: "Quarterfinal 3", Venue: "Lobby 3", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(3)), ScheduledAt: ptrTime(now.Add(6 * time.Hour))},
		models.Match{ID: seedids.MatchID(7), TournamentID: tA9, Name: "Quarterfinal 4", Venue: "Lobby 4", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(3)), ScheduledAt: ptrTime(now.Add(8 * time.Hour))},
	)
	entries = append(entries,
		newEntry(seedids.MatchID(4), tA9, "TM090001", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(4), tA9, "TM090002", strPtr("0"), false, nil),
		newEntry(seedids.MatchID(2), tA9, "TM090001", nil, false, nil),
		newEntry(seedids.MatchID(5), tA9, "TM090003", nil, false, nil),
		newEntry(seedids.MatchID(5), tA9, "TM090004", nil, false, nil),
		newEntry(seedids.MatchID(6), tA9, "TM090005", nil, false, nil),
		newEntry(seedids.MatchID(6), tA9, "TM090006", nil, false, nil),
		newEntry(seedids.MatchID(7), tA9, "TM090007", nil, false, nil),
		newEntry(seedids.MatchID(7), tA9, "TM090008", nil, false, nil),
	)

	// -------------------------------------------------------------------------
	// A10: [TEST · Bracket] Live Match & Swap (9 teams, 3 stages, uneven + stage skip)
	// Stage 1: Play-In Pod A (3->2 -> Stg2), Play-In Pod B (4->1 -> Stg2), Wildcard Duel (2->1 -> Stg3 skip!)
	// Stage 2: Upper Semifinal (3->2 -> Stg3)
	// Stage 3: Grand Final (3->1)
	// -------------------------------------------------------------------------
	tA10 := seedids.TournTestBracketLive
	matches = append(matches,
		models.Match{ID: seedids.MatchID(11), TournamentID: tA10, Name: "Grand Final Pod (3 -> 1)", Venue: "Main Stage", StageLevel: 3, TeamCount: 3, AdvancingCount: 1, ScheduledAt: ptrTime(now.Add(24 * time.Hour))},
		models.Match{ID: seedids.MatchID(12), TournamentID: tA10, Name: "Upper Semifinal (3 -> 2)", Venue: "Main Stage", StageLevel: 2, TeamCount: 3, AdvancingCount: 2, TargetMatchID: uuidPtr(seedids.MatchID(11)), ScheduledAt: ptrTime(now.Add(12 * time.Hour))},
		models.Match{ID: seedids.MatchID(13), TournamentID: tA10, Name: "Play-In Pod A (3 -> 2)", Venue: "Booth A", StageLevel: 1, TeamCount: 3, AdvancingCount: 2, TargetMatchID: uuidPtr(seedids.MatchID(12)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-8 * time.Hour))},
		models.Match{ID: seedids.MatchID(14), TournamentID: tA10, Name: "Play-In Pod B (4 -> 1)", Venue: "Booth B", StageLevel: 1, TeamCount: 4, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(12)), ScheduledAt: ptrTime(now.Add(4 * time.Hour))},
		models.Match{ID: seedids.MatchID(15), TournamentID: tA10, Name: "Wildcard Duel (Skips Stage 2)", Venue: "Booth C", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(11)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-6 * time.Hour))},
	)
	entries = append(entries,
		// Completed Play-In Pod A (3 -> 2)
		newEntry(seedids.MatchID(13), tA10, "TM100001", strPtr("18"), true, nil),
		newEntry(seedids.MatchID(13), tA10, "TM100002", strPtr("14"), true, nil),
		newEntry(seedids.MatchID(13), tA10, "TM100003", strPtr("6"), false, nil),
		// Completed Wildcard Duel skipping Stage 2 directly into Grand Final
		newEntry(seedids.MatchID(15), tA10, "TM100008", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(15), tA10, "TM100009", strPtr("0"), false, nil),
		// Seeded advanced entries waiting in Stage 2 and Stage 3
		newEntry(seedids.MatchID(12), tA10, "TM100001", nil, false, nil),
		newEntry(seedids.MatchID(12), tA10, "TM100002", nil, false, nil),
		newEntry(seedids.MatchID(11), tA10, "TM100008", nil, false, nil),
		// Unfinished 4-team Play-In Pod B (available for swaps/rescheduling/scoring)
		newEntry(seedids.MatchID(14), tA10, "TM100004", nil, false, nil),
		newEntry(seedids.MatchID(14), tA10, "TM100005", nil, false, nil),
		newEntry(seedids.MatchID(14), tA10, "TM100006", nil, false, nil),
		newEntry(seedids.MatchID(14), tA10, "TM100007", nil, false, nil),
	)

	// -------------------------------------------------------------------------
	// A11: [TEST · Bracket] Completed Archive (6 teams, 3 stages, uneven + stage skip)
	// Stage 1: Qualifying Pod (4->2 -> Stg2), Fast-Track Duel (2->1 -> Stg3 skip!)
	// Stage 2: Main Semifinal (2->1 -> Stg3)
	// Stage 3: Championship Final (2->1)
	// -------------------------------------------------------------------------
	tA11 := seedids.TournTestBracketCompleted
	matches = append(matches,
		models.Match{ID: seedids.MatchID(21), TournamentID: tA11, Name: "Championship Final (2 -> 1)", Venue: "Arena", StageLevel: 3, TeamCount: 2, AdvancingCount: 1, IsCompleted: true, ScheduledAt: ptrTime(now.Add(-12 * 24 * time.Hour))},
		models.Match{ID: seedids.MatchID(22), TournamentID: tA11, Name: "Main Semifinal (2 -> 1)", Venue: "Stage B", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(21)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-13 * 24 * time.Hour))},
		models.Match{ID: seedids.MatchID(23), TournamentID: tA11, Name: "Qualifying Pod (4 -> 2)", Venue: "Booth 1", StageLevel: 1, TeamCount: 4, AdvancingCount: 2, TargetMatchID: uuidPtr(seedids.MatchID(22)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-14 * 24 * time.Hour))},
		models.Match{ID: seedids.MatchID(24), TournamentID: tA11, Name: "Fast-Track Duel (Skips Stage 2)", Venue: "Booth 2", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(21)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-14 * 24 * time.Hour))},
	)
	entries = append(entries,
		newEntry(seedids.MatchID(23), tA11, "TM110001", strPtr("25"), true, nil),
		newEntry(seedids.MatchID(23), tA11, "TM110002", strPtr("19"), true, nil),
		newEntry(seedids.MatchID(23), tA11, "TM110003", strPtr("12"), false, intPtr(4)),
		newEntry(seedids.MatchID(23), tA11, "TM110004", strPtr("7"), false, intPtr(5)),
		newEntry(seedids.MatchID(24), tA11, "TM110005", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(24), tA11, "TM110006", strPtr("0"), false, intPtr(6)),
		newEntry(seedids.MatchID(22), tA11, "TM110001", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(22), tA11, "TM110002", strPtr("1"), false, intPtr(3)),
		newEntry(seedids.MatchID(21), tA11, "TM110001", strPtr("3"), false, intPtr(1)),
		newEntry(seedids.MatchID(21), tA11, "TM110005", strPtr("1"), false, intPtr(2)),
	)

	// -------------------------------------------------------------------------
	// B5: ASEAN Dota 2 Invitational 2026 (20 teams, 3 stages: 4×[5->2] -> 2×[4->2] -> [4->1])
	// MatchID 31 (GF 4->1), 32-33 (Heats 4->2), 34-37 (Groups A-D 5->2)
	// -------------------------------------------------------------------------
	tB5 := seedids.TournRealOngoingDota2
	matches = append(matches,
		models.Match{ID: seedids.MatchID(31), TournamentID: tB5, Name: "Grand Final Pod (4 -> 1)", Venue: "Main Stage A", StageLevel: 3, TeamCount: 4, AdvancingCount: 1, ScheduledAt: ptrTime(now.Add(48 * time.Hour))},
		models.Match{ID: seedids.MatchID(32), TournamentID: tB5, Name: "Playoff Heat 1 (4 -> 2)", Venue: "Main Stage A", StageLevel: 2, TeamCount: 4, AdvancingCount: 2, TargetMatchID: uuidPtr(seedids.MatchID(31)), ScheduledAt: ptrTime(now.Add(20 * time.Hour))},
		models.Match{ID: seedids.MatchID(33), TournamentID: tB5, Name: "Playoff Heat 2 (4 -> 2)", Venue: "Stream Booth 2", StageLevel: 2, TeamCount: 4, AdvancingCount: 2, TargetMatchID: uuidPtr(seedids.MatchID(31)), ScheduledAt: ptrTime(now.Add(24 * time.Hour))},
		models.Match{ID: seedids.MatchID(34), TournamentID: tB5, Name: "Group A (5 -> 2)", Venue: "Main Stage A", StageLevel: 1, TeamCount: 5, AdvancingCount: 2, TargetMatchID: uuidPtr(seedids.MatchID(32)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-14 * time.Hour))},
		models.Match{ID: seedids.MatchID(35), TournamentID: tB5, Name: "Group B (5 -> 2)", Venue: "Stream Booth 2", StageLevel: 1, TeamCount: 5, AdvancingCount: 2, TargetMatchID: uuidPtr(seedids.MatchID(32)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-10 * time.Hour))},
		models.Match{ID: seedids.MatchID(36), TournamentID: tB5, Name: "Group C (5 -> 2)", Venue: "Main Stage A", StageLevel: 1, TeamCount: 5, AdvancingCount: 2, TargetMatchID: uuidPtr(seedids.MatchID(33)), ScheduledAt: ptrTime(now.Add(4 * time.Hour))},
		models.Match{ID: seedids.MatchID(37), TournamentID: tB5, Name: "Group D (5 -> 2)", Venue: "Stream Booth 2", StageLevel: 1, TeamCount: 5, AdvancingCount: 2, TargetMatchID: uuidPtr(seedids.MatchID(33)), ScheduledAt: ptrTime(now.Add(7 * time.Hour))},
	)
	entries = append(entries,
		// Completed Group A (5 -> 2)
		newEntry(seedids.MatchID(34), tB5, "TM150001", strPtr("4-0"), true, nil),
		newEntry(seedids.MatchID(34), tB5, "TM150002", strPtr("3-1"), true, nil),
		newEntry(seedids.MatchID(34), tB5, "TM150003", strPtr("2-2"), false, nil),
		newEntry(seedids.MatchID(34), tB5, "TM150004", strPtr("1-3"), false, nil),
		newEntry(seedids.MatchID(34), tB5, "TM150005", strPtr("0-4"), false, nil),
		// Completed Group B (5 -> 2)
		newEntry(seedids.MatchID(35), tB5, "TM150006", strPtr("4-0"), true, nil),
		newEntry(seedids.MatchID(35), tB5, "TM150007", strPtr("3-1"), true, nil),
		newEntry(seedids.MatchID(35), tB5, "TM150008", strPtr("2-2"), false, nil),
		newEntry(seedids.MatchID(35), tB5, "TM150009", strPtr("1-3"), false, nil),
		newEntry(seedids.MatchID(35), tB5, "TM150010", strPtr("0-4"), false, nil),
		// Top 4 seeded into Playoff Heat 1 (4 -> 2)
		newEntry(seedids.MatchID(32), tB5, "TM150001", nil, false, nil),
		newEntry(seedids.MatchID(32), tB5, "TM150002", nil, false, nil),
		newEntry(seedids.MatchID(32), tB5, "TM150006", nil, false, nil),
		newEntry(seedids.MatchID(32), tB5, "TM150007", nil, false, nil),
		// Unfinished Group C (5 -> 2)
		newEntry(seedids.MatchID(36), tB5, "TM150011", nil, false, nil),
		newEntry(seedids.MatchID(36), tB5, "TM150012", nil, false, nil),
		newEntry(seedids.MatchID(36), tB5, "TM150013", nil, false, nil),
		newEntry(seedids.MatchID(36), tB5, "TM150014", nil, false, nil),
		newEntry(seedids.MatchID(36), tB5, "TM150015", nil, false, nil),
		// Unfinished Group D (5 -> 2)
		newEntry(seedids.MatchID(37), tB5, "TM150016", nil, false, nil),
		newEntry(seedids.MatchID(37), tB5, "TM150017", nil, false, nil),
		newEntry(seedids.MatchID(37), tB5, "TM150018", nil, false, nil),
		newEntry(seedids.MatchID(37), tB5, "TM150019", nil, false, nil),
		newEntry(seedids.MatchID(37), tB5, "TM150020", nil, false, nil),
	)

	// -------------------------------------------------------------------------
	// B6: Thailand Mobile Legends Championship S4 (8 teams, 3 stages, completed)
	// MatchID 51 (GF), 52-53 (SF), 54-57 (QF)
	// -------------------------------------------------------------------------
	tB6 := seedids.TournRealCompletedMLBB
	matches = append(matches,
		models.Match{ID: seedids.MatchID(51), TournamentID: tB6, Name: "Grand Final (Bo5)", Venue: "BCC Hall A", StageLevel: 3, TeamCount: 2, AdvancingCount: 1, IsCompleted: true, ScheduledAt: ptrTime(now.Add(-14 * 24 * time.Hour))},
		models.Match{ID: seedids.MatchID(52), TournamentID: tB6, Name: "Semifinal 1", Venue: "BCC Hall A", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(51)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-15 * 24 * time.Hour))},
		models.Match{ID: seedids.MatchID(53), TournamentID: tB6, Name: "Semifinal 2", Venue: "BCC Hall A", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(51)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-15 * 24 * time.Hour))},
		models.Match{ID: seedids.MatchID(54), TournamentID: tB6, Name: "Quarterfinal 1", Venue: "Stage 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(52)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-17 * 24 * time.Hour))},
		models.Match{ID: seedids.MatchID(55), TournamentID: tB6, Name: "Quarterfinal 2", Venue: "Stage 2", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(52)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-17 * 24 * time.Hour))},
		models.Match{ID: seedids.MatchID(56), TournamentID: tB6, Name: "Quarterfinal 3", Venue: "Stage 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(53)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-16 * 24 * time.Hour))},
		models.Match{ID: seedids.MatchID(57), TournamentID: tB6, Name: "Quarterfinal 4", Venue: "Stage 2", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetMatchID: uuidPtr(seedids.MatchID(53)), IsCompleted: true, ScheduledAt: ptrTime(now.Add(-16 * 24 * time.Hour))},
	)
	entries = append(entries,
		newEntry(seedids.MatchID(54), tB6, "TM160001", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(54), tB6, "TM160002", strPtr("0"), false, intPtr(5)),
		newEntry(seedids.MatchID(55), tB6, "TM160003", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(55), tB6, "TM160004", strPtr("1"), false, intPtr(6)),
		newEntry(seedids.MatchID(56), tB6, "TM160005", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(56), tB6, "TM160006", strPtr("0"), false, intPtr(7)),
		newEntry(seedids.MatchID(57), tB6, "TM160007", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(57), tB6, "TM160008", strPtr("1"), false, intPtr(8)),
		newEntry(seedids.MatchID(52), tB6, "TM160001", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(52), tB6, "TM160003", strPtr("1"), false, intPtr(3)),
		newEntry(seedids.MatchID(53), tB6, "TM160005", strPtr("2"), true, nil),
		newEntry(seedids.MatchID(53), tB6, "TM160007", strPtr("0"), false, intPtr(4)),
		newEntry(seedids.MatchID(51), tB6, "TM160001", strPtr("3"), false, intPtr(1)),
		newEntry(seedids.MatchID(51), tB6, "TM160005", strPtr("1"), false, intPtr(2)),
	)

	return matches, entries
}
