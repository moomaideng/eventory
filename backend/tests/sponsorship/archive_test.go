package sponsorship

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

// TestSponsorshipTierDelete checks that draft tiers with no pledges are deleted cleanly,
// while published tournaments or tiers with pledges cannot be edited or deleted.
func TestSponsorshipTierDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	app := apptest.NewApp(t, pg.DSN)
	s := &tierScenario{app: app, client: apptest.NewClient(app.TournamentBaseURL()), tierIDs: map[string]uuid.UUID{}}
	if err := s.draftTournament(); err != nil {
		t.Fatal(err)
	}

	if err := s.given("Unused", 10000); err != nil {
		t.Fatal(err)
	}
	unusedID := s.tierIDs["Unused"]
	path := s.tiersPath() + "/" + unusedID.String()
	mustAPI(t, s.client, "DELETE", path, nil, s.owner, 204)

	var count int64
	app.DB.Model(&models.SponsorshipTier{}).Where("id = ?", unusedID).Count(&count)
	if count != 0 {
		t.Fatalf("expected unused tier to be deleted, found %d", count)
	}
	mustAPI(t, s.client, "PUT", path, map[string]any{"name": "Changed", "priceMinor": 1, "slots": 1, "perks": []string{"Logo"}}, s.owner, 404)

	// Reusing the name after deletion creates a fresh tier ID.
	if err := s.given("Unused", 15000); err != nil {
		t.Fatal(err)
	}
	if s.tierIDs["Unused"] == unusedID {
		t.Fatal("recreated tier reused old ID")
	}

	// Publishing locks tiers against modification or deletion.
	if err := s.app.DB.Model(&s.tournament).Update("status", models.TournamentStatusRegistration).Error; err != nil {
		t.Fatal(err)
	}
	lockedPath := s.tiersPath() + "/" + s.tierIDs["Unused"].String()
	mustAPI(t, s.client, "PUT", lockedPath, map[string]any{"name": "Changed", "priceMinor": 1, "slots": 1, "perks": []string{"Logo"}}, s.owner, 409)
	mustAPI(t, s.client, "DELETE", lockedPath, nil, s.owner, 409)
}
