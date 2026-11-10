package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAnnouncementRepo struct {
	tournament    *models.Tournament
	tournamentErr error
	createErr     error
	audience      *repositories.AnnouncementAudience
	audienceErr   error
	createCalls   int
	audienceCalls int
	savedTitle    string
	savedBody     string
}

func (m *mockAnnouncementRepo) GetTournament(_ context.Context, tournamentID uuid.UUID) (*models.Tournament, error) {
	if m.tournamentErr != nil {
		return nil, m.tournamentErr
	}
	if m.tournament != nil {
		return m.tournament, nil
	}
	return nil, repositories.ErrTournamentNotFound
}

func (m *mockAnnouncementRepo) Create(
	_ context.Context,
	announcement *models.TournamentAnnouncement,
) error {
	m.createCalls++
	if m.createErr != nil {
		return m.createErr
	}
	m.savedTitle, m.savedBody = announcement.Title, announcement.Body
	if announcement.ID == uuid.Nil {
		announcement.ID = uuid.New()
	}
	if announcement.CreatedAt.IsZero() {
		announcement.CreatedAt = time.Now().UTC()
	}
	return nil
}

func (m *mockAnnouncementRepo) List(context.Context, uuid.UUID) ([]models.TournamentAnnouncement, error) {
	return nil, nil
}

func (m *mockAnnouncementRepo) Audience(context.Context, uuid.UUID) (*repositories.AnnouncementAudience, error) {
	m.audienceCalls++
	return m.audience, m.audienceErr
}

// alwaysFailingPublisher stands for a broker outage.
type alwaysFailingPublisher struct{ calls int }

func (p *alwaysFailingPublisher) Publish(context.Context, string, any) error {
	p.calls++
	return errors.New("broker down")
}

func (p *alwaysFailingPublisher) Close() error { return nil }

// failFirstPublisher fails the first publish, then behaves normally.
type failFirstPublisher struct {
	calls  int
	events []recordedEvent
}

func (p *failFirstPublisher) Publish(_ context.Context, routingKey string, event any) error {
	p.calls++
	if p.calls == 1 {
		return errors.New("one message was refused")
	}
	p.events = append(p.events, recordedEvent{routingKey: routingKey, event: event})
	return nil
}

func (p *failFirstPublisher) Close() error { return nil }

// announcementEvents unwraps what was published and checks every message went
// out under the announcement routing key.
func announcementEvents(t *testing.T, recorded []recordedEvent) []events.AnnouncementCreated {
	t.Helper()
	out := make([]events.AnnouncementCreated, 0, len(recorded))
	for _, r := range recorded {
		assert.Equal(t, events.AnnouncementCreatedType, r.routingKey)
		e, ok := r.event.(events.AnnouncementCreated)
		require.True(t, ok, "published value should be an events.AnnouncementCreated, got %T", r.event)
		out = append(out, e)
	}
	return out
}

func TestAnnouncementUseCase_Create(t *testing.T) {
	authorID := uuid.New()
	tournamentID := uuid.New()
	memberA, memberB, memberC := uuid.New(), uuid.New(), uuid.New()

	accSvc := newStubAccountService().
		withAccount(authorID, "host", "Organizer Ace").
		withAccountEmail(memberA, "a", "Player A", "a@example.com").
		withAccountEmail(memberB, "b", "Player B", "B@Example.com"). // normalized to lower case
		withAccountEmail(memberC, "c", "Player C", "c@example.com")

	newAudience := func() *repositories.AnnouncementAudience {
		return &repositories.AnnouncementAudience{
			TournamentName: "Cyber Championship",
			AccountIDs:     []uuid.UUID{memberA, memberB, memberC},
		}
	}

	ownedTournament := func() *models.Tournament {
		return &models.Tournament{
			ID: tournamentID,
			Staff: []models.TournamentStaff{
				{AccountID: authorID, Role: models.TournamentStaffRoleOwner},
			},
		}
	}

	t.Run("publishes one event per member of the accepted teams", func(t *testing.T) {
		repo := &mockAnnouncementRepo{tournament: ownedTournament(), audience: newAudience()}
		pub := &mockPublisher{}
		uc := usecases.NewAnnouncementUseCase(repo, accSvc, pub)

		got, err := uc.Create(context.Background(), authorID, tournamentID, "Venue change", "The final moves to room 301.")
		require.NoError(t, err)

		sent := announcementEvents(t, pub.events)
		require.Len(t, sent, 3)

		recipients := map[uuid.UUID]string{}
		eventIDs := map[uuid.UUID]bool{}
		for _, e := range sent {
			assert.Equal(t, sent[0].CorrelationID, e.CorrelationID, "one announcement shares one correlation ID")
			assert.False(t, eventIDs[e.ID], "every event needs its own ID")
			eventIDs[e.ID] = true

			assert.Equal(t, got.ID, e.Data.AnnouncementID)
			assert.Equal(t, tournamentID, e.Data.TournamentID)
			assert.Equal(t, "Cyber Championship", e.Data.TournamentName)
			assert.Equal(t, "Venue change", e.Data.Title)
			assert.Equal(t, "The final moves to room 301.", e.Data.Body)
			assert.Equal(t, "Organizer Ace", e.Data.AuthorName)
			recipients[e.Data.Recipient.AccountID] = e.Data.Recipient.Email
		}
		assert.Equal(t, map[uuid.UUID]string{
			memberA: "a@example.com",
			memberB: "b@example.com",
			memberC: "c@example.com",
		}, recipients)
	})

	t.Run("saves the trimmed title and message", func(t *testing.T) {
		repo := &mockAnnouncementRepo{tournament: ownedTournament(), audience: newAudience()}
		uc := usecases.NewAnnouncementUseCase(repo, accSvc, &mockPublisher{})

		_, err := uc.Create(context.Background(), authorID, tournamentID, "  Venue change  ", "\n The final moves. \n")
		require.NoError(t, err)

		assert.Equal(t, "Venue change", repo.savedTitle)
		assert.Equal(t, "The final moves.", repo.savedBody)
	})

	t.Run("rejects invalid input without saving or notifying", func(t *testing.T) {
		cases := []struct {
			name    string
			title   string
			body    string
			wantErr error
		}{
			{"empty title", "", "Body", usecases.ErrAnnouncementTitleRequired},
			{"title of only spaces", "   ", "Body", usecases.ErrAnnouncementTitleRequired},
			{"empty message", "Title", "", usecases.ErrAnnouncementBodyRequired},
			{"title over 150 characters", strings.Repeat("a", 151), "Body", usecases.ErrAnnouncementTitleTooLong},
			{"message over 2000 characters", "Title", strings.Repeat("a", 2001), usecases.ErrAnnouncementBodyTooLong},
			{"Thai title over 150 characters", strings.Repeat("ก", 151), "Body", usecases.ErrAnnouncementTitleTooLong},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				repo := &mockAnnouncementRepo{tournament: ownedTournament(), audience: newAudience()}
				pub := &mockPublisher{}
				uc := usecases.NewAnnouncementUseCase(repo, accSvc, pub)

				_, err := uc.Create(context.Background(), authorID, tournamentID, tc.title, tc.body)

				assert.ErrorIs(t, err, tc.wantErr)
				assert.Zero(t, repo.createCalls)
				assert.Empty(t, pub.events)
			})
		}
	})

	t.Run("counts characters, not bytes", func(t *testing.T) {
		// 150 Thai letters take 450 bytes, 2000 take 6000: both are within the limits.
		repo := &mockAnnouncementRepo{tournament: ownedTournament(), audience: newAudience()}
		uc := usecases.NewAnnouncementUseCase(repo, accSvc, &mockPublisher{})

		_, err := uc.Create(context.Background(), authorID, tournamentID, strings.Repeat("ก", 150), strings.Repeat("ก", 2000))

		require.NoError(t, err)
	})

	t.Run("notifies nobody when the post is refused", func(t *testing.T) {
		repo := &mockAnnouncementRepo{tournamentErr: repositories.ErrTournamentNotFound, audience: newAudience()}
		pub := &mockPublisher{}
		uc := usecases.NewAnnouncementUseCase(repo, accSvc, pub)

		_, err := uc.Create(context.Background(), authorID, tournamentID, "Title", "Body")

		assert.ErrorIs(t, err, usecases.ErrTournamentNotFound)
		assert.Zero(t, repo.audienceCalls)
		assert.Empty(t, pub.events)
	})

	t.Run("sends nothing when no team is accepted yet", func(t *testing.T) {
		repo := &mockAnnouncementRepo{tournament: ownedTournament(), audience: &repositories.AnnouncementAudience{TournamentName: "Cyber Championship"}}
		pub := &mockPublisher{}
		uc := usecases.NewAnnouncementUseCase(repo, accSvc, pub)

		_, err := uc.Create(context.Background(), authorID, tournamentID, "Title", "Body")

		require.NoError(t, err)
		assert.Empty(t, pub.events)
	})

	t.Run("still posts when publishing fails", func(t *testing.T) {
		repo := &mockAnnouncementRepo{tournament: ownedTournament(), audience: newAudience()}
		pub := &alwaysFailingPublisher{}
		uc := usecases.NewAnnouncementUseCase(repo, accSvc, pub)

		got, err := uc.Create(context.Background(), authorID, tournamentID, "Title", "Body")

		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, 3, pub.calls, "every recipient is still tried")
	})

	t.Run("one failed publish does not stop the others", func(t *testing.T) {
		repo := &mockAnnouncementRepo{tournament: ownedTournament(), audience: newAudience()}
		pub := &failFirstPublisher{}
		uc := usecases.NewAnnouncementUseCase(repo, accSvc, pub)

		_, err := uc.Create(context.Background(), authorID, tournamentID, "Title", "Body")

		require.NoError(t, err)
		assert.Len(t, announcementEvents(t, pub.events), 2)
	})

	t.Run("still posts when the audience lookup fails", func(t *testing.T) {
		repo := &mockAnnouncementRepo{tournament: ownedTournament(), audienceErr: errors.New("database unavailable")}
		pub := &mockPublisher{}
		uc := usecases.NewAnnouncementUseCase(repo, accSvc, pub)

		got, err := uc.Create(context.Background(), authorID, tournamentID, "Title", "Body")

		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, pub.events)
	})

	t.Run("works without a publisher", func(t *testing.T) {
		repo := &mockAnnouncementRepo{tournament: ownedTournament(), audience: newAudience()}
		uc := usecases.NewAnnouncementUseCase(repo, accSvc, nil)

		_, err := uc.Create(context.Background(), authorID, tournamentID, "Title", "Body")

		require.NoError(t, err)
	})
}
