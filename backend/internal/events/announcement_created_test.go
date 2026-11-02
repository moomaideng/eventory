package events

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func validAnnouncementEvent() AnnouncementCreated {
	return AnnouncementCreated{
		SpecVersion:   EventSpecVersion,
		ID:            uuid.New(),
		Type:          AnnouncementCreatedType,
		Source:        "tournament-service",
		OccurredAt:    time.Now().UTC(),
		CorrelationID: uuid.New(),
		Data: AnnouncementCreatedData{
			AnnouncementID: uuid.New(),
			TournamentID:   uuid.New(),
			TournamentName: "Cyber Championship",
			Title:          "Venue change",
			Body:           "The final moves to room 301.",
			AuthorName:     "Organizer Ace",
			Recipient: EventRecipient{
				AccountID:   uuid.New(),
				Email:       "player@example.com",
				DisplayName: "Player One",
			},
		},
	}
}

func TestAnnouncementCreatedValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		e := validAnnouncementEvent()
		if err := e.Validate(); err != nil {
			t.Fatalf("expected valid event, got %v", err)
		}
	})

	t.Run("valid without recipient email", func(t *testing.T) {
		e := validAnnouncementEvent()
		e.Data.Recipient.Email = ""
		if err := e.Validate(); err != nil {
			t.Fatalf("expected an event without email to stay valid, got %v", err)
		}
	})

	cases := []struct {
		name   string
		break_ func(e *AnnouncementCreated)
	}{
		{"unsupported spec version", func(e *AnnouncementCreated) { e.SpecVersion = "2.0" }},
		{"wrong type", func(e *AnnouncementCreated) { e.Type = "unknown.type" }},
		{"missing event id", func(e *AnnouncementCreated) { e.ID = uuid.Nil }},
		{"missing correlation id", func(e *AnnouncementCreated) { e.CorrelationID = uuid.Nil }},
		{"missing source", func(e *AnnouncementCreated) { e.Source = "" }},
		{"missing occurred at", func(e *AnnouncementCreated) { e.OccurredAt = time.Time{} }},
		{"missing announcement id", func(e *AnnouncementCreated) { e.Data.AnnouncementID = uuid.Nil }},
		{"missing tournament id", func(e *AnnouncementCreated) { e.Data.TournamentID = uuid.Nil }},
		{"missing tournament name", func(e *AnnouncementCreated) { e.Data.TournamentName = "" }},
		{"missing title", func(e *AnnouncementCreated) { e.Data.Title = "" }},
		{"missing body", func(e *AnnouncementCreated) { e.Data.Body = "" }},
		{"missing recipient account id", func(e *AnnouncementCreated) { e.Data.Recipient.AccountID = uuid.Nil }},
		{"invalid recipient email", func(e *AnnouncementCreated) { e.Data.Recipient.Email = "not-an-email" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := validAnnouncementEvent()
			tc.break_(&e)
			if err := e.Validate(); err == nil {
				t.Fatalf("expected an error for %s", tc.name)
			}
		})
	}
}

func TestAnnouncementCreatedNormalize(t *testing.T) {
	e := validAnnouncementEvent()
	e.Source = "  tournament-service  "
	e.Data.Title = "  Venue change  "
	e.Data.Body = "\n The final moves to room 301. \n"
	e.Data.Recipient.Email = "  Player@Example.COM "

	e.Normalize()

	if e.Source != "tournament-service" {
		t.Errorf("source not trimmed: %q", e.Source)
	}
	if e.Data.Title != "Venue change" {
		t.Errorf("title not trimmed: %q", e.Data.Title)
	}
	if e.Data.Body != "The final moves to room 301." {
		t.Errorf("body not trimmed: %q", e.Data.Body)
	}
	if e.Data.Recipient.Email != "player@example.com" {
		t.Errorf("email not lowercased and trimmed: %q", e.Data.Recipient.Email)
	}
	if err := e.Validate(); err != nil {
		t.Fatalf("expected a normalized event to be valid, got %v", err)
	}
}

func TestAnnouncementCreatedBlankTitleFailsAfterNormalize(t *testing.T) {
	e := validAnnouncementEvent()
	e.Data.Title = "   "

	e.Normalize()

	if err := e.Validate(); err == nil {
		t.Fatal("expected a title of only spaces to be rejected once normalized")
	}
}
