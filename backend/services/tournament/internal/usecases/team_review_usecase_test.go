package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockReviewRepo struct {
	reviewFunc func(
		ctx context.Context,
		accountID uuid.UUID,
		reviewerDisplayName string,
		tournamentID, teamID uuid.UUID,
		reason string,
		decide func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error),
	) (*models.TournamentTeamReview, *uuid.UUID, error)
}

func (m *mockReviewRepo) ApplyReview(
	ctx context.Context,
	accountID uuid.UUID,
	reviewerDisplayName string,
	tournamentID, teamID uuid.UUID,
	reason string,
	decide func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error),
) (*models.TournamentTeamReview, *uuid.UUID, error) {
	return m.reviewFunc(ctx, accountID, reviewerDisplayName, tournamentID, teamID, reason, decide)
}

func (m *mockReviewRepo) PendingRefunds(ctx context.Context) ([]models.TournamentTeam, error) {
	return nil, nil
}

func (m *mockReviewRepo) MarkRefunded(ctx context.Context, teamID, paymentID uuid.UUID) error {
	return nil
}

type mockRefundService struct{}

func (m *mockRefundService) RefundRegistration(ctx context.Context, entryID, paymentID uuid.UUID) (*ports.RegistrationPayment, error) {
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
	teamID := uuid.New()

	accSvc := newStubAccountService().
		withAccount(reviewerID, "organizer", "Admin Host").
		withAccountEmail(captainID, "captain", "Team Captain", "captain@example.com")

	t.Run("accepts application and publishes registration.approved", func(t *testing.T) {
		pub := &mockPublisher{}
		repo := &mockReviewRepo{
			reviewFunc: func(
				ctx context.Context,
				accountID uuid.UUID,
				reviewerDisplayName string,
				tournID, tID uuid.UUID,
				reason string,
				decide func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error),
			) (*models.TournamentTeamReview, *uuid.UUID, error) {
				tourn := &models.Tournament{
					ID:       tournID,
					Name:     "Cyber Invitational 2026",
					Status:   models.TournamentStatusRegistrationOpen,
					Capacity: 16,
				}
				team := &models.TournamentTeam{
					ID:           tID,
					TournamentID: tournID,
					Name:         "Cyber Phantoms",
					Status:       models.TournamentTeamStatusLocked,
					Members: []models.TournamentTeamMember{
						{AccountID: captainID, Role: models.TournamentTeamMemberRoleCaptain},
					},
				}
				newStatus, err := decide(tourn, team, 5)
				require.NoError(t, err)
				assert.Equal(t, models.TournamentTeamStatusAccepted, newStatus)

				return &models.TournamentTeamReview{
					ID:                  uuid.New(),
					TournamentID:        tournID,
					Tournament:          *tourn,
					TournamentTeamID:    tID,
					Team:                *team,
					FromStatus:          models.TournamentTeamStatusLocked,
					ToStatus:            newStatus,
					Reason:              reason,
					ReviewerAccountID:   accountID,
					ReviewerDisplayName: reviewerDisplayName,
					CreatedAt:           time.Now().UTC(),
				}, nil, nil
			},
		}

		uc := usecases.NewTeamReviewUseCase(repo, accSvc, &mockRefundService{}, pub)
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
		repo := &mockReviewRepo{
			reviewFunc: func(
				ctx context.Context,
				accountID uuid.UUID,
				reviewerDisplayName string,
				tournID, tID uuid.UUID,
				reason string,
				decide func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error),
			) (*models.TournamentTeamReview, *uuid.UUID, error) {
				tourn := &models.Tournament{
					ID:       tournID,
					Name:     "Cyber Invitational 2026",
					Status:   models.TournamentStatusRegistrationOpen,
					Capacity: 16,
				}
				team := &models.TournamentTeam{
					ID:           tID,
					TournamentID: tournID,
					Name:         "Cyber Phantoms",
					Status:       models.TournamentTeamStatusLocked,
					Members: []models.TournamentTeamMember{
						{AccountID: captainID, Role: models.TournamentTeamMemberRoleCaptain},
					},
				}
				newStatus, err := decide(tourn, team, 5)
				require.NoError(t, err)
				assert.Equal(t, models.TournamentTeamStatusRejected, newStatus)

				return &models.TournamentTeamReview{
					ID:                  uuid.New(),
					TournamentID:        tournID,
					Tournament:          *tourn,
					TournamentTeamID:    tID,
					Team:                *team,
					FromStatus:          models.TournamentTeamStatusLocked,
					ToStatus:            newStatus,
					Reason:              reason,
					ReviewerAccountID:   accountID,
					ReviewerDisplayName: reviewerDisplayName,
					CreatedAt:           time.Now().UTC(),
				}, nil, nil
			},
		}

		uc := usecases.NewTeamReviewUseCase(repo, accSvc, &mockRefundService{}, pub)
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
		repo := &mockReviewRepo{
			reviewFunc: func(
				ctx context.Context,
				accountID uuid.UUID,
				reviewerDisplayName string,
				tournID, tID uuid.UUID,
				reason string,
				decide func(tournament *models.Tournament, team *models.TournamentTeam, acceptedEntries int) (models.TournamentTeamStatus, error),
			) (*models.TournamentTeamReview, *uuid.UUID, error) {
				tourn := &models.Tournament{
					ID:       tournID,
					Status:   models.TournamentStatusRegistrationOpen,
					Capacity: 16,
				}
				team := &models.TournamentTeam{
					ID:     tID,
					Status: models.TournamentTeamStatusForming,
				}
				_, err := decide(tourn, team, 0)
				return nil, nil, err
			},
		}

		uc := usecases.NewTeamReviewUseCase(repo, accSvc, &mockRefundService{}, pub)
		_, err := uc.Review(context.Background(), reviewerID, tournamentID, teamID, "ACCEPT", "")
		assert.ErrorIs(t, err, usecases.ErrApplicationNotReviewable)
		assert.Empty(t, pub.events)
	})
}
