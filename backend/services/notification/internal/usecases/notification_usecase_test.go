package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/notification/internal/repositories"
	"github.com/moomaideng/eventory/services/notification/models"
)

type storeStub struct {
	delivery *models.Delivery
	status   models.DeliveryStatus
	hash     string
}

func (store *storeStub) Claim(_ context.Context, eventID uuid.UUID, eventType, recipient, template, payloadHash string, occurredAt, _ time.Time, _ time.Duration) (*models.Delivery, repositories.ClaimResult, error) {
	if store.delivery != nil && store.hash != payloadHash {
		return store.delivery, repositories.PayloadConflict, nil
	}
	if store.delivery != nil && store.status == models.DeliveryStatusSent {
		return store.delivery, repositories.AlreadyHandled, nil
	}
	store.hash = payloadHash
	store.delivery = &models.Delivery{ID: uuid.New(), EventID: eventID, EventType: eventType, RecipientEmail: recipient, TemplateKey: template, PayloadHash: payloadHash, EventOccurredAt: occurredAt}
	store.status = models.DeliveryStatusSending
	return store.delivery, repositories.Claimed, nil
}

func TestProcessRejectsReusedEventIDWithChangedPayload(t *testing.T) {
	store := &storeStub{}
	mailer := &mailerStub{}
	useCase := NewNotificationUseCase(store, mailer, "https://eventory.example", time.Minute)
	event := testEvent()

	if outcome, err := useCase.Process(context.Background(), event, false); err != nil || outcome != OutcomeSent {
		t.Fatalf("first delivery failed: outcome=%v err=%v", outcome, err)
	}
	event.Data.Recipient.Email = "other@example.com"
	if outcome, err := useCase.Process(context.Background(), event, false); err == nil || outcome != OutcomePermanentFailure {
		t.Fatalf("mutated duplicate should fail permanently: outcome=%v err=%v", outcome, err)
	}
	if len(mailer.messages) != 1 {
		t.Fatalf("mutated duplicate sent an email: count=%d", len(mailer.messages))
	}
}
func (store *storeStub) MarkSent(_ context.Context, _ uuid.UUID, _ time.Time) error {
	store.status = models.DeliveryStatusSent
	return nil
}
func (store *storeStub) MarkRetry(_ context.Context, _ uuid.UUID, _ string) error {
	store.status = models.DeliveryStatusRetryScheduled
	return nil
}
func (store *storeStub) MarkFailed(_ context.Context, _ uuid.UUID, _ string) error {
	store.status = models.DeliveryStatusFailed
	return nil
}

type mailerStub struct {
	messages []EmailMessage
	err      error
}

func (mailer *mailerStub) Send(_ context.Context, message EmailMessage) error {
	mailer.messages = append(mailer.messages, message)
	return mailer.err
}

type permanentSendError struct{ error }

func (permanentSendError) Permanent() bool { return true }

func testEvent() events.RegistrationSubmitted {
	teamID := uuid.New()
	teamName := "Phoenix <One>"
	return events.RegistrationSubmitted{
		SpecVersion: events.EventSpecVersion, ID: uuid.New(), Type: events.RegistrationSubmittedType,
		Source: "tournament-service", OccurredAt: time.Now().UTC(), CorrelationID: uuid.New(),
		Data: events.RegistrationSubmittedData{
			RegistrationID: uuid.New(), TournamentID: uuid.New(), TournamentName: "Eventory Cup",
			RegistrationMode: events.RegistrationModeTeam, TeamID: &teamID, TeamName: &teamName,
			Recipient: events.EventRecipient{AccountID: uuid.New(), Email: "CAPTAIN@example.com", DisplayName: "Captain <Ace>", Role: events.RecipientRoleCaptain},
		},
	}
}

func TestProcessSendsOnceAndRendersSafeHTML(t *testing.T) {
	store := &storeStub{}
	mailer := &mailerStub{}
	useCase := NewNotificationUseCase(store, mailer, "https://eventory.example/", time.Minute)
	event := testEvent()
	outcome, err := useCase.Process(context.Background(), event, false)
	if err != nil || outcome != OutcomeSent {
		t.Fatalf("expected sent outcome, got outcome=%v err=%v", outcome, err)
	}
	if len(mailer.messages) != 1 || mailer.messages[0].To != "captain@example.com" {
		t.Fatalf("unexpected sent messages: %#v", mailer.messages)
	}
	if strings.Contains(mailer.messages[0].HTMLBody, "Captain <Ace>") || !strings.Contains(mailer.messages[0].HTMLBody, "Captain &lt;Ace&gt;") {
		t.Fatal("expected recipient name to be HTML escaped")
	}
	outcome, err = useCase.Process(context.Background(), event, false)
	if err != nil || outcome != OutcomeAlreadyHandled || len(mailer.messages) != 1 {
		t.Fatalf("duplicate should be skipped, outcome=%v err=%v count=%d", outcome, err, len(mailer.messages))
	}
}

func TestProcessClassifiesFailures(t *testing.T) {
	t.Run("transient schedules retry", func(t *testing.T) {
		store := &storeStub{}
		useCase := NewNotificationUseCase(store, &mailerStub{err: errors.New("timeout")}, "https://eventory.example", time.Minute)
		outcome, _ := useCase.Process(context.Background(), testEvent(), false)
		if outcome != OutcomeRetry || store.status != models.DeliveryStatusRetryScheduled {
			t.Fatalf("outcome=%v status=%s", outcome, store.status)
		}
	})
	t.Run("last transient attempt fails permanently", func(t *testing.T) {
		store := &storeStub{}
		useCase := NewNotificationUseCase(store, &mailerStub{err: errors.New("timeout")}, "https://eventory.example", time.Minute)
		outcome, _ := useCase.Process(context.Background(), testEvent(), true)
		if outcome != OutcomePermanentFailure || store.status != models.DeliveryStatusFailed {
			t.Fatalf("outcome=%v status=%s", outcome, store.status)
		}
	})
	t.Run("permanent SMTP failure does not retry", func(t *testing.T) {
		store := &storeStub{}
		mailer := &mailerStub{err: permanentSendError{errors.New("550 rejected")}}
		useCase := NewNotificationUseCase(store, mailer, "https://eventory.example", time.Minute)
		outcome, _ := useCase.Process(context.Background(), testEvent(), false)
		if outcome != OutcomePermanentFailure || store.status != models.DeliveryStatusFailed {
			t.Fatalf("outcome=%v status=%s", outcome, store.status)
		}
	})
}

func TestRenderRegistrationEmailStripsHeaderNewlines(t *testing.T) {
	event := testEvent()
	event.Data.TournamentName = "Cup\r\nBcc: attacker@example.com"
	message, err := RenderRegistrationEmail(event, "https://eventory.example")
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(message.Subject, "\r\n") {
		t.Fatalf("subject contains a newline: %q", message.Subject)
	}
}
