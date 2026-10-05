package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

type OrganizerDashboardRepository interface {
	ListOwned(ctx context.Context, organizerID uuid.UUID, page, pageSize int) ([]models.Tournament, int64, error)
	GetOwned(ctx context.Context, organizerID, tournamentID uuid.UUID) (*models.Tournament, error)
}

type organizerDashboardRepository struct {
	db *gorm.DB
}

func NewOrganizerDashboardRepository(db *gorm.DB) OrganizerDashboardRepository {
	return &organizerDashboardRepository{db: db}
}

func (r *organizerDashboardRepository) owned(ctx context.Context, organizerID uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).Model(&models.Tournament{}).
		Where("organizer_id = ?", organizerID)
}

func (r *organizerDashboardRepository) ListOwned(ctx context.Context, organizerID uuid.UUID, page, pageSize int) ([]models.Tournament, int64, error) {
	var total int64
	if err := r.owned(ctx, organizerID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	tournaments := make([]models.Tournament, 0)
	err := r.owned(ctx, organizerID).
		Preload("Teams.Members").Preload("Funding").
		Order("start_at DESC, id ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).Find(&tournaments).Error
	return tournaments, total, err
}

func (r *organizerDashboardRepository) GetOwned(ctx context.Context, organizerID, tournamentID uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.owned(ctx, organizerID).
		Where("id = ?", tournamentID).
		Preload("Funding").
		Preload("Teams", func(db *gorm.DB) *gorm.DB { return db.Order("created_at DESC, id ASC") }).
		Preload("Teams.Members", func(db *gorm.DB) *gorm.DB { return db.Order("joined_at ASC, id ASC") }).
		First(&tournament).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tournament, nil
}
