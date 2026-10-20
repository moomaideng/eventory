package usecases

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
)

var (
	ErrBracketTournamentNotFound     = errors.New("tournament not found")
	ErrBracketCallerNotOwner         = errors.New("caller is not authorized: must be tournament owner")
	ErrBracketCallerNotStaff         = errors.New("caller is not authorized: must be tournament owner, admin, or referee")
	ErrBracketTournamentImmutable    = errors.New("bracket operations cannot be performed while tournament is ongoing, completed, or cancelled")
	ErrBracketEmpty                  = errors.New("bracket must contain at least one match")
	ErrBracketDuplicateTempID        = errors.New("duplicate temporary match identifier")
	ErrBracketEmptyMatchName         = errors.New("match name cannot be empty")
	ErrBracketInvalidStageLevel      = errors.New("stage level must be greater than or equal to 1")
	ErrBracketInvalidAdvancingCount  = errors.New("advancing count cannot be negative")
	ErrBracketAdvancingCountRequired = errors.New("matches with a target destination must advance at least 1 team")
	ErrBracketSelfTargetMatch        = errors.New("match cannot target itself")
	ErrBracketTargetMatchNotFound    = errors.New("target match not found in bracket definition")
	ErrBracketTargetMustBeHigher     = errors.New("bracket edge must point forward to a higher stage level")
	ErrBracketCycleDetected          = errors.New("bracket graph must be acyclic (no loops)")
	ErrBracketNoChampionshipMatch    = errors.New("bracket graph must contain at least one championship/final match with no target")
	ErrBracketAlreadyCompleted       = errors.New("bracket cannot be modified after matches have been completed")
	ErrBracketNoStage1Matches        = errors.New("no stage 1 matches found in bracket")
	ErrBracketNoAcceptedTeams        = errors.New("no accepted teams available to seed")
	ErrBracketInvalidSeedMatch       = errors.New("seeded match does not belong to stage 1")
	ErrBracketInvalidSeedTeam        = errors.New("seeded team is not an accepted team in this tournament")
	ErrBracketDuplicateSeedTeam      = errors.New("team cannot be seeded into multiple matches")
	ErrBracketMatchNotFound          = errors.New("match not found")
	ErrBracketMatchAlreadyCompleted  = errors.New("match is already completed")
	ErrBracketTargetMatchCompleted   = errors.New("target match is already completed")
	ErrBracketNoParticipants         = errors.New("match has no participants to record results for")
	ErrBracketMissingTeamPlacement   = errors.New("all match participants must have a placement")
	ErrBracketUnknownTeamPlacement   = errors.New("team in placement is not a participant of this match")
	ErrBracketInvalidPlacementRank   = errors.New("placement rank must be greater than or equal to 1")
	ErrBracketDuplicatePlacementTeam = errors.New("duplicate team in placement list")
	ErrBracketInvalidTeamCount       = errors.New("match team count must be at least 1")
	ErrBracketAdvancingExceedsTeams  = errors.New("advancing count cannot exceed team count")
	ErrBracketIncomingMismatchTarget = errors.New("total incoming advancing teams must match target match team count")
	ErrBracketIncomingExceedsTarget  = ErrBracketIncomingMismatchTarget
	ErrBracketSeedExceedsTeamCount   = errors.New("seeded teams exceed match team count")
	ErrBracketStage1CapacityMismatch = errors.New("total Stage 1 match slots must equal the number of accepted teams")
)

type DefineBracketMatchInput struct {
	TempID         string
	Name           string
	Venue          string
	StageLevel     int
	TeamCount      int
	AdvancingCount int
	TargetTempID   *string
}

type Stage1SeedAssignment struct {
	MatchID uuid.UUID
	TeamIDs []uuid.UUID
}

type TeamPlacementInput struct {
	TeamID    uuid.UUID
	FinalRank int
	Score     *string
}

type RecordMatchResultsResult struct {
	MatchID        uuid.UUID
	IsCompleted    bool
	AdvancingTeams []uuid.UUID
	TargetMatchID  *uuid.UUID
}

type BracketUseCase struct {
	repo repositories.BracketRepository
}

func NewBracketUseCase(repo repositories.BracketRepository) *BracketUseCase {
	return &BracketUseCase{repo: repo}
}

func (u *BracketUseCase) checkOwnerAuth(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.Tournament, error) {
	tournament, err := u.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		if errors.Is(err, repositories.ErrBracketTournamentNotFound) {
			return nil, ErrBracketTournamentNotFound
		}
		return nil, err
	}

	role, err := u.repo.GetStaffRole(ctx, tournamentID, accountID)
	if err != nil {
		if errors.Is(err, repositories.ErrBracketStaffNotFound) {
			return nil, ErrBracketCallerNotOwner
		}
		return nil, err
	}

	if *role != models.TournamentStaffRoleOwner {
		return nil, ErrBracketCallerNotOwner
	}

	return tournament, nil
}

func (u *BracketUseCase) checkStaffResultAuth(ctx context.Context, tournamentID, accountID uuid.UUID) (*models.Tournament, error) {
	tournament, err := u.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		if errors.Is(err, repositories.ErrBracketTournamentNotFound) {
			return nil, ErrBracketTournamentNotFound
		}
		return nil, err
	}

	role, err := u.repo.GetStaffRole(ctx, tournamentID, accountID)
	if err != nil {
		if errors.Is(err, repositories.ErrBracketStaffNotFound) {
			return nil, ErrBracketCallerNotStaff
		}
		return nil, err
	}

	if *role != models.TournamentStaffRoleOwner &&
		*role != models.TournamentStaffRoleAdmin &&
		*role != models.TournamentStaffRoleReferee {
		return nil, ErrBracketCallerNotStaff
	}

	return tournament, nil
}

func (u *BracketUseCase) ValidateBracketLayout(matches []DefineBracketMatchInput) error {
	if len(matches) == 0 {
		return ErrBracketEmpty
	}

	hasStage1 := false
	tempIDMap := make(map[string]DefineBracketMatchInput, len(matches))
	for _, m := range matches {
		tempID := strings.TrimSpace(m.TempID)
		if tempID == "" {
			return ErrBracketEmpty
		}
		if _, exists := tempIDMap[tempID]; exists {
			return ErrBracketDuplicateTempID
		}
		if strings.TrimSpace(m.Name) == "" {
			return ErrBracketEmptyMatchName
		}
		if m.StageLevel < 1 {
			return ErrBracketInvalidStageLevel
		}
		if m.StageLevel == 1 {
			hasStage1 = true
		}
		if m.TeamCount < 1 {
			return ErrBracketInvalidTeamCount
		}
		if m.AdvancingCount < 0 {
			return ErrBracketInvalidAdvancingCount
		}
		if m.AdvancingCount > m.TeamCount {
			return ErrBracketAdvancingExceedsTeams
		}
		if m.TargetTempID != nil && strings.TrimSpace(*m.TargetTempID) != "" {
			if m.AdvancingCount < 1 {
				return ErrBracketAdvancingCountRequired
			}
		}
		tempIDMap[tempID] = m
	}

	if !hasStage1 {
		return ErrBracketNoStage1Matches
	}

	hasChampionship := false
	inDegree := make(map[string]int, len(matches))
	adj := make(map[string][]string, len(matches))
	incomingAdvancing := make(map[string]int, len(matches))
	for tempID := range tempIDMap {
		inDegree[tempID] = 0
	}

	for _, m := range matches {
		tempID := strings.TrimSpace(m.TempID)
		if m.TargetTempID == nil || strings.TrimSpace(*m.TargetTempID) == "" {
			hasChampionship = true
			continue
		}

		targetID := strings.TrimSpace(*m.TargetTempID)
		if targetID == tempID {
			return ErrBracketSelfTargetMatch
		}

		targetMatch, exists := tempIDMap[targetID]
		if !exists {
			return ErrBracketTargetMatchNotFound
		}

		if targetMatch.StageLevel <= m.StageLevel {
			return ErrBracketTargetMustBeHigher
		}

		adj[tempID] = append(adj[tempID], targetID)
		inDegree[targetID]++
		incomingAdvancing[targetID] += m.AdvancingCount
	}

	if !hasChampionship {
		return ErrBracketNoChampionshipMatch
	}

	// Validate stages in ascending order (Stage 1 -> Stage 2 -> Stage 3...)
	sortedMatches := make([]DefineBracketMatchInput, 0, len(tempIDMap))
	for _, m := range tempIDMap {
		sortedMatches = append(sortedMatches, m)
	}
	sort.Slice(sortedMatches, func(i, j int) bool {
		if sortedMatches[i].StageLevel != sortedMatches[j].StageLevel {
			return sortedMatches[i].StageLevel < sortedMatches[j].StageLevel
		}
		return sortedMatches[i].TempID < sortedMatches[j].TempID
	})

	for _, m := range sortedMatches {
		tempID := strings.TrimSpace(m.TempID)
		if m.StageLevel > 1 {
			totalIncoming := incomingAdvancing[tempID]
			if totalIncoming != m.TeamCount {
				return ErrBracketIncomingMismatchTarget
			}
		}
	}

	// Topological sort for cycle detection (Kahn's algorithm)
	queue := make([]string, 0, len(matches))
	for tempID, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, tempID)
		}
	}

	visitedCount := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visitedCount++

		for _, neighbor := range adj[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if visitedCount != len(matches) {
		return ErrBracketCycleDetected
	}

	return nil
}

func (u *BracketUseCase) DefineBracket(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
	matches []DefineBracketMatchInput,
) ([]models.Match, error) {
	tournament, err := u.checkOwnerAuth(ctx, tournamentID, accountID)
	if err != nil {
		return nil, err
	}

	if tournament.Status == models.TournamentStatusOngoing ||
		tournament.Status == models.TournamentStatusCompleted ||
		tournament.Status == models.TournamentStatusCancelled {
		return nil, ErrBracketTournamentImmutable
	}

	hasCompleted, err := u.repo.HasCompletedMatches(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if hasCompleted {
		return nil, ErrBracketAlreadyCompleted
	}

	if len(matches) == 0 {
		if err := u.repo.SaveBracketLayout(ctx, tournamentID, []models.Match{}); err != nil {
			return nil, err
		}
		return []models.Match{}, nil
	}

	if err := u.ValidateBracketLayout(matches); err != nil {
		return nil, err
	}

	// Validate Stage 1 capacity against accepted teams count if any accepted teams exist
	acceptedTeams, err := u.repo.GetAcceptedTeams(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if len(acceptedTeams) > 0 {
		stage1TotalCap := 0
		for _, m := range matches {
			if m.StageLevel == 1 {
				stage1TotalCap += m.TeamCount
			}
		}
		if stage1TotalCap != len(acceptedTeams) {
			return nil, ErrBracketStage1CapacityMismatch
		}
	}

	// Assign real UUIDs to each match
	tempToUUID := make(map[string]uuid.UUID, len(matches))
	for _, m := range matches {
		tempToUUID[strings.TrimSpace(m.TempID)] = uuid.New()
	}

	modelsMatches := make([]models.Match, 0, len(matches))
	for _, m := range matches {
		tempID := strings.TrimSpace(m.TempID)
		id := tempToUUID[tempID]
		var targetUUID *uuid.UUID
		if m.TargetTempID != nil && strings.TrimSpace(*m.TargetTempID) != "" {
			tUUID := tempToUUID[strings.TrimSpace(*m.TargetTempID)]
			targetUUID = &tUUID
		}

		modelsMatches = append(modelsMatches, models.Match{
			ID:             id,
			TournamentID:   tournamentID,
			Name:           strings.TrimSpace(m.Name),
			Venue:          strings.TrimSpace(m.Venue),
			StageLevel:     m.StageLevel,
			TeamCount:      m.TeamCount,
			AdvancingCount: m.AdvancingCount,
			TargetMatchID:  targetUUID,
			IsCompleted:    false,
		})
	}

	if err := u.repo.SaveBracketLayout(ctx, tournamentID, modelsMatches); err != nil {
		if errors.Is(err, repositories.ErrBracketMatchCompleted) {
			return nil, ErrBracketAlreadyCompleted
		}
		return nil, err
	}

	return u.repo.GetBracket(ctx, tournamentID)
}

func (u *BracketUseCase) SeedStage1(
	ctx context.Context,
	accountID, tournamentID uuid.UUID,
	assignments []Stage1SeedAssignment,
) ([]models.Match, error) {
	tournament, err := u.checkOwnerAuth(ctx, tournamentID, accountID)
	if err != nil {
		return nil, err
	}

	if tournament.Status == models.TournamentStatusOngoing ||
		tournament.Status == models.TournamentStatusCompleted ||
		tournament.Status == models.TournamentStatusCancelled {
		return nil, ErrBracketTournamentImmutable
	}

	stage1Matches, err := u.repo.GetStage1Matches(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if len(stage1Matches) == 0 {
		return nil, ErrBracketNoStage1Matches
	}

	for _, m := range stage1Matches {
		if m.IsCompleted {
			return nil, ErrBracketAlreadyCompleted
		}
	}

	acceptedTeams, err := u.repo.GetAcceptedTeams(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if len(acceptedTeams) == 0 {
		return nil, ErrBracketNoAcceptedTeams
	}

	stage1MatchMap := make(map[uuid.UUID]bool, len(stage1Matches))
	stage1MatchCap := make(map[uuid.UUID]int, len(stage1Matches))
	stage1IDs := make([]uuid.UUID, 0, len(stage1Matches))
	stage1TotalCap := 0
	for _, m := range stage1Matches {
		stage1MatchMap[m.ID] = true
		stage1MatchCap[m.ID] = m.TeamCount
		stage1IDs = append(stage1IDs, m.ID)
		stage1TotalCap += m.TeamCount
	}

	if stage1TotalCap != len(acceptedTeams) {
		return nil, ErrBracketStage1CapacityMismatch
	}

	acceptedTeamMap := make(map[uuid.UUID]bool, len(acceptedTeams))
	for _, t := range acceptedTeams {
		acceptedTeamMap[t.ID] = true
	}

	entries := make([]models.MatchEntry, 0)
	if len(assignments) > 0 {
		assignedTeams := make(map[uuid.UUID]bool)
		for _, a := range assignments {
			if !stage1MatchMap[a.MatchID] {
				return nil, ErrBracketInvalidSeedMatch
			}
			if len(a.TeamIDs) > stage1MatchCap[a.MatchID] {
				return nil, ErrBracketSeedExceedsTeamCount
			}
			for _, teamID := range a.TeamIDs {
				if !acceptedTeamMap[teamID] {
					return nil, ErrBracketInvalidSeedTeam
				}
				if assignedTeams[teamID] {
					return nil, ErrBracketDuplicateSeedTeam
				}
				assignedTeams[teamID] = true
				entries = append(entries, models.MatchEntry{
					MatchID:  a.MatchID,
					TeamID:   teamID,
					Advances: false,
				})
			}
		}
	} else {
		// Auto-distribute accepted teams round-robin across Stage 1 matches
		for i, team := range acceptedTeams {
			targetMatch := stage1Matches[i%len(stage1Matches)]
			entries = append(entries, models.MatchEntry{
				MatchID:  targetMatch.ID,
				TeamID:   team.ID,
				Advances: false,
			})
		}
	}

	if err := u.repo.SeedStage1(ctx, tournamentID, stage1IDs, entries); err != nil {
		return nil, err
	}

	return u.repo.GetBracket(ctx, tournamentID)
}

func (u *BracketUseCase) RecordMatchResults(
	ctx context.Context,
	accountID, tournamentID, matchID uuid.UUID,
	placements []TeamPlacementInput,
) (*RecordMatchResultsResult, error) {
	tournament, err := u.checkStaffResultAuth(ctx, tournamentID, accountID)
	if err != nil {
		return nil, err
	}

	if tournament.Status == models.TournamentStatusCompleted ||
		tournament.Status == models.TournamentStatusCancelled {
		return nil, ErrBracketTournamentImmutable
	}

	match, err := u.repo.GetMatch(ctx, tournamentID, matchID)
	if err != nil {
		if errors.Is(err, repositories.ErrBracketMatchNotFound) {
			return nil, ErrBracketMatchNotFound
		}
		return nil, err
	}

	if match.IsCompleted {
		return nil, ErrBracketMatchAlreadyCompleted
	}

	if len(match.Entries) == 0 {
		return nil, ErrBracketNoParticipants
	}

	participantSet := make(map[uuid.UUID]bool, len(match.Entries))
	for _, entry := range match.Entries {
		participantSet[entry.TeamID] = true
	}

	if len(placements) != len(participantSet) {
		return nil, ErrBracketMissingTeamPlacement
	}

	rankUpdates := make(map[uuid.UUID]int, len(placements))
	scoreUpdates := make(map[uuid.UUID]*string, len(placements))
	for _, p := range placements {
		if !participantSet[p.TeamID] {
			return nil, ErrBracketUnknownTeamPlacement
		}
		if _, seen := rankUpdates[p.TeamID]; seen {
			return nil, ErrBracketDuplicatePlacementTeam
		}
		if p.FinalRank < 1 {
			return nil, ErrBracketInvalidPlacementRank
		}
		rankUpdates[p.TeamID] = p.FinalRank
		scoreUpdates[p.TeamID] = p.Score
	}

	advanceFlags := make(map[uuid.UUID]bool, len(match.Entries))
	advancingTeamIDs := make([]uuid.UUID, 0)
	for teamID, rank := range rankUpdates {
		advances := rank <= match.AdvancingCount
		advanceFlags[teamID] = advances
		if advances && match.TargetMatchID != nil {
			advancingTeamIDs = append(advancingTeamIDs, teamID)
		}
	}

	if match.TargetMatchID != nil {
		targetMatch, err := u.repo.GetMatch(ctx, tournamentID, *match.TargetMatchID)
		if err != nil {
			if errors.Is(err, repositories.ErrBracketMatchNotFound) {
				return nil, ErrBracketTargetMatchNotFound
			}
			return nil, err
		}
		if targetMatch.IsCompleted {
			return nil, ErrBracketTargetMatchCompleted
		}
	}

	if err := u.repo.RecordMatchResults(
		ctx,
		tournamentID, matchID,
		rankUpdates, scoreUpdates, advanceFlags,
		match.TargetMatchID, advancingTeamIDs,
	); err != nil {
		return nil, err
	}

	return &RecordMatchResultsResult{
		MatchID:        matchID,
		IsCompleted:    true,
		AdvancingTeams: advancingTeamIDs,
		TargetMatchID:  match.TargetMatchID,
	}, nil
}

func (u *BracketUseCase) GetBracket(ctx context.Context, tournamentID uuid.UUID) ([]models.Match, error) {
	_, err := u.repo.GetTournament(ctx, tournamentID)
	if err != nil {
		if errors.Is(err, repositories.ErrBracketTournamentNotFound) {
			return nil, ErrBracketTournamentNotFound
		}
		return nil, err
	}

	return u.repo.GetBracket(ctx, tournamentID)
}
