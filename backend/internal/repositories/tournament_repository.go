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
	Status      models.TournamentStatus
	Sort        string
	Page        int
	PageSize    int
}

type TournamentSearchItem struct {
	Tournament      models.Tournament
	RegisteredCount int64
}

type TournamentRepository interface {
	Search(ctx context.Context, filters TournamentFilters) ([]TournamentSearchItem, int64, error)
	GetPublishedByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error)
	Create(ctx context.Context, tournament *models.Tournament, funding *models.TournamentFunding) error
	Update(ctx context.Context, tournament *models.Tournament) error
	GetActiveTeamCount(ctx context.Context, tournamentID uuid.UUID) (acceptedCount int64, lockedOrAcceptedCount int64, err error)
}

type tournamentRepositoryImpl struct {
	db *gorm.DB
}

func NewTournamentRepository(db *gorm.DB) TournamentRepository {
	return &tournamentRepositoryImpl{db: db}
}

func (r *tournamentRepositoryImpl) Search(
	ctx context.Context,
	filters TournamentFilters,
) ([]TournamentSearchItem, int64, error) {
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

	items := make([]TournamentSearchItem, len(tournaments))
	if len(tournaments) == 0 {
		return items, total, nil
	}

	tournamentIDs := make([]uuid.UUID, len(tournaments))
	for index, tournament := range tournaments {
		tournamentIDs[index] = tournament.ID
		items[index].Tournament = tournament
	}

	type tournamentTeamCount struct {
		TournamentID uuid.UUID
		Count        int64
	}
	var counts []tournamentTeamCount
	if err := r.db.WithContext(ctx).
		Model(&models.TournamentTeam{}).
		Select("tournament_id, COUNT(*) AS count").
		Where("tournament_id IN ? AND status = ?", tournamentIDs, models.TournamentTeamStatusAccepted).
		Group("tournament_id").
		Scan(&counts).Error; err != nil {
		return nil, 0, err
	}

	countsByTournamentID := make(map[uuid.UUID]int64, len(counts))
	for _, count := range counts {
		countsByTournamentID[count.TournamentID] = count.Count
	}
	for index := range items {
		items[index].RegisteredCount = countsByTournamentID[items[index].Tournament.ID]
	}

	return items, total, nil
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
			return db.Where("status = ?", models.TournamentTeamStatusAccepted).Order("name ASC")
		}).
		Preload("Teams.Members").
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

func (r *tournamentRepositoryImpl) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Preload("Organizer").
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

func (r *tournamentRepositoryImpl) Create(
	ctx context.Context,
	tournament *models.Tournament,
	funding *models.TournamentFunding,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(tournament).Error; err != nil {
			return err
		}
		if funding != nil {
			funding.TournamentID = tournament.ID
			if err := tx.Create(funding).Error; err != nil {
				return err
			}
			tournament.Funding = funding
		}
		return nil
	})
}

func (r *tournamentRepositoryImpl) Update(
	ctx context.Context,
	tournament *models.Tournament,
) error {
	return r.db.WithContext(ctx).Save(tournament).Error
}

func (r *tournamentRepositoryImpl) GetActiveTeamCount(
	ctx context.Context,
	tournamentID uuid.UUID,
) (int64, int64, error) {
	var acceptedCount int64
	err := r.db.WithContext(ctx).Model(&models.TournamentTeam{}).
		Where("tournament_id = ? AND status = ?", tournamentID, models.TournamentTeamStatusAccepted).
		Count(&acceptedCount).Error
	if err != nil {
		return 0, 0, err
	}

	var lockedOrAcceptedCount int64
	err = r.db.WithContext(ctx).Model(&models.TournamentTeam{}).
		Where("tournament_id = ? AND status IN ?", tournamentID, []models.TournamentTeamStatus{
			models.TournamentTeamStatusAccepted,
			models.TournamentTeamStatusLocked,
		}).
		Count(&lockedOrAcceptedCount).Error
	if err != nil {
		return 0, 0, err
	}

	return acceptedCount, lockedOrAcceptedCount, nil
}

func escapeLikePattern(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")
	return replacer.Replace(value)
}
