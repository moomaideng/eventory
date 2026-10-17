package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrTournamentNotFound     = errors.New("tournament not found")
	ErrTournamentStateChanged = errors.New("tournament status changed during update")
)

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

// TournamentStatusHistory pairs a tournament's current status with its override
// trail so the modal can render both from a single request.
type TournamentStatusHistory struct {
	CurrentStatus models.TournamentStatus
	Changes       []models.TournamentStatusChange
}

type TournamentRepository interface {
	Search(ctx context.Context, filters TournamentFilters) ([]TournamentSearchItem, int64, error)
	GetPublishedByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error)
	Create(ctx context.Context, tournament *models.Tournament, funding *models.TournamentFunding) error
	Update(ctx context.Context, tournament *models.Tournament) error
	GetActiveTeamCount(ctx context.Context, tournamentID uuid.UUID) (acceptedCount int64, lockedOrAcceptedCount int64, err error)
	ListOwned(ctx context.Context, organizerID uuid.UUID, page, pageSize int) ([]models.Tournament, int64, error)
	GetOwned(ctx context.Context, organizerID, tournamentID uuid.UUID) (*models.Tournament, error)
	ApplyOverride(
		ctx context.Context,
		accountID uuid.UUID,
		actorDisplayName string,
		tournamentID uuid.UUID,
		reason string,
		mutate func(current *models.Tournament) (models.TournamentStatus, error),
	) (*models.TournamentStatusChange, error)
	ListHistory(ctx context.Context, accountID, tournamentID uuid.UUID) (*TournamentStatusHistory, error)
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
		Preload("Staff", orderedStaff).
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
		Preload("Staff", orderedStaff).
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
		Preload("Staff", orderedStaff).
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Tournament{}).
			Where("id = ? AND status = ?", tournament.ID, tournament.Status).
			Updates(map[string]any{
				"name":                  tournament.Name,
				"description":           tournament.Description,
				"game":                  tournament.Game,
				"location":              tournament.Location,
				"start_at":              tournament.StartsAt,
				"end_at":                tournament.EndsAt,
				"registration_deadline": tournament.RegistrationDeadline,
				"entry_fee":             tournament.EntryFee,
				"registration_mode":     tournament.RegistrationMode,
				"min_team_size":         tournament.MinTeamSize,
				"max_team_size":         tournament.MaxTeamSize,
				"capacity":              tournament.Capacity,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrTournamentStateChanged
		}
		if tournament.Funding != nil {
			return tx.Model(&models.TournamentFunding{}).
				Where("tournament_id = ?", tournament.ID).
				Updates(map[string]any{
					"goal_amount": tournament.Funding.GoalAmount,
					"deadline":    tournament.Funding.Deadline,
				}).Error
		}
		return nil
	})
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
			models.TournamentTeamStatusPaymentPending,
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

// orderedStaff preloads a tournament's staff in a stable order so the owner
// shown as the organizer is deterministic.
func orderedStaff(db *gorm.DB) *gorm.DB {
	return db.Order("tournament_staff.created_at ASC, tournament_staff.account_id ASC")
}

func (r *tournamentRepositoryImpl) staffOf(ctx context.Context, accountID uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).Model(&models.Tournament{}).
		Joins("JOIN tournament_staff ON tournament_staff.tournament_id = tournaments.id").
		Where("tournament_staff.account_id = ?", accountID)
}

func (r *tournamentRepositoryImpl) ListOwned(ctx context.Context, accountID uuid.UUID, page, pageSize int) ([]models.Tournament, int64, error) {
	var total int64
	if err := r.staffOf(ctx, accountID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	tournaments := make([]models.Tournament, 0)
	err := r.staffOf(ctx, accountID).
		Preload("Staff", orderedStaff).Preload("Teams.Members").Preload("Funding").
		Order("tournaments.start_at DESC, tournaments.id ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).Find(&tournaments).Error
	return tournaments, total, err
}

func (r *tournamentRepositoryImpl) GetOwned(ctx context.Context, accountID, tournamentID uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.staffOf(ctx, accountID).
		Where("tournaments.id = ?", tournamentID).
		Preload("Staff", orderedStaff).Preload("Funding").
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

// ownedTournaments scopes queries to tournaments where the user is an owner or admin.
func ownedTournaments(tx *gorm.DB, accountID uuid.UUID) *gorm.DB {
	return tx.Model(&models.Tournament{}).
		Joins("JOIN tournament_staff ON tournament_staff.tournament_id = tournaments.id").
		Where("tournament_staff.account_id = ? AND tournament_staff.role IN ?", accountID, []models.TournamentStaffRole{models.TournamentStaffRoleOwner, models.TournamentStaffRoleAdmin})
}

func (r *tournamentRepositoryImpl) ApplyOverride(
	ctx context.Context,
	accountID uuid.UUID,
	actorDisplayName string,
	tournamentID uuid.UUID,
	reason string,
	mutate func(current *models.Tournament) (models.TournamentStatus, error),
) (*models.TournamentStatusChange, error) {
	var change models.TournamentStatusChange

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tournament models.Tournament

		// Lock the row for the rest of the transaction so two concurrent
		// overrides cannot both read the same FromStatus and clobber each
		// other.
		err := ownedTournaments(tx, accountID).
			Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "tournaments"}}).
			Where("id = ?", tournamentID).
			Preload("Staff", orderedStaff).
			First(&tournament).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTournamentNotFound
		}
		if err != nil {
			return err
		}
		if tournament.Status == models.TournamentStatusCrowdfunding {
			var funding models.TournamentFunding
			err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("tournament_id = ?", tournament.ID).First(&funding).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err == nil {
				tournament.Funding = &funding
			}
		}

		to, err := mutate(&tournament)
		if err != nil {
			return err
		}

		from := tournament.Status
		if err := tx.Model(&tournament).Update("status", to).Error; err != nil {
			return err
		}

		change = models.TournamentStatusChange{
			TournamentID:     tournament.ID,
			FromStatus:       from,
			ToStatus:         to,
			Reason:           reason,
			ActorAccountID:   accountID,
			ActorDisplayName: actorDisplayName,
		}
		return tx.Create(&change).Error
	})
	if err != nil {
		return nil, err
	}
	return &change, nil
}

func (r *tournamentRepositoryImpl) ListHistory(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
) (*TournamentStatusHistory, error) {
	var tournament models.Tournament
	err := ownedTournaments(r.db.WithContext(ctx), accountID).
		Select("id", "status").
		Where("id = ?", tournamentID).
		First(&tournament).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTournamentNotFound
	}
	if err != nil {
		return nil, err
	}

	changes := make([]models.TournamentStatusChange, 0)
	err = r.db.WithContext(ctx).
		Where("tournament_id = ?", tournamentID).
		Order("created_at DESC, id ASC").
		Find(&changes).Error
	if err != nil {
		return nil, err
	}
	return &TournamentStatusHistory{CurrentStatus: tournament.Status, Changes: changes}, nil
}
