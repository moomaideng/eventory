package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrTeamLobbyNotFound            = errors.New("team lobby not found")
	ErrTeamLobbyNotForming          = errors.New("team lobby is not forming")
	ErrTeamLobbyFull                = errors.New("team lobby is full")
	ErrRosterBelowMinimum           = errors.New("roster does not meet the tournament minimum")
	ErrAlreadyInTournamentLobby     = errors.New("account is already in a team for this tournament")
	ErrTournamentRegistrationClosed = errors.New("tournament registration is not open")
	ErrCannotRemoveCaptain          = errors.New("captain cannot be removed")
	ErrCannotDisbandLobby           = errors.New("only a forming team can be disbanded")
)

// TeamLobbyRepository uses TournamentTeam as the canonical persistence model.
// “Lobby” is only the workflow name used while a team is still forming.
type TeamLobbyRepository interface {
	FindTournamentByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.TournamentTeam, error)
	FindByInviteCode(ctx context.Context, inviteCode string) (*models.TournamentTeam, error)
	Create(ctx context.Context, team *models.TournamentTeam, captain *models.TournamentTeamMember) (*models.TournamentTeam, error)
	Join(ctx context.Context, teamID, accountID uuid.UUID) (*models.TournamentTeam, error)
	RegenerateInvite(ctx context.Context, teamID uuid.UUID, inviteCode string) (*models.TournamentTeam, error)
	Lock(ctx context.Context, teamID uuid.UUID) (*models.TournamentTeam, error)
	RemoveMember(ctx context.Context, teamID, memberID uuid.UUID) (*models.TournamentTeam, error)
	Disband(ctx context.Context, teamID uuid.UUID) error
}

type teamLobbyRepository struct {
	db *gorm.DB
}

func NewTeamLobbyRepository(db *gorm.DB) TeamLobbyRepository {
	return &teamLobbyRepository{db: db}
}

func (r *teamLobbyRepository) FindTournamentByID(ctx context.Context, id uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	err := r.db.WithContext(ctx).
		Preload("Organizer").
		Where("id = ? AND published = ?", id, true).
		First(&tournament).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tournament, nil
}

func (r *teamLobbyRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.TournamentTeam, error) {
	return r.find(ctx, r.db.WithContext(ctx).Where("id = ?", id))
}

func (r *teamLobbyRepository) FindByInviteCode(ctx context.Context, inviteCode string) (*models.TournamentTeam, error) {
	return r.find(ctx, r.db.WithContext(ctx).Where("invite_code = ?", inviteCode))
}

func (r *teamLobbyRepository) find(ctx context.Context, query *gorm.DB) (*models.TournamentTeam, error) {
	var team models.TournamentTeam
	err := query.
		Preload("Tournament.Organizer").
		Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Order("joined_at ASC") }).
		Preload("Members.Account").
		First(&team).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTeamLobbyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *teamLobbyRepository) Create(ctx context.Context, team *models.TournamentTeam, captain *models.TournamentTeamMember) (*models.TournamentTeam, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := accountHasActiveTournamentTeam(tx, team.TournamentID, captain.AccountID); err != nil {
			return err
		}
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		captain.TournamentTeamID = team.ID
		return tx.Create(captain).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, team.ID)
}

func (r *teamLobbyRepository) Join(ctx context.Context, teamID, accountID uuid.UUID) (*models.TournamentTeam, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		team, err := r.lockTeam(tx, teamID)
		if err != nil {
			return err
		}
		if team.Status != models.TournamentTeamStatusForming {
			return ErrTeamLobbyNotForming
		}
		if team.Tournament.Status != models.TournamentStatusRegistrationOpen {
			return ErrTournamentRegistrationClosed
		}
		if len(team.Members) >= team.Tournament.MaxTeamSize {
			return ErrTeamLobbyFull
		}
		if err := accountHasActiveTournamentTeam(tx, team.TournamentID, accountID); err != nil {
			return err
		}
		return tx.Create(&models.TournamentTeamMember{
			ID:               uuid.New(),
			TournamentTeamID: team.ID,
			AccountID:        accountID,
			Role:             models.TournamentTeamMemberRoleMember,
			JoinedAt:         time.Now().UTC(),
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, teamID)
}

func (r *teamLobbyRepository) RegenerateInvite(ctx context.Context, teamID uuid.UUID, inviteCode string) (*models.TournamentTeam, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		team, err := r.lockTeam(tx, teamID)
		if err != nil {
			return err
		}
		if team.Status != models.TournamentTeamStatusForming {
			return ErrTeamLobbyNotForming
		}
		return tx.Model(&models.TournamentTeam{}).
			Where("id = ?", team.ID).
			Update("invite_code", inviteCode).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, teamID)
}

func (r *teamLobbyRepository) Lock(ctx context.Context, teamID uuid.UUID) (*models.TournamentTeam, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		team, err := r.lockTeam(tx, teamID)
		if err != nil {
			return err
		}
		if team.Status != models.TournamentTeamStatusForming {
			return ErrTeamLobbyNotForming
		}
		if team.Tournament.Status != models.TournamentStatusRegistrationOpen {
			return ErrTournamentRegistrationClosed
		}
		if len(team.Members) < team.Tournament.MinTeamSize {
			return ErrRosterBelowMinimum
		}
		now := time.Now().UTC()
		return tx.Model(&models.TournamentTeam{}).Where("id = ?", team.ID).Updates(map[string]any{
			"status":    models.TournamentTeamStatusLocked,
			"locked_at": now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, teamID)
}

func (r *teamLobbyRepository) RemoveMember(ctx context.Context, teamID, memberID uuid.UUID) (*models.TournamentTeam, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		team, err := r.lockTeam(tx, teamID)
		if err != nil {
			return err
		}
		if team.Status != models.TournamentTeamStatusForming {
			return ErrTeamLobbyNotForming
		}
		var member models.TournamentTeamMember
		if err := tx.First(&member, "id = ? AND tournament_team_id = ?", memberID, teamID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrTeamLobbyNotFound
			}
			return err
		}
		if member.Role == models.TournamentTeamMemberRoleCaptain {
			return ErrCannotRemoveCaptain
		}
		return tx.Delete(&member).Error
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, teamID)
}

func (r *teamLobbyRepository) Disband(ctx context.Context, teamID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		team, err := r.lockTeam(tx, teamID)
		if err != nil {
			return err
		}
		if team.Status != models.TournamentTeamStatusForming {
			return ErrCannotDisbandLobby
		}
		if err := tx.Where("tournament_team_id = ?", team.ID).Delete(&models.TournamentTeamMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.TournamentTeam{}, "id = ?", team.ID).Error
	})
}

func (r *teamLobbyRepository) lockTeam(tx *gorm.DB, teamID uuid.UUID) (*models.TournamentTeam, error) {
	var team models.TournamentTeam
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Tournament").
		Preload("Members").
		First(&team, "id = ?", teamID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTeamLobbyNotFound
	}
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func accountHasActiveTournamentTeam(tx *gorm.DB, tournamentID, accountID uuid.UUID) error {
	var count int64
	err := tx.Model(&models.TournamentTeamMember{}).
		Joins("JOIN tournament_teams ON tournament_teams.id = tournament_team_members.tournament_team_id").
		Where("tournament_teams.tournament_id = ? AND tournament_team_members.account_id = ?", tournamentID, accountID).
		Where("tournament_teams.status IN ?", []models.TournamentTeamStatus{
			models.TournamentTeamStatusForming,
			models.TournamentTeamStatusLocked,
			models.TournamentTeamStatusAccepted,
		}).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrAlreadyInTournamentLobby
	}
	return nil
}
