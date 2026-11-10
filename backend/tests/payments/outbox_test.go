package payments_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/payment/config"
	paymentdb "github.com/moomaideng/eventory/services/payment/database"
	"github.com/moomaideng/eventory/services/payment/models"
	paymentv1 "github.com/moomaideng/eventory/services/payment/pkg/proto/payment/v1"
	"github.com/moomaideng/eventory/services/payment/server"
	"github.com/moomaideng/eventory/tests/internal/apptest"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// flakyPublisher records what reaches the broker and rejects the first
// attempt for each routing key, as an unreachable broker would.
type flakyPublisher struct {
	mu       sync.Mutex
	failed   map[string]bool
	received map[string][]json.RawMessage
}

func (p *flakyPublisher) Publish(_ context.Context, routingKey string, event any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.failed[routingKey] {
		p.failed[routingKey] = true
		return errors.New("broker unavailable")
	}
	p.received[routingKey] = append(p.received[routingKey], event.(json.RawMessage))
	return nil
}

func (p *flakyPublisher) Close() error { return nil }

func (p *flakyPublisher) count(routingKey string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.received[routingKey])
}

// An invoice outcome and its event commit together, a failed publish is
// retried, and expiry (which no request triggers) is announced exactly once.
func TestPaymentEventsAreDeliveredFromTheOutbox(t *testing.T) {
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
	publisher := &flakyPublisher{failed: map[string]bool{}, received: map[string][]json.RawMessage{}}
	app := server.NewApp(db, config.Config{Environment: "production", SupabaseURL: jwt.URL(), ServiceKey: key}, publisher)
	client := paymentClient(t, app.GRPC)

	payer := uuid.New()
	token := "Bearer " + jwt.SignJWT(payer, "payer@example.test")
	create := func(reference string, expiresAt time.Time) string {
		t.Helper()
		ctx := metadata.AppendToOutgoingContext(context.Background(), "x-payment-service-key", key)
		res, err := client.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{
			PayerAccountId: payer.String(), Source: "TEAM_REGISTRATION", Reference: reference, IdempotencyKey: reference,
			Description: "Entry fee", AmountMinor: 15000, Currency: "THB", ExpiresAt: timestamppb.New(expiresAt),
		})
		if err != nil {
			t.Fatal(err)
		}
		return res.GetPayment().GetId()
	}
	complete := func(id, result string) {
		t.Helper()
		r := httptest.NewRequest("POST", "/api/v1/payments/"+id+"/mock-result", strings.NewReader(`{"result":"`+result+`"}`))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		app.HTTP.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("complete %s: %d %s", result, w.Code, w.Body.String())
		}
	}

	future := time.Now().Add(time.Hour)
	paid := create("team-paid:attempt", future)
	complete(paid, "SUCCEEDED")
	complete(paid, "SUCCEEDED") // a replay must not queue a second event
	cancelled := create("team-cancelled:attempt", future)
	complete(cancelled, "FAILED")
	expired := create("team-expired:attempt", time.Now().Add(-time.Minute))
	create("team-open:attempt", future) // still payable: no event

	var queued int64
	db.Model(&models.OutboxEvent{}).Where("routing_key = ?", events.PaymentConfirmedType).Count(&queued)
	if queued != 1 {
		t.Fatalf("queued %d payment.confirmed events for one payment", queued)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go app.RunBackground(ctx)

	deadline := time.Now().Add(15 * time.Second)
	for publisher.count(events.PaymentConfirmedType) < 1 || publisher.count(events.PaymentCancelledType) < 1 || publisher.count(events.PaymentExpiredType) < 1 {
		if time.Now().After(deadline) {
			t.Fatalf("events not delivered: confirmed=%d cancelled=%d expired=%d",
				publisher.count(events.PaymentConfirmedType), publisher.count(events.PaymentCancelledType), publisher.count(events.PaymentExpiredType))
		}
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(2500 * time.Millisecond) // two more relay passes: nothing may repeat

	for routingKey, wantID := range map[string]string{events.PaymentConfirmedType: paid, events.PaymentCancelledType: cancelled, events.PaymentExpiredType: expired} {
		if got := publisher.count(routingKey); got != 1 {
			t.Fatalf("%s delivered %d times", routingKey, got)
		}
		var event struct {
			Data struct {
				PaymentID     string `json:"payment_id"`
				InvoiceSource string `json:"invoice_source"`
				Reference     string `json:"reference"`
				AmountMinor   int64  `json:"amount_minor"`
			} `json:"data"`
		}
		if err := json.Unmarshal(publisher.received[routingKey][0], &event); err != nil {
			t.Fatal(err)
		}
		if event.Data.PaymentID != wantID || event.Data.InvoiceSource != "TEAM_REGISTRATION" || event.Data.AmountMinor != 15000 || !strings.HasSuffix(event.Data.Reference, ":attempt") {
			t.Fatalf("%s carries %+v", routingKey, event.Data)
		}
	}

	var unpublished int64
	db.Model(&models.OutboxEvent{}).Where("published_at IS NULL").Count(&unpublished)
	if unpublished != 0 {
		t.Fatalf("%d events left unpublished", unpublished)
	}
}
