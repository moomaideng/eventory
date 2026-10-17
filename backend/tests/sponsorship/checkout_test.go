package sponsorship

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	accountmodels "github.com/moomaideng/eventory/services/account/models"
	paymentmodels "github.com/moomaideng/eventory/services/payment/models"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

// Real PostgreSQL, account gRPC, multipart HTTP, and the mock payment service.
func TestSponsorshipCheckout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	app := apptest.NewApp(t, pg.DSN)
	s := &tierScenario{app: app, client: apptest.NewClient(app.TournamentBaseURL()), tierIDs: map[string]uuid.UUID{}}
	if err := s.publishedTournament(); err != nil {
		t.Fatal(err)
	}
	if err := s.given("Gold", 0); err != nil {
		t.Fatal(err)
	}
	if err := app.DB.Model(&models.SponsorshipTier{}).Where("id = ?", s.tierIDs["Gold"]).Update("slots", 1).Error; err != nil {
		t.Fatal(err)
	}
	var imageBuffer bytes.Buffer
	if err := png.Encode(&imageBuffer, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	logo := imageBuffer.Bytes()
	sponsor := createSponsor(t, app, "First Brand", true)
	other := createSponsor(t, app, "Second Brand", true)
	withoutProfile := createSponsor(t, app, "No Profile", false)
	startPath := app.TournamentBaseURL() + s.tiersPath() + "/" + s.tierIDs["Gold"].String() + "/checkout"
	post := func(user *apptest.TestUser, id uuid.UUID, data []byte) *apptest.Response {
		return uploadPledge(t, startPath, user, id, data)
	}
	if got := post(withoutProfile, uuid.New(), logo).StatusCode; got != 403 {
		t.Fatalf("missing sponsor profile: %d", got)
	}
	if got := post(sponsor, uuid.New(), []byte("not an image")).StatusCode; got != 422 {
		t.Fatalf("spoofed image: %d", got)
	}
	missingTier := app.TournamentBaseURL() + s.tiersPath() + "/" + uuid.NewString() + "/checkout"
	if got := uploadPledge(t, missingTier, sponsor, uuid.New(), logo).StatusCode; got != 404 {
		t.Fatalf("missing package should map to 404, got %d", got)
	}
	mustAPI(t, s.client, "GET", "/tournaments/"+uuid.NewString()+"/sponsors", nil, nil, 404)
	mustAPI(t, s.client, "GET", "/tournaments/sponsorships/"+uuid.NewString()+"/logo", nil, nil, 404)
	firstID, secondID := uuid.New(), uuid.New()
	var first, second *apptest.Response
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); first = post(sponsor, firstID, logo) }()
	go func() { defer wg.Done(); second = post(other, secondID, logo) }()
	wg.Wait()
	winner, loser := sponsor, other
	winnerID := firstID
	if first.StatusCode == 409 {
		first, second = second, first
		winner, loser = other, sponsor
		winnerID = secondID
	}
	if first.StatusCode != 200 || second.StatusCode != 409 {
		t.Fatalf("last-slot race: %d %s / %d %s", first.StatusCode, first.Body, second.StatusCode, second.Body)
	}
	var pledge struct {
		ID        uuid.UUID `json:"id"`
		PaymentID uuid.UUID `json:"paymentId"`
		Status    string    `json:"status"`
	}
	if err := first.JSON(&pledge); err != nil || pledge.ID != winnerID || pledge.PaymentID == uuid.Nil || pledge.Status != "PENDING" {
		t.Fatalf("checkout: %s, %v", first.Body, err)
	}
	mustAPI(t, s.client, "PUT", s.tiersPath()+"/"+s.tierIDs["Gold"].String(), map[string]any{"name": "Changed", "price": 1, "slots": 9, "perks": []string{"Changed"}}, s.owner, 409)
	mustAPI(t, s.client, "DELETE", s.tiersPath()+"/"+s.tierIDs["Gold"].String(), nil, s.owner, 409)
	if retried := post(winner, winnerID, logo); retried.StatusCode != 200 {
		t.Fatalf("retry: %s", retried.Body)
	}
	var invoiceCount int64
	app.DB.Model(&paymentmodels.Invoice{}).Where("reference = ?", winnerID.String()).Count(&invoiceCount)
	if invoiceCount != 1 {
		t.Fatalf("created %d invoices for one attempt", invoiceCount)
	}
	refreshPath := "/tournaments/sponsorships/" + winnerID.String() + "/refresh"
	minePath := "/tournaments/sponsorships/mine"
	mustAPI(t, s.client, "GET", minePath, nil, nil, 401)
	var mine struct {
		Items []struct {
			ID             uuid.UUID `json:"id"`
			TournamentName string    `json:"tournamentName"`
			TierName       string    `json:"tierName"`
			AmountMinor    int64     `json:"amountMinor"`
		} `json:"items"`
		Total int64 `json:"total"`
	}
	owned := mustAPI(t, s.client, "GET", minePath+"?limit=1", nil, winner, 200)
	if err := owned.JSON(&mine); err != nil || mine.Total != 1 || len(mine.Items) != 1 || mine.Items[0].ID != winnerID || mine.Items[0].TournamentName != s.tournament.Name || mine.Items[0].TierName != "Gold" || mine.Items[0].AmountMinor != 1000000 {
		t.Fatalf("owned sponsorship list: %s, %v", owned.Body, err)
	}
	for _, path := range []string{minePath, minePath + "?page=2&limit=1"} {
		user := loser
		if path != minePath {
			user = winner
		}
		response := mustAPI(t, s.client, "GET", path, nil, user, 200)
		if err := response.JSON(&mine); err != nil || len(mine.Items) != 0 {
			t.Fatalf("ownership or pagination leaked sponsorship: %s", response.Body)
		}
	}
	mustAPI(t, s.client, "GET", minePath+"?limit=101", nil, winner, 422)
	mustAPI(t, s.client, "POST", refreshPath, nil, loser, 404)
	logoPath := "/tournaments/sponsorships/" + winnerID.String() + "/logo"
	mustAPI(t, s.client, "GET", logoPath, nil, nil, 404)
	public := mustAPI(t, s.client, "GET", "/tournaments/"+s.tournament.ID.String()+"/sponsors", nil, nil, 200)
	if !bytes.Contains(public.Body, []byte(`"items":[]`)) {
		t.Fatalf("unpaid sponsor published: %s", public.Body)
	}
	// Paid invoices are immutable; repeated successes must not double-count funding.
	payments := apptest.NewClient(app.PaymentServer.URL + "/api/v1")
	for range 2 {
		mustAPI(t, payments, "POST", "/payments/"+pledge.PaymentID.String()+"/mock-result", map[string]string{"result": "SUCCEEDED"}, winner, 200)
		confirmed := mustAPI(t, s.client, "POST", refreshPath, nil, winner, 200)
		if !bytes.Contains(confirmed.Body, []byte(`"status":"CONFIRMED"`)) {
			t.Fatalf("confirmation: %s", confirmed.Body)
		}
	}
	var funding models.TournamentFunding
	if err := app.DB.First(&funding, "tournament_id = ?", s.tournament.ID).Error; err != nil || funding.RaisedAmount != 10000 || funding.SupporterCount != 1 {
		t.Fatalf("funding: %+v, %v", funding, err)
	}
	public = mustAPI(t, s.client, "GET", "/tournaments/"+s.tournament.ID.String()+"/sponsors", nil, nil, 200)
	if !bytes.Contains(public.Body, []byte("Brand")) {
		t.Fatalf("paid sponsor missing: %s", public.Body)
	}
	storedLogo := mustAPI(t, s.client, "GET", logoPath, nil, nil, 200)
	if !bytes.Equal(storedLogo.Body, logo) || storedLogo.Header.Get("Content-Type") != "image/png" || storedLogo.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("logo response mismatch")
	}
	mustAPI(t, s.client, "PUT", s.tiersPath()+"/"+s.tierIDs["Gold"].String(), map[string]any{"name": "Changed", "price": 1, "slots": 9, "perks": []string{"Changed"}}, s.owner, 409)
	mustAPI(t, s.client, "DELETE", s.tiersPath()+"/"+s.tierIDs["Gold"].String(), nil, s.owner, 409)

	for _, result := range []string{"FAILED", "CANCELLED"} {
		if err := s.given(result, 1); err != nil {
			t.Fatal(err)
		}
		path := app.TournamentBaseURL() + s.tiersPath() + "/" + s.tierIDs[result].String() + "/checkout"
		attemptID := uuid.New()
		resp := uploadPledge(t, path, sponsor, attemptID, logo)
		if err := resp.JSON(&pledge); err != nil || resp.StatusCode != 200 {
			t.Fatalf("start %s: %s", result, resp.Body)
		}
		mustAPI(t, payments, "POST", "/payments/"+pledge.PaymentID.String()+"/mock-result", map[string]string{"result": result}, sponsor, 200)
		closed := mustAPI(t, s.client, "POST", "/tournaments/sponsorships/"+attemptID.String()+"/refresh", nil, sponsor, 200)
		if !bytes.Contains(closed.Body, []byte(`"status":"CANCELLED"`)) {
			t.Fatalf("cancel failed: %s", closed.Body)
		}
		mustAPI(t, s.client, "GET", "/tournaments/sponsorships/"+attemptID.String()+"/logo", nil, nil, 404)
		if retry := uploadPledge(t, path, sponsor, uuid.New(), logo); retry.StatusCode != 200 {
			t.Fatalf("slot not released: %s", retry.Body)
		}
	}
	// Recover an attempt after an outage that lasted beyond its deadline.
	expiredID := uuid.New()
	expired := models.SponsorshipPledge{ID: expiredID, TournamentID: s.tournament.ID, TierID: s.tierIDs["Gold"], PayerAccountID: sponsor.ID, SponsorName: "Expired", WebsiteURL: "https://brand.example", TierName: "Gold", AmountMinor: 10000, Currency: "THB", Status: models.PledgePending, ExpiresAt: time.Now().Add(-time.Minute).UTC().Truncate(time.Microsecond)}
	if err := app.DB.Omit("Tier").Create(&expired).Error; err != nil {
		t.Fatal(err)
	}
	closed := mustAPI(t, s.client, "POST", "/tournaments/sponsorships/"+expiredID.String()+"/refresh", nil, sponsor, 200)
	if !bytes.Contains(closed.Body, []byte(`"status":"EXPIRED"`)) {
		t.Fatalf("expired recovery: %s", closed.Body)
	}
}

func TestSponsorshipCrowdfundingLaunch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	app := apptest.NewApp(t, pg.DSN)
	s := &tierScenario{app: app, client: apptest.NewClient(app.TournamentBaseURL()), tierIDs: map[string]uuid.UUID{}}
	owner, err := s.organizer("Campaign Host")
	if err != nil {
		t.Fatal(err)
	}
	s.owner = owner
	now := time.Now().UTC().Truncate(time.Second)
	deadline := now.Add(24 * time.Hour)
	created := mustAPI(t, s.client, "POST", "/tournaments", map[string]any{
		"name": "Crowdfunded Sponsor Cup", "description": "A funded tournament", "game": "Chess", "location": "Online",
		"startAt": now.Add(72 * time.Hour), "endAt": now.Add(80 * time.Hour), "registrationDeadline": now.Add(48 * time.Hour),
		"registrationMode": "SOLO", "entryFee": 0, "minTeamSize": 1, "maxTeamSize": 1, "capacity": 8,
		"fundingGoal": 10000, "fundingDeadline": deadline,
	}, owner, http.StatusCreated)
	if err := created.JSON(&s.tournament); err != nil || s.tournament.Status != models.TournamentStatusCrowdfunding {
		t.Fatalf("create crowdfunding tournament: %s, %v", created.Body, err)
	}
	if err := s.given("Gold", 0); err != nil {
		t.Fatal(err)
	}
	path := "/tournaments/" + s.tournament.ID.String()
	launch := map[string]any{"status": "REGISTRATION_OPEN", "reason": "Event details are ready for registration"}
	mustAPI(t, s.client, "PATCH", path+"/status", launch, owner, http.StatusUnprocessableEntity)

	var imageBuffer bytes.Buffer
	if err := png.Encode(&imageBuffer, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	sponsor := createSponsor(t, app, "Campaign Sponsor", true)
	checkout := uploadPledge(t, app.TournamentBaseURL()+s.tiersPath()+"/"+s.tierIDs["Gold"].String()+"/checkout", sponsor, uuid.New(), imageBuffer.Bytes())
	var pledge struct {
		ID        uuid.UUID `json:"id"`
		PaymentID uuid.UUID `json:"paymentId"`
	}
	if err := checkout.JSON(&pledge); err != nil || checkout.StatusCode != http.StatusOK || pledge.PaymentID == uuid.Nil {
		t.Fatalf("crowdfunding checkout: %s, %v", checkout.Body, err)
	}
	payments := apptest.NewClient(app.PaymentServer.URL + "/api/v1")
	mustAPI(t, payments, "POST", "/payments/"+pledge.PaymentID.String()+"/mock-result", map[string]string{"result": "SUCCEEDED"}, sponsor, http.StatusOK)
	for range 2 {
		mustAPI(t, s.client, "POST", "/tournaments/sponsorships/"+pledge.ID.String()+"/refresh", nil, sponsor, http.StatusOK)
	}
	mustAPI(t, s.client, "PATCH", path, map[string]string{"description": "Final event details confirmed"}, owner, http.StatusOK)
	var funding models.TournamentFunding
	if err := app.DB.First(&funding, "tournament_id = ?", s.tournament.ID).Error; err != nil {
		t.Fatal(err)
	}
	if funding.GoalAmount != 10000 || funding.RaisedAmount != 10000 || funding.SupporterCount != 1 || funding.Deadline == nil || !funding.Deadline.Equal(deadline) {
		t.Fatalf("payment or event edit changed the campaign terms or totals: %+v", funding)
	}
	mustAPI(t, s.client, "PATCH", path+"/status", launch, owner, http.StatusOK)
}

func createSponsor(t *testing.T, app *apptest.App, name string, profile bool) *apptest.TestUser {
	t.Helper()
	u := apptest.NewTestUser()
	account := accountmodels.Account{ID: u.ID, Email: u.Email, Handle: u.Handle, DisplayName: u.DisplayName, Status: accountmodels.AccountStatusActive}
	if err := app.DB.Create(&account).Error; err != nil {
		t.Fatal(err)
	}
	if profile {
		if err := app.DB.Create(&accountmodels.SponsorProfile{AccountID: u.ID, SponsorName: name}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return u
}

func uploadPledge(t *testing.T, path string, user *apptest.TestUser, id uuid.UUID, logo []byte) *apptest.Response {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	if err := w.WriteField("requestId", id.String()); err != nil {
		t.Error(err)
		return &apptest.Response{}
	}
	if err := w.WriteField("websiteUrl", "https://brand.example"); err != nil {
		t.Error(err)
		return &apptest.Response{}
	}
	header := textproto.MIMEHeader{"Content-Disposition": {`form-data; name="logo"; filename="logo.png"`}, "Content-Type": {"image/png"}}
	part, err := w.CreatePart(header)
	if err != nil {
		t.Error(err)
		return &apptest.Response{}
	}
	if _, err := part.Write(logo); err != nil {
		t.Error(err)
		return &apptest.Response{}
	}
	if err := w.Close(); err != nil {
		t.Error(err)
		return &apptest.Response{}
	}
	req, err := http.NewRequest("POST", path, &body)
	if err != nil {
		t.Error(err)
		return &apptest.Response{}
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+user.Token)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Error(err)
		return &apptest.Response{}
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return &apptest.Response{}
	}
	return &apptest.Response{StatusCode: resp.StatusCode, Header: resp.Header, Body: data}
}

func mustAPI(t *testing.T, client *apptest.Client, method, path string, body any, user *apptest.TestUser, status int) *apptest.Response {
	t.Helper()
	var headers map[string]string
	if user != nil {
		headers = user.AuthHeaders()
	}
	resp, err := client.Do(method, path, body, headers)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != status {
		b, _ := json.Marshal(body)
		t.Fatalf("%s %s (%s): %d, want %d: %s", method, path, b, resp.StatusCode, status, resp.Body)
	}
	return resp
}
