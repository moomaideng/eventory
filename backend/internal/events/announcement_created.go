package events

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

const AnnouncementCreatedType = "announcement.created"

// AnnouncementCreated is published by Tournament Service after an organizer or
// admin posts an announcement (US3-4). One event is published per recipient,
// so every event in one announcement shares a CorrelationID but has its own ID.
type AnnouncementCreated struct {
	SpecVersion   string                  `json:"spec_version"`
	ID            uuid.UUID               `json:"id"`
	Type          string                  `json:"type"`
	Source        string                  `json:"source"`
	OccurredAt    time.Time               `json:"occurred_at"`
	CorrelationID uuid.UUID               `json:"correlation_id"`
	Data          AnnouncementCreatedData `json:"data"`
}

type AnnouncementCreatedData struct {
	AnnouncementID uuid.UUID      `json:"announcement_id"`
	TournamentID   uuid.UUID      `json:"tournament_id"`
	TournamentName string         `json:"tournament_name"`
	Title          string         `json:"title"`
	Body           string         `json:"body"`
	AuthorName     string         `json:"author_name"`
	Recipient      EventRecipient `json:"recipient"`
}

func (event *AnnouncementCreated) Normalize() {
	event.Source = strings.TrimSpace(event.Source)
	event.Data.TournamentName = strings.TrimSpace(event.Data.TournamentName)
	event.Data.Title = strings.TrimSpace(event.Data.Title)
	event.Data.Body = strings.TrimSpace(event.Data.Body)
	event.Data.AuthorName = strings.TrimSpace(event.Data.AuthorName)
	event.Data.Recipient.Email = strings.ToLower(strings.TrimSpace(event.Data.Recipient.Email))
	event.Data.Recipient.DisplayName = strings.TrimSpace(event.Data.Recipient.DisplayName)
}

func (event AnnouncementCreated) Validate() error {
	if event.SpecVersion != EventSpecVersion {
		return fmt.Errorf("unsupported spec_version %q", event.SpecVersion)
	}
	if event.Type != AnnouncementCreatedType {
		return fmt.Errorf("unexpected event type %q", event.Type)
	}
	if event.ID == uuid.Nil || event.CorrelationID == uuid.Nil {
		return errors.New("event id and correlation_id are required")
	}
	if event.Source == "" || event.OccurredAt.IsZero() {
		return errors.New("source and occurred_at are required")
	}
	if event.Data.AnnouncementID == uuid.Nil || event.Data.TournamentID == uuid.Nil {
		return errors.New("announcement_id and tournament_id are required")
	}
	if event.Data.TournamentName == "" || event.Data.Title == "" || event.Data.Body == "" {
		return errors.New("tournament_name, title and body are required")
	}
	if event.Data.Recipient.AccountID == uuid.Nil {
		return errors.New("recipient account_id is required")
	}
	if event.Data.Recipient.Email != "" {
		address, err := mail.ParseAddress(event.Data.Recipient.Email)
		if err != nil || !strings.EqualFold(address.Address, event.Data.Recipient.Email) {
			return errors.New("recipient email is invalid")
		}
	}
	return nil
}
