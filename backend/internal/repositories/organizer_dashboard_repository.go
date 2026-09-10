package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"gorm.io/gorm"
)

type OrganizerDashboardRepository interface {
	ListOwned(ctx context.Context, accountID uuid.UUID, page, pageSize int) ([]models.Tournament, int64, error)
	GetOwned(ctx context.Context, accountID, tournamentID uuid.UUID) (*models.Tournament, error)
}

type organizerDashboardRepository struct {
	db *gorm.DB
}

func NewOrganizerDashboardRepository(db *gorm.DB) OrganizerDashboardRepository {
	return &organizerDashboardRepository{db: db}
}

func (r *organizerDashboardRepository) owned(ctx context.Context, accountID uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).Model(&models.Tournament{}).
		Joins("JOIN organizer_profiles ON organizer_profiles.id = tournaments.organizer_id").
		Where("organizer_profiles.account_id = ?", accountID)
}

func (r *organizerDashboardRepository) ListOwned(ctx context.Context, accountID uuid.UUID, page, pageSize int) ([]models.Tournament, int64, error) {
	var total int64
	if err := r.owned(ctx, accountID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	tournaments := make([]models.Tournament, 0)
	err := r.owned(ctx, accountID).Select("tournaments.*").
		Preload("Organizer").Preload("Teams.Members").Preload("Funding").
		Order("tournaments.start_at DESC, tournaments.id ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).Find(&tournaments).Error
	return tournaments, total, err
}

func (r *organizerDashboardRepository) GetOwned(ctx context.Context, accountID, tournamentID uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	// Scope the root query before loading private registration data, including drafts.
	err := r.owned(ctx, accountID).Select("tournaments.*").
		Where("tournaments.id = ?", tournamentID).
		Preload("Organizer").Preload("Funding").
		Preload("Teams", func(db *gorm.DB) *gorm.DB { return db.Order("created_at DESC, id ASC") }).
		Preload("Teams.Members", func(db *gorm.DB) *gorm.DB { return db.Order("joined_at ASC, id ASC") }).
		Preload("Teams.Members.Account").First(&tournament).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tournament, nil
}
