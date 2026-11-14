package usecases

import (
	"testing"
	"time"

	"github.com/moomaideng/eventory/services/tournament/models"
)

func TestSponsorshipOpenAfterCrowdfunding(t *testing.T) {
	for _, status := range []models.TournamentStatus{models.TournamentStatusCrowdfunding, models.TournamentStatusRegistration, models.TournamentStatusOngoing, models.TournamentStatusDraft, models.TournamentStatusCompleted} {
		now := time.Now()
		tournament := models.Tournament{Status: status, EndsAt: now.Add(time.Hour)}
		want := status != models.TournamentStatusDraft && status != models.TournamentStatusCompleted
		if sponsorshipOpen(&tournament, now) != want {
			t.Errorf("unexpected eligibility for %s", status)
		}
		tournament.EndsAt = now
		if sponsorshipOpen(&tournament, now) {
			t.Errorf("ended %s accepted", status)
		}
	}
}

func TestSponsorshipOpenWithoutEndDate(t *testing.T) {
	now := time.Now()
	open := models.Tournament{Status: models.TournamentStatusCrowdfunding}
	if !sponsorshipOpen(&open, now) {
		t.Error("crowdfunding tournament without an end date rejected sponsorship")
	}
	draft := models.Tournament{Status: models.TournamentStatusDraft}
	if sponsorshipOpen(&draft, now) {
		t.Error("draft tournament accepted sponsorship")
	}
}
