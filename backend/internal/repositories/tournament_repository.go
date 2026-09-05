package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"gorm.io/gorm"
)

var ErrTournamentNotFound = errors.New("tournament not found")

type TournamentFilters struct {
	Query       string
	StartFrom   *time.Time
	StartTo     *time.Time
	MinEntryFee *int64
	MaxEntryFee *int64
	Status      string
	Sort        string
	Page        int
	PageSize    int
}

type TournamentRepository interface {
	Search(ctx context.Context, filters TournamentFilters) ([]models.Tournament, int64, error)
	GetPublishedByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error)
}

type tournamentRepositoryImpl struct {
	db *gorm.DB
}

func NewTournamentRepository(db *gorm.DB) TournamentRepository {
	return &tournamentRepositoryImpl{db: db}
}

// basically convert query into SQL
func (r *tournamentRepositoryImpl) Search(
	ctx context.Context,
	filters TournamentFilters,
) ([]models.Tournament, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.Tournament{}).
		Where("published = ?", true)

	if filters.Query != "" {
		pattern := "%" + escapeLikePattern(filters.Query) + "%"
		query = query.Where(
			"(name || ' ' || game || ' ' || description) ILIKE ? ESCAPE '\\'",
			pattern,
		)
	}
	if filters.StartFrom != nil {
		query = query.Where("start_at >= ?", *filters.StartFrom)
	}
	if filters.StartTo != nil {
		query = query.Where("start_at <= ?", *filters.StartTo)
	}
	if filters.MinEntryFee != nil {
		query = query.Where("entry_fee >= ?", *filters.MinEntryFee)
	}
	if filters.MaxEntryFee != nil {
		query = query.Where("entry_fee <= ?", *filters.MaxEntryFee)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := map[string]string{
		"start_asc":  "start_at ASC, id ASC",
		"start_desc": "start_at DESC, id ASC",
		"fee_asc":    "entry_fee ASC, start_at ASC, id ASC",
		"fee_desc":   "entry_fee DESC, start_at ASC, id ASC",
	}[filters.Sort]
	if orderBy == "" {
		return nil, 0, fmt.Errorf("unsupported tournament sort %q", filters.Sort)
	}

	var tournaments []models.Tournament
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Session(&gorm.Session{}).
		Preload("Organizer").
		Order(orderBy).
		Limit(filters.PageSize).
		Offset(offset).
		Find(&tournaments).Error; err != nil {
		return nil, 0, err
	}

	return tournaments, total, nil
}

func (r *tournamentRepositoryImpl) GetPublishedByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.db.WithContext(ctx).
		Where("id = ? AND published = ?", id, true).
		Preload("Organizer").
		Preload("Teams", func(db *gorm.DB) *gorm.DB {
			return db.Order("name ASC")
		}).
		Preload("Funding").
		First(&tournament).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	return &tournament, nil
}

func escapeLikePattern(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")
	return replacer.Replace(value)
}
