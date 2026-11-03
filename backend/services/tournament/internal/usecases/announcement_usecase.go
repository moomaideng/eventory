package usecases

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

const (
	announcementTitleMaxLength = 150
	announcementBodyMaxLength  = 2000

	// announcementNotifyTimeout bounds how long a post can be held up by
	// publishing notifications, so a slow broker never stalls the organizer.
	announcementNotifyTimeout = 10 * time.Second
)

var (
	ErrAnnouncementTitleRequired = errors.New("title is required")
	ErrAnnouncementTitleTooLong  = errors.New("title must be at most 150 characters")
	ErrAnnouncementBodyRequired  = errors.New("body is required")
	ErrAnnouncementBodyTooLong   = errors.New("body must be at most 2000 characters")
)

// AnnouncementUseCase posts and lists tournament announcements (US3-4).
type AnnouncementUseCase struct {
	repo       repositories.AnnouncementRepository
	accountSvc ports.AccountService
	publisher  events.Publisher
}

// NewAnnouncementUseCase builds the use case. A nil publisher means no
// notifications are sent.
func NewAnnouncementUseCase(
	repo repositories.AnnouncementRepository,
	accountSvc ports.AccountService,
	publisher events.Publisher,
) *AnnouncementUseCase {
	if publisher == nil {
		publisher = events.NoopPublisher{}
	}
	return &AnnouncementUseCase{repo: repo, accountSvc: accountSvc, publisher: publisher}
}

// Create validates and stores an announcement. Only an owner or admin of the
// tournament may post; anyone else gets ErrTournamentNotFound.
func (u *AnnouncementUseCase) Create(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
	rawTitle, rawBody string,
) (*models.TournamentAnnouncement, error) {
	title := strings.TrimSpace(rawTitle)
	body := strings.TrimSpace(rawBody)
	switch {
	case title == "":
		return nil, ErrAnnouncementTitleRequired
	case utf8.RuneCountInString(title) > announcementTitleMaxLength:
		return nil, ErrAnnouncementTitleTooLong
	case body == "":
		return nil, ErrAnnouncementBodyRequired
	case utf8.RuneCountInString(body) > announcementBodyMaxLength:
		return nil, ErrAnnouncementBodyTooLong
	}

	author, err := u.accountSvc.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	announcement, err := u.repo.Create(ctx, accountID, author.DisplayName, tournamentID, title, body)
	if err != nil {
		return nil, translateAnnouncementError(err)
	}

	u.notify(ctx, announcement)
	return announcement, nil
}

// List returns a published tournament's announcements, newest first.
func (u *AnnouncementUseCase) List(
	ctx context.Context,
	tournamentID uuid.UUID,
) ([]models.TournamentAnnouncement, error) {
	announcements, err := u.repo.List(ctx, tournamentID)
	if err != nil {
		return nil, translateAnnouncementError(err)
	}
	return announcements, nil
}

func translateAnnouncementError(err error) error {
	if errors.Is(err, repositories.ErrTournamentNotFound) {
		return ErrTournamentNotFound
	}
	return err
}

// notify publishes one announcement.created event per member of an accepted
// team. The announcement is already saved, so every failure here is logged and
// never turns the post into an error: a broker outage must not cost the
// organizer their announcement.
func (u *AnnouncementUseCase) notify(ctx context.Context, announcement *models.TournamentAnnouncement) {
	ctx, cancel := context.WithTimeout(ctx, announcementNotifyTimeout)
	defer cancel()

	audience, err := u.repo.Audience(ctx, announcement.TournamentID)
	if err != nil {
		slog.WarnContext(ctx, "announcement notifications skipped: audience lookup failed",
			"announcement_id", announcement.ID, "error", err)
		return
	}
	if len(audience.AccountIDs) == 0 {
		return
	}

	accounts, err := u.accountSvc.BatchGetAccounts(ctx, audience.AccountIDs)
	if err != nil {
		slog.WarnContext(ctx, "announcement notifications skipped: account lookup failed",
			"announcement_id", announcement.ID, "error", err)
		return
	}

	// One correlation ID ties together every event of this announcement.
	correlationID := uuid.New()
	published, failed := 0, 0
	for _, accountID := range audience.AccountIDs {
		account, ok := accounts[accountID]
		if !ok {
			failed++
			continue
		}

		event := events.AnnouncementCreated{
			SpecVersion:   events.EventSpecVersion,
			ID:            uuid.New(),
			Type:          events.AnnouncementCreatedType,
			Source:        "tournament-service",
			OccurredAt:    announcement.CreatedAt,
			CorrelationID: correlationID,
			Data: events.AnnouncementCreatedData{
				AnnouncementID: announcement.ID,
				TournamentID:   announcement.TournamentID,
				TournamentName: audience.TournamentName,
				Title:          announcement.Title,
				Body:           announcement.Body,
				AuthorName:     announcement.AuthorDisplayName,
				Recipient: events.EventRecipient{
					AccountID:   account.ID,
					Email:       account.Email,
					DisplayName: account.DisplayName,
				},
			},
		}
		event.Normalize()
		if err := event.Validate(); err != nil {
			slog.WarnContext(ctx, "announcement event invalid, not published",
				"announcement_id", announcement.ID, "account_id", accountID, "error", err)
			failed++
			continue
		}
		if err := u.publisher.Publish(ctx, events.AnnouncementCreatedType, event); err != nil {
			slog.WarnContext(ctx, "announcement event publish failed",
				"announcement_id", announcement.ID, "account_id", accountID, "error", err)
			failed++
			continue
		}
		published++
	}

	slog.InfoContext(ctx, "announcement notifications published",
		"announcement_id", announcement.ID, "published", published, "failed", failed)
}
