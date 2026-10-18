package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/uuid"
	sharedmiddlewares "github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/handlers/rest"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBracketRepoForHandler struct {
	tournament     *models.Tournament
	staffRole      models.TournamentStaffRole
	matches        []models.Match
	stage1Matches  []models.Match
	acceptedTeams  []models.TournamentTeam
	matchMap       map[uuid.UUID]*models.Match
}

func (m *mockBracketRepoForHandler) GetTournament(_ context.Context, id uuid.UUID) (*models.Tournament, error) {
	if m.tournament != nil && m.tournament.ID == id {
		return m.tournament, nil
	}
	return nil, repositories.ErrBracketTournamentNotFound
}

func (m *mockBracketRepoForHandler) GetStaffRole(_ context.Context, _, _ uuid.UUID) (*models.TournamentStaffRole, error) {
	return &m.staffRole, nil
}

func (m *mockBracketRepoForHandler) HasCompletedMatches(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (m *mockBracketRepoForHandler) SaveBracketLayout(_ context.Context, _ uuid.UUID, matches []models.Match) error {
	m.matches = matches
	return nil
}

func (m *mockBracketRepoForHandler) GetBracket(_ context.Context, _ uuid.UUID) ([]models.Match, error) {
	return m.matches, nil
}

func (m *mockBracketRepoForHandler) GetStage1Matches(_ context.Context, _ uuid.UUID) ([]models.Match, error) {
	return m.stage1Matches, nil
}

func (m *mockBracketRepoForHandler) GetAcceptedTeams(_ context.Context, _ uuid.UUID) ([]models.TournamentTeam, error) {
	return m.acceptedTeams, nil
}

func (m *mockBracketRepoForHandler) SeedStage1(_ context.Context, _ uuid.UUID, _ []uuid.UUID, entries []models.MatchEntry) error {
	for i := range m.matches {
		m.matches[i].Entries = nil
		for _, e := range entries {
			if e.MatchID == m.matches[i].ID {
				m.matches[i].Entries = append(m.matches[i].Entries, e)
			}
		}
	}
	return nil
}

func (m *mockBracketRepoForHandler) GetMatch(_ context.Context, _, matchID uuid.UUID) (*models.Match, error) {
	match, ok := m.matchMap[matchID]
	if !ok {
		return nil, repositories.ErrBracketMatchNotFound
	}
	return match, nil
}

func (m *mockBracketRepoForHandler) RecordMatchResults(
	_ context.Context,
	_, matchID uuid.UUID,
	rankUpdates map[uuid.UUID]int,
	scoreUpdates map[uuid.UUID]*string,
	advanceFlags map[uuid.UUID]bool,
	targetMatchID *uuid.UUID,
	advancingTeamIDs []uuid.UUID,
) error {
	if match, ok := m.matchMap[matchID]; ok {
		match.IsCompleted = true
		for i := range match.Entries {
			tID := match.Entries[i].TeamID
			if r, exists := rankUpdates[tID]; exists {
				rank := r
				match.Entries[i].FinalRank = &rank
			}
			if scoreUpdates != nil {
				match.Entries[i].Score = scoreUpdates[tID]
			}
			match.Entries[i].Advances = advanceFlags[tID]
		}
	}
	return nil
}

func TestBracketRoutes(t *testing.T) {
	ownerAccountID := uuid.New()
	tournamentID := uuid.New()

	tourney := &models.Tournament{
		ID:     tournamentID,
		Status: models.TournamentStatusRegistrationClosed,
	}

	repo := &mockBracketRepoForHandler{
		tournament: tourney,
		staffRole:  models.TournamentStaffRoleOwner,
		matchMap:   make(map[uuid.UUID]*models.Match),
	}
	useCase := usecases.NewBracketUseCase(repo)

	_, api := humatest.New(t)

	authMW := func(ctx huma.Context, next func(huma.Context)) {
		newCtx := context.WithValue(ctx.Context(), sharedmiddlewares.UserSubContextKey, ownerAccountID.String())
		ctx = huma.WithContext(ctx, newCtx)
		next(ctx)
	}

	rest.RegisterBracketRoutes(api, authMW, useCase)

	t.Run("POST /tournaments/{id}/bracket creates layout", func(t *testing.T) {
		body := map[string]any{
			"matches": []map[string]any{
				{
					"tempId":         "semi-1",
					"name":           "Semifinal 1",
					"venue":          "Arena A",
					"stageLevel":     1,
					"teamCount":      2,
					"advancingCount": 1,
					"targetMatchId":  "final",
				},
				{
					"tempId":         "semi-2",
					"name":           "Semifinal 2",
					"venue":          "Arena B",
					"stageLevel":     1,
					"teamCount":      2,
					"advancingCount": 1,
					"targetMatchId":  "final",
				},
				{
					"tempId":         "final",
					"name":           "Championship",
					"venue":          "Main Stage",
					"stageLevel":     2,
					"teamCount":      2,
					"advancingCount": 0,
				},
			},
		}

		resp := api.Post("/api/v1/tournaments/"+tournamentID.String()+"/bracket", body)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var res struct {
			Matches []rest.MatchResponse `json:"matches"`
		}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &res))
		assert.Equal(t, 3, len(res.Matches))
	})

	t.Run("POST /tournaments/{id}/bracket resets bracket with empty matches", func(t *testing.T) {
		body := map[string]any{
			"matches": []map[string]any{},
		}

		resp := api.Post("/api/v1/tournaments/"+tournamentID.String()+"/bracket", body)
		assert.Equal(t, http.StatusCreated, resp.Code)

		var res struct {
			Matches []rest.MatchResponse `json:"matches"`
		}
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &res))
		assert.Equal(t, 0, len(res.Matches))
	})

	t.Run("POST /tournaments/{id}/seed auto-distributes teams", func(t *testing.T) {
		stage1MatchID := uuid.New()
		team1ID := uuid.New()
		team2ID := uuid.New()

		repo.stage1Matches = []models.Match{
			{ID: stage1MatchID, TournamentID: tournamentID, Name: "Heat 1", StageLevel: 1, TeamCount: 2},
		}
		repo.acceptedTeams = []models.TournamentTeam{
			{ID: team1ID, TournamentID: tournamentID, Name: "Team 1", Status: models.TournamentTeamStatusAccepted},
			{ID: team2ID, TournamentID: tournamentID, Name: "Team 2", Status: models.TournamentTeamStatusAccepted},
		}
		repo.matches = repo.stage1Matches

		resp := api.Post("/api/v1/tournaments/"+tournamentID.String()+"/seed", map[string]any{})
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("POST /tournaments/{id}/matches/{matchId}/results records results", func(t *testing.T) {
		matchID := uuid.New()
		team1ID := uuid.New()
		team2ID := uuid.New()

		targetMatchID := uuid.New()
		repo.matchMap[targetMatchID] = &models.Match{
			ID:           targetMatchID,
			TournamentID: tournamentID,
			Name:         "Final",
			StageLevel:   2,
			IsCompleted:  false,
		}

		repo.matchMap[matchID] = &models.Match{
			ID:             matchID,
			TournamentID:   tournamentID,
			Name:           "Heat 1",
			StageLevel:     1,
			AdvancingCount: 1,
			TargetMatchID:  &targetMatchID,
			IsCompleted:    false,
			Entries: []models.MatchEntry{
				{MatchID: matchID, TeamID: team1ID},
				{MatchID: matchID, TeamID: team2ID},
			},
		}

		body := map[string]any{
			"placements": []map[string]any{
				{"teamId": team1ID.String(), "finalRank": 1, "score": "2"},
				{"teamId": team2ID.String(), "finalRank": 2, "score": "1"},
			},
		}

		resp := api.Post("/api/v1/tournaments/"+tournamentID.String()+"/matches/"+matchID.String()+"/results", body)
		assert.Equal(t, http.StatusOK, resp.Code)

		var res rest.RecordMatchResultsResponse
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &res))
		assert.True(t, res.IsCompleted)
		assert.Equal(t, 1, len(res.AdvancingTeams))
		assert.Equal(t, team1ID, res.AdvancingTeams[0])
		require.NotNil(t, repo.matchMap[matchID].Entries[0].Score)
		assert.Equal(t, "2", *repo.matchMap[matchID].Entries[0].Score)
		require.NotNil(t, repo.matchMap[matchID].Entries[1].Score)
		assert.Equal(t, "1", *repo.matchMap[matchID].Entries[1].Score)
	})

	t.Run("GET /tournaments/{id}/bracket returns current bracket", func(t *testing.T) {
		resp := api.Get("/api/v1/tournaments/" + tournamentID.String() + "/bracket")
		assert.Equal(t, http.StatusOK, resp.Code)
	})
}
