package events

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func validReviewedEvent() RegistrationReviewed {
	return RegistrationReviewed{
		SpecVersion:   EventSpecVersion,
		ID:            uuid.New(),
		Type:          RegistrationApprovedType,
		Source:        "tournament-service",
		OccurredAt:    time.Now().UTC(),
		CorrelationID: uuid.New(),
		Data: RegistrationReviewedData{
			TournamentID:   uuid.New(),
			TournamentName: "Cyber Championship",
			TeamID:         uuid.New(),
			TeamName:       "Phoenix Squad",
			Recipient: EventRecipient{
				AccountID:   uuid.New(),
				Email:       "captain@example.com",
				DisplayName: "Captain Ace",
				Role:        RecipientRoleCaptain,
			},
		},
	}
}

func TestRegistrationReviewedValidate(t *testing.T) {
	t.Run("valid approved", func(t *testing.T) {
		e := validReviewedEvent()
		if err := e.Validate(); err != nil {
			t.Fatalf("expected valid event, got %v", err)
		}
	})

	t.Run("valid rejected with reason", func(t *testing.T) {
		e := validReviewedEvent()
		e.Type = RegistrationRejectedType
		e.Data.Reason = "Team roster incomplete"
		if err := e.Validate(); err != nil {
			t.Fatalf("expected valid rejected event, got %v", err)
		}
	})

	t.Run("invalid spec version", func(t *testing.T) {
		e := validReviewedEvent()
		e.SpecVersion = "2.0"
		if err := e.Validate(); err == nil {
			t.Fatal("expected error for invalid spec version")
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		e := validReviewedEvent()
		e.Type = "unknown.type"
		if err := e.Validate(); err == nil {
			t.Fatal("expected error for invalid type")
		}
	})

	t.Run("missing tournament id", func(t *testing.T) {
		e := validReviewedEvent()
		e.Data.TournamentID = uuid.Nil
		if err := e.Validate(); err == nil {
			t.Fatal("expected error for missing tournament id")
		}
	})

	t.Run("missing recipient account id", func(t *testing.T) {
		e := validReviewedEvent()
		e.Data.Recipient.AccountID = uuid.Nil
		if err := e.Validate(); err == nil {
			t.Fatal("expected error for missing recipient account id")
		}
	})
}
