package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/notification/config"
	"github.com/moomaideng/eventory/services/notification/internal/adapters/rabbitmq"
)

func main() {
	email := flag.String("email", "captain@example.test", "recipient email captured by Mailpit")
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	broker, err := rabbitmq.Connect(cfg.BrokerURL, cfg.RetryDelays, cfg.ConsumerPrefetch)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer broker.Close()

	teamID := uuid.New()
	teamName := "Phoenix Squad"
	now := time.Now().UTC()
	event := events.RegistrationSubmitted{
		SpecVersion: events.EventSpecVersion, ID: uuid.New(), Type: events.RegistrationSubmittedType,
		Source: "notification-smoke-publisher", OccurredAt: now, CorrelationID: uuid.New(),
		Data: events.RegistrationSubmittedData{
			RegistrationID: uuid.New(), TournamentID: uuid.New(), TournamentName: "Eventory Cup",
			RegistrationMode: events.RegistrationModeTeam, TeamID: &teamID, TeamName: &teamName,
			Recipient: events.EventRecipient{AccountID: uuid.New(), Email: *email, DisplayName: "Demo Captain", Role: events.RecipientRoleCaptain},
		},
	}
	body, err := json.Marshal(event)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := broker.PublishRegistration(ctx, body); err != nil {
		log.Fatalf("failed to publish event: %v", err)
	}
	log.Printf("published registration.submitted event_id=%s", event.ID)
}
