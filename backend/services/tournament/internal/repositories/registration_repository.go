package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/registration"
	"github.com/moomaideng/eventory/services/tournament/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *teamLobbyRepository) FindRegistrationForm(ctx context.Context, id uuid.UUID) (*models.RegistrationForm, error) {
	return findRegistrationForm(r.db.WithContext(ctx), id)
}

func findRegistrationForm(db *gorm.DB, id uuid.UUID) (*models.RegistrationForm, error) {
	form := models.EmptyRegistrationForm(id)
	err := db.First(form, "tournament_id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.EmptyRegistrationForm(id), nil
	}
	return form, err
}

// The tournament row serializes form edits, create and join operations. This
// also prevents the same account from simultaneously joining different teams.
func lockRegistrationTournament(tx *gorm.DB, id uuid.UUID) (*models.Tournament, error) {
	var tournament models.Tournament
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&tournament, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tournament, nil
}

func saveMemberRegistration(tx *gorm.DB, id uuid.UUID, member *models.TournamentTeamMember, submission models.RegistrationSubmission) error {
	form, err := findRegistrationForm(tx, id)
	if err != nil {
		return err
	}
	if err := registration.ValidateSubmission(form, submission); err != nil {
		return err
	}
	now := time.Now().UTC()
	member.FormVersion = form.Version
	member.RegistrationAnswers = submission.Answers
	member.RegistrationQuestions = form.Questions
	member.ConsentNotice = form.ConsentNotice
	member.ConsentedAt = &now
	return nil
}

func (r *teamLobbyRepository) SaveRegistrationForm(ctx context.Context, id, organizerID uuid.UUID, form *models.RegistrationForm) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tournament, err := lockRegistrationTournament(tx, id)
		if err != nil {
			return err
		}
		if tournament.OrganizerID != organizerID {
			return ErrFormForbidden
		}
		old, err := findRegistrationForm(tx, id)
		if err != nil {
			return err
		}
		form.TournamentID, form.Version = id, old.Version+1
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tournament_id"}}, UpdateAll: true}).Create(form).Error
	})
}

var ErrFormForbidden = errors.New("registration data access denied")

// Raw answers and files never appear in a shared lobby response.
func (r *teamLobbyRepository) FindMemberRegistration(ctx context.Context, memberID, viewerID, organizerID uuid.UUID) (*models.TournamentTeamMember, error) {
	var member models.TournamentTeamMember
	if err := r.db.WithContext(ctx).First(&member, "id = ?", memberID).Error; err != nil {
		return nil, err
	}
	if member.AccountID != viewerID {
		var team models.TournamentTeam
		if err := r.db.WithContext(ctx).Preload("Tournament").First(&team, "id = ?", member.TournamentTeamID).Error; err != nil {
			return nil, err
		}
		if team.Tournament.OrganizerID != organizerID || organizerID == uuid.Nil {
			return nil, ErrFormForbidden
		}
	}
	return &member, nil
}
