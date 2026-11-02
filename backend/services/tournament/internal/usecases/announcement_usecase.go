package usecases

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

const (
	announcementTitleMaxLength = 150
	announcementBodyMaxLength  = 5000
)

var (
	ErrAnnouncementTitleRequired = errors.New("title is required")
	ErrAnnouncementTitleTooLong  = errors.New("title must be at most 150 characters")
	ErrAnnouncementBodyRequired  = errors.New("body is required")
	ErrAnnouncementBodyTooLong   = errors.New("body must be at most 5000 characters")
)

// AnnouncementUseCase posts and lists tournament announcements (US3-4).
type AnnouncementUseCase struct {
	repo       repositories.AnnouncementRepository
	accountSvc ports.AccountService
}

func NewAnnouncementUseCase(
	repo repositories.AnnouncementRepository,
	accountSvc ports.AccountService,
) *AnnouncementUseCase {
	return &AnnouncementUseCase{repo: repo, accountSvc: accountSvc}
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

	// TODO(US9-1): after the announcement is saved, publish announcement.created
	// to the members of accepted teams (added in the notification step of this PR).
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
