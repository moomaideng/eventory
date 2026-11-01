package events

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	RegistrationApprovedType = "registration.approved"
	RegistrationRejectedType = "registration.rejected"
)

// RegistrationReviewed is published by Tournament Service when an organizer
// accepts or rejects a team or solo registration application.
type RegistrationReviewed struct {
	SpecVersion   string                   `json:"spec_version"`
	ID            uuid.UUID                `json:"id"`
	Type          string                   `json:"type"` // RegistrationApprovedType or RegistrationRejectedType
	Source        string                   `json:"source"`
	OccurredAt    time.Time                `json:"occurred_at"`
	CorrelationID uuid.UUID                `json:"correlation_id"`
	Data          RegistrationReviewedData `json:"data"`
}

type RegistrationReviewedData struct {
	TournamentID   uuid.UUID      `json:"tournament_id"`
	TournamentName string         `json:"tournament_name"`
	TeamID         uuid.UUID      `json:"team_id"`
	TeamName       string         `json:"team_name"`
	Recipient      EventRecipient `json:"recipient"`
	Reason         string         `json:"reason,omitempty"`
}

func (event *RegistrationReviewed) Normalize() {
	event.Source = strings.TrimSpace(event.Source)
	event.Data.TournamentName = strings.TrimSpace(event.Data.TournamentName)
	event.Data.TeamName = strings.TrimSpace(event.Data.TeamName)
	event.Data.Recipient.Email = strings.ToLower(strings.TrimSpace(event.Data.Recipient.Email))
	event.Data.Recipient.DisplayName = strings.TrimSpace(event.Data.Recipient.DisplayName)
	event.Data.Reason = strings.TrimSpace(event.Data.Reason)
}

func (event RegistrationReviewed) Validate() error {
	if event.SpecVersion != EventSpecVersion {
		return fmt.Errorf("unsupported spec_version %q", event.SpecVersion)
	}
	if event.Type != RegistrationApprovedType && event.Type != RegistrationRejectedType {
		return fmt.Errorf("unexpected event type %q", event.Type)
	}
	if event.ID == uuid.Nil || event.CorrelationID == uuid.Nil {
		return errors.New("event id and correlation_id are required")
	}
	if event.Source == "" || event.OccurredAt.IsZero() {
		return errors.New("source and occurred_at are required")
	}
	if event.Data.TournamentID == uuid.Nil || event.Data.TeamID == uuid.Nil {
		return errors.New("tournament_id and team_id are required")
	}
	if event.Data.TournamentName == "" || event.Data.TeamName == "" {
		return errors.New("tournament_name and team_name are required")
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
