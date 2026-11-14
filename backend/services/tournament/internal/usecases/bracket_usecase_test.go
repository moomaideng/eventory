package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
	lastRecordedRankUpdates map[string]int
	lastRecordedScores      map[string]*string
	lastRecordedAdvances    map[string]bool
	lastAdvancingTeamIDs    []string
	lastSwap                []string
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

func (m *mockBracketRepository) UpdateMatchSchedule(_ context.Context, _, matchID uuid.UUID, scheduledAt time.Time, venue string) error {
	match := m.matchesMap[matchID]
	match.ScheduledAt = &scheduledAt
	match.Venue = venue
	return nil
}

func (m *mockBracketRepository) SwapMatchEntries(_ context.Context, matchA uuid.UUID, teamA string, matchB uuid.UUID, teamB string) error {
	m.lastSwap = []string{matchA.String(), teamA, matchB.String(), teamB}
	return nil
}

func (m *mockBracketRepository) RecordMatchResults(
	_ context.Context,
	_, _ uuid.UUID,
	rankUpdates map[string]int,
	scoreUpdates map[string]*string,
	advanceFlags map[string]bool,
	_ *uuid.UUID,
	advancingTeamIDs []string,
) error {
	if m.recordResultsErr != nil {
		return m.recordResultsErr
	}
	m.lastRecordedRankUpdates = rankUpdates
	m.lastRecordedScores = scoreUpdates
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
			{TempID: "   ", Name: "Match 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1},
		})
		if err != usecases.ErrBracketEmpty {
			t.Fatalf("expected ErrBracketEmpty, got %v", err)
		}
	})

	t.Run("duplicate temp ID", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1},
			{TempID: "m1", Name: "Match 2", StageLevel: 1, TeamCount: 2, AdvancingCount: 1},
		})
		if err != usecases.ErrBracketDuplicateTempID {
			t.Fatalf("expected ErrBracketDuplicateTempID, got %v", err)
		}
	})

	t.Run("empty match name", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "   ", StageLevel: 1, TeamCount: 2, AdvancingCount: 1},
		})
		if err != usecases.ErrBracketEmptyMatchName {
			t.Fatalf("expected ErrBracketEmptyMatchName, got %v", err)
		}
	})

	t.Run("invalid stage level", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 0, TeamCount: 2, AdvancingCount: 1},
		})
		if err != usecases.ErrBracketInvalidStageLevel {
			t.Fatalf("expected ErrBracketInvalidStageLevel, got %v", err)
		}
	})

	t.Run("invalid advancing count", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, TeamCount: 2, AdvancingCount: -1},
		})
		if err != usecases.ErrBracketInvalidAdvancingCount {
			t.Fatalf("expected ErrBracketInvalidAdvancingCount, got %v", err)
		}
	})

	t.Run("advancing count must be >= 1 if target is specified", func(t *testing.T) {
		target := "final"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Semi 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 0, TargetTempID: &target},
			{TempID: "final", Name: "Final", StageLevel: 2, TeamCount: 2, AdvancingCount: 0},
		})
		if err != usecases.ErrBracketAdvancingCountRequired {
			t.Fatalf("expected ErrBracketAdvancingCountRequired, got %v", err)
		}
	})

	t.Run("self-targeting match", func(t *testing.T) {
		self := "m1"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &self},
		})
		if err != usecases.ErrBracketSelfTargetMatch {
			t.Fatalf("expected ErrBracketSelfTargetMatch, got %v", err)
		}
	})

	t.Run("target not found", func(t *testing.T) {
		target := "m99"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &target},
		})
		if err != usecases.ErrBracketTargetMatchNotFound {
			t.Fatalf("expected ErrBracketTargetMatchNotFound, got %v", err)
		}
	})

	t.Run("backward edge (lower stage level)", func(t *testing.T) {
		target := "m1"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 0},
			{TempID: "m2", Name: "Match 2", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetTempID: &target},
		})
		if err != usecases.ErrBracketTargetMustBeHigher {
			t.Fatalf("expected ErrBracketTargetMustBeHigher, got %v", err)
		}
	})

	t.Run("same stage level edge", func(t *testing.T) {
		target := "m2"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &target},
			{TempID: "m2", Name: "Match 2", StageLevel: 1, TeamCount: 2, AdvancingCount: 0},
		})
		if err != usecases.ErrBracketTargetMustBeHigher {
			t.Fatalf("expected ErrBracketTargetMustBeHigher, got %v", err)
		}
	})

	t.Run("no championship match (every node points to another)", func(t *testing.T) {
		m2 := "m2"
		m3 := "m3"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &m2},
			{TempID: "m2", Name: "Match 2", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetTempID: &m3},
		})
		// m3 is missing
		if err != usecases.ErrBracketTargetMatchNotFound {
			t.Fatalf("expected ErrBracketTargetMatchNotFound, got %v", err)
		}
	})

	t.Run("invalid team count less than 1", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, TeamCount: 0, AdvancingCount: 0},
		})
		if err != usecases.ErrBracketInvalidTeamCount {
			t.Fatalf("expected ErrBracketInvalidTeamCount, got %v", err)
		}
	})

	t.Run("advancing count cannot exceed team count", func(t *testing.T) {
		final := "final"
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "m1", Name: "Match 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 3, TargetTempID: &final},
			{TempID: "final", Name: "Final", StageLevel: 2, TeamCount: 3, AdvancingCount: 0},
		})
		if err != usecases.ErrBracketAdvancingExceedsTeams {
			t.Fatalf("expected ErrBracketAdvancingExceedsTeams, got %v", err)
		}
	})

	t.Run("valid single-team bye match advancing 1 team", func(t *testing.T) {
		final := "final"
		matches := []usecases.DefineBracketMatchInput{
			{TempID: "bye-1", Name: "Bye Match 1", StageLevel: 1, TeamCount: 1, AdvancingCount: 1, TargetTempID: &final},
			{TempID: "semi-2", Name: "Semifinal 2", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &final},
			{TempID: "final", Name: "Grand Final", StageLevel: 2, TeamCount: 2, AdvancingCount: 0},
		}
		if err := uc.ValidateBracketLayout(matches); err != nil {
			t.Fatalf("expected valid layout with bye match, got %v", err)
		}
	})

	t.Run("incoming advancing teams exceed target team count", func(t *testing.T) {
		final := "final"
		matches := []usecases.DefineBracketMatchInput{
			{TempID: "heat-1", Name: "Heat 1", StageLevel: 1, TeamCount: 4, AdvancingCount: 2, TargetTempID: &final},
			{TempID: "heat-2", Name: "Heat 2", StageLevel: 1, TeamCount: 4, AdvancingCount: 2, TargetTempID: &final},
			{TempID: "final", Name: "Final", StageLevel: 2, TeamCount: 2, AdvancingCount: 0}, // Target team count is 2, but 2+2=4 incoming!
		}
		err := uc.ValidateBracketLayout(matches)
		if err != usecases.ErrBracketIncomingMismatchTarget {
			t.Fatalf("expected ErrBracketIncomingMismatchTarget, got %v", err)
		}
	})

	t.Run("incoming advancing teams less than target team count", func(t *testing.T) {
		final := "final"
		matches := []usecases.DefineBracketMatchInput{
			{TempID: "semi-1", Name: "Semifinal 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &final},
			{TempID: "final", Name: "Final", StageLevel: 2, TeamCount: 2, AdvancingCount: 0}, // Target team count is 2, but only 1 incoming!
		}
		err := uc.ValidateBracketLayout(matches)
		if err != usecases.ErrBracketIncomingMismatchTarget {
			t.Fatalf("expected ErrBracketIncomingMismatchTarget, got %v", err)
		}
	})

	t.Run("no stage 1 matches", func(t *testing.T) {
		err := uc.ValidateBracketLayout([]usecases.DefineBracketMatchInput{
			{TempID: "final", Name: "Final", StageLevel: 2, TeamCount: 2, AdvancingCount: 0},
		})
		if err != usecases.ErrBracketNoStage1Matches {
			t.Fatalf("expected ErrBracketNoStage1Matches, got %v", err)
		}
	})

	t.Run("stage ordering catches stage 2 mismatch before stage 3 mismatch", func(t *testing.T) {
		semi := "semi"
		final := "final"
		// Semi (Stage 2) has incoming 1 != teamCount 2.
		// Final (Stage 3) has incoming 0 != teamCount 2.
		// Stage 2 mismatch should be returned first.
		matches := []usecases.DefineBracketMatchInput{
			{TempID: "qf-1", Name: "QF 1", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &semi},
			{TempID: "semi", Name: "Semi", StageLevel: 2, TeamCount: 2, AdvancingCount: 1, TargetTempID: &final},
			{TempID: "final", Name: "Final", StageLevel: 3, TeamCount: 2, AdvancingCount: 0},
		}
		err := uc.ValidateBracketLayout(matches)
		if err != usecases.ErrBracketIncomingMismatchTarget {
			t.Fatalf("expected ErrBracketIncomingMismatchTarget, got %v", err)
		}
	})

	t.Run("valid 4-team single elimination bracket", func(t *testing.T) {
		final := "final"
		matches := []usecases.DefineBracketMatchInput{
			{TempID: "semi-1", Name: "Semifinal 1", Venue: "Court A", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &final},
			{TempID: "semi-2", Name: "Semifinal 2", Venue: "Court B", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &final},
			{TempID: "final", Name: "Grand Final", Venue: "Main Arena", StageLevel: 2, TeamCount: 2, AdvancingCount: 0, TargetTempID: nil},
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
		ID:                   tournamentID,
		Status:               models.TournamentStatusRegistration,
		RegistrationDeadline: time.Now().Add(-time.Hour),
	}

	final := "final"
	validMatches := []usecases.DefineBracketMatchInput{
		{TempID: "semi-1", Name: "Semifinal 1", Venue: "Court A", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &final},
		{TempID: "semi-2", Name: "Semifinal 2", Venue: "Court B", StageLevel: 1, TeamCount: 2, AdvancingCount: 1, TargetTempID: &final},
		{TempID: "final", Name: "Grand Final", Venue: "Main Arena", StageLevel: 2, TeamCount: 2, AdvancingCount: 0, TargetTempID: nil},
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

	t.Run("allowed while registration is still open", func(t *testing.T) {
		openTourn := &models.Tournament{
			ID:                   tournamentID,
			Status:               models.TournamentStatusRegistration,
			RegistrationDeadline: time.Now().Add(time.Hour),
		}
		repo := &mockBracketRepository{
			tournament: openTourn,
			staffRole:  &ownerRole,
		}
		uc := usecases.NewBracketUseCase(repo)

		result, err := uc.DefineBracket(context.Background(), ownerID, tournamentID, validMatches)
		if err != nil || len(result) != 3 {
			t.Fatalf("expected the bracket to be saved before the deadline, got %d matches, err %v", len(result), err)
		}
	})

	t.Run("forbidden if tournament draft", func(t *testing.T) {
		draftTourn := &models.Tournament{
			ID:                   tournamentID,
			Status:               models.TournamentStatusDraft,
			RegistrationDeadline: time.Now().Add(-time.Hour),
		}
		repo := &mockBracketRepository{
			tournament: draftTourn,
			staffRole:  &ownerRole,
		}
		uc := usecases.NewBracketUseCase(repo)

		_, err := uc.DefineBracket(context.Background(), ownerID, tournamentID, validMatches)
		if !errors.Is(err, usecases.ErrBracketTournamentImmutable) {
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

	t.Run("success resetting/clearing bracket with empty matches", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament: baseTournament,
			staffRole:  &ownerRole,
		}
		uc := usecases.NewBracketUseCase(repo)

		result, err := uc.DefineBracket(context.Background(), ownerID, tournamentID, []usecases.DefineBracketMatchInput{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("expected 0 saved matches, got %d", len(result))
		}
	})

	t.Run("fails when stage 1 capacity does not match accepted teams count", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament: baseTournament,
			staffRole:  &ownerRole,
			acceptedTeams: []models.TournamentTeam{
				{ID: "TM000001", Name: "T1", Status: models.TournamentTeamStatusAccepted},
				{ID: "TM000002", Name: "T2", Status: models.TournamentTeamStatusAccepted},
				{ID: "TM000003", Name: "T3", Status: models.TournamentTeamStatusAccepted},
			},
		}
		uc := usecases.NewBracketUseCase(repo)

		_, err := uc.DefineBracket(context.Background(), ownerID, tournamentID, validMatches)
		if !errors.Is(err, usecases.ErrBracketStage1CapacityMismatch) {
			t.Fatalf("expected ErrBracketStage1CapacityMismatch, got %v", err)
		}
	})
}

func TestSeedStage1(t *testing.T) {
	ownerRole := models.TournamentStaffRoleOwner
	tournamentID := uuid.New()
	ownerID := uuid.New()

	tourn := &models.Tournament{
		ID:                   tournamentID,
		Status:               models.TournamentStatusRegistration,
		RegistrationDeadline: time.Now().Add(-time.Hour),
	}

	match1ID := uuid.New()
	match2ID := uuid.New()
	stage1Matches := []models.Match{
		{ID: match1ID, TournamentID: tournamentID, Name: "Match 1", StageLevel: 1, TeamCount: 2},
		{ID: match2ID, TournamentID: tournamentID, Name: "Match 2", StageLevel: 1, TeamCount: 2},
	}

	team1ID := "TM000001"
	team2ID := "TM000002"
	team3ID := "TM000003"
	team4ID := "TM000004"
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
			{MatchID: match1ID, TeamIDs: []string{team1ID, team4ID}},
			{MatchID: match2ID, TeamIDs: []string{team2ID, team3ID}},
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
			{MatchID: match1ID, TeamIDs: []string{team1ID, team2ID}},
			{MatchID: match2ID, TeamIDs: []string{team1ID, team3ID}},
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

		unknownTeam := "TM999999"
		assignments := []usecases.Stage1SeedAssignment{
			{MatchID: match1ID, TeamIDs: []string{unknownTeam}},
		}

		_, err := uc.SeedStage1(context.Background(), ownerID, tournamentID, assignments)
		if err != usecases.ErrBracketInvalidSeedTeam {
			t.Fatalf("expected ErrBracketInvalidSeedTeam, got %v", err)
		}
	})

	t.Run("rejects assignment exceeding match team count", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament:    tourn,
			staffRole:     &ownerRole,
			stage1Matches: stage1Matches,
			acceptedTeams: acceptedTeams,
		}
		uc := usecases.NewBracketUseCase(repo)

		assignments := []usecases.Stage1SeedAssignment{
			{MatchID: match1ID, TeamIDs: []string{team1ID, team2ID, team3ID}},
		}

		_, err := uc.SeedStage1(context.Background(), ownerID, tournamentID, assignments)
		if err != usecases.ErrBracketSeedExceedsTeamCount {
			t.Fatalf("expected ErrBracketSeedExceedsTeamCount, got %v", err)
		}
	})

	t.Run("fails when stage 1 capacity does not match accepted teams count", func(t *testing.T) {
		repo := &mockBracketRepository{
			tournament:    tourn,
			staffRole:     &ownerRole,
			stage1Matches: stage1Matches,
			acceptedTeams: []models.TournamentTeam{
				{ID: team1ID, Name: "Team Alpha", Status: models.TournamentTeamStatusAccepted},
				{ID: team2ID, Name: "Team Beta", Status: models.TournamentTeamStatusAccepted},
			},
		}
		uc := usecases.NewBracketUseCase(repo)

		_, err := uc.SeedStage1(context.Background(), ownerID, tournamentID, nil)
		if !errors.Is(err, usecases.ErrBracketStage1CapacityMismatch) {
			t.Fatalf("expected ErrBracketStage1CapacityMismatch, got %v", err)
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
	team1ID := "TM000001"
	team2ID := "TM000002"
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

		score1 := "2"
		score2 := "1"
		placements := []usecases.TeamPlacementInput{
			{TeamID: team1ID, FinalRank: 1, Score: &score1},
			{TeamID: team2ID, FinalRank: 2, Score: &score2},
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
		if repo.lastRecordedScores[team1ID] == nil || *repo.lastRecordedScores[team1ID] != "2" {
			t.Errorf("expected team 1 score to be '2', got %v", repo.lastRecordedScores[team1ID])
		}
		if repo.lastRecordedScores[team2ID] == nil || *repo.lastRecordedScores[team2ID] != "1" {
			t.Errorf("expected team 2 score to be '1', got %v", repo.lastRecordedScores[team2ID])
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

func TestScheduleMatch(t *testing.T) {
	referee := models.TournamentStaffRoleReferee
	tournamentID, matchID := uuid.New(), uuid.New()
	future := time.Now().Add(48 * time.Hour)

	tests := []struct {
		name    string
		role    *models.TournamentStaffRole
		at      time.Time
		wantErr error
	}{
		{"staff sets future time", &referee, future, nil},
		{"past time rejected", &referee, time.Now().Add(-24 * time.Hour), usecases.ErrBracketScheduleInPast},
		{"non-staff rejected", nil, future, usecases.ErrBracketCallerNotStaff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBracketRepository{
				tournament: &models.Tournament{ID: tournamentID, Status: models.TournamentStatusOngoing},
				staffRole:  tt.role,
				matchesMap: map[uuid.UUID]*models.Match{
					matchID: {ID: matchID, TournamentID: tournamentID},
				},
			}
			if tt.role == nil {
				repo.staffErr = repositories.ErrBracketStaffNotFound
			}

			match, err := usecases.NewBracketUseCase(repo).ScheduleMatch(context.Background(), uuid.New(), tournamentID, matchID, tt.at, " Field 2 ")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && (match.Venue != "Field 2" || !match.ScheduledAt.Equal(tt.at)) {
				t.Errorf("schedule not saved: venue=%q at=%v", match.Venue, match.ScheduledAt)
			}
		})
	}
}

func TestSwapTeams(t *testing.T) {
	referee := models.TournamentStaffRoleReferee
	tournamentID := uuid.New()
	m1, m2, m3 := uuid.New(), uuid.New(), uuid.New()
	t1, t2, t3 := "TM000001", "TM000002", "TM000003"
	newRepo := func(role *models.TournamentStaffRole, m2Completed bool) *mockBracketRepository {
		repo := &mockBracketRepository{
			tournament: &models.Tournament{ID: tournamentID, Status: models.TournamentStatusOngoing},
			staffRole:  role,
			matchesMap: map[uuid.UUID]*models.Match{
				m1: {ID: m1, StageLevel: 1, Entries: []models.MatchEntry{{TeamID: t1}}},
				m2: {ID: m2, StageLevel: 1, IsCompleted: m2Completed, Entries: []models.MatchEntry{{TeamID: t2}}},
				m3: {ID: m3, StageLevel: 2, Entries: []models.MatchEntry{{TeamID: t3}}},
			},
		}
		if role == nil {
			repo.staffErr = repositories.ErrBracketStaffNotFound
		}
		return repo
	}

	tests := []struct {
		name    string
		repo    *mockBracketRepository
		second  usecases.SwapSlot
		wantErr error
	}{
		{"swaps teams in same stage", newRepo(&referee, false), usecases.SwapSlot{MatchID: m2, TeamID: t2}, nil},
		{"completed match rejected", newRepo(&referee, true), usecases.SwapSlot{MatchID: m2, TeamID: t2}, usecases.ErrBracketMatchAlreadyCompleted},
		{"different stage rejected", newRepo(&referee, false), usecases.SwapSlot{MatchID: m3, TeamID: t3}, usecases.ErrBracketSwapStageMismatch},
		{"non-staff rejected", newRepo(nil, false), usecases.SwapSlot{MatchID: m2, TeamID: t2}, usecases.ErrBracketCallerNotStaff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := usecases.NewBracketUseCase(tt.repo).SwapTeams(context.Background(), uuid.New(), tournamentID,
				usecases.SwapSlot{MatchID: m1, TeamID: t1}, tt.second)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want error %v, got %v", tt.wantErr, err)
			}
			if (tt.wantErr == nil) != (tt.repo.lastSwap != nil) {
				t.Errorf("swap saved = %v, want %v", tt.repo.lastSwap != nil, tt.wantErr == nil)
			}
		})
	}
}
