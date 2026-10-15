package usecases_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type mockBracketRepository struct {
	tournament              *models.Tournament
	tournamentErr           error
	staffRole               *models.TournamentStaffRole
	staffErr                error
	hasCompletedMatches     bool
	hasCompletedMatchesErr  error
	savedMatches            []models.Match
	saveBracketLayoutErr    error
	bracketMatches          []models.Match
	bracketErr              error
	stage1Matches           []models.Match
	stage1Err               error
	acceptedTeams           []models.TournamentTeam
	acceptedTeamsErr        error
	seededEntries           []models.MatchEntry
	seedErr                 error
	matchesMap              map[uuid.UUID]*models.Match
	getMatchErr             error
	recordResultsErr        error
	lastRecordedRankUpdates map[uuid.UUID]int
	lastRecordedAdvances    map[uuid.UUID]bool
	lastAdvancingTeamIDs    []uuid.UUID
}

func (m *mockBracketRepository) GetTournament(_ context.Context, _ uuid.UUID) (*models.Tournament, error) {
	if m.tournamentErr != nil {
		return nil, m.tournamentErr
	}
	return m.tournament, nil
}

func (m *mockBracketRepository) GetStaffRole(_ context.Context, _, _ uuid.UUID) (*models.TournamentStaffRole, error) {
	if m.staffErr != nil {
		return nil, m.staffErr
	}
	return m.staffRole, nil
}

func (m *mockBracketRepository) HasCompletedMatches(_ context.Context, _ uuid.UUID) (bool, error) {
	return m.hasCompletedMatches, m.hasCompletedMatchesErr
}

func (m *mockBracketRepository) SaveBracketLayout(_ context.Context, _ uuid.UUID, matches []models.Match) error {
	if m.saveBracketLayoutErr != nil {
		return m.saveBracketLayoutErr
	}
	m.savedMatches = matches
	m.bracketMatches = matches
	return nil
}

func (m *mockBracketRepository) GetBracket(_ context.Context, _ uuid.UUID) ([]models.Match, error) {
	if m.bracketErr != nil {
		return nil, m.bracketErr
	}
	return m.bracketMatches, nil
}

func (m *mockBracketRepository) GetStage1Matches(_ context.Context, _ uuid.UUID) ([]models.Match, error) {
	if m.stage1Err != nil {
		return nil, m.stage1Err
	}
	return m.stage1Matches, nil
}

func (m *mockBracketRepository) GetAcceptedTeams(_ context.Context, _ uuid.UUID) ([]models.TournamentTeam, error) {
	if m.acceptedTeamsErr != nil {
		return nil, m.acceptedTeamsErr
	}
	return m.acceptedTeams, nil
}

func (m *mockBracketRepository) SeedStage1(_ context.Context, _ uuid.UUID, _ []uuid.UUID, entries []models.MatchEntry) error {
	if m.seedErr != nil {
		return m.seedErr
	}
	m.seededEntries = entries
	return nil
}

func (m *mockBracketRepository) GetMatch(_ context.Context, _, matchID uuid.UUID) (*models.Match, error) {
	if m.getMatchErr != nil {
		return nil, m.getMatchErr
	}
	match, ok := m.matchesMap[matchID]
	if !ok {
		return nil, repositories.ErrBracketMatchNotFound
	}
	return match, nil
}

func (m *mockBracketRepository) RecordMatchResults(
	_ context.Context,
	_, _ uuid.UUID,
	rankUpdates map[uuid.UUID]int,
	advanceFlags map[uuid.UUID]bool,
	_ *uuid.UUID,
	advancingTeamIDs []uuid.UUID,
) error {
	if m.recordResultsErr != nil {
		return m.recordResultsErr
	}
	m.lastRecordedRankUpdates = rankUpdates
	m.lastRecordedAdvances = advanceFlags
	m.lastAdvancingTeamIDs = advancingTeamIDs
	return nil
}

func TestDAGValidation(t *testing.T) {
	uc := usecases.NewBracketUseCase(&mockBracketRepository{})

	t.Run("empty match list", func(t *testing.T) {
		err := uc.ValidateBracketLayout(nil)
		if err != usecases.ErrBracketEmpty {
			t.Fatalf("expected ErrBracketEmpty, got %v", err)
		}
	})

	t.Run("empty temp ID", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "   ", Name: "Match 1", StageLevel: 1, AdvancingCount: 1},
		})
		if err != usecases.ErrBracketEmpty {
			t.Fatalf("expected ErrBracketEmpty, got %v", err)
		}
	})

	t.Run("duplicate temp ID", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, AdvancingCount: 1},
			{TempID: "m1", Name: "Match 2", StageLevel: 1, AdvancingCount: 1},
		})
		if err != usecases.ErrBracketDuplicateTempID {
			t.Fatalf("expected ErrBracketDuplicateTempID, got %v", err)
		}
	})

	t.Run("empty match name", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "   ", StageLevel: 1, AdvancingCount: 1},
		})
		if err != usecases.ErrBracketEmptyMatchName {
			t.Fatalf("expected ErrBracketEmptyMatchName, got %v", err)
		}
	})

	t.Run("invalid stage level", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 0, AdvancingCount: 1},
		})
		if err != usecases.ErrBracketInvalidStageLevel {
			t.Fatalf("expected ErrBracketInvalidStageLevel, got %v", err)
		}
	})

	t.Run("invalid advancing count", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, AdvancingCount: -1},
		})
		if err != usecases.ErrBracketInvalidAdvancingCount {
			t.Fatalf("expected ErrBracketInvalidAdvancingCount, got %v", err)
		}
	})

	t.Run("advancing count must be >= 1 if target is specified", func(t *testing.T) {
		target := "final"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Semi 1", StageLevel: 1, AdvancingCount: 0, TargetTempID: &target},
			{TempID: "final", Name: "Final", StageLevel: 2, AdvancingCount: 0},
		})
		if err != usecases.ErrBracketAdvancingCountRequired {
			t.Fatalf("expected ErrBracketAdvancingCountRequired, got %v", err)
		}
	})

	t.Run("self-targeting match", func(t *testing.T) {
		self := "m1"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, AdvancingCount: 1, TargetTempID: &self},
		})
		if err != usecases.ErrBracketSelfTargetMatch {
			t.Fatalf("expected ErrBracketSelfTargetMatch, got %v", err)
		}
	})

	t.Run("target not found", func(t *testing.T) {
		target := "m99"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, AdvancingCount: 1, TargetTempID: &target},
		})
		if err != usecases.ErrBracketTargetMatchNotFound {
			t.Fatalf("expected ErrBracketTargetMatchNotFound, got %v", err)
		}
	})

	t.Run("backward edge (lower stage level)", func(t *testing.T) {
		target := "m1"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, AdvancingCount: 0},
			{TempID: "m2", Name: "Match 2", StageLevel: 2, AdvancingCount: 1, TargetTempID: &target},
		})
		if err != usecases.ErrBracketTargetMustBeHigher {
			t.Fatalf("expected ErrBracketTargetMustBeHigher, got %v", err)
		}
	})

	t.Run("same stage level edge", func(t *testing.T) {
		target := "m2"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, AdvancingCount: 1, TargetTempID: &target},
			{TempID: "m2", Name: "Match 2", StageLevel: 1, AdvancingCount: 0},
		})
		if err != usecases.ErrBracketTargetMustBeHigher {
			t.Fatalf("expected ErrBracketTargetMustBeHigher, got %v", err)
		}
	})

	t.Run("no championship match (every node points to another)", func(t *testing.T) {
		m2 := "m2"
		m3 := "m3"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, AdvancingCount: 1, TargetTempID: &m2},
			{TempID: "m2", Name: "Match 2", StageLevel: 2, AdvancingCount: 1, TargetTempID: &m3},
		})
		// m3 is missing
		if err != usecases.ErrBracketTargetMatchNotFound {
			t.Fatalf("expected ErrBracketTargetMatchNotFound, got %v", err)
		}
	})

	t.Run("valid 4-team single elimination bracket", func(t *testing.T) {
		final := "final"
		matches := []usecases.DefineBracketMatchInput{
			{TempID: "semi-1", Name: "Semifinal 1", Venue: "Court A", StageLevel: 1, AdvancingCount: 1, TargetTempID: &final},
			{TempID: "semi-2", Name: "Semifinal 2", Venue: "Court B", StageLevel: 1, AdvancingCount: 1, TargetTempID: &final},
			{TempID: "final", Name: "Grand Final", Venue: "Main Arena", StageLevel: 2, AdvancingCount: 0, TargetTempID: nil},
		}
		if err := uc.ValidateBracketLayout(matches); err != nil {
			t.Fatalf("expected valid layout, got %v", err)
		}
	})
}

func TestDefineBracket(t *testing.T) {
	ownerRole := models.TournamentStaffRoleOwner
	adminRole := models.TournamentStaffRoleAdmin
	tournamentID := uuid.New()
	ownerID := uuid.New()
	adminID := uuid.New()

	baseTournament := &models.Tournament{
		ID:     tournamentID,
		Status: models.TournamentStatusRegistrationClosed,
	}

	final := "final"
	validMatches := []usecases.DefineBracketMatchInput{
		{TempID: "semi-1", Name: "Semifinal 1", Venue: "Court A", StageLevel: 1, AdvancingCount: 1, TargetTempID: &final},
		{TempID: "semi-2", Name: "Semifinal 2", Venue: "Court B", StageLevel: 1, AdvancingCount: 1, TargetTempID: &final},
		{TempID: "final", Name: "Grand Final", Venue: "Main Arena", StageLevel: 2, AdvancingCount: 0, TargetTempID: nil},
	}

	t.Run("forbidden if not owner", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament: baseTournament,
			staffRole:  &adminRole,
		}
		uc := usecases.NewBracketUseCase(repo)

		_, err := uc.DefineBracket(context.Background(), adminID, tournamentID, validMatches)
		if err != usecases.ErrBracketCallerNotOwner {
			t.Fatalf("expected ErrBracketCallerNotOwner, got %v", err)
		}
	})

	t.Run("forbidden if tournament ongoing", func(t *testing.T) {
		ongoingTourn := &models.Tournament{
			ID:     tournamentID,
			Status: models.TournamentStatusOngoing,
		}
		repo := &mockBracketRepository{
			tournament: ongoingTourn,
			staffRole:  &ownerRole,
		}
		uc := usecases.NewBracketUseCase(repo)

		_, err := uc.DefineBracket(context.Background(), ownerID, tournamentID, validMatches)
		if err != usecases.ErrBracketTournamentImmutable {
			t.Fatalf("expected ErrBracketTournamentImmutable, got %v", err)
		}
	})

	t.Run("forbidden if matches already completed", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament:          baseTournament,
			staffRole:           &ownerRole,
			hasCompletedMatches: true,
		}
		uc := usecases.NewBracketUseCase(repo)

		_, err := uc.DefineBracket(context.Background(), ownerID, tournamentID, validMatches)
		if err != usecases.ErrBracketAlreadyCompleted {
			t.Fatalf("expected ErrBracketAlreadyCompleted, got %v", err)
		}
	})

	t.Run("success as owner", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament: baseTournament,
			staffRole:  &ownerRole,
		}
		uc := usecases.NewBracketUseCase(repo)

		result, err := uc.DefineBracket(context.Background(), ownerID, tournamentID, validMatches)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 3 {
			t.Fatalf("expected 3 saved matches, got %d", len(result))
		}

		// Verify target linkages
		var finalMatch *models.Match
		for i := range result {
			if result[i].Name == "Grand Final" {
				finalMatch = &result[i]
				break
			}
		}
		if finalMatch == nil {
			t.Fatal("expected to find Grand Final match")
		}

		for _, m := range result {
			if m.Name != "Grand Final" {
				if m.TargetMatchID == nil || *m.TargetMatchID != finalMatch.ID {
					t.Fatalf("expected match %s to target final match ID %s, got %v", m.Name, finalMatch.ID, m.TargetMatchID)
				}
			}
		}
	})
}

func TestSeedStage1(t *testing.T) {
	ownerRole := models.TournamentStaffRoleOwner
	tournamentID := uuid.New()
	ownerID := uuid.New()

	tourn := &models.Tournament{
		ID:     tournamentID,
		Status: models.TournamentStatusRegistrationClosed,
	}

	match1ID := uuid.New()
	match2ID := uuid.New()
	stage1Matches := []models.Match{
		{ID: match1ID, TournamentID: tournamentID, Name: "Match 1", StageLevel: 1},
		{ID: match2ID, TournamentID: tournamentID, Name: "Match 2", StageLevel: 1},
	}

	team1ID := uuid.New()
	team2ID := uuid.New()
	team3ID := uuid.New()
	team4ID := uuid.New()
	acceptedTeams := []models.TournamentTeam{
		{ID: team1ID, Name: "Team Alpha", Status: models.TournamentTeamStatusAccepted},
		{ID: team2ID, Name: "Team Beta", Status: models.TournamentTeamStatusAccepted},
		{ID: team3ID, Name: "Team Gamma", Status: models.TournamentTeamStatusAccepted},
		{ID: team4ID, Name: "Team Delta", Status: models.TournamentTeamStatusAccepted},
	}

	t.Run("auto-distributes teams round-robin when assignments omitted", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament:    tourn,
			staffRole:     &ownerRole,
			stage1Matches: stage1Matches,
			acceptedTeams: acceptedTeams,
		}
		uc := usecases.NewBracketUseCase(repo)

		_, err := uc.SeedStage1(context.Background(), ownerID, tournamentID, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repo.seededEntries) != 4 {
			t.Fatalf("expected 4 seeded entries, got %d", len(repo.seededEntries))
		}

		// Team 0 -> match 0 (Match 1)
		// Team 1 -> match 1 (Match 2)
		// Team 2 -> match 0 (Match 1)
		// Team 3 -> match 1 (Match 2)
		if repo.seededEntries[0].MatchID != match1ID || repo.seededEntries[0].TeamID != team1ID {
			t.Errorf("entry 0 unexpected: %+v", repo.seededEntries[0])
		}
		if repo.seededEntries[1].MatchID != match2ID || repo.seededEntries[1].TeamID != team2ID {
			t.Errorf("entry 1 unexpected: %+v", repo.seededEntries[1])
		}
		if repo.seededEntries[2].MatchID != match1ID || repo.seededEntries[2].TeamID != team3ID {
			t.Errorf("entry 2 unexpected: %+v", repo.seededEntries[2])
		}
		if repo.seededEntries[3].MatchID != match2ID || repo.seededEntries[3].TeamID != team4ID {
			t.Errorf("entry 3 unexpected: %+v", repo.seededEntries[3])
		}
	})

	t.Run("explicit assignment seeding", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament:    tourn,
			staffRole:     &ownerRole,
			stage1Matches: stage1Matches,
			acceptedTeams: acceptedTeams,
		}
		uc := usecases.NewBracketUseCase(repo)

		assignments := []usecases.Stage1SeedAssignment{
			{MatchID: match1ID, TeamIDs: []uuid.UUID{team1ID, team4ID}},
			{MatchID: match2ID, TeamIDs: []uuid.UUID{team2ID, team3ID}},
		}

		_, err := uc.SeedStage1(context.Background(), ownerID, tournamentID, assignments)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repo.seededEntries) != 4 {
			t.Fatalf("expected 4 seeded entries, got %d", len(repo.seededEntries))
		}
	})

	t.Run("rejects duplicate team in explicit assignments", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament:    tourn,
			staffRole:     &ownerRole,
			stage1Matches: stage1Matches,
			acceptedTeams: acceptedTeams,
		}
		uc := usecases.NewBracketUseCase(repo)

		assignments := []usecases.Stage1SeedAssignment{
			{MatchID: match1ID, TeamIDs: []uuid.UUID{team1ID, team2ID}},
			{MatchID: match2ID, TeamIDs: []uuid.UUID{team1ID, team3ID}},
		}

		_, err := uc.SeedStage1(context.Background(), ownerID, tournamentID, assignments)
		if err != usecases.ErrBracketDuplicateSeedTeam {
			t.Fatalf("expected ErrBracketDuplicateSeedTeam, got %v", err)
		}
	})

	t.Run("rejects non-accepted team", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament:    tourn,
			staffRole:     &ownerRole,
			stage1Matches: stage1Matches,
			acceptedTeams: acceptedTeams,
		}
		uc := usecases.NewBracketUseCase(repo)

		unknownTeam := uuid.New()
		assignments := []usecases.Stage1SeedAssignment{
			{MatchID: match1ID, TeamIDs: []uuid.UUID{unknownTeam}},
		}

		_, err := uc.SeedStage1(context.Background(), ownerID, tournamentID, assignments)
		if err != usecases.ErrBracketInvalidSeedTeam {
			t.Fatalf("expected ErrBracketInvalidSeedTeam, got %v", err)
		}
	})
}

func TestRecordMatchResults(t *testing.T) {
	refereeRole := models.TournamentStaffRoleReferee
	tournamentID := uuid.New()
	refereeID := uuid.New()

	tourn := &models.Tournament{
		ID:     tournamentID,
		Status: models.TournamentStatusOngoing,
	}

	targetMatchID := uuid.New()
	targetMatch := &models.Match{
		ID:           targetMatchID,
		TournamentID: tournamentID,
		Name:         "Final",
		StageLevel:   2,
		IsCompleted:  false,
	}

	matchID := uuid.New()
	team1ID := uuid.New()
	team2ID := uuid.New()
	match := &models.Match{
		ID:             matchID,
		TournamentID:   tournamentID,
		Name:           "Semifinal 1",
		StageLevel:     1,
		AdvancingCount: 1,
		TargetMatchID:  &targetMatchID,
		IsCompleted:    false,
		Entries: []models.MatchEntry{
			{MatchID: matchID, TeamID: team1ID},
			{MatchID: matchID, TeamID: team2ID},
		},
	}

	t.Run("records placements and advances winner to target match", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament: tourn,
			staffRole:  &refereeRole,
			matchesMap: map[uuid.UUID]*models.Match{
				matchID:       match,
				targetMatchID: targetMatch,
			},
		}
		uc := usecases.NewBracketUseCase(repo)

		placements := []usecases.TeamPlacementInput{
			{TeamID: team1ID, FinalRank: 1},
			{TeamID: team2ID, FinalRank: 2},
		}

		result, err := uc.RecordMatchResults(context.Background(), refereeID, tournamentID, matchID, placements)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.IsCompleted {
			t.Errorf("expected result to indicate match completed")
		}
		if len(result.AdvancingTeams) != 1 || result.AdvancingTeams[0] != team1ID {
			t.Errorf("expected advancing teams to be [%s], got %v", team1ID, result.AdvancingTeams)
		}
		if repo.lastRecordedAdvances[team1ID] != true {
			t.Errorf("expected team 1 to have advances=true")
		}
		if repo.lastRecordedAdvances[team2ID] != false {
			t.Errorf("expected team 2 to have advances=false")
		}
	})

	t.Run("cannot record results for already completed match", func(t *testing.T) {
		completedMatch := &models.Match{
			ID:           matchID,
			TournamentID: tournamentID,
			IsCompleted:  true,
			Entries: []models.MatchEntry{
				{MatchID: matchID, TeamID: team1ID},
			},
		}
		repo := &mockBracketRepository{
			tournament: tourn,
			staffRole:  &refereeRole,
			matchesMap: map[uuid.UUID]*models.Match{
				matchID: completedMatch,
			},
		}
		uc := usecases.NewBracketUseCase(repo)

		_, err := uc.RecordMatchResults(context.Background(), refereeID, tournamentID, matchID, []usecases.TeamPlacementInput{
			{TeamID: team1ID, FinalRank: 1},
		})
		if err != usecases.ErrBracketMatchAlreadyCompleted {
			t.Fatalf("expected ErrBracketMatchAlreadyCompleted, got %v", err)
		}
	})

	t.Run("cannot record results if target match is already completed", func(t *testing.T) {
		completedTarget := &models.Match{
			ID:           targetMatchID,
			TournamentID: tournamentID,
			IsCompleted:  true,
		}
		repo := &mockBracketRepository{
			tournament: tourn,
			staffRole:  &refereeRole,
			matchesMap: map[uuid.UUID]*models.Match{
				matchID:       match,
				targetMatchID: completedTarget,
			},
		}
		uc := usecases.NewBracketUseCase(repo)

		placements := []usecases.TeamPlacementInput{
			{TeamID: team1ID, FinalRank: 1},
			{TeamID: team2ID, FinalRank: 2},
		}

		_, err := uc.RecordMatchResults(context.Background(), refereeID, tournamentID, matchID, placements)
		if err != usecases.ErrBracketTargetMatchCompleted {
			t.Fatalf("expected ErrBracketTargetMatchCompleted, got %v", err)
		}
	})

	t.Run("final match completes without target match", func(t *testing.T) {
		finalMatchID := uuid.New()
		finalMatch := &models.Match{
			ID:             finalMatchID,
			TournamentID:   tournamentID,
			Name:           "Grand Final",
			StageLevel:     2,
			AdvancingCount: 0,
			TargetMatchID:  nil,
			IsCompleted:    false,
			Entries: []models.MatchEntry{
				{MatchID: finalMatchID, TeamID: team1ID},
				{MatchID: finalMatchID, TeamID: team2ID},
			},
		}
		repo := &mockBracketRepository{
			tournament: tourn,
			staffRole:  &refereeRole,
			matchesMap: map[uuid.UUID]*models.Match{
				finalMatchID: finalMatch,
			},
		}
		uc := usecases.NewBracketUseCase(repo)

		placements := []usecases.TeamPlacementInput{
			{TeamID: team1ID, FinalRank: 1},
			{TeamID: team2ID, FinalRank: 2},
		}

		result, err := uc.RecordMatchResults(context.Background(), refereeID, tournamentID, finalMatchID, placements)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.TargetMatchID != nil {
			t.Errorf("expected TargetMatchID to be nil for final")
		}
		if len(result.AdvancingTeams) != 0 {
			t.Errorf("expected 0 advancing teams for final, got %v", result.AdvancingTeams)
		}
	})
}
