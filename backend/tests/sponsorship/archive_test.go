package sponsorship

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	paymentmodels "github.com/moomaideng/eventory/services/payment/models"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

func TestSponsorshipArchive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	app := apptest.NewApp(t, pg.DSN)
	s := &tierScenario{app: app, client: apptest.NewClient(app.TournamentBaseURL()), tierIDs: map[string]uuid.UUID{}}
	if err := s.publishedTournament(); err != nil {
		t.Fatal(err)
	}
	var imageBuffer bytes.Buffer
	if err := png.Encode(&imageBuffer, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	logo := imageBuffer.Bytes()
	payer := createSponsor(t, app, "Archive test", true)
	payments := apptest.NewClient(app.PaymentServer.URL + "/api/v1")

	for _, result := range []string{"UNUSED", "CANCELLED", "FAILED", "EXPIRED"} {
		t.Run(result, func(t *testing.T) {
			if err := s.given(result, 0); err != nil {
				t.Fatal(err)
			}
			tierID := s.tierIDs[result]
			path := s.tiersPath() + "/" + tierID.String()
			checkoutURL := app.TournamentBaseURL() + path + "/checkout"
			var pledge struct {
				ID        uuid.UUID `json:"id"`
				PaymentID uuid.UUID `json:"paymentId"`
			}
			closedStatus := "CANCELLED"
			if result != "UNUSED" {
				response := uploadPledge(t, checkoutURL, payer, uuid.New(), logo)
				if err := response.JSON(&pledge); err != nil || response.StatusCode != 200 {
					t.Fatalf("checkout: %s, %v", response.Body, err)
				}
				if result == "EXPIRED" {
					// Both services own matching immutable terms; fast-forward them together.
					deadline := time.Now().Add(-time.Minute).UTC().Truncate(time.Microsecond)
					if err := app.DB.Model(&models.SponsorshipPledge{}).Where("id = ?", pledge.ID).Update("expires_at", deadline).Error; err != nil {
						t.Fatal(err)
					}
					if err := app.DB.Model(&paymentmodels.Invoice{}).Where("id = ?", pledge.PaymentID).Update("expires_at", deadline).Error; err != nil {
						t.Fatal(err)
					}
					closedStatus = "EXPIRED"
				} else {
					mustAPI(t, payments, "POST", "/payments/"+pledge.PaymentID.String()+"/mock-result", map[string]string{"result": result}, payer, 200)
				}
				mustAPI(t, s.client, "POST", "/tournaments/sponsorships/"+pledge.ID.String()+"/refresh", nil, payer, 200)
			}

			mustAPI(t, s.client, "DELETE", path, nil, s.owner, 204)
			var tier models.SponsorshipTier
			if err := app.DB.First(&tier, "id = ?", tierID).Error; err != nil || !tier.Archived {
				t.Fatalf("package must stay stored and archived: %+v, %v", tier, err)
			}
			for _, list := range []string{s.tiersPath(), s.tiersPath() + "/manage"} {
				if response := mustAPI(t, s.client, "GET", list, nil, s.owner, 200); bytes.Contains(response.Body, []byte(tierID.String())) {
					t.Fatalf("archived package still listed: %s", response.Body)
				}
			}
			if response := uploadPledge(t, checkoutURL, payer, uuid.New(), logo); response.StatusCode != 404 {
				t.Fatalf("new checkout must reject an archived package: %d %s", response.StatusCode, response.Body)
			}
			mustAPI(t, s.client, "PUT", path, map[string]any{"name": "Changed", "price": 1, "slots": 1, "perks": []string{"Logo"}}, s.owner, 404)
			mustAPI(t, s.client, "DELETE", path, nil, s.owner, 404)

			if result != "UNUSED" {
				response := mustAPI(t, s.client, "POST", "/tournaments/sponsorships/"+pledge.ID.String()+"/refresh", nil, payer, 200)
				if !bytes.Contains(response.Body, []byte(`"status":"`+closedStatus+`"`)) {
					t.Fatalf("closed checkout context was lost: %s", response.Body)
				}
				if response := uploadPledge(t, checkoutURL, payer, pledge.ID, logo); response.StatusCode != 200 {
					t.Fatalf("old request must still return its original checkout: %s", response.Body)
				}
				if response := mustAPI(t, s.client, "GET", "/tournaments/sponsorships/mine", nil, payer, 200); !bytes.Contains(response.Body, []byte(pledge.ID.String())) {
					t.Fatalf("sponsorship history was lost: %s", response.Body)
				}
				var storedLogo models.SponsorshipLogo
				if err := app.DB.First(&storedLogo, "pledge_id = ?", pledge.ID).Error; err != nil || !bytes.Equal(storedLogo.Data, logo) {
					t.Fatalf("original branding was lost: %v", err)
				}
				mustAPI(t, payments, "GET", "/payments/"+pledge.PaymentID.String(), nil, payer, 200)
				mustAPI(t, payments, "GET", "/payments/"+pledge.PaymentID.String()+"/receipt", nil, payer, 404)
			}
			// A new package can reuse an archived name without changing old purchase snapshots.
			if err := s.given(result, 0); err != nil {
				t.Fatal(err)
			}
			if s.tierIDs[result] == tierID {
				t.Fatal("replacement package reused the archived ID")
			}
		})
	}

	t.Run("archive and checkout use the same lock", func(t *testing.T) {
		if err := s.given("Race", 0); err != nil {
			t.Fatal(err)
		}
		path := s.tiersPath() + "/" + s.tierIDs["Race"].String()
		var checkout, archive *apptest.Response
		var archiveErr error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			checkout = uploadPledge(t, app.TournamentBaseURL()+path+"/checkout", payer, uuid.New(), logo)
		}()
		go func() {
			defer wg.Done()
			archive, archiveErr = s.client.Do("DELETE", path, nil, s.owner.AuthHeaders())
		}()
		wg.Wait()
		if archiveErr != nil {
			t.Fatal(archiveErr)
		}
		if !((archive.StatusCode == 204 && checkout.StatusCode == 404) || (archive.StatusCode == 409 && checkout.StatusCode == 200)) {
			t.Fatalf("archive and checkout both succeeded or failed unexpectedly: archive=%d %s checkout=%d %s", archive.StatusCode, archive.Body, checkout.StatusCode, checkout.Body)
		}
	})
}
