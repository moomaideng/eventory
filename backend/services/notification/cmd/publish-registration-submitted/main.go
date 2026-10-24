// Command publish-registration-submitted publishes one sample
// registration.submitted event. It is also the reference for how a real
// producer should trace a publish: start a span for the business operation,
// then publish with that span's context. The broker injects the trace context
// into the message headers, so the notification worker's span joins this trace.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/internal/telemetry"
	"github.com/moomaideng/eventory/services/notification/config"
	"github.com/moomaideng/eventory/services/notification/internal/adapters/rabbitmq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func main() {
	email := flag.String("email", "captain@example.test", "recipient email captured by Mailpit")
	flag.Parse()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Tracing is off unless OTEL_EXPORTER_OTLP_ENDPOINT is set and reachable from where this runs.
	shutdownTelemetry, err := telemetry.Setup(context.Background(), telemetry.Config{ServiceName: "notification-smoke-publisher", Environment: cfg.Environment, Endpoint: cfg.OTel.ExporterOTLPEndpoint})
	if err != nil {
		log.Printf("failed to set up telemetry; continuing without tracing: %v", err)
	}

	publishErr := run(cfg, *email)

	// Flush buffered spans before exiting; log.Fatal would skip this.
	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdownTelemetry(flushCtx); err != nil {
		log.Printf("telemetry shutdown failed: %v", err)
	}

	if publishErr != nil {
		log.Fatal(publishErr)
	}
}

func run(cfg config.Config, email string) error {
	broker, err := rabbitmq.Connect(cfg.BrokerURL, cfg.RetryDelays, cfg.ConsumerPrefetch)
	if err != nil {
		return err
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
			Recipient: events.EventRecipient{AccountID: uuid.New(), Email: email, DisplayName: "Demo Captain", Role: events.RecipientRoleCaptain},
		},
	}
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// In a real producer this span is the request or use case that submits the
	// registration, so the publish and the email delivery appear under it.
	ctx, span := otel.Tracer("notification-smoke-publisher").Start(ctx, "submit registration")
	defer span.End()
	if err := broker.PublishRegistration(ctx, body); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	log.Printf("published registration.submitted event_id=%s trace_id=%s", event.ID, span.SpanContext().TraceID())
	return nil
}
