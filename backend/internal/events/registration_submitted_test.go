package events

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func validRegistrationEvent() RegistrationSubmitted {
	teamID := uuid.New()
	teamName := "Phoenix"
	return RegistrationSubmitted{
		SpecVersion: EventSpecVersion, ID: uuid.New(), Type: RegistrationSubmittedType,
		Source: "tournament-service", OccurredAt: time.Now().UTC(), CorrelationID: uuid.New(),
		Data: RegistrationSubmittedData{
			RegistrationID: uuid.New(), TournamentID: uuid.New(), TournamentName: "Eventory Cup",
			RegistrationMode: RegistrationModeTeam, TeamID: &teamID, TeamName: &teamName,
			Recipient: EventRecipient{AccountID: uuid.New(), Email: "captain@example.com", DisplayName: "Captain", Role: RecipientRoleCaptain},
		},
	}
}

func TestRegistrationSubmittedValidate(t *testing.T) {
	event := validRegistrationEvent()
	if err := event.Validate(); err != nil {
		t.Fatalf("expected valid event, got %v", err)
	}
	event.Data.Recipient.Role = RecipientRoleApplicant
	if err := event.Validate(); err == nil {
		t.Fatal("expected team recipient role validation error")
	}
	event = validRegistrationEvent()
	event.Data.RegistrationMode = RegistrationModeSolo
	event.Data.TeamID = nil
	event.Data.TeamName = nil
	event.Data.Recipient.Role = RecipientRoleApplicant
	if err := event.Validate(); err != nil {
		t.Fatalf("expected valid solo event, got %v", err)
	}
}
