package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockReviewRepo struct {
	tournament      *models.Tournament
	team            *models.TournamentTeam
	acceptedEntries int
}

type mockReviewTx struct {
	repo *mockReviewRepo
}

func (m *mockReviewRepo) InTransaction(ctx context.Context, fn func(repositories.TeamReviewTransaction) error) error {
	return fn(&mockReviewTx{repo: m})
}

func (tx *mockReviewTx) LockTournament(tournamentID uuid.UUID) (*models.Tournament, error) {
	if tx.repo.tournament == nil {
		return nil, repositories.ErrTournamentNotFound
	}
	return tx.repo.tournament, nil
}

func (tx *mockReviewTx) LockTeamAndAcceptedCount(tournamentID uuid.UUID, teamID string) (*models.TournamentTeam, int, error) {
	if tx.repo.team == nil {
		return nil, 0, repositories.ErrTeamEntryNotFound
	}
	return tx.repo.team, tx.repo.acceptedEntries, nil
}

func (tx *mockReviewTx) SaveReview(team *models.TournamentTeam, toStatus models.TournamentTeamStatus, review *models.TournamentTeamReview) error {
	team.Status = toStatus
	review.ID = uuid.New()
	review.CreatedAt = time.Now().UTC()
	return nil
}

func (m *mockReviewRepo) PendingRefunds(ctx context.Context) ([]models.TournamentTeam, error) {
	return nil, nil
}

func (m *mockReviewRepo) MarkRefunded(ctx context.Context, tournamentID uuid.UUID, teamID string, paymentID uuid.UUID) error {
	return nil
}

type mockRefundService struct{}

func (m *mockRefundService) RefundRegistration(ctx context.Context, paymentID uuid.UUID, entryID string) (*ports.RegistrationPayment, error) {
	return nil, nil
}

type recordedEvent struct {
	routingKey string
	event      any
}

type mockPublisher struct {
	events []recordedEvent
}

func (m *mockPublisher) Publish(ctx context.Context, routingKey string, event any) error {
	m.events = append(m.events, recordedEvent{routingKey: routingKey, event: event})
	return nil
}

func (m *mockPublisher) Close() error {
	return nil
}

func TestTeamReviewUseCase_Review(t *testing.T) {
	reviewerID := uuid.New()
	captainID := uuid.New()
	tournamentID := uuid.New()
	teamID := "Ab12Cd34"

	accSvc := newStubAccountService().
		withAccount(reviewerID, "organizer", "Admin Host").
		withAccountEmail(captainID, "captain", "Team Captain", "captain@example.com")

	newLockedRepo := func() *mockReviewRepo {
		return &mockReviewRepo{
			tournament: &models.Tournament{
				ID:       tournamentID,
				Name:     "Cyber Invitational 2026",
				Status:   models.TournamentStatusRegistration,
				Capacity: 16,
				Staff: []models.TournamentStaff{
					{AccountID: reviewerID, Role: models.TournamentStaffRoleOwner},
				},
			},
			team: &models.TournamentTeam{
				ID:           teamID,
				TournamentID: tournamentID,
				Name:         "Cyber Phantoms",
				Status:       models.TournamentTeamStatusLocked,
				Members: []models.TournamentTeamMember{
					{AccountID: captainID, Role: models.TournamentTeamMemberRoleCaptain},
				},
			},
			acceptedEntries: 5,
		}
	}

	t.Run("accepts application and publishes registration.approved", func(t *testing.T) {
		pub := &mockPublisher{}
		uc := usecases.NewTeamReviewUseCase(newLockedRepo(), accSvc, &mockRefundService{}, pub)
		res, err := uc.Review(context.Background(), reviewerID, tournamentID, teamID, "ACCEPT", "")
		require.NoError(t, err)
		assert.Equal(t, "ACCEPTED", res.ToStatus)

		require.Len(t, pub.events, 1)
		assert.Equal(t, events.RegistrationApprovedType, pub.events[0].routingKey)
		revEvent, ok := pub.events[0].event.(events.RegistrationReviewed)
		require.True(t, ok)
		assert.Equal(t, events.RegistrationApprovedType, revEvent.Type)
		assert.Equal(t, tournamentID, revEvent.Data.TournamentID)
		assert.Equal(t, "Cyber Invitational 2026", revEvent.Data.TournamentName)
		assert.Equal(t, teamID, revEvent.Data.TeamID)
		assert.Equal(t, "Cyber Phantoms", revEvent.Data.TeamName)
		assert.Equal(t, captainID, revEvent.Data.Recipient.AccountID)
		assert.Equal(t, "Team Captain", revEvent.Data.Recipient.DisplayName)
		assert.Equal(t, "captain@example.com", revEvent.Data.Recipient.Email)
	})

	t.Run("rejects application and publishes registration.rejected", func(t *testing.T) {
		pub := &mockPublisher{}
		uc := usecases.NewTeamReviewUseCase(newLockedRepo(), accSvc, &mockRefundService{}, pub)
		res, err := uc.Review(context.Background(), reviewerID, tournamentID, teamID, "REJECT", "Incomplete roster")
		require.NoError(t, err)
		assert.Equal(t, "REJECTED", res.ToStatus)

		require.Len(t, pub.events, 1)
		assert.Equal(t, events.RegistrationRejectedType, pub.events[0].routingKey)
		revEvent, ok := pub.events[0].event.(events.RegistrationReviewed)
		require.True(t, ok)
		assert.Equal(t, events.RegistrationRejectedType, revEvent.Type)
		assert.Equal(t, "captain@example.com", revEvent.Data.Recipient.Email)
		assert.Equal(t, "Incomplete roster", revEvent.Data.Reason)
	})

	t.Run("returns error when application is not in locked status", func(t *testing.T) {
		pub := &mockPublisher{}
		repo := newLockedRepo()
		repo.team.Status = models.TournamentTeamStatusForming

		uc := usecases.NewTeamReviewUseCase(repo, accSvc, &mockRefundService{}, pub)
		_, err := uc.Review(context.Background(), reviewerID, tournamentID, teamID, "ACCEPT", "")
		assert.ErrorIs(t, err, usecases.ErrApplicationNotReviewable)
		assert.Empty(t, pub.events)
	})
}
