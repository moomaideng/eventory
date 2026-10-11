package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

var (
	ErrStaffAlreadyExists = errors.New("staff member already exists")
	ErrStaffNotFound      = errors.New("staff member not found")
)

type TournamentStaffRepository interface {
	TournamentExists(ctx context.Context, tournamentID uuid.UUID) (bool, error)
	ListStaff(ctx context.Context, tournamentID uuid.UUID) ([]models.TournamentStaff, error)
	GetStaff(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.TournamentStaff, error)
	AddStaff(ctx context.Context, staff *models.TournamentStaff) error
	UpdateStaffRole(ctx context.Context, tournamentID, accountID uuid.UUID, role models.TournamentStaffRole) error
	RemoveStaff(ctx context.Context, tournamentID, accountID uuid.UUID) error
}

type tournamentStaffRepositoryImpl struct {
	db *gorm.DB
}

func NewTournamentStaffRepository(db *gorm.DB) TournamentStaffRepository {
	return &tournamentStaffRepositoryImpl{db: db}
}

func (r *tournamentStaffRepositoryImpl) TournamentExists(ctx context.Context, tournamentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Tournament{}).
		Where("id = ?", tournamentID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *tournamentStaffRepositoryImpl) ListStaff(ctx context.Context, tournamentID uuid.UUID) ([]models.TournamentStaff, error) {
	var staff []models.TournamentStaff
	err := r.db.WithContext(ctx).
		Where("tournament_id = ?", tournamentID).
		Order("created_at ASC, account_id ASC").
		Find(&staff).Error
	if err != nil {
		return nil, err
	}
	return staff, nil
}

func (r *tournamentStaffRepositoryImpl) GetStaff(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.TournamentStaff, error) {
	var staff models.TournamentStaff
	err := r.db.WithContext(ctx).
		Where("tournament_id = ? AND account_id = ?", tournamentID, accountID).
		First(&staff).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &staff, nil
}

func (r *tournamentStaffRepositoryImpl) AddStaff(ctx context.Context, staff *models.TournamentStaff) error {
	err := r.db.WithContext(ctx).Create(staff).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(err.Error(), "23505") {
			return ErrStaffAlreadyExists
		}
		return err
	}
	return nil
}

func (r *tournamentStaffRepositoryImpl) UpdateStaffRole(ctx context.Context, tournamentID, accountID uuid.UUID, role models.TournamentStaffRole) error {
	res := r.db.WithContext(ctx).
		Model(&models.TournamentStaff{}).
		Where("tournament_id = ? AND account_id = ?", tournamentID, accountID).
		Updates(map[string]interface{}{
			"role":       role,
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrStaffNotFound
	}
	return nil
}

func (r *tournamentStaffRepositoryImpl) RemoveStaff(ctx context.Context, tournamentID, accountID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("tournament_id = ? AND account_id = ?", tournamentID, accountID).
		Delete(&models.TournamentStaff{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrStaffNotFound
	}
	return nil
}
