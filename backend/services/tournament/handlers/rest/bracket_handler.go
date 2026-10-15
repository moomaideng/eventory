package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type DefineBracketMatchItem struct {
	TempID         string  `json:"tempId" minLength:"1" maxLength:"100" doc:"Temporary identifier for this node (e.g. 'heat-1.1', 'semi-a')"`
	Name           string  `json:"name" minLength:"1" maxLength:"160" doc:"Display label (e.g. 'Heat 1.1', 'Quarterfinal B')"`
	Venue          string  `json:"venue" maxLength:"160" doc:"Location/court/station details (e.g. 'Field 2', 'Arena Room A')"`
	StageLevel     int     `json:"stageLevel" minimum:"1" doc:"Integer representing progression depth (1, 2, 3...)"`
	AdvancingCount int     `json:"advancingCount" minimum:"0" doc:"Integer N specifying how many top teams advance to the target match"`
	TargetMatchID  *string `json:"targetMatchId,omitempty" doc:"Temporary identifier pointing to the target match node (null for final/championship match)"`
	TargetTempID   *string `json:"targetTempId,omitempty" doc:"Alias for targetMatchId"`
}

type DefineBracketRequest struct {
	Matches []DefineBracketMatchItem `json:"matches" minItems:"1" doc:"List of matches defining the bracket DAG"`
}

type DefineBracketInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
	Body         DefineBracketRequest
}

type MatchEntryResponse struct {
	ID        uuid.UUID `json:"id"`
	MatchID   uuid.UUID `json:"matchId"`
	TeamID    uuid.UUID `json:"teamId"`
	TeamName  string    `json:"teamName,omitempty"`
	FinalRank *int      `json:"finalRank,omitempty"`
	Advances  bool      `json:"advances"`
	CreatedAt time.Time `json:"createdAt"`
}

type MatchResponse struct {
	ID             uuid.UUID            `json:"id"`
	TournamentID   uuid.UUID            `json:"tournamentId"`
	Name           string               `json:"name"`
	Venue          string               `json:"venue"`
	StageLevel     int                  `json:"stageLevel"`
	AdvancingCount int                  `json:"advancingCount"`
	TargetMatchID  *uuid.UUID           `json:"targetMatchId,omitempty"`
	IsCompleted    bool                 `json:"isCompleted"`
	Entries        []MatchEntryResponse `json:"entries"`
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
}

type BracketResponse struct {
	Matches []MatchResponse `json:"matches"`
}

type BracketOutput struct {
	Body BracketResponse
}

type CreateBracketOutput struct {
	Status int `default:"201"`
	Body   BracketResponse
}

type MatchSeedAssignmentItem struct {
	MatchID uuid.UUID   `json:"matchId" doc:"Stage 1 match ID"`
	TeamIDs []uuid.UUID `json:"teamIds" doc:"Accepted tournament team IDs assigned to this match"`
}

type SeedTournamentRequest struct {
	Assignments []MatchSeedAssignmentItem `json:"assignments,omitempty" doc:"Optional explicit assignments of teams to Stage 1 matches. If omitted or empty, teams are auto-distributed round-robin."`
}

type SeedTournamentInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
	Body         SeedTournamentRequest
}

type SeedTournamentOutput struct {
	Body BracketResponse
}

type TeamPlacementItem struct {
	TeamID    uuid.UUID `json:"teamId" doc:"Participating team ID"`
	FinalRank int       `json:"finalRank" minimum:"1" doc:"Match placement (1st, 2nd, 3rd, etc.)"`
}

type RecordMatchResultsRequest struct {
	Placements []TeamPlacementItem `json:"placements" minItems:"1" doc:"Ordered placements of the match's existing participants"`
}

type RecordMatchResultsInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
	MatchID      uuid.UUID `path:"matchId" doc:"Match ID"`
	Body         RecordMatchResultsRequest
}

type RecordMatchResultsResponse struct {
	MatchID        uuid.UUID   `json:"matchId"`
	IsCompleted    bool        `json:"isCompleted"`
	AdvancingTeams []uuid.UUID `json:"advancingTeams"`
	TargetMatchID  *uuid.UUID  `json:"targetMatchId,omitempty"`
}

type RecordMatchResultsOutput struct {
	Body RecordMatchResultsResponse
}

type GetBracketInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
}

func RegisterBracketRoutes(
	api huma.API,
	auth func(huma.Context, func(huma.Context)),
	bracketUseCase *usecases.BracketUseCase,
) {
	// 1. Define Bracket Layout: POST /api/v1/tournaments/{id}/bracket
	huma.Register(api, huma.Operation{
		OperationID:   "define-tournament-bracket",
		Method:        http.MethodPost,
		Path:          "/api/v1/tournaments/{id}/bracket",
		Summary:       "Define bracket layout",
		Description:   "Bulk-creates the entire bracket match graph structure before launch. The graph must be an acyclic converging DAG where edges point forward to higher stage levels.",
		Tags:          []string{"Tournament Bracket"},
		Middlewares:   huma.Middlewares{auth},
		Security:      []map[string][]string{{"bearer": {}}},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, input *DefineBracketInput) (*CreateBracketOutput, error) {
		callerID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		matchesInput := make([]usecases.DefineBracketMatchInput, 0, len(input.Body.Matches))
		for _, m := range input.Body.Matches {
			target := m.TargetMatchID
			if target == nil || strings.TrimSpace(*target) == "" {
				target = m.TargetTempID
			}

			matchesInput = append(matchesInput, usecases.DefineBracketMatchInput{
				TempID:         m.TempID,
				Name:           m.Name,
				Venue:          m.Venue,
				StageLevel:     m.StageLevel,
				AdvancingCount: m.AdvancingCount,
				TargetTempID:   target,
			})
		}

		matches, err := bracketUseCase.DefineBracket(ctx, callerID, input.TournamentID, matchesInput)
		if err != nil {
			return nil, bracketHTTPError(err)
		}

		return &CreateBracketOutput{
			Status: http.StatusCreated,
			Body:   BracketResponse{Matches: toMatchResponses(matches)},
		}, nil
	})

	// 2. Seed Stage 1: POST /api/v1/tournaments/{id}/seed
	huma.Register(api, huma.Operation{
		OperationID: "seed-tournament-stage1",
		Method:      http.MethodPost,
		Path:        "/api/v1/tournaments/{id}/seed",
		Summary:     "Seed Stage 1 matches",
		Description: "Distributes the tournament's accepted teams across Stage 1 matches. If assignments are omitted, teams are auto-distributed round-robin.",
		Tags:        []string{"Tournament Bracket"},
		Middlewares: huma.Middlewares{auth},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *SeedTournamentInput) (*SeedTournamentOutput, error) {
		callerID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		assignments := make([]usecases.Stage1SeedAssignment, 0, len(input.Body.Assignments))
		for _, a := range input.Body.Assignments {
			assignments = append(assignments, usecases.Stage1SeedAssignment{
				MatchID: a.MatchID,
				TeamIDs: a.TeamIDs,
			})
		}

		matches, err := bracketUseCase.SeedStage1(ctx, callerID, input.TournamentID, assignments)
		if err != nil {
			return nil, bracketHTTPError(err)
		}

		return &SeedTournamentOutput{
			Body: BracketResponse{Matches: toMatchResponses(matches)},
		}, nil
	})

	// 3. Record Results & Advance: POST /api/v1/tournaments/{id}/matches/{matchId}/results
	huma.Register(api, huma.Operation{
		OperationID: "record-match-results",
		Method:      http.MethodPost,
		Path:        "/api/v1/tournaments/{id}/matches/{matchId}/results",
		Summary:     "Record match results and advance winners",
		Description: "Records final match rankings, flags the top N teams as advancing, automatically inserts them into the target match node, and marks the current match completed.",
		Tags:        []string{"Tournament Bracket"},
		Middlewares: huma.Middlewares{auth},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *RecordMatchResultsInput) (*RecordMatchResultsOutput, error) {
		callerID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		placements := make([]usecases.TeamPlacementInput, 0, len(input.Body.Placements))
		for _, p := range input.Body.Placements {
			placements = append(placements, usecases.TeamPlacementInput{
				TeamID:    p.TeamID,
				FinalRank: p.FinalRank,
			})
		}

		result, err := bracketUseCase.RecordMatchResults(ctx, callerID, input.TournamentID, input.MatchID, placements)
		if err != nil {
			return nil, bracketHTTPError(err)
		}

		return &RecordMatchResultsOutput{
			Body: RecordMatchResultsResponse{
				MatchID:        result.MatchID,
				IsCompleted:    result.IsCompleted,
				AdvancingTeams: result.AdvancingTeams,
				TargetMatchID:  result.TargetMatchID,
			},
		}, nil
	})

	// 4. View Bracket: GET /api/v1/tournaments/{id}/bracket
	huma.Register(api, huma.Operation{
		OperationID: "get-tournament-bracket",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments/{id}/bracket",
		Summary:     "View tournament bracket",
		Description: "Returns all bracket matches and team participant entries for a tournament.",
		Tags:        []string{"Tournament Bracket"},
	}, func(ctx context.Context, input *GetBracketInput) (*BracketOutput, error) {
		matches, err := bracketUseCase.GetBracket(ctx, input.TournamentID)
		if err != nil {
			return nil, bracketHTTPError(err)
		}

		return &BracketOutput{
			Body: BracketResponse{Matches: toMatchResponses(matches)},
		}, nil
	})
}

func toMatchResponses(matches []models.Match) []MatchResponse {
	resp := make([]MatchResponse, 0, len(matches))
	for _, m := range matches {
		entries := make([]MatchEntryResponse, 0, len(m.Entries))
		for _, e := range m.Entries {
			teamName := ""
			if e.Team.Name != "" {
				teamName = e.Team.Name
			}
			entries = append(entries, MatchEntryResponse{
				ID:        e.ID,
				MatchID:   e.MatchID,
				TeamID:    e.TeamID,
				TeamName:  teamName,
				FinalRank: e.FinalRank,
				Advances:  e.Advances,
				CreatedAt: e.CreatedAt,
			})
		}

		resp = append(resp, MatchResponse{
			ID:             m.ID,
			TournamentID:   m.TournamentID,
			Name:           m.Name,
			Venue:          m.Venue,
			StageLevel:     m.StageLevel,
			AdvancingCount: m.AdvancingCount,
			TargetMatchID:  m.TargetMatchID,
			IsCompleted:    m.IsCompleted,
			Entries:        entries,
			CreatedAt:      m.CreatedAt,
			UpdatedAt:      m.UpdatedAt,
		})
	}
	return resp
}

func bracketHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, usecases.ErrBracketTournamentNotFound):
		return huma.Error404NotFound("Tournament not found", err)
	case errors.Is(err, usecases.ErrBracketMatchNotFound):
		return huma.Error404NotFound("Match not found", err)
	case errors.Is(err, usecases.ErrBracketCallerNotOwner):
		return huma.Error403Forbidden("Only tournament owner can define or seed the bracket", err)
	case errors.Is(err, usecases.ErrBracketCallerNotStaff):
		return huma.Error403Forbidden("Only tournament staff (owner, admin, referee) can record match results", err)
	case errors.Is(err, usecases.ErrBracketTournamentImmutable):
		return huma.Error409Conflict("Bracket operations cannot be performed while tournament is ongoing, completed, or cancelled", err)
	case errors.Is(err, usecases.ErrBracketAlreadyCompleted):
		return huma.Error409Conflict("Bracket cannot be modified after matches have been completed", err)
	case errors.Is(err, usecases.ErrBracketMatchAlreadyCompleted):
		return huma.Error409Conflict("Match is already completed", err)
	case errors.Is(err, usecases.ErrBracketTargetMatchCompleted):
		return huma.Error409Conflict("Target match is already completed", err)
	case errors.Is(err, usecases.ErrBracketEmpty),
		errors.Is(err, usecases.ErrBracketDuplicateTempID),
		errors.Is(err, usecases.ErrBracketEmptyMatchName),
		errors.Is(err, usecases.ErrBracketInvalidStageLevel),
		errors.Is(err, usecases.ErrBracketInvalidAdvancingCount),
		errors.Is(err, usecases.ErrBracketAdvancingCountRequired),
		errors.Is(err, usecases.ErrBracketSelfTargetMatch),
		errors.Is(err, usecases.ErrBracketTargetMatchNotFound),
		errors.Is(err, usecases.ErrBracketTargetMustBeHigher),
		errors.Is(err, usecases.ErrBracketCycleDetected),
		errors.Is(err, usecases.ErrBracketNoChampionshipMatch),
		errors.Is(err, usecases.ErrBracketNoStage1Matches),
		errors.Is(err, usecases.ErrBracketNoAcceptedTeams),
		errors.Is(err, usecases.ErrBracketInvalidSeedMatch),
		errors.Is(err, usecases.ErrBracketInvalidSeedTeam),
		errors.Is(err, usecases.ErrBracketDuplicateSeedTeam),
		errors.Is(err, usecases.ErrBracketNoParticipants),
		errors.Is(err, usecases.ErrBracketMissingTeamPlacement),
		errors.Is(err, usecases.ErrBracketUnknownTeamPlacement),
		errors.Is(err, usecases.ErrBracketInvalidPlacementRank),
		errors.Is(err, usecases.ErrBracketDuplicatePlacementTeam):
		return huma.Error400BadRequest(err.Error(), err)
	default:
		return huma.Error500InternalServerError("Tournament bracket operation failed", err)
	}
}
