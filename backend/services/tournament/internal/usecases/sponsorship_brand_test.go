package usecases

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"time"

	"github.com/moomaideng/eventory/services/tournament/models"
)

func TestSponsorshipBrandValidation(t *testing.T) {
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	valid := buffer.Bytes()
	for _, tc := range []struct {
		name, website string
		logo          []byte
		valid         bool
	}{
		{"valid", " https://brand.example/about ", valid, true},
		{"http", "http://brand.example", valid, true},
		{"script link", "javascript:alert(1)", valid, false},
		{"relative link", "/brand", valid, false},
		{"credentials", "https://user:pass@brand.example", valid, false},
		{"no host", "https://", valid, false},
		{"no logo", "https://brand.example", nil, false},
		{"spoofed PNG", "https://brand.example", []byte("<svg onload='alert(1)'/>"), false},
		{"oversize logo", "https://brand.example", make([]byte, MaxSponsorLogoBytes+1), false},
		{"truncated image", "https://brand.example", valid[:len(valid)-20], false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			website, contentType, err := validateSponsorshipBrand(tc.website, tc.logo)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
			if tc.valid && (website == "" || contentType != "image/png") {
				t.Fatalf("invalid normalized brand: %q %q", website, contentType)
			}
		})
	}
}

func TestSponsorshipOpenAfterCrowdfunding(t *testing.T) {
	// Covered end-to-end for registration-open tournaments; this guards the other lifecycle stages.
	for _, status := range []models.TournamentStatus{models.TournamentStatusCrowdfunding, models.TournamentStatusRegistrationOpen, models.TournamentStatusRegistrationClosed, models.TournamentStatusOngoing, models.TournamentStatusDraft, models.TournamentStatusCompleted, models.TournamentStatusCancelled} {
		now := time.Now()
		tournament := models.Tournament{Published: true, Status: status, EndsAt: now.Add(time.Hour)}
		want := status != models.TournamentStatusDraft && status != models.TournamentStatusCompleted && status != models.TournamentStatusCancelled
		if sponsorshipOpen(&tournament, now) != want {
			t.Errorf("unexpected eligibility for %s", status)
		}
		tournament.Published = false
		if sponsorshipOpen(&tournament, now) {
			t.Errorf("unpublished %s accepted", status)
		}
		tournament.Published = true
		tournament.EndsAt = now
		if sponsorshipOpen(&tournament, now) {
			t.Errorf("ended %s accepted", status)
		}
	}
}
