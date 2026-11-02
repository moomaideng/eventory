package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

// AnnouncementAudience is who an announcement is delivered to.
type AnnouncementAudience struct {
	TournamentName string
	// AccountIDs are the members of every accepted team. A solo player is a
	// one-person team, so they are included too. An account can be in only one
	// team per tournament, so an ID appears at most once.
	AccountIDs []uuid.UUID
}

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

	// Audience returns the tournament's name and the accounts that should be
	// notified of a new announcement.
	Audience(ctx context.Context, tournamentID uuid.UUID) (*AnnouncementAudience, error)
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

func (r *announcementRepository) Audience(
	ctx context.Context,
	tournamentID uuid.UUID,
) (*AnnouncementAudience, error) {
	var tournament models.Tournament
	err := r.db.WithContext(ctx).
		Select("id", "name").
		Where("id = ?", tournamentID).
		First(&tournament).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	var accountIDs []uuid.UUID
	err = r.db.WithContext(ctx).
		Model(&models.TournamentTeamMember{}).
		Joins("JOIN tournament_teams ON tournament_teams.id = tournament_team_members.tournament_team_id").
		Where("tournament_teams.tournament_id = ? AND tournament_teams.status = ?",
			tournamentID, models.TournamentTeamStatusAccepted).
		Pluck("tournament_team_members.account_id", &accountIDs).Error
	if err != nil {
		return nil, err
	}

	return &AnnouncementAudience{TournamentName: tournament.Name, AccountIDs: accountIDs}, nil
}
