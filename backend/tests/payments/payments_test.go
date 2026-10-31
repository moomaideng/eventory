package payments_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/payment/config"
	paymentdb "github.com/moomaideng/eventory/services/payment/database"
	"github.com/moomaideng/eventory/services/payment/handlers/rest"
	"github.com/moomaideng/eventory/services/payment/models"
	"github.com/moomaideng/eventory/services/payment/server"
	"github.com/moomaideng/eventory/tests/internal/apptest"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// These checks use a disposable PostgreSQL database, not a mocked repository.
func TestPaymentAPI(t *testing.T) {
	pg := apptest.StartPostgres(context.Background(), t)
	db, err := gorm.Open(postgres.Open(pg.DSN), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := paymentdb.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	jwt := apptest.StartMockJWKSServer()
	key := "test-private-payment-service-key-123456"
	app := server.NewApp(db, config.Config{Environment: "production", SupabaseURL: jwt.URL(), ServiceKey: key})
	owner, other := uuid.New(), uuid.New()
	ownerToken := "Bearer " + jwt.SignJWT(owner, "payer@example.test")
	otherToken := "Bearer " + jwt.SignJWT(other, "other@example.test")
	request := func(method, path, body, token, serviceKey string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", token)
		r.Header.Set("X-Payment-Service-Key", serviceKey)
		w := httptest.NewRecorder()
		app.HTTP.ServeHTTP(w, r)
		return w
	}
	expect := func(t *testing.T, w *httptest.ResponseRecorder, code int) {
		t.Helper()
		if w.Code != code {
			t.Fatalf("Expected %d, got %d: %s", code, w.Code, w.Body.String())
		}
	}
	decode := func(t *testing.T, w *httptest.ResponseRecorder) rest.PaymentResponse {
		t.Helper()
		var payment rest.PaymentResponse
		if err := json.Unmarshal(w.Body.Bytes(), &payment); err != nil {
			t.Fatal(err)
		}
		return payment
	}
	deadline := time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond)
	body := func(reference, idempotency string, amount int64) string {
		return fmt.Sprintf(`{"payerAccountId":%q,"source":%q,"reference":%q,"idempotencyKey":%q,"description":"Sample fee","amountMinor":%d,"currency":"THB","expiresAt":%q}`, owner.String(), owner.String(), reference, idempotency, amount, deadline.Format(time.RFC3339Nano))
	}
	create := func(t *testing.T, reference, idempotency string) rest.PaymentResponse {
		t.Helper()
		w := request("POST", "/internal/v1/payments", body(reference, idempotency, 15000), "", key)
		expect(t, w, 200)
		return decode(t, w)
	}

	t.Run("trusted creation and validation", func(t *testing.T) {
		expect(t, request("POST", "/internal/v1/payments", body("a", "a", 100), ownerToken, ""), 401)
		expect(t, request("POST", "/api/v1/payments", body("a", "a", 100), ownerToken, ""), 405)
		for _, amount := range []int64{0, -1, 9007199254740992} {
			expect(t, request("POST", "/internal/v1/payments", body("invalid", "invalid", amount), "", key), 422)
		}
		expect(t, request("POST", "/internal/v1/payments", strings.Replace(body("a", "a", 100), "THB", "JPY", 1), "", key), 422)
		expect(t, request("POST", "/internal/v1/payments", strings.Replace(body("a", "a", 100), "Sample fee", "   ", 1), "", key), 422)
	})
	payment := create(t, "owned", "owned")
	path := "/api/v1/payments/" + payment.ID.String()
	t.Run("service status query", func(t *testing.T) {
		internalPath := "/internal/v1/payments/" + payment.ID.String()
		expect(t, request("GET", internalPath, "", ownerToken, ""), 401)
		w := request("GET", internalPath, "", "", key)
		expect(t, w, 200)
		if decode(t, w).ID != payment.ID {
			t.Fatal("Wrong payment returned to caller")
		}
	})
	t.Run("ownership and authentication", func(t *testing.T) {
		expect(t, request("GET", path, "", "", ""), 401)
		expect(t, request("GET", path, "", otherToken, ""), 404)
		expect(t, request("POST", path+"/mock-result", `{"result":"SUCCEEDED"}`, otherToken, ""), 404)
		expect(t, request("GET", path, "", "Bearer dev-token", ""), 401)
		w := request("GET", path, "", ownerToken, "")
		expect(t, w, 200)
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Error("Private history must not be cached")
		}
	})
	t.Run("idempotent create and reference protection", func(t *testing.T) {
		retry := create(t, "owned", "owned")
		if retry.ID != payment.ID {
			t.Fatal("Retry created another payment")
		}
		expect(t, request("POST", "/internal/v1/payments", body("owned", "owned", 15001), "", key), 409)
		expect(t, request("POST", "/internal/v1/payments", body("owned", "different-key", 15000), "", key), 409)
	})
	t.Run("final result cannot change", func(t *testing.T) {
		expect(t, request("POST", path+"/mock-result", `{"result":"PENDING"}`, ownerToken, ""), 422)
		w := request("POST", path+"/mock-result", `{"result":"SUCCEEDED"}`, ownerToken, "")
		expect(t, w, 200)
		completed := decode(t, w)
		if completed.Status != models.Paid || completed.PaidAt == nil || completed.Provider != "MOCK" {
			t.Fatalf("Wrong result: %+v", completed)
		}
		w = request("POST", path+"/mock-result", `{"result":"SUCCEEDED"}`, ownerToken, "")
		expect(t, w, 200)
		if !decode(t, w).PaidAt.Equal(*completed.PaidAt) {
			t.Fatal("Replay changed completion time")
		}
		expect(t, request("POST", path+"/mock-result", `{"result":"FAILED"}`, ownerToken, ""), 409)
		expect(t, request("POST", "/internal/v1/payments", body("owned", "new-key", 15000), "", key), 409)
	})
	t.Run("rejected registration refund", func(t *testing.T) {
		entryID := uuid.New()
		reference := entryID.String() + ":" + uuid.NewString()
		createBody := fmt.Sprintf(`{"payerAccountId":%q,"source":"TEAM_REGISTRATION","reference":%q,"idempotencyKey":%q,"description":"Team registration","amountMinor":15000,"currency":"THB","expiresAt":%q}`, owner.String(), reference, uuid.NewString(), deadline.Format(time.RFC3339Nano))
		w := request("POST", "/internal/v1/payments", createBody, "", key)
		expect(t, w, 200)
		invoice := decode(t, w)
		refundPath := "/internal/v1/payments/" + invoice.ID.String() + "/refund"
		refundBody := fmt.Sprintf(`{"entryId":%q}`, entryID.String())
		expect(t, request("POST", refundPath, refundBody, ownerToken, ""), 401)
		expect(t, request("POST", refundPath, refundBody, "", key), 409)
		expect(t, request("POST", refundPath, fmt.Sprintf(`{"entryId":%q}`, uuid.NewString()), "", key), 409)
		expect(t, request("POST", "/api/v1/payments/"+invoice.ID.String()+"/mock-result", `{"result":"SUCCEEDED"}`, ownerToken, ""), 200)
		expect(t, request("POST", refundPath, fmt.Sprintf(`{"entryId":%q}`, uuid.NewString()), "", key), 409)
		expect(t, request("POST", "/internal/v1/payments/"+payment.ID.String()+"/refund", refundBody, "", key), 409)
		w = request("POST", refundPath, refundBody, "", key)
		expect(t, w, 200)
		refunded := decode(t, w)
		if refunded.Status != models.Refunded || refunded.RefundedAt == nil || refunded.PaidAt == nil {
			t.Fatalf("wrong refund result: %+v", refunded)
		}
		w = request("POST", refundPath, refundBody, "", key)
		expect(t, w, 200)
		if got := decode(t, w); got.RefundedAt == nil || !got.RefundedAt.Equal(*refunded.RefundedAt) {
			t.Fatal("duplicate request changed refund timestamp")
		}
		w = request("GET", "/api/v1/payments/"+invoice.ID.String(), "", ownerToken, "")
		expect(t, w, 200)
		if decode(t, w).Status != models.Refunded {
			t.Fatal("payer cannot see refunded status")
		}
		expect(t, request("GET", "/api/v1/payments/"+invoice.ID.String()+"/receipt", "", ownerToken, ""), 200)
	})
	t.Run("failed and cancelled attempts remain retryable", func(t *testing.T) {
		for _, status := range []models.MockResult{models.Failed, models.Cancelled} {
			ref := string(status)
			first := create(t, ref, ref+"-1")
			url := "/api/v1/payments/" + first.ID.String()
			w := request("POST", url+"/mock-result", fmt.Sprintf(`{"result":%q}`, status), ownerToken, "")
			expect(t, w, 200)
			if decode(t, w).Status != models.Unpaid {
				t.Fatal("failure must leave invoice unpaid")
			}
			expect(t, request("GET", url+"/receipt", "", ownerToken, ""), 404)
			expect(t, request("GET", url+"/receipt.pdf", "", ownerToken, ""), 404)
			if create(t, ref, ref+"-1").ID != first.ID {
				t.Fatal("retry created another invoice")
			}
			expect(t, request("POST", url+"/mock-result", `{"result":"SUCCEEDED"}`, ownerToken, ""), 200)
		}
	})
	t.Run("concurrent creation", func(t *testing.T) {
		results := make(chan *httptest.ResponseRecorder, 8)
		for range 8 {
			go func() {
				results <- request("POST", "/internal/v1/payments", body("parallel", "parallel", 15000), "", key)
			}()
		}
		var id uuid.UUID
		for range 8 {
			w := <-results
			expect(t, w, 200)
			p := decode(t, w)
			if id != uuid.Nil && id != p.ID {
				t.Fatal("Parallel retries produced multiple payments")
			}
			id = p.ID
		}
	})
	t.Run("concurrent success creates one receipt", func(t *testing.T) {
		p := create(t, "race", "race")
		results := make(chan *httptest.ResponseRecorder, 8)
		for range 8 {
			go func() {
				results <- request("POST", "/api/v1/payments/"+p.ID.String()+"/mock-result", `{"result":"SUCCEEDED"}`, ownerToken, "")
			}()
		}
		var paidAt *time.Time
		for range 8 {
			w := <-results
			expect(t, w, 200)
			invoice := decode(t, w)
			if invoice.PaidAt == nil || (paidAt != nil && !paidAt.Equal(*invoice.PaidAt)) {
				t.Fatal("success replay changed payment time")
			}
			paidAt = invoice.PaidAt
		}
		var count int64
		if err := db.Model(&models.Receipt{}).Where("invoice_id = ?", p.ID).Count(&count).Error; err != nil || count != 1 {
			t.Fatalf("receipt count %d, error %v", count, err)
		}
	})
	t.Run("different keys cannot create two active attempts for one reference", func(t *testing.T) {
		results := make(chan *httptest.ResponseRecorder, 8)
		for i := range 8 {
			go func() {
				results <- request("POST", "/internal/v1/payments", body("parallel-reference", fmt.Sprintf("key-%d", i), 15000), "", key)
			}()
		}
		wins := 0
		for range 8 {
			w := <-results
			if w.Code == 200 {
				wins++
			} else {
				expect(t, w, 409)
			}
		}
		if wins != 1 {
			t.Fatalf("Expected one active attempt, got %d", wins)
		}
	})
	t.Run("expiry is derived and enforced without writes", func(t *testing.T) {
		p := create(t, "expired", "expired")
		expires := time.Now().Add(-time.Second).UTC().Truncate(time.Microsecond)
		if err := db.Model(&models.Invoice{}).Where("id = ?", p.ID).Update("expires_at", expires).Error; err != nil {
			t.Fatal(err)
		}
		var before, after models.Invoice
		db.First(&before, "id = ?", p.ID)
		url := "/api/v1/payments/" + p.ID.String()
		w := request("GET", url, "", ownerToken, "")
		expect(t, w, 200)
		if decode(t, w).Status != models.Expired {
			t.Fatal("expired invoice not derived")
		}
		db.First(&after, "id = ?", p.ID)
		if !before.UpdatedAt.Equal(after.UpdatedAt) || after.PaidAt != nil {
			t.Fatal("viewing changed invoice")
		}
		expect(t, request("POST", url+"/mock-result", `{"result":"SUCCEEDED"}`, ownerToken, ""), 409)
		expect(t, request("GET", url+"/receipt.pdf", "", ownerToken, ""), 404)
		payload := strings.Replace(body("expired", "expired", 15000), deadline.Format(time.RFC3339Nano), expires.Format(time.RFC3339Nano), 1)
		expect(t, request("POST", "/internal/v1/payments", payload, "", key), 200)
		payload = strings.Replace(payload, `"expired"`, `"new-expired"`, -1)
		expect(t, request("POST", "/internal/v1/payments", payload, "", key), 422)
		if db.Migrator().HasColumn(&models.Invoice{}, "status") {
			t.Fatal("derived status must not be stored")
		}
	})
	t.Run("deadline is checked after waiting for the invoice lock", func(t *testing.T) {
		p := create(t, "lock-expiry", "lock-expiry")
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatal(tx.Error)
		}
		defer tx.Rollback()
		var locked models.Invoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "id = ?", p.ID).Error; err != nil {
			t.Fatal(err)
		}
		result := make(chan *httptest.ResponseRecorder, 1)
		go func() {
			result <- request("POST", "/api/v1/payments/"+p.ID.String()+"/mock-result", `{"result":"SUCCEEDED"}`, ownerToken, "")
		}()
		waiting := false
		until := time.Now().Add(5 * time.Second)
		for time.Now().Before(until) {
			if err := db.Raw(`SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE wait_event_type = 'Lock' AND query LIKE '%payment_invoices%')`).Scan(&waiting).Error; err != nil {
				t.Fatal(err)
			}
			if waiting {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if !waiting {
			t.Fatal("payment request did not wait for invoice lock")
		}
		// Move the deadline into the past while the payment request is blocked.
		if err := tx.Model(&locked).Update("expires_at", time.Now().Add(-time.Second)).Error; err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatal(err)
		}
		expect(t, <-result, 409)
		var count int64
		if err := db.Model(&models.Receipt{}).Where("invoice_id = ?", p.ID).Count(&count).Error; err != nil || count != 0 {
			t.Fatalf("expired invoice produced a receipt: %d, %v", count, err)
		}
	})
	t.Run("owned immutable receipt and direct PDF", func(t *testing.T) {
		var receipt rest.ReceiptResponse
		w := request("GET", path+"/receipt", "", ownerToken, "")
		expect(t, w, 200)
		if err := json.Unmarshal(w.Body.Bytes(), &receipt); err != nil {
			t.Fatal(err)
		}
		if receipt.InvoiceID != payment.ID || receipt.AmountMinor != 15000 || receipt.PayerAccountID != owner || receipt.ID == uuid.Nil {
			t.Fatalf("wrong receipt: %+v", receipt)
		}
		for _, suffix := range []string{"/receipt", "/receipt.pdf"} {
			expect(t, request("GET", path+suffix, "", "", ""), 401)
			expect(t, request("GET", path+suffix, "", otherToken, ""), 404)
		}
		w = request("GET", path+"/receipt.pdf", "", ownerToken, "")
		expect(t, w, 200)
		if w.Header().Get("Content-Type") != "application/pdf" || !strings.HasPrefix(w.Body.String(), "%PDF-") || !strings.Contains(w.Header().Get("Content-Disposition"), receipt.ID.String()+".pdf") {
			t.Fatal("invalid PDF response")
		}
		if err := db.Model(&models.Invoice{}).Where("id = ?", payment.ID).Update("expires_at", time.Now().Add(-time.Hour)).Error; err != nil {
			t.Fatal(err)
		}
		w = request("POST", path+"/mock-result", `{"result":"SUCCEEDED"}`, ownerToken, "")
		expect(t, w, 200)
		if decode(t, w).Status != models.Paid {
			t.Fatal("paid invoice expired")
		}
		w = request("GET", path+"/receipt", "", ownerToken, "")
		expect(t, w, 200)
		var replay rest.ReceiptResponse
		if err := json.Unmarshal(w.Body.Bytes(), &replay); err != nil {
			t.Fatal(err)
		}
		if replay.ID != receipt.ID || !replay.PaidAt.Equal(receipt.PaidAt) {
			t.Fatal("receipt changed on replay")
		}
	})
	t.Run("history filters pagination and isolation", func(t *testing.T) {
		date := time.Date(2020, 2, 29, 23, 59, 59, 0, time.UTC)
		old := models.Invoice{ID: uuid.New(), PayerAccountID: owner, Source: owner.String(), Reference: "old", IdempotencyKey: "old", Description: "Old invoice", AmountMinor: 100, Currency: "THB", Provider: "MOCK", ExpiresAt: deadline, CreatedAt: date}
		if err := db.Create(&old).Error; err != nil {
			t.Fatal(err)
		}
		otherPayment := old
		otherPayment.ID, otherPayment.PayerAccountID, otherPayment.Reference, otherPayment.IdempotencyKey = uuid.New(), other, "other", "other"
		if err := db.Create(&otherPayment).Error; err != nil {
			t.Fatal(err)
		}
		var history struct {
			Items []rest.PaymentResponse `json:"items"`
			Total int64                  `json:"total"`
		}
		w := request("GET", "/api/v1/payments?from=2020-02-29&to=2020-02-29", "", ownerToken, "")
		expect(t, w, 200)
		if err := json.Unmarshal(w.Body.Bytes(), &history); err != nil {
			t.Fatal(err)
		}
		if history.Total != 1 || len(history.Items) != 1 || history.Items[0].ID != old.ID {
			t.Fatalf("Wrong filtered history: %+v", history)
		}
		w = request("GET", "/api/v1/payments?limit=1&page=2", "", ownerToken, "")
		expect(t, w, 200)
		if err := json.Unmarshal(w.Body.Bytes(), &history); err != nil {
			t.Fatal(err)
		}
		if len(history.Items) != 1 || history.Total < 2 {
			t.Fatal("Pagination did not keep total")
		}
		for _, query := range []string{"from=2025-02-29", "from=2021-01-01&to=2020-01-01", "page=0", "limit=101"} {
			expect(t, request("GET", "/api/v1/payments?"+query, "", ownerToken, ""), 422)
		}
		w = request("GET", "/api/v1/payments?from=2020-02-29&to=2020-02-29", "", otherToken, "")
		expect(t, w, 200)
		if err := json.Unmarshal(w.Body.Bytes(), &history); err != nil {
			t.Fatal(err)
		}
		if history.Total != 1 || history.Items[0].ID != otherPayment.ID {
			t.Fatal("Another user's payment leaked")
		}
	})
}
