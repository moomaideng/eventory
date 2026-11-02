package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

// AnnouncementRepository persists tournament announcements (US3-4).
//
// A tournament the caller may not see or manage reports ErrTournamentNotFound,
// so a draft tournament looks identical to one that does not exist.
type AnnouncementRepository interface {
	// Create stores an announcement. Only an owner or admin of the tournament
	// may post; anyone else gets ErrTournamentNotFound.
	Create(
		ctx context.Context,
		authorAccountID uuid.UUID,
		authorDisplayName string,
		tournamentID uuid.UUID,
		title, body string,
	) (*models.TournamentAnnouncement, error)

	// List returns the tournament's announcements, newest first. Only a
	// published tournament is visible; a draft reports ErrTournamentNotFound,
	// so it looks identical to one that does not exist.
	List(ctx context.Context, tournamentID uuid.UUID) ([]models.TournamentAnnouncement, error)
}

type announcementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) AnnouncementRepository {
	return &announcementRepository{db: db}
}

func (r *announcementRepository) Create(
	ctx context.Context,
	authorAccountID uuid.UUID,
	authorDisplayName string,
	tournamentID uuid.UUID,
	title, body string,
) (*models.TournamentAnnouncement, error) {
	// Scoping through ownedTournaments also enforces who may post.
	var tournament models.Tournament
	err := ownedTournaments(r.db.WithContext(ctx), authorAccountID).
		Where("tournaments.id = ?", tournamentID).
		First(&tournament).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	announcement := models.TournamentAnnouncement{
		TournamentID:      tournament.ID,
		Title:             title,
		Body:              body,
		AuthorAccountID:   authorAccountID,
		AuthorDisplayName: authorDisplayName,
	}
	if err := r.db.WithContext(ctx).Create(&announcement).Error; err != nil {
		return nil, err
	}
	return &announcement, nil
}

func (r *announcementRepository) List(
	ctx context.Context,
	tournamentID uuid.UUID,
) ([]models.TournamentAnnouncement, error) {
	// Filtering on published also hides drafts: both a missing tournament and
	// a draft come back as "not found".
	var tournament models.Tournament
	err := r.db.WithContext(ctx).
		Where("id = ? AND published = ?", tournamentID, true).
		First(&tournament).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	var announcements []models.TournamentAnnouncement
	err = r.db.WithContext(ctx).
		Where("tournament_id = ?", tournamentID).
		Order("created_at DESC").
		Find(&announcements).Error
	if err != nil {
		return nil, err
	}
	return announcements, nil
}
