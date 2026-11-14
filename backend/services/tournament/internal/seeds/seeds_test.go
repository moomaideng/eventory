package seeds

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/seedids"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

func TestSeedFixturesConsistency(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tournaments := buildTournaments(now)
	if len(tournaments) != 17 {
		t.Fatalf("expected 17 tournaments, got %d", len(tournaments))
	}

	staff := buildStaff(now)
	alexCount := 0
	for _, s := range staff {
		if s.AccountID == seedids.AlexID {
			alexCount++
		}
	}
	if alexCount != 12 {
		t.Fatalf("expected Alex to hold 12 staff memberships (fits page 1), got %d", alexCount)
	}

	teams, members := buildTeamsAndMembers(now)
	inviteCodes := map[string]bool{}
	teamKeys := map[string]models.TournamentTeam{}
	acceptedCounts := map[uuid.UUID]int{}
	for _, tm := range teams {
		teamKeys[tm.TournamentID.String()+":"+tm.ID] = tm
		if tm.Status == models.TournamentTeamStatusAccepted {
			acceptedCounts[tm.TournamentID]++
		}
		if tm.InviteCode != nil {
			if inviteCodes[*tm.InviteCode] {
				t.Fatalf("duplicate invite code %s", *tm.InviteCode)
			}
			inviteCodes[*tm.InviteCode] = true
		}
	}

	// Ensure no account appears twice in the same tournament's teams
	memberSeen := map[string]bool{}
	for _, m := range members {
		key := m.TournamentID.String() + ":" + m.AccountID.String()
		if memberSeen[key] {
			t.Fatalf("duplicate account %s in tournament %s", m.AccountID, m.TournamentID)
		}
		memberSeen[key] = true
	}

	// Ensure every RegistrationInvoiceFixture points to an existing seeded team
	for _, inv := range seedids.RegistrationInvoices() {
		key := inv.TournamentID.String() + ":" + inv.TeamID
		tm, ok := teamKeys[key]
		if !ok {
			t.Fatalf("registration invoice references missing team %s", key)
		}
		if tm.PaymentID == nil || *tm.PaymentID != inv.InvoiceID {
			t.Fatalf("team %s missing matching payment ID %s", key, inv.InvoiceID)
		}
	}

	// Validate all seeded brackets against the production bracket layout validator
	matches, entries := buildBrackets(now)
	matchesByTourn := map[uuid.UUID][]models.Match{}
	matchByID := map[uuid.UUID]models.Match{}
	for _, m := range matches {
		matchesByTourn[m.TournamentID] = append(matchesByTourn[m.TournamentID], m)
		matchByID[m.ID] = m
	}

	var validator usecases.BracketUseCase
	for tournID, tMatches := range matchesByTourn {
		inputs := make([]usecases.DefineBracketMatchInput, 0, len(tMatches))
		stage1Cap := 0
		for _, m := range tMatches {
			var targetTempID *string
			if m.TargetMatchID != nil {
				s := m.TargetMatchID.String()
				targetTempID = &s
			}
			if m.StageLevel == 1 {
				stage1Cap += m.TeamCount
			}
			inputs = append(inputs, usecases.DefineBracketMatchInput{
				TempID:         m.ID.String(),
				Name:           m.Name,
				Venue:          m.Venue,
				StageLevel:     m.StageLevel,
				TeamCount:      m.TeamCount,
				AdvancingCount: m.AdvancingCount,
				TargetTempID:   targetTempID,
			})
		}
		if err := validator.ValidateBracketLayout(inputs); err != nil {
			t.Fatalf("invalid seeded bracket for tournament %s: %v", tournID, err)
		}
		if stage1Cap != acceptedCounts[tournID] {
			t.Fatalf("tournament %s stage 1 capacity %d != accepted team count %d", tournID, stage1Cap, acceptedCounts[tournID])
		}
	}

	entryCountByMatch := map[uuid.UUID]int{}
	for _, e := range entries {
		m, ok := matchByID[e.MatchID]
		if !ok {
			t.Fatalf("match entry references missing match %s", e.MatchID)
		}
		if m.TournamentID != e.TournamentID {
			t.Fatalf("match entry tournament %s mismatches match tournament %s", e.TournamentID, m.TournamentID)
		}
		tm, ok := teamKeys[e.TournamentID.String()+":"+e.TeamID]
		if !ok || tm.Status != models.TournamentTeamStatusAccepted {
			t.Fatalf("match entry references non-accepted or missing team %s in %s", e.TeamID, e.TournamentID)
		}
		entryCountByMatch[e.MatchID]++
		if entryCountByMatch[e.MatchID] > m.TeamCount {
			t.Fatalf("match %s (%s) has %d entries exceeding TeamCount %d", m.ID, m.Name, entryCountByMatch[e.MatchID], m.TeamCount)
		}
	}

	// Validate all MongoDB registration forms & answers
	forms, responses, err := BuildFormFixtures()
	if err != nil {
		t.Fatalf("BuildFormFixtures failed: %v", err)
	}
	if len(forms) != 5 {
		t.Fatalf("expected 5 registration forms, got %d", len(forms))
	}
	if len(responses) == 0 {
		t.Fatal("expected non-empty seeded form responses")
	}
}
