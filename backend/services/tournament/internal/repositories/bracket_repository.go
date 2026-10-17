package repositories

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrBracketTournamentNotFound = errors.New("tournament not found")
	ErrBracketStaffNotFound      = errors.New("staff member not found")
	ErrBracketMatchNotFound      = errors.New("match not found")
	ErrBracketMatchCompleted     = errors.New("match is already completed")
)

type BracketRepository interface {
	GetTournament(ctx context.Context, tournamentID uuid.UUID) (*models.Tournament, error)
	GetStaffRole(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.TournamentStaffRole, error)
	HasCompletedMatches(ctx context.Context, tournamentID uuid.UUID) (bool, error)
	SaveBracketLayout(ctx context.Context, tournamentID uuid.UUID, matches []models.Match) error
	GetBracket(ctx context.Context, tournamentID uuid.UUID) ([]models.Match, error)
	GetStage1Matches(ctx context.Context, tournamentID uuid.UUID) ([]models.Match, error)
	GetAcceptedTeams(ctx context.Context, tournamentID uuid.UUID) ([]models.TournamentTeam, error)
	SeedStage1(ctx context.Context, tournamentID uuid.UUID, stage1MatchIDs []uuid.UUID, entries []models.MatchEntry) error
	GetMatch(ctx context.Context, tournamentID, matchID uuid.UUID) (*models.Match, error)
	RecordMatchResults(
		ctx context.Context,
		tournamentID, matchID uuid.UUID,
		rankUpdates map[uuid.UUID]int,
		scoreUpdates map[uuid.UUID]*string,
		advanceFlags map[uuid.UUID]bool,
		targetMatchID *uuid.UUID,
		advancingTeamIDs []uuid.UUID,
	) error
}

type bracketRepository struct {
	db *gorm.DB
}

func NewBracketRepository(db *gorm.DB) BracketRepository {
	return &bracketRepository{db: db}
}

func (r *bracketRepository) GetTournament(ctx context.Context, tournamentID uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	if err := r.db.WithContext(ctx).First(&tournament, "id = ?", tournamentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBracketTournamentNotFound
		}
		return nil, err
	}
	return &tournament, nil
}

func (r *bracketRepository) GetStaffRole(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.TournamentStaffRole, error) {
	var staff models.TournamentStaff
	if err := r.db.WithContext(ctx).
		First(&staff, "tournament_id = ? AND account_id = ?", tournamentID, accountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBracketStaffNotFound
		}
		return nil, err
	}
	return &staff.Role, nil
}

func (r *bracketRepository) HasCompletedMatches(ctx context.Context, tournamentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Match{}).
		Where("tournament_id = ? AND is_completed = true", tournamentID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *bracketRepository) SaveBracketLayout(ctx context.Context, tournamentID uuid.UUID, matches []models.Match) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var completedCount int64
		if err := tx.Model(&models.Match{}).
			Where("tournament_id = ? AND is_completed = true", tournamentID).
			Count(&completedCount).Error; err != nil {
			return err
		}
		if completedCount > 0 {
			return ErrBracketMatchCompleted
		}

		// Delete existing match entries for matches of this tournament
		subQuery := tx.Model(&models.Match{}).Select("id").Where("tournament_id = ?", tournamentID)
		if err := tx.Where("match_id IN (?)", subQuery).Delete(&models.MatchEntry{}).Error; err != nil {
			return err
		}

		// Delete existing matches for this tournament
		if err := tx.Where("tournament_id = ?", tournamentID).Delete(&models.Match{}).Error; err != nil {
			return err
		}

		if len(matches) == 0 {
			return nil
		}

		// To respect self-referencing foreign keys, insert target matches (higher stage levels) first.
		sortedMatches := make([]models.Match, len(matches))
		copy(sortedMatches, matches)
		sort.Slice(sortedMatches, func(i, j int) bool {
			return sortedMatches[i].StageLevel > sortedMatches[j].StageLevel
		})

		// Insert nodes with target_match_id = nil first to prevent any cyclic dependency issues
		for i := range sortedMatches {
			targetID := sortedMatches[i].TargetMatchID
			sortedMatches[i].TargetMatchID = nil
			if err := tx.Create(&sortedMatches[i]).Error; err != nil {
				return err
			}
			sortedMatches[i].TargetMatchID = targetID
		}

		// Now update target_match_id for those that have one
		for _, m := range sortedMatches {
			if m.TargetMatchID != nil {
				if err := tx.Model(&models.Match{}).
					Where("id = ?", m.ID).
					Update("target_match_id", m.TargetMatchID).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (r *bracketRepository) GetBracket(ctx context.Context, tournamentID uuid.UUID) ([]models.Match, error) {
	var matches []models.Match
	err := r.db.WithContext(ctx).
		Preload("Entries", func(db *gorm.DB) *gorm.DB {
			return db.Order("match_entries.final_rank ASC NULLS LAST, match_entries.created_at ASC")
		}).
		Preload("Entries.Team").
		Where("tournament_id = ?", tournamentID).
		Order("stage_level ASC, name ASC").
		Find(&matches).Error
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *bracketRepository) GetStage1Matches(ctx context.Context, tournamentID uuid.UUID) ([]models.Match, error) {
	var matches []models.Match
	err := r.db.WithContext(ctx).
		Where("tournament_id = ? AND stage_level = 1", tournamentID).
		Order("name ASC, id ASC").
		Find(&matches).Error
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (r *bracketRepository) GetAcceptedTeams(ctx context.Context, tournamentID uuid.UUID) ([]models.TournamentTeam, error) {
	var teams []models.TournamentTeam
	err := r.db.WithContext(ctx).
		Where("tournament_id = ? AND status = ?", tournamentID, models.TournamentTeamStatusAccepted).
		Order("name ASC, id ASC").
		Find(&teams).Error
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *bracketRepository) SeedStage1(ctx context.Context, tournamentID uuid.UUID, stage1MatchIDs []uuid.UUID, entries []models.MatchEntry) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(stage1MatchIDs) > 0 {
			// Clear existing entries for stage 1 matches
			if err := tx.Where("match_id IN ?", stage1MatchIDs).Delete(&models.MatchEntry{}).Error; err != nil {
				return err
			}
		}

		if len(entries) > 0 {
			if err := tx.Create(&entries).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *bracketRepository) GetMatch(ctx context.Context, tournamentID, matchID uuid.UUID) (*models.Match, error) {
	var match models.Match
	err := r.db.WithContext(ctx).
		Preload("Entries", func(db *gorm.DB) *gorm.DB {
			return db.Order("final_rank ASC NULLS LAST")
		}).
		Preload("Entries.Team").
		Preload("TargetMatch").
		First(&match, "id = ? AND tournament_id = ?", matchID, tournamentID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBracketMatchNotFound
		}
		return nil, err
	}
	return &match, nil
}

func (r *bracketRepository) RecordMatchResults(
	ctx context.Context,
	tournamentID, matchID uuid.UUID,
	rankUpdates map[uuid.UUID]int,
	scoreUpdates map[uuid.UUID]*string,
	advanceFlags map[uuid.UUID]bool,
	targetMatchID *uuid.UUID,
	advancingTeamIDs []uuid.UUID,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update final_rank, score, and advances on current match entries
		for teamID, rank := range rankUpdates {
			advances := advanceFlags[teamID]
			updates := map[string]any{
				"final_rank": rank,
				"advances":   advances,
			}
			if scoreUpdates != nil {
				updates["score"] = scoreUpdates[teamID]
			}
			if err := tx.Model(&models.MatchEntry{}).
				Where("match_id = ? AND team_id = ?", matchID, teamID).
				Updates(updates).Error; err != nil {
				return err
			}
		}

		// Insert advancing teams into target match if one exists
		if targetMatchID != nil && len(advancingTeamIDs) > 0 {
			targetEntries := make([]models.MatchEntry, 0, len(advancingTeamIDs))
			for _, teamID := range advancingTeamIDs {
				targetEntries = append(targetEntries, models.MatchEntry{
					MatchID:   *targetMatchID,
					TeamID:    teamID,
					FinalRank: nil,
					Advances:  false,
				})
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&targetEntries).Error; err != nil {
				return err
			}
		}

		// Mark match completed
		if err := tx.Model(&models.Match{}).
			Where("id = ? AND tournament_id = ?", matchID, tournamentID).
			Update("is_completed", true).Error; err != nil {
			return err
		}

		return nil
	})
}
