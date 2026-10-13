package solo_registration

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	accountmodels "github.com/moomaideng/eventory/services/account/models"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/moomaideng/eventory/tests/internal/apptest"
)

func TestSoloRegistrationPaymentAndEligibility(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pg := apptest.StartPostgres(ctx, t)
	app := apptest.NewApp(t, pg.DSN)
	client := apptest.NewClient(app.TournamentBaseURL())
	payments := apptest.NewClient(app.PaymentServer.URL + "/api/v1/payments")
	user := func() *apptest.TestUser {
		u := apptest.NewTestUser()
		if err := app.DB.Create(&accountmodels.Account{ID: u.ID, Email: u.Email, Handle: u.Handle, DisplayName: u.DisplayName, Status: accountmodels.AccountStatusActive}).Error; err != nil {
			t.Fatal(err)
		}
		return u
	}
	owner, first, second := user(), user(), user()
	if err := app.DB.Create(&accountmodels.OrganizerProfile{AccountID: owner.ID, OrganizerName: "Solo host", OrganizerEmail: owner.Email}).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	tournament := models.Tournament{ID: uuid.New(), Name: "Solo cup", Description: "Solo registration test", Game: "Chess", Location: "Online", StartsAt: now.Add(72 * time.Hour), EndsAt: now.Add(80 * time.Hour), RegistrationDeadline: now.Add(48 * time.Hour), EntryFee: 100, Currency: "THB", RegistrationMode: models.TournamentRegistrationModeSolo, MinTeamSize: 1, MaxTeamSize: 1, Capacity: 1, Status: models.TournamentStatusRegistrationOpen, Published: true, Staff: []models.TournamentStaff{{AccountID: owner.ID, Role: models.TournamentStaffRoleOwner}}}
	if err := app.DB.Create(&tournament).Error; err != nil {
		t.Fatal(err)
	}
	path := "/tournaments/" + tournament.ID.String() + "/solo-registration"
	call := func(c *apptest.Client, method, endpoint string, body any, u *apptest.TestUser, expected int) *apptest.Response {
		t.Helper()
		resp, err := c.Do(method, endpoint, body, u.AuthHeaders())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != expected {
			t.Fatalf("%s %s: %d want %d: %s", method, endpoint, resp.StatusCode, expected, resp.Body)
		}
		return resp
	}
	var pending struct {
		Status    string    `json:"status"`
		PaymentID uuid.UUID `json:"paymentId"`
	}
	firstResponse := call(client, http.MethodPost, path, map[string]any{"answers": []any{}}, first, 200)
	if err := firstResponse.JSON(&pending); err != nil {
		t.Fatal(err)
	}
	if pending.Status != "PAYMENT_PENDING" || pending.PaymentID == uuid.Nil {
		t.Fatalf("unexpected pending result: %+v", pending)
	}
	var duplicate struct {
		PaymentID uuid.UUID `json:"paymentId"`
	}
	if err := call(client, http.MethodPost, path, map[string]any{"answers": []any{}}, first, 200).JSON(&duplicate); err != nil {
		t.Fatal(err)
	}
	if duplicate.PaymentID != pending.PaymentID {
		t.Fatal("duplicate submission created a new invoice")
	}
	call(client, http.MethodPost, path, map[string]any{"answers": []any{}}, second, 409)
	call(payments, http.MethodPost, "/"+pending.PaymentID.String()+"/mock-result", map[string]string{"result": "FAILED"}, first, 200)
	var failed struct {
		Status string `json:"status"`
	}
	if err := call(client, http.MethodPost, path+"/refresh", nil, first, 200).JSON(&failed); err != nil {
		t.Fatal(err)
	}
	if failed.Status != "PAYMENT_FAILED" {
		t.Fatalf("unexpected failed status: %s", failed.Status)
	}
	firstInvoice := pending.PaymentID
	if err := call(client, http.MethodPost, path, map[string]any{"answers": []any{}}, first, 200).JSON(&pending); err != nil {
		t.Fatal(err)
	}
	if pending.Status != "PAYMENT_PENDING" || pending.PaymentID == firstInvoice {
		t.Fatalf("retry did not create a fresh invoice: %+v", pending)
	}
	call(payments, http.MethodPost, "/"+pending.PaymentID.String()+"/mock-result", map[string]string{"result": "CANCELLED"}, first, 200)
	call(client, http.MethodPost, path+"/refresh", nil, first, 200)
	if err := call(client, http.MethodPost, path, map[string]any{"answers": []any{}}, second, 200).JSON(&pending); err != nil {
		t.Fatal(err)
	}
	call(payments, http.MethodPost, "/"+pending.PaymentID.String()+"/mock-result", map[string]string{"result": "SUCCEEDED"}, second, 200)
	var registered struct {
		Status string `json:"status"`
	}
	if err := call(client, http.MethodPost, path+"/refresh", nil, second, 200).JSON(&registered); err != nil {
		t.Fatal(err)
	}
	if registered.Status != "REGISTERED" {
		t.Fatalf("unexpected registered status: %s", registered.Status)
	}
	if err := call(client, http.MethodPost, path, map[string]any{"answers": []any{}}, second, 200).JSON(&registered); err != nil {
		t.Fatal(err)
	}
	if registered.Status != "REGISTERED" {
		t.Fatal("duplicate registration changed completed status")
	}
	call(client, http.MethodPost, path, map[string]any{"answers": []any{}}, first, 409)
	var details struct {
		Tournament struct {
			RegisteredCount int `json:"registeredCount"`
		} `json:"tournament"`
	}
	if err := call(client, http.MethodGet, "/tournaments/"+tournament.ID.String(), nil, first, 200).JSON(&details); err != nil {
		t.Fatal(err)
	}
	if details.Tournament.RegisteredCount != 1 {
		t.Fatalf("registered count = %d", details.Tournament.RegisteredCount)
	}
	free := tournament
	free.ID = uuid.New()
	free.Name = "Free solo cup"
	free.EntryFee = 0
	free.Staff = []models.TournamentStaff{{AccountID: owner.ID, Role: models.TournamentStaffRoleOwner}}
	if err := app.DB.Create(&free).Error; err != nil {
		t.Fatal(err)
	}
	freePath := "/tournaments/" + free.ID.String() + "/solo-registration"
	var freeResult struct {
		Status    string     `json:"status"`
		PaymentID *uuid.UUID `json:"paymentId"`
	}
	if err := call(client, http.MethodPost, freePath, map[string]any{"answers": []any{}}, first, 200).JSON(&freeResult); err != nil {
		t.Fatal(err)
	}
	if freeResult.Status != "REGISTERED" || freeResult.PaymentID != nil {
		t.Fatalf("unexpected free result: %+v", freeResult)
	}
	var eligibility struct {
		Reason      string `json:"reason"`
		CanRegister bool   `json:"canRegister"`
	}
	if err := call(client, http.MethodGet, freePath+"/eligibility", nil, first, 200).JSON(&eligibility); err != nil {
		t.Fatal(err)
	}
	if eligibility.Reason != "ALREADY_REGISTERED" || eligibility.CanRegister {
		t.Fatalf("unexpected eligibility: %+v", eligibility)
	}
	if err := call(client, http.MethodGet, freePath+"/eligibility", nil, second, 200).JSON(&eligibility); err != nil {
		t.Fatal(err)
	}
	if eligibility.Reason != "TOURNAMENT_FULL" || eligibility.CanRegister {
		t.Fatalf("unexpected full result: %+v", eligibility)
	}
	closed := tournament
	closed.ID = uuid.New()
	closed.Name = "Closed solo cup"
	closed.Status = models.TournamentStatusRegistrationClosed
	closed.Staff = []models.TournamentStaff{{AccountID: owner.ID, Role: models.TournamentStaffRoleOwner}}
	if err := app.DB.Create(&closed).Error; err != nil {
		t.Fatal(err)
	}
	closedPath := "/tournaments/" + closed.ID.String() + "/solo-registration"
	if err := call(client, http.MethodGet, closedPath+"/eligibility", nil, first, 200).JSON(&eligibility); err != nil {
		t.Fatal(err)
	}
	if eligibility.Reason != "REGISTRATION_CLOSED" || eligibility.CanRegister {
		t.Fatalf("unexpected closed result: %+v", eligibility)
	}
	call(client, http.MethodPost, closedPath, map[string]any{"answers": []any{}}, first, 409)
	race := free
	race.ID = uuid.New()
	race.Name = "One seat race"
	race.Staff = []models.TournamentStaff{{AccountID: owner.ID, Role: models.TournamentStaffRoleOwner}}
	if err := app.DB.Create(&race).Error; err != nil {
		t.Fatal(err)
	}
	racePath := "/tournaments/" + race.ID.String() + "/solo-registration"
	results := make(chan int, 2)
	var wait sync.WaitGroup
	for _, competitor := range []*apptest.TestUser{first, second} {
		wait.Add(1)
		go func(u *apptest.TestUser) {
			defer wait.Done()
			response, err := client.Do(http.MethodPost, racePath, map[string]any{"answers": []any{}}, u.AuthHeaders())
			if err != nil {
				results <- 0
				return
			}
			results <- response.StatusCode
		}(competitor)
	}
	wait.Wait()
	close(results)
	counts := map[int]int{}
	for status := range results {
		counts[status]++
	}
	if counts[200] != 1 || counts[409] != 1 {
		t.Fatalf("concurrent one-seat results: %v", counts)
	}
}
