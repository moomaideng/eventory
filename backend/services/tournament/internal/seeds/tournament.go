package seeds

import (
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/seedids"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedTournaments populates PostgreSQL with 17 deterministic showcase tournaments
// (11 feature-test fixtures and 6 realistic multi-stage tournaments), along with
// staff assignments, sponsorship tiers, pledges, and announcements.
func SeedTournaments(db *gorm.DB) error {
	now := time.Now().UTC().Truncate(time.Second)

	tournaments := buildTournaments(now)
	for i := range tournaments {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(&tournaments[i]).Error; err != nil {
			return err
		}
	}

	staff := buildStaff(now)
	for i := range staff {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tournament_id"}, {Name: "account_id"}},
			UpdateAll: true,
		}).Create(&staff[i]).Error; err != nil {
			return err
		}
	}

	tiers := buildSponsorshipTiers()
	for i := range tiers {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(&tiers[i]).Error; err != nil {
			return err
		}
	}

	pledges := buildSponsorshipPledges()
	for i := range pledges {
		// Preserve state changes if the developer completes or cancels a seeded pledge.
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&pledges[i]).Error; err != nil {
			return err
		}
	}

	announcements := buildAnnouncements(now)
	for i := range announcements {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(&announcements[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

func ptrTime(t time.Time) *time.Time { return &t }

func buildTournaments(now time.Time) []models.Tournament {
	return []models.Tournament{
		// ---------------------------------------------------------------------
		// Group A: Feature-Test Showcase Tournaments ([TEST · ...])
		// ---------------------------------------------------------------------
		{
			ID:                   seedids.TournTestDraft,
			Name:                 "[TEST · Draft] Setup & Publish Gate",
			Description:          "Draft sandbox with empty form and zero tiers for testing tournament editing, form creation, tier setup, and DRAFT transitions.",
			Detail:               "### Purpose\nUse this tournament to test:\n- Editing basic metadata while in `DRAFT`\n- Adding sponsorship tiers & custom MongoDB form questions\n- Publishing to `CROWDFUNDING` or directly to `REGISTRATION`",
			Game:                 "Valorant",
			Location:             "Online (SEA)",
			StartsAt:             now.Add(30 * 24 * time.Hour),
			EndsAt:               now.Add(32 * 24 * time.Hour),
			RegistrationDeadline: now.Add(20 * 24 * time.Hour),
			EntryFeeMinor:        0,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          2,
			MaxTeamSize:          5,
			Capacity:             16,
			Status:               models.TournamentStatusDraft,
			CreatedAt:            now.Add(-2 * 24 * time.Hour),
			UpdatedAt:            now.Add(-1 * time.Hour),
		},
		{
			ID:                   seedids.TournTestCrowdGoalReached,
			Name:                 "[TEST · Crowdfunding] Goal Reached (Normal Open)",
			Description:          "Crowdfunding goal already exceeded (35,000 / 30,000 THB) to verify standard CROWDFUNDING -> REGISTRATION transition without force.",
			Detail:               "### Crowdfunding Test\nTwo confirmed sponsorships total **35,000 THB** against a **30,000 THB** goal. Opening registration succeeds cleanly via the standard transition.",
			Game:                 "Counter-Strike 2",
			Location:             "Bangkok Arena",
			StartsAt:             now.Add(35 * 24 * time.Hour),
			EndsAt:               now.Add(37 * 24 * time.Hour),
			RegistrationDeadline: now.Add(25 * 24 * time.Hour),
			EntryFeeMinor:        50000,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          3,
			MaxTeamSize:          5,
			Capacity:             16,
			Status:               models.TournamentStatusCrowdfunding,
			Crowdfunding: models.CrowdfundingInfo{
				GoalAmountMinor: 3000000,
				Deadline:        ptrTime(now.Add(10 * 24 * time.Hour)),
			},
			CreatedAt: now.Add(-5 * 24 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:                   seedids.TournTestCrowdUnderfunded,
			Name:                 "[TEST · Crowdfunding] Underfunded & Tier Locks",
			Description:          "Tests ErrCrowdfundingGoalNotReached vs force-open, plus sold-out/pending locked tiers alongside untouched editable/deletable tiers.",
			Detail:               "### Tier & Goal Edge Cases\n- **Gold (Sold Out):** 1/1 confirmed (`Carter Fundings`)\n- **Silver (Locked):** 1 pending checkout (`Maya Works Studio`)\n- **Community (Editable):** only expired/cancelled attempts\n- **Supporter (Fresh):** zero pledges, safe to edit or delete",
			Game:                 "Street Fighter 6",
			Location:             "Online (Thailand)",
			StartsAt:             now.Add(35 * 24 * time.Hour),
			EndsAt:               now.Add(37 * 24 * time.Hour),
			RegistrationDeadline: now.Add(25 * 24 * time.Hour),
			EntryFeeMinor:        20000,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeSolo,
			MinTeamSize:          1,
			MaxTeamSize:          1,
			Capacity:             32,
			Status:               models.TournamentStatusCrowdfunding,
			Crowdfunding: models.CrowdfundingInfo{
				GoalAmountMinor: 5000000,
				Deadline:        ptrTime(now.Add(10 * 24 * time.Hour)),
			},
			CreatedAt: now.Add(-6 * 24 * time.Hour),
			UpdatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID:                   seedids.TournTestSponsorCheckout,
			Name:                 "[TEST · Sponsor] Checkout Playground",
			Description:          "Maya-owned crowdfunding sandbox with open tiers so Alex Dev can immediately run Sponsor Checkout and Mock Payment.",
			Detail:               "### Sponsor Checkout Playground\nOwned by **Maya Works** so `Alex Dev` has no ownership conflict and no existing active pledge. Switch to Sponsor mode or browse this page to sponsor any open tier.",
			Game:                 "Valorant",
			Location:             "True Digital Park, Bangkok",
			StartsAt:             now.Add(40 * 24 * time.Hour),
			EndsAt:               now.Add(41 * 24 * time.Hour),
			RegistrationDeadline: now.Add(28 * 24 * time.Hour),
			EntryFeeMinor:        0,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeSolo,
			MinTeamSize:          1,
			MaxTeamSize:          1,
			Capacity:             32,
			Status:               models.TournamentStatusCrowdfunding,
			Crowdfunding: models.CrowdfundingInfo{
				GoalAmountMinor: 4000000,
				Deadline:        ptrTime(now.Add(14 * 24 * time.Hour)),
			},
			CreatedAt: now.Add(-4 * 24 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour),
		},
		{
			ID:                   seedids.TournTestReviewQueue,
			Name:                 "[TEST · Review] Queue, Forms & Refunds",
			Description:          "Paid team tournament with FORMING, PAYMENT_PENDING, LOCKED (paid), ACCEPTED, and REJECTED (refunded) entries plus a 5-type custom form.",
			Detail:               "### Review Queue & Refund Playground\n- Join lobby **`TEST01`** or inspect incomplete form blocking on **`TM050002`**\n- Rejecting a paid `LOCKED` team (`TM050004`+) triggers a real refund in Payment Service\n- Includes pre-seeded `TournamentTeamReview` audit trail rows",
			Game:                 "Valorant",
			Location:             "Online (SEA)",
			StartsAt:             now.Add(25 * 24 * time.Hour),
			EndsAt:               now.Add(27 * 24 * time.Hour),
			RegistrationDeadline: now.Add(14 * 24 * time.Hour),
			EntryFeeMinor:        50000,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          2,
			MaxTeamSize:          3,
			Capacity:             8,
			Status:               models.TournamentStatusRegistration,
			CreatedAt:            now.Add(-8 * 24 * time.Hour),
			UpdatedAt:            now.Add(-2 * time.Hour),
		},
		{
			ID:                   seedids.TournTestSoloPaid,
			Name:                 "[TEST · Solo] Paid Entry & Form Playground",
			Description:          "Maya-owned paid solo registration with all 5 custom form field types. Alex Dev is unregistered and ready to apply and pay.",
			Detail:               "### Solo Paid Registration Test\nLog in as **Alex Dev** to fill out the 5-field custom registration form and complete the **250 THB** entry-fee checkout.",
			Game:                 "Street Fighter 6",
			Location:             "MBK Center, Bangkok",
			StartsAt:             now.Add(26 * 24 * time.Hour),
			EndsAt:               now.Add(27 * 24 * time.Hour),
			RegistrationDeadline: now.Add(15 * 24 * time.Hour),
			EntryFeeMinor:        25000,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeSolo,
			MinTeamSize:          1,
			MaxTeamSize:          1,
			Capacity:             32,
			Status:               models.TournamentStatusRegistration,
			Crowdfunding: models.CrowdfundingInfo{
				GoalAmountMinor: 1000000,
				Deadline:        ptrTime(now.Add(-5 * 24 * time.Hour)),
			},
			CreatedAt: now.Add(-12 * 24 * time.Hour),
			UpdatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID:                   seedids.TournTestBracketSetup,
			Name:                 "[TEST · Bracket] Setup & Start Gate",
			Description:          "16 accepted teams, zero unresolved applications, and a past registration deadline — ready for bracket generation and launching ONGOING.",
			Detail:               "### Bracket Builder & Start Gate\n- **16 `ACCEPTED` teams**, 0 `FORMING`/`LOCKED` teams\n- Registration deadline has already elapsed\n- Use the **Bracket** tab to build stages, seed teams, schedule matches, and click **Start Tournament**",
			Game:                 "Dota 2",
			Location:             "Bangkok LAN Studio",
			StartsAt:             now.Add(2 * 24 * time.Hour),
			EndsAt:               now.Add(4 * 24 * time.Hour),
			RegistrationDeadline: now.Add(-1 * 24 * time.Hour),
			EntryFeeMinor:        0,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          1,
			MaxTeamSize:          3,
			Capacity:             16,
			Status:               models.TournamentStatusRegistration,
			CreatedAt:            now.Add(-14 * 24 * time.Hour),
			UpdatedAt:            now.Add(-4 * time.Hour),
		},
		{
			ID:                   seedids.TournTestPermAdmin,
			Name:                 "[TEST · Permissions] Admin Role View",
			Description:          "Owned by Bob Events with Alex Dev assigned as ADMIN and Maya as REFEREE to verify non-owner Admin permissions.",
			Detail:               "### Admin Role Matrix\nAs **Alex Dev** (`ADMIN`), you can review teams, manage brackets, and post announcements, while Owner-only tabs (`Form`, `Sponsorship`, `Staff Invite/Remove`, `Edit Details`) remain restricted.",
			Game:                 "Counter-Strike 2",
			Location:             "Online (Bangkok)",
			StartsAt:             now.Add(22 * 24 * time.Hour),
			EndsAt:               now.Add(24 * 24 * time.Hour),
			RegistrationDeadline: now.Add(12 * 24 * time.Hour),
			EntryFeeMinor:        0,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          2,
			MaxTeamSize:          4,
			Capacity:             12,
			Status:               models.TournamentStatusRegistration,
			CreatedAt:            now.Add(-7 * 24 * time.Hour),
			UpdatedAt:            now.Add(-2 * time.Hour),
		},
		{
			ID:                   seedids.TournTestPermReferee,
			Name:                 "[TEST · Permissions] Referee Role View",
			Description:          "Ongoing 8-team tournament owned by Bob Events with Alex Dev assigned as REFEREE for match scheduling, swapping, and score reporting.",
			Detail:               "### Referee Role Matrix\nAs **Alex Dev** (`REFEREE`), verify that Home is read-only and only Bracket operations (scheduling, swapping unfinished slots, reporting scores) plus Leave Tournament in Staff are available.",
			Game:                 "Valorant",
			Location:             "Thailand Pro Studio",
			StartsAt:             now.Add(-1 * 24 * time.Hour),
			EndsAt:               now.Add(3 * 24 * time.Hour),
			RegistrationDeadline: now.Add(-5 * 24 * time.Hour),
			EntryFeeMinor:        0,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          2,
			MaxTeamSize:          3,
			Capacity:             8,
			Status:               models.TournamentStatusOngoing,
			CreatedAt:            now.Add(-15 * 24 * time.Hour),
			UpdatedAt:            now.Add(-6 * time.Hour),
		},
		{
			ID:                   seedids.TournTestBracketLive,
			Name:                 "[TEST · Bracket] Live Match & Swap",
			Description:          "Ongoing 9-team uneven stage-skipping bracket (3->2, 4->1, plus a 2->1 Wildcard Duel skipping Stage 2) for testing multi-team scores, swaps, and completion.",
			Detail:               "### Live Uneven & Stage-Skipping Bracket\n- **Stage 1:** `Play-In Pod A (3->2)` (completed, 2 advanced to Stage 2), `Play-In Pod B (4->1)` (unfinished, swap/score here), and `Wildcard Duel (2->1)` (**skips Stage 2** directly into the Grand Final!)\n- **Stage 2:** `Upper Semifinal (3->2)`\n- **Stage 3:** `Grand Final (3->1)`\n- Also has an open tier for late sponsorships during `ONGOING`",
			Game:                 "Valorant",
			Location:             "Garena Esports Studio, Bangkok",
			StartsAt:             now.Add(-2 * 24 * time.Hour),
			EndsAt:               now.Add(4 * 24 * time.Hour),
			RegistrationDeadline: now.Add(-6 * 24 * time.Hour),
			EntryFeeMinor:        0,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          1,
			MaxTeamSize:          3,
			Capacity:             9,
			Status:               models.TournamentStatusOngoing,
			CreatedAt:            now.Add(-16 * 24 * time.Hour),
			UpdatedAt:            now.Add(-3 * time.Hour),
		},
		{
			ID:                   seedids.TournTestBracketCompleted,
			Name:                 "[TEST · Bracket] Completed Archive",
			Description:          "Completed 6-team uneven stage-skipping bracket (4->2 Pod + 2->1 Fast-Track skipping Stage 2 into the 2->1 Championship Final) with final ranks.",
			Detail:               "### Completed Archive\nAll matches have concluded (`4->2` -> `2->1`, plus a `2->1` Fast-Track skipping Stage 2) with `FinalRank` 1–6 recorded. Mutations to teams, brackets, forms, and sponsorships are locked.",
			Game:                 "Mobile Legends",
			Location:             "Online (Thailand)",
			StartsAt:             now.Add(-15 * 24 * time.Hour),
			EndsAt:               now.Add(-12 * 24 * time.Hour),
			RegistrationDeadline: now.Add(-20 * 24 * time.Hour),
			EntryFeeMinor:        0,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          1,
			MaxTeamSize:          3,
			Capacity:             6,
			Status:               models.TournamentStatusCompleted,
			CreatedAt:            now.Add(-30 * 24 * time.Hour),
			UpdatedAt:            now.Add(-12 * 24 * time.Hour),
		},

		// ---------------------------------------------------------------------
		// Group B: "Real-Looking" Tournaments Across Every Stage
		// ---------------------------------------------------------------------
		{
			ID:                   seedids.TournRealDraftCU,
			Name:                 "Chula Inter-Faculty Valorant Cup 2026",
			Description:          "Annual university Valorant championship across Chulalongkorn University faculties, complete with pre-configured registration form and sponsor packages.",
			Detail:               "### Event Overview\nOpen to all active undergraduate and graduate rosters representing their faculty.\n\n### Format & Prizes\n- **Format:** 16 teams, Best-of-3 Single Elimination (Grand Finals Bo5)\n- **Prize Pool:** 45,000 THB funded via campus & tech sponsors\n- **LAN Finals:** CU Engineering Auditorium",
			Game:                 "Valorant",
			Location:             "CU Engineering Auditorium, Bangkok",
			StartsAt:             now.Add(45 * 24 * time.Hour),
			EndsAt:               now.Add(48 * 24 * time.Hour),
			RegistrationDeadline: now.Add(30 * 24 * time.Hour),
			EntryFeeMinor:        100000,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          5,
			MaxTeamSize:          5,
			Capacity:             16,
			Status:               models.TournamentStatusDraft,
			Crowdfunding: models.CrowdfundingInfo{
				GoalAmountMinor: 4500000,
				Deadline:        ptrTime(now.Add(15 * 24 * time.Hour)),
			},
			CreatedAt: now.Add(-3 * 24 * time.Hour),
			UpdatedAt: now.Add(-4 * time.Hour),
		},
		{
			ID:                   seedids.TournRealCrowdBCA,
			Name:                 "Bangkok Cyber Arena 2026",
			Description:          "Premier Thailand Counter-Strike 2 LAN championship seeking community and hardware sponsors for a 100,000 THB main-stage production.",
			Detail:               "### Event Overview\nHosted by **Bob Events** at Samyan Mitrtown Hall, bringing together Thailand's top 16 CS2 squads.\n\n### Production & Streaming\n- 120Hz stage monitors & dedicated broadcast desk\n- Live bilingual commentary (TH/EN)",
			Game:                 "Counter-Strike 2",
			Location:             "Samyan Mitrtown Hall, Bangkok",
			StartsAt:             now.Add(42 * 24 * time.Hour),
			EndsAt:               now.Add(45 * 24 * time.Hour),
			RegistrationDeadline: now.Add(28 * 24 * time.Hour),
			EntryFeeMinor:        150000,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          5,
			MaxTeamSize:          5,
			Capacity:             16,
			Status:               models.TournamentStatusCrowdfunding,
			Crowdfunding: models.CrowdfundingInfo{
				GoalAmountMinor: 10000000,
				Deadline:        ptrTime(now.Add(14 * 24 * time.Hour)),
			},
			CreatedAt: now.Add(-10 * 24 * time.Hour),
			UpdatedAt: now.Add(-5 * time.Hour),
		},
		{
			ID:                   seedids.TournRealRegValorant,
			Name:                 "SEA Valorant Challengers Open",
			Description:          "Regional 5v5 qualifier on Singapore & Bangkok servers with verified Riot ID intake, team lobbies, and active sponsor backing.",
			Detail:               "### Qualifier Rules\n- **Check-in:** 45 minutes prior to Round 1\n- **Server Preference:** Singapore / Bangkok coin-flip veto\n- **Roster Lock:** All 5 players must complete the Valorant intake form before locking",
			Game:                 "Valorant",
			Location:             "Online (SG / BKK Servers)",
			StartsAt:             now.Add(28 * 24 * time.Hour),
			EndsAt:               now.Add(31 * 24 * time.Hour),
			RegistrationDeadline: now.Add(18 * 24 * time.Hour),
			EntryFeeMinor:        80000,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          5,
			MaxTeamSize:          5,
			Capacity:             16,
			Status:               models.TournamentStatusRegistration,
			Crowdfunding: models.CrowdfundingInfo{
				GoalAmountMinor: 2000000,
				Deadline:        ptrTime(now.Add(-3 * 24 * time.Hour)),
			},
			CreatedAt: now.Add(-18 * 24 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:                   seedids.TournRealRegSF6Solo,
			Name:                 "TH Street Fighter 6 Community Clash",
			Description:          "Free-to-enter Bangkok FGC offline monthly for up to 64 solo competitors at MBK Center Esports Zone.",
			Detail:               "### FGC Monthly Ruleset\n- **Format:** FT2 until Top 8, FT3 for Winners/Losers/Grand Finals\n- **Hardware:** PS5 stations (BYO arcade stick / leverless / pad)\n- **Entry Fee:** Free (community-supported)",
			Game:                 "Street Fighter 6",
			Location:             "MBK Center Esports Zone, Bangkok",
			StartsAt:             now.Add(21 * 24 * time.Hour),
			EndsAt:               now.Add(22 * 24 * time.Hour),
			RegistrationDeadline: now.Add(14 * 24 * time.Hour),
			EntryFeeMinor:        0,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeSolo,
			MinTeamSize:          1,
			MaxTeamSize:          1,
			Capacity:             64,
			Status:               models.TournamentStatusRegistration,
			CreatedAt:            now.Add(-9 * 24 * time.Hour),
			UpdatedAt:            now.Add(-1 * time.Hour),
		},
		{
			ID:                   seedids.TournRealOngoingDota2,
			Name:                 "ASEAN Dota 2 Invitational 2026",
			Description:          "20-team SEA LAN Invitational live at True Digital Park — 4×5-team Groups feeding two 4-team Playoff Heats and a 4-team Grand Final Pod.",
			Detail:               "### Live Coverage\n- **20 Teams · Multi-Pod Bracket:** Stage 1 `5->2` (×4 Groups) → Stage 2 `4->2` (×2 Heats) → Stage 3 `4->1` Grand Final Pod\n- Groups A & B have concluded; Groups C & D run live on Main Stage A & Stream Booth 2",
			Game:                 "Dota 2",
			Location:             "True Digital Park, Bangkok",
			StartsAt:             now.Add(-2 * 24 * time.Hour),
			EndsAt:               now.Add(4 * 24 * time.Hour),
			RegistrationDeadline: now.Add(-5 * 24 * time.Hour),
			EntryFeeMinor:        200000,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          5,
			MaxTeamSize:          5,
			Capacity:             20,
			Status:               models.TournamentStatusOngoing,
			Crowdfunding: models.CrowdfundingInfo{
				GoalAmountMinor: 8000000,
				Deadline:        ptrTime(now.Add(-15 * 24 * time.Hour)),
			},
			CreatedAt: now.Add(-35 * 24 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour),
		},
		{
			ID:                   seedids.TournRealCompletedMLBB,
			Name:                 "Thailand Mobile Legends Championship S4",
			Description:          "Completed Season 4 8-team national Mobile Legends: Bang Bang finals with full bracket archive, scores, and sponsor roster.",
			Detail:               "### Season 4 Archive\nConcluded in front of a packed Bangkok Convention Center crowd with **Siam Titans** taking the 3-1 Grand Final victory.",
			Game:                 "Mobile Legends",
			Location:             "Bangkok Convention Center",
			StartsAt:             now.Add(-18 * 24 * time.Hour),
			EndsAt:               now.Add(-14 * 24 * time.Hour),
			RegistrationDeadline: now.Add(-25 * 24 * time.Hour),
			EntryFeeMinor:        50000,
			Currency:             "THB",
			RegistrationMode:     models.TournamentRegistrationModeTeam,
			MinTeamSize:          5,
			MaxTeamSize:          5,
			Capacity:             8,
			Status:               models.TournamentStatusCompleted,
			Crowdfunding: models.CrowdfundingInfo{
				GoalAmountMinor: 4000000,
				Deadline:        ptrTime(now.Add(-40 * 24 * time.Hour)),
			},
			CreatedAt: now.Add(-55 * 24 * time.Hour),
			UpdatedAt: now.Add(-14 * 24 * time.Hour),
		},
	}
}

func buildStaff(now time.Time) []models.TournamentStaff {
	staffEntry := func(tournID, accID uuid.UUID, role models.TournamentStaffRole) models.TournamentStaff {
		return models.TournamentStaff{
			TournamentID: tournID,
			AccountID:    accID,
			Role:         role,
			CreatedAt:    now.Add(-5 * 24 * time.Hour),
			UpdatedAt:    now.Add(-5 * 24 * time.Hour),
		}
	}

	// Alex has exactly 12 staff memberships (9 OWNER, 2 ADMIN, 1 REFEREE), fitting cleanly on Page 1 of /organizer.
	return []models.TournamentStaff{
		// A1
		staffEntry(seedids.TournTestDraft, seedids.AlexID, models.TournamentStaffRoleOwner),
		// A2
		staffEntry(seedids.TournTestCrowdGoalReached, seedids.AlexID, models.TournamentStaffRoleOwner),
		// A3
		staffEntry(seedids.TournTestCrowdUnderfunded, seedids.AlexID, models.TournamentStaffRoleOwner),
		// A4
		staffEntry(seedids.TournTestSponsorCheckout, seedids.MayaID, models.TournamentStaffRoleOwner),
		// A5
		staffEntry(seedids.TournTestReviewQueue, seedids.AlexID, models.TournamentStaffRoleOwner),
		staffEntry(seedids.TournTestReviewQueue, seedids.MayaID, models.TournamentStaffRoleAdmin),
		staffEntry(seedids.TournTestReviewQueue, seedids.PlayerID(1), models.TournamentStaffRoleReferee),
		// A6
		staffEntry(seedids.TournTestSoloPaid, seedids.MayaID, models.TournamentStaffRoleOwner),
		// A7
		staffEntry(seedids.TournTestBracketSetup, seedids.AlexID, models.TournamentStaffRoleOwner),
		// A8
		staffEntry(seedids.TournTestPermAdmin, seedids.BobID, models.TournamentStaffRoleOwner),
		staffEntry(seedids.TournTestPermAdmin, seedids.AlexID, models.TournamentStaffRoleAdmin),
		staffEntry(seedids.TournTestPermAdmin, seedids.MayaID, models.TournamentStaffRoleReferee),
		// A9
		staffEntry(seedids.TournTestPermReferee, seedids.BobID, models.TournamentStaffRoleOwner),
		staffEntry(seedids.TournTestPermReferee, seedids.MayaID, models.TournamentStaffRoleAdmin),
		staffEntry(seedids.TournTestPermReferee, seedids.AlexID, models.TournamentStaffRoleReferee),
		// A10
		staffEntry(seedids.TournTestBracketLive, seedids.AlexID, models.TournamentStaffRoleOwner),
		// A11
		staffEntry(seedids.TournTestBracketCompleted, seedids.AlexID, models.TournamentStaffRoleOwner),

		// B1
		staffEntry(seedids.TournRealDraftCU, seedids.AlexID, models.TournamentStaffRoleOwner),
		// B2
		staffEntry(seedids.TournRealCrowdBCA, seedids.BobID, models.TournamentStaffRoleOwner),
		staffEntry(seedids.TournRealCrowdBCA, seedids.PlayerID(1), models.TournamentStaffRoleAdmin),
		// B3
		staffEntry(seedids.TournRealRegValorant, seedids.AlexID, models.TournamentStaffRoleOwner),
		staffEntry(seedids.TournRealRegValorant, seedids.BobID, models.TournamentStaffRoleAdmin),
		staffEntry(seedids.TournRealRegValorant, seedids.PlayerID(2), models.TournamentStaffRoleReferee),
		// B4
		staffEntry(seedids.TournRealRegSF6Solo, seedids.BobID, models.TournamentStaffRoleOwner),
		// B5
		staffEntry(seedids.TournRealOngoingDota2, seedids.BobID, models.TournamentStaffRoleOwner),
		staffEntry(seedids.TournRealOngoingDota2, seedids.AlexID, models.TournamentStaffRoleAdmin),
		staffEntry(seedids.TournRealOngoingDota2, seedids.MayaID, models.TournamentStaffRoleReferee),
		// B6
		staffEntry(seedids.TournRealCompletedMLBB, seedids.MayaID, models.TournamentStaffRoleOwner),
		staffEntry(seedids.TournRealCompletedMLBB, seedids.PlayerID(3), models.TournamentStaffRoleAdmin),
	}
}

func buildSponsorshipTiers() []models.SponsorshipTier {
	return []models.SponsorshipTier{
		// A2: Goal Reached
		{ID: seedids.TierID(1), TournamentID: seedids.TournTestCrowdGoalReached, Name: "Platinum", PriceMinor: 2000000, Slots: 1, Perks: []string{"Main stage branding", "Logo on live overlay", "VIP passes"}},
		{ID: seedids.TierID(2), TournamentID: seedids.TournTestCrowdGoalReached, Name: "Gold", PriceMinor: 1500000, Slots: 2, Perks: []string{"Broadcast break shoutout", "Website logo placement"}},
		{ID: seedids.TierID(3), TournamentID: seedids.TournTestCrowdGoalReached, Name: "Silver", PriceMinor: 500000, Slots: 3, Perks: []string{"Social media thank-you", "Bracket footer badge"}},

		// A3: Underfunded & Tier Locks
		{ID: seedids.TierID(4), TournamentID: seedids.TournTestCrowdUnderfunded, Name: "Gold (Sold Out)", PriceMinor: 1500000, Slots: 1, Perks: []string{"Locked & sold out with 1/1 confirmed pledge"}},
		{ID: seedids.TierID(5), TournamentID: seedids.TournTestCrowdUnderfunded, Name: "Silver (Locked)", PriceMinor: 500000, Slots: 3, Perks: []string{"Locked by active pending pledge"}},
		{ID: seedids.TierID(6), TournamentID: seedids.TournTestCrowdUnderfunded, Name: "Community (Editable)", PriceMinor: 100000, Slots: 5, Perks: []string{"Only historical expired/cancelled pledges"}},
		{ID: seedids.TierID(7), TournamentID: seedids.TournTestCrowdUnderfunded, Name: "Supporter (Fresh)", PriceMinor: 250000, Slots: 4, Perks: []string{"Zero pledges — safe to edit or delete"}},

		// A4: Sponsor Checkout Playground
		{ID: seedids.TierID(8), TournamentID: seedids.TournTestSponsorCheckout, Name: "Title Partner", PriceMinor: 3000000, Slots: 1, Perks: []string{"Naming rights", "On-stage trophy handover", "Full overlay suite"}},
		{ID: seedids.TierID(9), TournamentID: seedids.TournTestSponsorCheckout, Name: "Gold", PriceMinor: 1500000, Slots: 2, Perks: []string{"Stream segment sponsor", "Logo on player jerseys"}},
		{ID: seedids.TierID(10), TournamentID: seedids.TournTestSponsorCheckout, Name: "Silver", PriceMinor: 500000, Slots: 3, Perks: []string{"Lobby banner", "Discord partner role"}},
		{ID: seedids.TierID(11), TournamentID: seedids.TournTestSponsorCheckout, Name: "Bronze", PriceMinor: 100000, Slots: 5, Perks: []string{"Supporter wall credit", "Bracket listing"}},

		// A6: Solo Paid Supporter Tier
		{ID: seedids.TierID(12), TournamentID: seedids.TournTestSoloPaid, Name: "Community Supporter", PriceMinor: 300000, Slots: 5, Perks: []string{"Top-8 stream shoutout", "Partner badge on tournament page"}},

		// A10: Live Match Late Sponsorship Tier
		{ID: seedids.TierID(13), TournamentID: seedids.TournTestBracketLive, Name: "Late Broadcast Sponsor", PriceMinor: 500000, Slots: 2, Perks: []string{"Finals broadcast slate", "Winner graphic logo"}},

		// B1: Draft CU Cup Tiers
		{ID: seedids.TierID(14), TournamentID: seedids.TournRealDraftCU, Name: "Faculty Title Partner", PriceMinor: 3000000, Slots: 1, Perks: []string{"Main auditorium stage naming", "Opening ceremony keynote slot"}},
		{ID: seedids.TierID(15), TournamentID: seedids.TournRealDraftCU, Name: "Stage Sponsor", PriceMinor: 1500000, Slots: 2, Perks: []string{"Booth space at CU courtyard", "Stream break video reel"}},
		{ID: seedids.TierID(16), TournamentID: seedids.TournRealDraftCU, Name: "Booth Partner", PriceMinor: 500000, Slots: 4, Perks: []string{"Campus booth table", "Tournament page logo"}},

		// B2: Bangkok Cyber Arena 2026
		{ID: seedids.TierID(17), TournamentID: seedids.TournRealCrowdBCA, Name: "Title Partner", PriceMinor: 3500000, Slots: 1, Perks: []string{"Co-branded trophy & main stage LED", "All-access VIP hospitality"}},
		{ID: seedids.TierID(18), TournamentID: seedids.TournRealCrowdBCA, Name: "Hardware Partner", PriceMinor: 2000000, Slots: 2, Perks: []string{"Stage PC & monitor branding", "Product demo booth"}},
		{ID: seedids.TierID(19), TournamentID: seedids.TournRealCrowdBCA, Name: "Server Partner", PriceMinor: 1000000, Slots: 4, Perks: []string{"Match server MOTD credit", "Broadcast lower-third logo"}},

		// B3: SEA Valorant Challengers Open
		{ID: seedids.TierID(20), TournamentID: seedids.TournRealRegValorant, Name: "Gold", PriceMinor: 1500000, Slots: 1, Perks: []string{"Bracket header branding", "Finals MVP award presenter"}},
		{ID: seedids.TierID(21), TournamentID: seedids.TournRealRegValorant, Name: "Silver", PriceMinor: 750000, Slots: 2, Perks: []string{"Stream ticker rotation", "Discord partner announcement"}},
		{ID: seedids.TierID(22), TournamentID: seedids.TournRealRegValorant, Name: "Community", PriceMinor: 250000, Slots: 4, Perks: []string{"Official supporter badge"}},

		// B5: ASEAN Dota 2 Invitational 2026
		{ID: seedids.TierID(23), TournamentID: seedids.TournRealOngoingDota2, Name: "Platinum", PriceMinor: 4000000, Slots: 1, Perks: []string{"In-game draft overlay sponsor", "Main Stage A LED backdrop"}},
		{ID: seedids.TierID(24), TournamentID: seedids.TournRealOngoingDota2, Name: "Gold", PriceMinor: 3000000, Slots: 1, Perks: []string{"Post-match interview backdrop", "VIP arena passes"}},
		{ID: seedids.TierID(25), TournamentID: seedids.TournRealOngoingDota2, Name: "Broadcast Partner", PriceMinor: 2000000, Slots: 2, Perks: []string{"Broadcast fiber & replay wipe branding"}},

		// B6: Thailand Mobile Legends Championship S4
		{ID: seedids.TierID(26), TournamentID: seedids.TournRealCompletedMLBB, Name: "Finals Partner", PriceMinor: 2500000, Slots: 1, Perks: []string{"Grand Final stage presentation", "Winner check branding"}},
		{ID: seedids.TierID(27), TournamentID: seedids.TournRealCompletedMLBB, Name: "Payment Partner", PriceMinor: 1500000, Slots: 1, Perks: []string{"Official checkout & merchandise partner"}},
		{ID: seedids.TierID(28), TournamentID: seedids.TournRealCompletedMLBB, Name: "Stream Partner", PriceMinor: 1000000, Slots: 2, Perks: []string{"VOD & highlight reel outro logo"}},
	}
}

func buildSponsorshipPledges() []models.SponsorshipPledge {
	tierByPledge := map[uuid.UUID]struct {
		tournID uuid.UUID
		tierID  uuid.UUID
	}{
		seedids.PledgeID(1):  {seedids.TournTestCrowdGoalReached, seedids.TierID(1)},
		seedids.PledgeID(2):  {seedids.TournTestCrowdGoalReached, seedids.TierID(2)},
		seedids.PledgeID(3):  {seedids.TournTestCrowdUnderfunded, seedids.TierID(4)},
		seedids.PledgeID(4):  {seedids.TournTestCrowdUnderfunded, seedids.TierID(5)},
		seedids.PledgeID(5):  {seedids.TournTestCrowdUnderfunded, seedids.TierID(6)},
		seedids.PledgeID(6):  {seedids.TournTestCrowdUnderfunded, seedids.TierID(6)},
		seedids.PledgeID(7):  {seedids.TournTestSponsorCheckout, seedids.TierID(11)},
		seedids.PledgeID(8):  {seedids.TournTestSoloPaid, seedids.TierID(12)},
		seedids.PledgeID(9):  {seedids.TournRealCrowdBCA, seedids.TierID(17)},
		seedids.PledgeID(10): {seedids.TournRealCrowdBCA, seedids.TierID(18)},
		seedids.PledgeID(11): {seedids.TournRealCrowdBCA, seedids.TierID(19)},
		seedids.PledgeID(12): {seedids.TournRealRegValorant, seedids.TierID(20)},
		seedids.PledgeID(13): {seedids.TournRealRegValorant, seedids.TierID(21)},
		seedids.PledgeID(14): {seedids.TournRealOngoingDota2, seedids.TierID(23)},
		seedids.PledgeID(15): {seedids.TournRealOngoingDota2, seedids.TierID(24)},
		seedids.PledgeID(16): {seedids.TournRealOngoingDota2, seedids.TierID(25)},
		seedids.PledgeID(17): {seedids.TournRealCompletedMLBB, seedids.TierID(26)},
		seedids.PledgeID(18): {seedids.TournRealCompletedMLBB, seedids.TierID(27)},
		seedids.PledgeID(19): {seedids.TournRealCompletedMLBB, seedids.TierID(28)},
	}

	invoices := seedids.SponsorshipInvoices()
	pledges := make([]models.SponsorshipPledge, 0, len(invoices))
	for _, inv := range invoices {
		meta := tierByPledge[inv.PledgeID]
		paymentID := inv.InvoiceID

		var status models.PledgeStatus
		switch inv.Status {
		case "PAID":
			status = models.PledgeConfirmed
		case "UNPAID":
			status = models.PledgePending
		case "EXPIRED":
			status = models.PledgeExpired
		default:
			status = models.PledgeCancelled
		}

		pledges = append(pledges, models.SponsorshipPledge{
			ID:             inv.PledgeID,
			TournamentID:   meta.tournID,
			TierID:         meta.tierID,
			PayerAccountID: inv.PayerAccountID,
			AmountMinor:    inv.AmountMinor,
			Currency:       "THB",
			Status:         status,
			PaymentID:      &paymentID,
		})
	}
	return pledges
}

func buildAnnouncements(now time.Time) []models.TournamentAnnouncement {
	return []models.TournamentAnnouncement{
		{
			ID:                seedids.AnnouncementID(1),
			TournamentID:      seedids.TournTestPermAdmin,
			Title:             "Team Check-in & Server Assignment",
			Body:              "Captains must confirm their 4-player max roster 30 minutes before registration closes. Admins can approve pending teams from the Registrations tab.",
			AuthorAccountID:   seedids.BobID,
			AuthorDisplayName: "Bob Dev",
			CreatedAt:         now.Add(-36 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(2),
			TournamentID:      seedids.TournTestBracketLive,
			Title:             "Play-In Pod A & Wildcard Duel Complete",
			Body:              "Play-In Pod A (3->2) winners advanced to Upper Semifinal, and the Wildcard Duel (2->1) winner jumped directly to Stage 3 Grand Final. Play-In Pod B (4->1) remains open for swaps and scoring.",
			AuthorAccountID:   seedids.AlexID,
			AuthorDisplayName: "Alex Dev",
			CreatedAt:         now.Add(-12 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(3),
			TournamentID:      seedids.TournRealCrowdBCA,
			Title:             "Venue Confirmation: Samyan Mitrtown Hall",
			Body:              "We have officially locked Hall 2 at Samyan Mitrtown with spectator seating for 400 fans and dedicated player warm-up booths.",
			AuthorAccountID:   seedids.BobID,
			AuthorDisplayName: "Bob Dev",
			CreatedAt:         now.Add(-72 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(4),
			TournamentID:      seedids.TournRealCrowdBCA,
			Title:             "Welcome HyperVolt, Apex Gear & Cloud9TH!",
			Body:              "Thank you to our three confirmed partners bringing us to 65% of our stage production goal. Two Server Partner slots and one Hardware slot remain open.",
			AuthorAccountID:   seedids.BobID,
			AuthorDisplayName: "Bob Dev",
			CreatedAt:         now.Add(-20 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(5),
			TournamentID:      seedids.TournRealRegValorant,
			Title:             "Updated Map Pool & Riot ID Verification",
			Body:              "Please make sure every team member enters an exact Riot ID with #TAG in the registration form so admins can clear rosters quickly.",
			AuthorAccountID:   seedids.AlexID,
			AuthorDisplayName: "Alex Dev",
			CreatedAt:         now.Add(-96 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(6),
			TournamentID:      seedids.TournRealRegValorant,
			Title:             "8 Slots Remaining Before Roster Lock",
			Body:              "Six teams are already verified and two locked rosters are currently in the review queue. Use invite code VALO26 if you are testing team lobby join.",
			AuthorAccountID:   seedids.AlexID,
			AuthorDisplayName: "Alex Dev",
			CreatedAt:         now.Add(-16 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(7),
			TournamentID:      seedids.TournRealRegSF6Solo,
			Title:             "PS5 Setup & Controller Compatibility Notice",
			Body:              "All stations will run PS5 builds of Street Fighter 6. Bring a Brook Wingman FGC converter if your arcade stick is PS4/PC only.",
			AuthorAccountID:   seedids.BobID,
			AuthorDisplayName: "Bob Dev",
			CreatedAt:         now.Add(-60 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(8),
			TournamentID:      seedids.TournRealRegSF6Solo,
			Title:             "Pool Seeding Closes Soon",
			Body:              "Fill out your CFN ID and self-rated tier accurately so referees can balance early pools and avoid local region collisions.",
			AuthorAccountID:   seedids.BobID,
			AuthorDisplayName: "Bob Dev",
			CreatedAt:         now.Add(-10 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(9),
			TournamentID:      seedids.TournRealOngoingDota2,
			Title:             "Groups A & B Wrap — Top 4 Advance to Playoff Heat 1",
			Body:              "Groups A and B (5->2 pods) have finished on schedule, seeding Bangkok Guardians, Manila Roshan, Saigon Storm, and Mekong Dire into Playoff Heat 1 (4->2). Groups C & D play next on Main Stage A and Stream Booth 2.",
			AuthorAccountID:   seedids.BobID,
			AuthorDisplayName: "Bob Dev",
			CreatedAt:         now.Add(-30 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(10),
			TournamentID:      seedids.TournRealOngoingDota2,
			Title:             "Groups C & D Stage Schedule",
			Body:              "Group C (5->2) kicks off on Main Stage A followed by Group D (5->2) in Stream Booth 2. Doors open at 12:00 ICT.",
			AuthorAccountID:   seedids.AlexID,
			AuthorDisplayName: "Alex Dev",
			CreatedAt:         now.Add(-6 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(11),
			TournamentID:      seedids.TournRealOngoingDota2,
			Title:             "Crowdfunding Milestone Reached (112%)",
			Body:              "Thanks to Alex Ventures, Carter Fundings, and Siam Fiber, bonus broadcast segments and player lounge refreshments are fully funded.",
			AuthorAccountID:   seedids.BobID,
			AuthorDisplayName: "Bob Dev",
			CreatedAt:         now.Add(-18 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(12),
			TournamentID:      seedids.TournRealCompletedMLBB,
			Title:             "Playoff Bracket Locked — Welcome to Season 4",
			Body:              "Eight invited and qualified teams have checked in at Bangkok Convention Center. Good luck in the Quarterfinals!",
			AuthorAccountID:   seedids.MayaID,
			AuthorDisplayName: "Maya",
			CreatedAt:         now.Add(-18 * 24 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(13),
			TournamentID:      seedids.TournRealCompletedMLBB,
			Title:             "Grand Final Matchup: Siam Titans vs Chao Phraya Five",
			Body:              "After two thrilling Semifinals, Siam Titans and Chao Phraya Five clash on the main stage in a Best-of-5 for the Season 4 crown.",
			AuthorAccountID:   seedids.MayaID,
			AuthorDisplayName: "Maya",
			CreatedAt:         now.Add(-15 * 24 * time.Hour),
		},
		{
			ID:                seedids.AnnouncementID(14),
			TournamentID:      seedids.TournRealCompletedMLBB,
			Title:             "Congratulations to Our Season 4 Champions — Siam Titans!",
			Body:              "Siam Titans clinch a 3-1 Grand Final victory! Full bracket results, match scores, and sponsor credits are permanently archived below.",
			AuthorAccountID:   seedids.MayaID,
			AuthorDisplayName: "Maya",
			CreatedAt:         now.Add(-14 * 24 * time.Hour),
		},
	}
}
