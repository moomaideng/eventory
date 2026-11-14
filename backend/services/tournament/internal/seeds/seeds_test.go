package seeds

import (
	"testing"
	"time"

	"github.com/moomaideng/eventory/internal/seedids"
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
	for _, tm := range teams {
		teamKeys[tm.TournamentID.String()+":"+tm.ID] = tm
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
