package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/notification/internal/repositories"
	"github.com/moomaideng/eventory/services/notification/models"
)

type Outcome int

const (
	OutcomeSent Outcome = iota
	OutcomeAlreadyHandled
	OutcomeRetry
	OutcomePermanentFailure
)

type DeliveryStore interface {
	Claim(context.Context, uuid.UUID, string, string, string, string, time.Time, time.Time, time.Duration) (*models.Delivery, repositories.ClaimResult, error)
	MarkSent(context.Context, uuid.UUID, time.Time) error
	MarkRetry(context.Context, uuid.UUID, string) error
	MarkFailed(context.Context, uuid.UUID, string) error
}

type MailSender interface {
	Send(context.Context, EmailMessage) error
}

type permanentError interface{ Permanent() bool }

type NotificationUseCase struct {
	store        DeliveryStore
	mailer       MailSender
	publicWebURL string
	lease        time.Duration
	now          func() time.Time
}

func NewNotificationUseCase(store DeliveryStore, mailer MailSender, publicWebURL string, lease time.Duration) *NotificationUseCase {
	return &NotificationUseCase{store: store, mailer: mailer, publicWebURL: publicWebURL, lease: lease, now: func() time.Time { return time.Now().UTC() }}
}

func (useCase *NotificationUseCase) Process(ctx context.Context, event events.RegistrationSubmitted, finalAttempt bool) (Outcome, error) {
	event.Normalize()
	if err := event.Validate(); err != nil {
		return OutcomePermanentFailure, err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return OutcomePermanentFailure, fmt.Errorf("marshal normalized event: %w", err)
	}
	digest := sha256.Sum256(payload)
	payloadHash := hex.EncodeToString(digest[:])
	now := useCase.now()
	delivery, claim, err := useCase.store.Claim(ctx, event.ID, event.Type, event.Data.Recipient.Email, registrationTemplateKey, payloadHash, event.OccurredAt, now, useCase.lease)
	if err != nil {
		return OutcomeRetry, err
	}
	if claim == repositories.PayloadConflict {
		return OutcomePermanentFailure, errors.New("event ID was reused with a different payload")
	}
	if claim == repositories.AlreadyHandled || claim == repositories.AlreadyProcessing {
		if claim == repositories.AlreadyHandled {
			return OutcomeAlreadyHandled, nil
		}
		// The previous consumer may still be active or may have crashed. Defer the
		// duplicate instead of acknowledging it while the processing lease is live.
		return OutcomeRetry, errors.New("delivery processing lease is still active")
	}

	message, err := RenderRegistrationEmail(event, useCase.publicWebURL)
	if err != nil {
		_ = useCase.store.MarkFailed(ctx, delivery.ID, "email template rendering failure")
		return OutcomePermanentFailure, err
	}
	if err := useCase.mailer.Send(ctx, message); err != nil {
		var classified permanentError
		if errors.As(err, &classified) && classified.Permanent() {
			_ = useCase.store.MarkFailed(ctx, delivery.ID, "permanent SMTP failure")
			return OutcomePermanentFailure, err
		}
		if finalAttempt {
			_ = useCase.store.MarkFailed(ctx, delivery.ID, "retry budget exhausted after transient SMTP failure")
			return OutcomePermanentFailure, err
		}
		if storeErr := useCase.store.MarkRetry(ctx, delivery.ID, "transient SMTP failure"); storeErr != nil {
			return OutcomeRetry, errors.Join(err, storeErr)
		}
		return OutcomeRetry, err
	}
	if err := useCase.store.MarkSent(ctx, delivery.ID, useCase.now()); err != nil {
		return OutcomeRetry, err
	}
	return OutcomeSent, nil
}
