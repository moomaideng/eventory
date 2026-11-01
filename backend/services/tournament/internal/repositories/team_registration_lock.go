package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
)

// TeamRegistrationLock holds only a PostgreSQL row lock, never answer data.
// The use case checks membership and state while this lock prevents concurrent
// submission, removal, or disband from racing a MongoDB answer update.
type TeamRegistrationLock interface {
	Team() *models.TournamentTeam
	Release() error
}

type teamRegistrationLock struct {
	tx   *gorm.DB
	team *models.TournamentTeam
}

func (l *teamRegistrationLock) Team() *models.TournamentTeam { return l.team }
func (l *teamRegistrationLock) Release() error               { return l.tx.Rollback().Error }

func (r *teamLobbyRepository) LockRegistration(ctx context.Context, teamID uuid.UUID) (TeamRegistrationLock, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	team, err := r.lockTeam(tx, teamID)
	if err != nil {
		return nil, errors.Join(err, tx.Rollback().Error)
	}
	return &teamRegistrationLock{tx: tx, team: team}, nil
}
