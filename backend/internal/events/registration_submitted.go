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
	RegistrationSubmittedType = "registration.submitted"
	EventSpecVersion          = "1.0"
)

type RegistrationMode string

const (
	RegistrationModeSolo RegistrationMode = "SOLO"
	RegistrationModeTeam RegistrationMode = "TEAM"
)

type RecipientRole string

const (
	RecipientRoleApplicant RecipientRole = "APPLICANT"
	RecipientRoleCaptain   RecipientRole = "CAPTAIN"
)

// RegistrationSubmitted is the versioned integration event consumed by the
// Notification Service. Producers must publish it only after the registration
// transaction has committed successfully.
type RegistrationSubmitted struct {
	SpecVersion   string                    `json:"spec_version"`
	ID            uuid.UUID                 `json:"id"`
	Type          string                    `json:"type"`
	Source        string                    `json:"source"`
	OccurredAt    time.Time                 `json:"occurred_at"`
	CorrelationID uuid.UUID                 `json:"correlation_id"`
	Data          RegistrationSubmittedData `json:"data"`
}

type RegistrationSubmittedData struct {
	RegistrationID   uuid.UUID        `json:"registration_id"`
	TournamentID     uuid.UUID        `json:"tournament_id"`
	TournamentName   string           `json:"tournament_name"`
	RegistrationMode RegistrationMode `json:"registration_mode"`
	TeamID           *uuid.UUID       `json:"team_id,omitempty"`
	TeamName         *string          `json:"team_name,omitempty"`
	Recipient        EventRecipient   `json:"recipient"`
}

type EventRecipient struct {
	AccountID   uuid.UUID     `json:"account_id"`
	Email       string        `json:"email"`
	DisplayName string        `json:"display_name"`
	Role        RecipientRole `json:"role"`
}

func (event *RegistrationSubmitted) Normalize() {
	event.Source = strings.TrimSpace(event.Source)
	event.Data.TournamentName = strings.TrimSpace(event.Data.TournamentName)
	event.Data.Recipient.Email = strings.ToLower(strings.TrimSpace(event.Data.Recipient.Email))
	event.Data.Recipient.DisplayName = strings.TrimSpace(event.Data.Recipient.DisplayName)
	if event.Data.TeamName != nil {
		trimmed := strings.TrimSpace(*event.Data.TeamName)
		event.Data.TeamName = &trimmed
	}
}

func (event RegistrationSubmitted) Validate() error {
	if event.SpecVersion != EventSpecVersion {
		return fmt.Errorf("unsupported spec_version %q", event.SpecVersion)
	}
	if event.Type != RegistrationSubmittedType {
		return fmt.Errorf("unexpected event type %q", event.Type)
	}
	if event.ID == uuid.Nil || event.CorrelationID == uuid.Nil {
		return errors.New("event id and correlation_id are required")
	}
	if event.Source == "" || event.OccurredAt.IsZero() {
		return errors.New("source and occurred_at are required")
	}
	if event.Data.RegistrationID == uuid.Nil || event.Data.TournamentID == uuid.Nil {
		return errors.New("registration_id and tournament_id are required")
	}
	if event.Data.TournamentName == "" {
		return errors.New("tournament_name is required")
	}
	if event.Data.Recipient.AccountID == uuid.Nil || event.Data.Recipient.DisplayName == "" {
		return errors.New("recipient account_id and display_name are required")
	}
	address, err := mail.ParseAddress(event.Data.Recipient.Email)
	if err != nil || !strings.EqualFold(address.Address, event.Data.Recipient.Email) {
		return errors.New("recipient email is invalid")
	}
	switch event.Data.RegistrationMode {
	case RegistrationModeSolo:
		if event.Data.Recipient.Role != RecipientRoleApplicant {
			return errors.New("solo registration recipient must have APPLICANT role")
		}
	case RegistrationModeTeam:
		if event.Data.TeamID == nil || *event.Data.TeamID == uuid.Nil || event.Data.TeamName == nil || *event.Data.TeamName == "" {
			return errors.New("team_id and team_name are required for team registration")
		}
		if event.Data.Recipient.Role != RecipientRoleCaptain {
			return errors.New("team registration recipient must have CAPTAIN role")
		}
	default:
		return fmt.Errorf("unsupported registration_mode %q", event.Data.RegistrationMode)
	}
	return nil
}
