package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TournamentStatusRepository persists manual status overrides (US3-3) together
// with their audit trail.
type TournamentStatusRepository interface {
	// ApplyOverride writes the new status and its history row in one
	// transaction. mutate receives the tournament's current state and returns
	// the status to move to; returning an error aborts the whole override, so
	// transition rules stay in the use case while the row stays locked.
	ApplyOverride(
		ctx context.Context,
		organizerID uuid.UUID,
		actorAccountID uuid.UUID,
		actorDisplayName string,
		tournamentID uuid.UUID,
		reason string,
		mutate func(current *models.Tournament) (models.TournamentStatus, error),
	) (*models.TournamentStatusChange, error)

	// ListHistory returns an owned tournament's current status and its override
	// trail, newest first.
	ListHistory(ctx context.Context, organizerID, tournamentID uuid.UUID) (*TournamentStatusHistory, error)
}

// TournamentStatusHistory pairs a tournament's current status with its override
// trail so the modal can render both from a single request.
type TournamentStatusHistory struct {
	CurrentStatus models.TournamentStatus
	Changes       []models.TournamentStatusChange
}

type tournamentStatusRepository struct {
	db *gorm.DB
}

func NewTournamentStatusRepository(db *gorm.DB) TournamentStatusRepository {
	return &tournamentStatusRepository{db: db}
}

// ownedTournaments scopes queries to tournaments owned by the given
// organizer profile (tournaments.organizer_id).
func ownedTournaments(tx *gorm.DB, organizerID uuid.UUID) *gorm.DB {
	return tx.Model(&models.Tournament{}).Where("organizer_id = ?", organizerID)
}

func (r *tournamentStatusRepository) ApplyOverride(
	ctx context.Context,
	organizerID uuid.UUID,
	actorAccountID uuid.UUID,
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
		err := ownedTournaments(tx, organizerID).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", tournamentID).
			First(&tournament).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTournamentNotFound
		}
		if err != nil {
			return err
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
			ActorAccountID:   actorAccountID,
			ActorDisplayName: actorDisplayName,
		}
		return tx.Create(&change).Error
	})
	if err != nil {
		return nil, err
	}
	return &change, nil
}

func (r *tournamentStatusRepository) ListHistory(
	ctx context.Context,
	organizerID, tournamentID uuid.UUID,
) (*TournamentStatusHistory, error) {
	// Confirm ownership before exposing the trail; an unowned tournament must
	// look identical to a missing one.
	var tournament models.Tournament
	err := ownedTournaments(r.db.WithContext(ctx), organizerID).
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
