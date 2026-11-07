package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/notification/models"
)

func testAnnouncementEvent() events.AnnouncementCreated {
	return events.AnnouncementCreated{
		SpecVersion: events.EventSpecVersion, ID: uuid.New(), Type: events.AnnouncementCreatedType,
		Source: "tournament-service", OccurredAt: time.Now().UTC(), CorrelationID: uuid.New(),
		Data: events.AnnouncementCreatedData{
			AnnouncementID: uuid.New(), TournamentID: uuid.New(), TournamentName: "Eventory Cup",
			Title: "Venue change", Body: "The final moves to room 301.", AuthorName: "Organizer Ace",
			Recipient: events.EventRecipient{AccountID: uuid.New(), Email: "PLAYER@example.com", DisplayName: "Player One"},
		},
	}
}

// failingInAppStub stands for a notification box that cannot be written to.
type failingInAppStub struct{ calls int }

func (c *failingInAppStub) Create(context.Context, *models.Notification) (*models.Notification, error) {
	c.calls++
	return nil, errors.New("notification database unavailable")
}

func TestProcessAnnouncementCreatedSendsEmailAndNotification(t *testing.T) {
	store := &storeStub{}
	mailer := &mailerStub{}
	inApp := &inAppCreatorStub{}
	useCase := NewNotificationUseCase(store, mailer, "https://eventory.example/", time.Minute).WithInAppCreator(inApp)
	event := testAnnouncementEvent()

	outcome, err := useCase.ProcessAnnouncementCreated(context.Background(), event, false)
	if err != nil || outcome != OutcomeSent {
		t.Fatalf("expected sent outcome, got outcome=%v err=%v", outcome, err)
	}

	if len(mailer.messages) != 1 {
		t.Fatalf("expected 1 email, got %d", len(mailer.messages))
	}
	message := mailer.messages[0]
	if message.To != "player@example.com" {
		t.Fatalf("expected a lower-cased recipient, got %q", message.To)
	}
	if message.Subject != "[Eventory Cup] Venue change" {
		t.Fatalf("unexpected subject %q", message.Subject)
	}
	wantURL := "https://eventory.example/tournaments/" + event.Data.TournamentID.String()
	if !strings.Contains(message.TextBody, wantURL) || !strings.Contains(message.HTMLBody, wantURL) {
		t.Fatalf("expected both bodies to link to %s", wantURL)
	}

	if len(inApp.created) != 1 {
		t.Fatalf("expected 1 in-app notification, got %d", len(inApp.created))
	}
	noti := inApp.created[0]
	if noti.UserID != event.Data.Recipient.AccountID || noti.Type != models.NotificationTypeAnnouncementCreated {
		t.Fatalf("unexpected in-app notification: %#v", noti)
	}
	if noti.Title != "Venue change" || noti.Link != "/tournaments/"+event.Data.TournamentID.String()+"/announcements" {
		t.Fatalf("unexpected title or link: %#v", noti)
	}
	if !strings.Contains(noti.Message, "Eventory Cup") || !strings.Contains(noti.Message, "The final moves to room 301.") {
		t.Fatalf("expected tournament name and text in the message, got %q", noti.Message)
	}

	// The same event again must not send or create anything a second time.
	outcome, err = useCase.ProcessAnnouncementCreated(context.Background(), event, false)
	if err != nil || outcome != OutcomeAlreadyHandled {
		t.Fatalf("duplicate should be skipped, outcome=%v err=%v", outcome, err)
	}
	if len(mailer.messages) != 1 || len(inApp.created) != 1 {
		t.Fatalf("duplicate must not resend: emails=%d notifications=%d", len(mailer.messages), len(inApp.created))
	}
}

func TestProcessAnnouncementCreatedWithoutEmailOnlyNotifiesInApp(t *testing.T) {
	store := &storeStub{}
	mailer := &mailerStub{}
	inApp := &inAppCreatorStub{}
	useCase := NewNotificationUseCase(store, mailer, "https://eventory.example/", time.Minute).WithInAppCreator(inApp)
	event := testAnnouncementEvent()
	event.Data.Recipient.Email = ""

	outcome, err := useCase.ProcessAnnouncementCreated(context.Background(), event, false)

	if err != nil || outcome != OutcomeSent {
		t.Fatalf("expected sent outcome, got outcome=%v err=%v", outcome, err)
	}
	if len(mailer.messages) != 0 {
		t.Fatalf("expected no email for a recipient without an address, got %#v", mailer.messages)
	}
	if store.delivery == nil || store.delivery.RecipientEmail != "" {
		t.Fatalf("expected no made-up address to be recorded, got %#v", store.delivery)
	}
	if len(inApp.created) != 1 {
		t.Fatalf("expected the in-app notification to be created, got %d", len(inApp.created))
	}
}

func TestProcessAnnouncementCreatedRejectsInvalidEvent(t *testing.T) {
	store := &storeStub{}
	mailer := &mailerStub{}
	inApp := &inAppCreatorStub{}
	useCase := NewNotificationUseCase(store, mailer, "https://eventory.example/", time.Minute).WithInAppCreator(inApp)
	event := testAnnouncementEvent()
	event.Data.Title = "   "

	outcome, err := useCase.ProcessAnnouncementCreated(context.Background(), event, false)

	if err == nil || outcome != OutcomePermanentFailure {
		t.Fatalf("expected a permanent failure, got outcome=%v err=%v", outcome, err)
	}
	if len(mailer.messages) != 0 || len(inApp.created) != 0 {
		t.Fatal("an invalid event must not send or create anything")
	}
}

func TestProcessAnnouncementCreatedClassifiesEmailFailures(t *testing.T) {
	cases := []struct {
		name         string
		sendErr      error
		finalAttempt bool
		want         Outcome
		wantStatus   models.DeliveryStatus
	}{
		{"transient failure is retried", errors.New("connection timed out"), false, OutcomeRetry, ""},
		{"transient failure on the last attempt gives up", errors.New("connection timed out"), true, OutcomePermanentFailure, models.DeliveryStatusFailed},
		{"permanent failure gives up at once", permanentSendError{errors.New("550 mailbox unavailable")}, false, OutcomePermanentFailure, models.DeliveryStatusFailed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &storeStub{}
			mailer := &mailerStub{err: tc.sendErr}
			inApp := &inAppCreatorStub{}
			useCase := NewNotificationUseCase(store, mailer, "https://eventory.example/", time.Minute).WithInAppCreator(inApp)

			outcome, err := useCase.ProcessAnnouncementCreated(context.Background(), testAnnouncementEvent(), tc.finalAttempt)

			if err == nil || outcome != tc.want {
				t.Fatalf("expected outcome %v with an error, got outcome=%v err=%v", tc.want, outcome, err)
			}
			if tc.wantStatus != "" && store.status != tc.wantStatus {
				t.Fatalf("expected delivery status %s, got %s", tc.wantStatus, store.status)
			}
			if len(inApp.created) != 0 {
				t.Fatal("no in-app notification until the email has gone out")
			}
		})
	}
}

func TestProcessAnnouncementCreatedSurvivesFailedNotificationBox(t *testing.T) {
	store := &storeStub{}
	mailer := &mailerStub{}
	inApp := &failingInAppStub{}
	useCase := NewNotificationUseCase(store, mailer, "https://eventory.example/", time.Minute).WithInAppCreator(inApp)

	outcome, err := useCase.ProcessAnnouncementCreated(context.Background(), testAnnouncementEvent(), false)

	if err != nil || outcome != OutcomeSent {
		t.Fatalf("the email already went out, so this stays sent: outcome=%v err=%v", outcome, err)
	}
	if inApp.calls != 1 || len(mailer.messages) != 1 {
		t.Fatalf("expected one attempt each, got notifications=%d emails=%d", inApp.calls, len(mailer.messages))
	}
}

func TestRenderAnnouncementEmailIsSafe(t *testing.T) {
	event := testAnnouncementEvent()
	event.Data.Title = "<b>Venue</b> change\r\nBcc: attacker@example.com"
	event.Data.Body = "<script>alert(1)</script>\nSecond line"

	message, err := RenderAnnouncementEmail(event, "https://eventory.example")
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if strings.ContainsAny(message.Subject, "\r\n") {
		t.Fatalf("subject contains a line break: %q", message.Subject)
	}
	if strings.Contains(message.HTMLBody, "<script>") || strings.Contains(message.HTMLBody, "<b>Venue</b>") {
		t.Fatal("expected the title and message to be HTML escaped")
	}
	if !strings.Contains(message.HTMLBody, "&lt;script&gt;") {
		t.Fatal("expected the escaped text to still be shown")
	}
	if !strings.Contains(message.TextBody, "Second line") {
		t.Fatal("expected the plain-text body to keep the message")
	}
}

func TestRenderAnnouncementEmailFallsBackWhenNamesAreMissing(t *testing.T) {
	event := testAnnouncementEvent()
	event.Data.Recipient.DisplayName = ""
	event.Data.AuthorName = ""

	message, err := RenderAnnouncementEmail(event, "https://eventory.example")
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(message.TextBody, "Hello there,") || !strings.Contains(message.TextBody, "The organizer posted") {
		t.Fatalf("expected readable fallbacks, got %q", message.TextBody)
	}
}

func TestAnnouncementPreview(t *testing.T) {
	if got := announcementPreview("Short text"); got != "Short text" {
		t.Fatalf("a short text stays as it is, got %q", got)
	}
	if got := announcementPreview("line one\n\nline   two"); got != "line one line two" {
		t.Fatalf("line breaks and repeated spaces collapse, got %q", got)
	}

	long := announcementPreview(strings.Repeat("a", 200))
	if long != strings.Repeat("a", 140)+"…" {
		t.Fatalf("a long text is cut at 140 characters, got %d runes", len([]rune(long)))
	}

	thai := announcementPreview(strings.Repeat("ก", 200))
	if len([]rune(thai)) != 141 || !strings.HasSuffix(thai, "…") {
		t.Fatalf("the limit counts characters, not bytes: got %d runes", len([]rune(thai)))
	}
}
