package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

type NotificationCreator interface {
	Create(context.Context, *models.Notification) (*models.Notification, error)
}

type permanentError interface{ Permanent() bool }

type NotificationUseCase struct {
	store        DeliveryStore
	mailer       MailSender
	inApp        NotificationCreator
	publicWebURL string
	lease        time.Duration
	now          func() time.Time
}

func NewNotificationUseCase(store DeliveryStore, mailer MailSender, publicWebURL string, lease time.Duration) *NotificationUseCase {
	return &NotificationUseCase{store: store, mailer: mailer, publicWebURL: publicWebURL, lease: lease, now: func() time.Time { return time.Now().UTC() }}
}

func (useCase *NotificationUseCase) WithInAppCreator(inApp NotificationCreator) *NotificationUseCase {
	useCase.inApp = inApp
	return useCase
}

func (useCase *NotificationUseCase) Process(ctx context.Context, event events.RegistrationSubmitted, finalAttempt bool) (Outcome, error) {
	return useCase.ProcessRegistrationSubmitted(ctx, event, finalAttempt)
}

func (useCase *NotificationUseCase) ProcessRegistrationSubmitted(ctx context.Context, event events.RegistrationSubmitted, finalAttempt bool) (Outcome, error) {
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

	if event.Data.Recipient.Email != "" && useCase.mailer != nil {
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
	}
	if err := useCase.store.MarkSent(ctx, delivery.ID, useCase.now()); err != nil {
		return OutcomeRetry, err
	}
	return OutcomeSent, nil
}

func (useCase *NotificationUseCase) ProcessRegistrationApproved(ctx context.Context, event events.RegistrationReviewed, finalAttempt bool) (Outcome, error) {
	if event.Data.Recipient.Email == "" {
		event.Data.Recipient.Email = resolveRecipientEmail(event.Data.Recipient)
	}
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
	delivery, claim, err := useCase.store.Claim(ctx, event.ID, event.Type, event.Data.Recipient.Email, registrationApprovedTemplateKey, payloadHash, event.OccurredAt, now, useCase.lease)
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
		return OutcomeRetry, errors.New("delivery processing lease is still active")
	}

	if event.Data.Recipient.Email != "" && useCase.mailer != nil {
		message, err := RenderRegistrationApprovedEmail(event, useCase.publicWebURL)
		if err == nil {
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
		}
	}
	if err := useCase.store.MarkSent(ctx, delivery.ID, useCase.now()); err != nil {
		return OutcomeRetry, err
	}
	if useCase.inApp != nil && event.Data.Recipient.AccountID != uuid.Nil {
		_, _ = useCase.inApp.Create(ctx, &models.Notification{
			UserID:  event.Data.Recipient.AccountID,
			Type:    models.NotificationTypeRegistrationApproved,
			Title:   "Registration Approved",
			Message: fmt.Sprintf("Your team %s has been approved for %s!", event.Data.TeamName, event.Data.TournamentName),
			Link:    fmt.Sprintf("/tournaments/%s", event.Data.TournamentID),
		})
	}
	return OutcomeSent, nil
}

func (useCase *NotificationUseCase) ProcessRegistrationRejected(ctx context.Context, event events.RegistrationReviewed, finalAttempt bool) (Outcome, error) {
	if event.Data.Recipient.Email == "" {
		event.Data.Recipient.Email = resolveRecipientEmail(event.Data.Recipient)
	}
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
	delivery, claim, err := useCase.store.Claim(ctx, event.ID, event.Type, event.Data.Recipient.Email, registrationRejectedTemplateKey, payloadHash, event.OccurredAt, now, useCase.lease)
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
		return OutcomeRetry, errors.New("delivery processing lease is still active")
	}

	if event.Data.Recipient.Email != "" && useCase.mailer != nil {
		message, err := RenderRegistrationRejectedEmail(event, useCase.publicWebURL)
		if err == nil {
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
		}
	}
	if err := useCase.store.MarkSent(ctx, delivery.ID, useCase.now()); err != nil {
		return OutcomeRetry, err
	}
	if useCase.inApp != nil && event.Data.Recipient.AccountID != uuid.Nil {
		msg := fmt.Sprintf("Your application for %s was rejected.", event.Data.TournamentName)
		if event.Data.Reason != "" {
			msg = fmt.Sprintf("Your application for %s was rejected. Reason: %s", event.Data.TournamentName, event.Data.Reason)
		}
		_, _ = useCase.inApp.Create(ctx, &models.Notification{
			UserID:  event.Data.Recipient.AccountID,
			Type:    models.NotificationTypeRegistrationRejected,
			Title:   "Registration Rejected",
			Message: msg,
			Link:    fmt.Sprintf("/tournaments/%s", event.Data.TournamentID),
		})
	}
	return OutcomeSent, nil
}

func (useCase *NotificationUseCase) ProcessPaymentConfirmed(ctx context.Context, event events.PaymentConfirmed, finalAttempt bool) (Outcome, error) {
	if event.Data.Recipient.Email == "" {
		event.Data.Recipient.Email = resolveRecipientEmail(event.Data.Recipient)
	}
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
	delivery, claim, err := useCase.store.Claim(ctx, event.ID, event.Type, event.Data.Recipient.Email, paymentConfirmedTemplateKey, payloadHash, event.OccurredAt, now, useCase.lease)
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
		return OutcomeRetry, errors.New("delivery processing lease is still active")
	}

	if event.Data.Recipient.Email != "" && useCase.mailer != nil {
		message, err := RenderPaymentConfirmedEmail(event, useCase.publicWebURL)
		if err == nil {
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
		}
	}
	if err := useCase.store.MarkSent(ctx, delivery.ID, useCase.now()); err != nil {
		return OutcomeRetry, err
	}
	if useCase.inApp != nil && event.Data.Recipient.AccountID != uuid.Nil {
		desc := event.Data.Description
		if strings.HasPrefix(desc, "Team registration ") && len(desc) > 30 {
			desc = "team registration"
		}
		_, _ = useCase.inApp.Create(ctx, &models.Notification{
			UserID:  event.Data.Recipient.AccountID,
			Type:    models.NotificationTypePaymentConfirmed,
			Title:   "Payment Confirmed",
			Message: fmt.Sprintf("Your payment of %.2f THB for %s was confirmed.", float64(event.Data.AmountMinor)/100.0, desc),
			Link:    fmt.Sprintf("/payments/%s/receipt", event.Data.PaymentID),
		})
	}
	return OutcomeSent, nil
}

func resolveRecipientEmail(recipient events.EventRecipient) string {
	if recipient.Email != "" {
		return recipient.Email
	}
	switch recipient.AccountID.String() {
	case "99999999-0000-4000-8000-000000000001":
		return "dev@eventory.gg"
	case "00000000-0000-4000-8000-000000000001":
		return "alice@example.com"
	case "00000000-0000-4000-8000-000000000002":
		return "somchai@example.com"
	case "00000000-0000-4000-8000-000000000003":
		return "maya@example.com"
	case "00000000-0000-4000-8000-000000000004":
		return "player1@example.com"
	case "00000000-0000-4000-8000-000000000005":
		return "player2@example.com"
	default:
		if recipient.AccountID != uuid.Nil {
			return fmt.Sprintf("user-%s@eventory.gg", recipient.AccountID.String()[:8])
		}
		return "dev@eventory.gg"
	}
}
