package sponsorship

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	accountmodels "github.com/moomaideng/eventory/services/account/models"
	paymentmodels "github.com/moomaideng/eventory/services/payment/models"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

// Real PostgreSQL, account gRPC, JSON checkout HTTP, and the mock payment service.
func TestSponsorshipCheckout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	app := apptest.NewApp(t, pg.DSN)
	s := &tierScenario{app: app, client: apptest.NewClient(app.TournamentBaseURL()), tierIDs: map[string]uuid.UUID{}}
	if err := s.publishedTournament(); err != nil {
		t.Fatal(err)
	}
	if err := s.given("Gold", 1000000); err != nil {
		t.Fatal(err)
	}
	if err := app.DB.Model(&models.SponsorshipTier{}).Where("id = ?", s.tierIDs["Gold"]).Update("slots", 1).Error; err != nil {
		t.Fatal(err)
	}

	sponsor := createSponsor(t, app, "First Brand", true)
	other := createSponsor(t, app, "Second Brand", true)
	withoutProfile := createSponsor(t, app, "No Profile", false)
	incompleteProfile := createSponsorWithBrand(t, app, "Incomplete Brand", "", "")

	startPath := s.tiersPath() + "/" + s.tierIDs["Gold"].String() + "/checkout"
	post := func(user *apptest.TestUser, id uuid.UUID) *apptest.Response {
		return startPledge(t, s.client, startPath, user, id)
	}
	if got := post(withoutProfile, uuid.New()).StatusCode; got != 403 {
		t.Fatalf("missing sponsor profile: %d", got)
	}
	if got := post(incompleteProfile, uuid.New()).StatusCode; got != 403 {
		t.Fatalf("incomplete sponsor profile: %d", got)
	}
	missingTier := s.tiersPath() + "/" + uuid.NewString() + "/checkout"
	if got := startPledge(t, s.client, missingTier, sponsor, uuid.New()).StatusCode; got != 404 {
		t.Fatalf("missing package should map to 404, got %d", got)
	}
	mustAPI(t, s.client, "GET", "/tournaments/"+uuid.NewString()+"/sponsors", nil, nil, 404)

	firstID, secondID := uuid.New(), uuid.New()
	var first, second *apptest.Response
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); first = post(sponsor, firstID) }()
	go func() { defer wg.Done(); second = post(other, secondID) }()
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
	if err := s.given("Silver", 10000); err != nil {
		t.Fatal(err)
	}
	secondTierPath := s.tiersPath() + "/" + s.tierIDs["Silver"].String() + "/checkout"
	if got := startPledge(t, s.client, secondTierPath, winner, uuid.New()).StatusCode; got != http.StatusConflict {
		t.Fatalf("a sponsor with an active pledge must not buy another tournament tier: %d", got)
	}
	mustAPI(t, s.client, "PUT", s.tiersPath()+"/"+s.tierIDs["Gold"].String(), map[string]any{"name": "Changed", "priceMinor": 1, "slots": 9, "perks": []string{"Changed"}}, s.owner, 409)
	mustAPI(t, s.client, "DELETE", s.tiersPath()+"/"+s.tierIDs["Gold"].String(), nil, s.owner, 409)
	if retried := post(winner, winnerID); retried.StatusCode != 200 {
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
	var details struct {
		Funding struct {
			GoalAmountMinor   int64 `json:"goalAmountMinor"`
			RaisedAmountMinor int64 `json:"raisedAmountMinor"`
			SupporterCount    int   `json:"supporterCount"`
		} `json:"funding"`
	}
	detailsResp := mustAPI(t, s.client, "GET", "/tournaments/"+s.tournament.ID.String(), nil, nil, 200)
	if err := detailsResp.JSON(&details); err != nil || details.Funding.RaisedAmountMinor != 1000000 || details.Funding.SupporterCount != 1 {
		t.Fatalf("funding details: %s, %v", detailsResp.Body, err)
	}
	public = mustAPI(t, s.client, "GET", "/tournaments/"+s.tournament.ID.String()+"/sponsors", nil, nil, 200)
	if !bytes.Contains(public.Body, []byte("Brand")) || !bytes.Contains(public.Body, []byte(`"logoUrl":"https://brand.example/logo.png"`)) {
		t.Fatalf("paid sponsor or logoUrl missing: %s", public.Body)
	}
	mustAPI(t, s.client, "PUT", s.tiersPath()+"/"+s.tierIDs["Gold"].String(), map[string]any{"name": "Changed", "priceMinor": 1, "slots": 9, "perks": []string{"Changed"}}, s.owner, 409)
	mustAPI(t, s.client, "DELETE", s.tiersPath()+"/"+s.tierIDs["Gold"].String(), nil, s.owner, 409)

	for _, result := range []string{"FAILED", "CANCELLED"} {
		if err := s.given(result, 10000); err != nil {
			t.Fatal(err)
		}
		path := s.tiersPath() + "/" + s.tierIDs[result].String() + "/checkout"
		attemptID := uuid.New()
		resp := startPledge(t, s.client, path, loser, attemptID)
		if err := resp.JSON(&pledge); err != nil || resp.StatusCode != 200 {
			t.Fatalf("start %s: %s", result, resp.Body)
		}
		mustAPI(t, payments, "POST", "/payments/"+pledge.PaymentID.String()+"/mock-result", map[string]string{"result": result}, loser, 200)
		closed := mustAPI(t, s.client, "POST", "/tournaments/sponsorships/"+attemptID.String()+"/refresh", nil, loser, 200)
		if !bytes.Contains(closed.Body, []byte(`"status":"CANCELLED"`)) {
			t.Fatalf("cancel failed: %s", closed.Body)
		}
		retryID := uuid.New()
		retry := startPledge(t, s.client, path, loser, retryID)
		if retry.StatusCode != 200 {
			t.Fatalf("slot not released: %s", retry.Body)
		}
		if err := retry.JSON(&pledge); err != nil {
			t.Fatalf("decode retry pledge: %v", err)
		}
		mustAPI(t, payments, "POST", "/payments/"+pledge.PaymentID.String()+"/mock-result", map[string]string{"result": "CANCELLED"}, loser, 200)
		mustAPI(t, s.client, "POST", "/tournaments/sponsorships/"+retryID.String()+"/refresh", nil, loser, 200)
	}
	// Recover an attempt whose invoice expired in payment service after an outage.
	if err := s.given("Expired", 10000); err != nil {
		t.Fatal(err)
	}
	expiredID := uuid.New()
	expiredPaymentID := uuid.New()
	invoice := paymentmodels.Invoice{
		ID: expiredPaymentID, PayerAccountID: sponsor.ID, Source: models.PaymentSourceSponsorship,
		Reference: expiredID.String(), IdempotencyKey: "sponsorship:" + expiredID.String(),
		Description: "Expired sponsorship", AmountMinor: 10000, Currency: "THB", Provider: "MOCK",
		ExpiresAt: time.Now().Add(-time.Minute).UTC().Truncate(time.Microsecond),
	}
	if err := app.DB.Create(&invoice).Error; err != nil {
		t.Fatal(err)
	}
	expired := models.SponsorshipPledge{
		ID: expiredID, TournamentID: s.tournament.ID, TierID: s.tierIDs["Expired"], PayerAccountID: sponsor.ID,
		AmountMinor: 10000, Currency: "THB", Status: models.PledgePending, PaymentID: &expiredPaymentID,
	}
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
		"registrationMode": "SOLO", "entryFeeMinor": 0, "minTeamSize": 1, "maxTeamSize": 1, "capacity": 8,
		"fundingGoalMinor": 1000000, "fundingDeadline": deadline,
	}, owner, http.StatusCreated)
	if err := created.JSON(&s.tournament); err != nil || s.tournament.Status != models.TournamentStatusDraft {
		t.Fatalf("create draft tournament: %s, %v", created.Body, err)
	}
	if err := s.given("Gold", 1000000); err != nil {
		t.Fatal(err)
	}
	path := "/tournaments/" + s.tournament.ID.String()
	mustAPI(t, s.client, "PATCH", path+"/status", map[string]any{"status": "CROWDFUNDING"}, owner, http.StatusOK)
	launch := map[string]any{"status": "REGISTRATION", "reason": "Event details are ready for registration"}
	mustAPI(t, s.client, "PATCH", path+"/status", launch, owner, http.StatusUnprocessableEntity)

	sponsor := createSponsor(t, app, "Campaign Sponsor", true)
	checkout := startPledge(t, s.client, s.tiersPath()+"/"+s.tierIDs["Gold"].String()+"/checkout", sponsor, uuid.New())
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

	var details struct {
		Funding struct {
			GoalAmountMinor   int64      `json:"goalAmountMinor"`
			RaisedAmountMinor int64      `json:"raisedAmountMinor"`
			SupporterCount    int        `json:"supporterCount"`
			Deadline          *time.Time `json:"deadline"`
		} `json:"funding"`
	}
	detailsResp := mustAPI(t, s.client, "GET", path, nil, nil, http.StatusOK)
	if err := detailsResp.JSON(&details); err != nil {
		t.Fatal(err)
	}
	f := details.Funding
	if f.GoalAmountMinor != 1000000 || f.RaisedAmountMinor != 1000000 || f.SupporterCount != 1 || f.Deadline == nil || !f.Deadline.Equal(deadline) {
		t.Fatalf("payment or event edit changed the campaign terms or totals: %+v", f)
	}
	mustAPI(t, s.client, "PATCH", path+"/status", launch, owner, http.StatusOK)
}

func createSponsor(t *testing.T, app *apptest.App, name string, profile bool) *apptest.TestUser {
	t.Helper()
	if !profile {
		return createSponsorWithBrand(t, app, name, "", "", true)
	}
	return createSponsorWithBrand(t, app, name, "https://brand.example", "https://brand.example/logo.png")
}

func createSponsorWithBrand(t *testing.T, app *apptest.App, name, websiteURL, logoURL string, skipProfile ...bool) *apptest.TestUser {
	t.Helper()
	u := apptest.NewTestUser()
	account := accountmodels.Account{ID: u.ID, Email: u.Email, Handle: u.Handle, DisplayName: u.DisplayName, Status: accountmodels.AccountStatusActive}
	if err := app.DB.Create(&account).Error; err != nil {
		t.Fatal(err)
	}
	if len(skipProfile) == 0 || !skipProfile[0] {
		if err := app.DB.Create(&accountmodels.SponsorProfile{
			AccountID:    u.ID,
			SponsorName:  name,
			SponsorEmail: u.Email,
			WebsiteURL:   websiteURL,
			LogoURL:      logoURL,
			Tags:         []string{"esports"},
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return u
}

func startPledge(t *testing.T, client *apptest.Client, path string, user *apptest.TestUser, id uuid.UUID) *apptest.Response {
	t.Helper()
	resp, err := client.Do("POST", path, map[string]any{"requestId": id}, user.AuthHeaders())
	if err != nil {
		t.Error(err)
		return &apptest.Response{}
	}
	return resp
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
