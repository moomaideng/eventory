package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/usecases"
)

type TournamentOutput struct {
	Body TournamentResponse
}

type CreateTournamentRequest struct {
	Name                 string    `json:"name" minLength:"1" maxLength:"160" doc:"Tournament name"`
	Description          string    `json:"description" doc:"Tournament description or rules"`
	Game                 string    `json:"game" minLength:"1" maxLength:"80" doc:"Game title"`
	Location             string    `json:"location" minLength:"1" maxLength:"160" doc:"Location or 'Online'"`
	StartAt              time.Time `json:"startAt" doc:"Tournament start timestamp (RFC 3339)"`
	EndAt                time.Time `json:"endAt" doc:"Tournament end timestamp (RFC 3339)"`
	RegistrationDeadline time.Time `json:"registrationDeadline" doc:"Registration deadline timestamp (RFC 3339)"`
	EntryFee             int64     `json:"entryFee" minimum:"0" default:"0" doc:"Entry fee in whole currency units"`
	RegistrationMode     string    `json:"registrationMode" enum:"SOLO,TEAM" doc:"Registration mode: SOLO or TEAM"`
	MinTeamSize          int       `json:"minTeamSize" minimum:"1" default:"1" doc:"Minimum team size (coerced to 1 for SOLO)"`
	MaxTeamSize          int       `json:"maxTeamSize" minimum:"1" default:"1" doc:"Maximum team size (coerced to 1 for SOLO)"`
	Capacity             int       `json:"capacity" minimum:"1" doc:"Maximum participant or team capacity"`
}

type CreateTournamentInput struct {
	Body CreateTournamentRequest
}

type CreateTournamentOutput struct {
	Status int `default:"201"`
	Body   TournamentResponse
}

type UpdateTournamentRequest struct {
	Name                 *string    `json:"name,omitempty" minLength:"1" maxLength:"160" doc:"New tournament name"`
	Description          *string    `json:"description,omitempty" doc:"New tournament description"`
	Game                 *string    `json:"game,omitempty" minLength:"1" maxLength:"80" doc:"New game title"`
	Location             *string    `json:"location,omitempty" minLength:"1" maxLength:"160" doc:"New location"`
	StartAt              *time.Time `json:"startAt,omitempty" doc:"New start timestamp"`
	EndAt                *time.Time `json:"endAt,omitempty" doc:"New end timestamp"`
	RegistrationDeadline *time.Time `json:"registrationDeadline,omitempty" doc:"New registration deadline timestamp"`
	EntryFee             *int64     `json:"entryFee,omitempty" minimum:"0" doc:"New entry fee"`
	RegistrationMode     *string    `json:"registrationMode,omitempty" enum:"SOLO,TEAM" doc:"New registration mode"`
	MinTeamSize          *int       `json:"minTeamSize,omitempty" minimum:"1" doc:"New minimum team size"`
	MaxTeamSize          *int       `json:"maxTeamSize,omitempty" minimum:"1" doc:"New maximum team size"`
	Capacity             *int       `json:"capacity,omitempty" minimum:"1" doc:"New capacity"`
}

type UpdateTournamentInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	Body         UpdateTournamentRequest
}

type TournamentResponse struct {
	ID                   uuid.UUID `json:"id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	Game                 string    `json:"game"`
	Location             string    `json:"location"`
	OrganizerName        string    `json:"organizerName"`
	StartAt              time.Time `json:"startAt"`
	EndAt                time.Time `json:"endAt"`
	RegistrationDeadline time.Time `json:"registrationDeadline"`
	EntryFee             int64     `json:"entryFee" doc:"Entry fee in whole currency units"`
	Currency             string    `json:"currency"`
	RegistrationMode     string    `json:"registrationMode" enum:"SOLO,TEAM"`
	MinTeamSize          int       `json:"minTeamSize"`
	MaxTeamSize          int       `json:"maxTeamSize"`
	Capacity             int       `json:"capacity"`
	RegisteredCount      int       `json:"registeredCount"`
	Status               string    `json:"status"`
}

type SearchTournamentsInput struct {
	Q           string `query:"q" maxLength:"100" doc:"Case-insensitive name, game, or description search"`
	StartFrom   string `query:"startFrom" doc:"Earliest tournament start date (YYYY-MM-DD)"`
	StartTo     string `query:"startTo" doc:"Latest tournament start date (YYYY-MM-DD)"`
	MinEntryFee int64  `query:"minEntryFee" default:"-1" minimum:"-1" doc:"Minimum entry fee; omit to disable"`
	MaxEntryFee int64  `query:"maxEntryFee" default:"-1" minimum:"-1" doc:"Maximum entry fee; omit to disable"`
	Status      string `query:"status" enum:"REGISTRATION_OPEN,REGISTRATION_CLOSED,ONGOING,COMPLETED"`
	Sort        string `query:"sort" default:"start_asc" enum:"start_asc,start_desc,fee_asc,fee_desc"`
	Page        int    `query:"page" default:"1" minimum:"1"`
	PageSize    int    `query:"pageSize" default:"12" minimum:"1" maximum:"100"`
}

type TournamentListBody struct {
	Items    []TournamentResponse `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

type SearchTournamentsOutput struct {
	Body TournamentListBody
}

type TournamentTeamResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	MemberCount int       `json:"memberCount"`
}

type TournamentFundingResponse struct {
	GoalAmount      int64   `json:"goalAmount" doc:"Funding goal in whole currency units"`
	RaisedAmount    int64   `json:"raisedAmount" doc:"Amount raised in whole currency units"`
	RemainingAmount int64   `json:"remainingAmount" doc:"Amount remaining to reach the goal"`
	SupporterCount  int     `json:"supporterCount"`
	Percentage      float64 `json:"percentage" doc:"Percentage of the funding goal raised"`
	Currency        string  `json:"currency"`
}

type TournamentDetailsBody struct {
	Tournament TournamentResponse        `json:"tournament"`
	Teams      []TournamentTeamResponse  `json:"teams"`
	Funding    TournamentFundingResponse `json:"funding"`
}

type GetTournamentDetailsInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament ID"`
}

type GetTournamentDetailsOutput struct {
	Body TournamentDetailsBody
}

func RegisterTournamentRoutes(api huma.API, organizerGroup huma.API, tournamentUseCase *usecases.TournamentUseCase) {
	huma.Register(api, huma.Operation{
		OperationID: "search-tournaments",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments",
		Summary:     "Browse and filter tournaments",
		Description: "Returns published tournaments matching schedule, budget, status, and text filters.",
		Tags:        []string{"Tournaments"},
	}, func(ctx context.Context, input *SearchTournamentsInput) (*SearchTournamentsOutput, error) {
		result, err := tournamentUseCase.Search(ctx, usecases.SearchTournamentsInput{
			Query: input.Q, StartFrom: input.StartFrom, StartTo: input.StartTo,
			MinEntryFee: optionalFee(input.MinEntryFee), MaxEntryFee: optionalFee(input.MaxEntryFee),
			Status: input.Status, Sort: input.Sort, Page: input.Page, PageSize: input.PageSize,
		})
		if err != nil {
			if errors.Is(err, usecases.ErrInvalidTournamentFilters) {
				return nil, huma.Error400BadRequest("Invalid tournament filters", err)
			}
			return nil, huma.Error500InternalServerError("Failed to search tournaments", err)
		}

		items := make([]TournamentResponse, 0, len(result.Items))
		for _, item := range result.Items {
			items = append(items, toTournamentResponse(item.Tournament, item.RegisteredCount))
		}
		return &SearchTournamentsOutput{Body: TournamentListBody{
			Items: items, Total: result.Total, Page: result.Page, PageSize: result.PageSize,
		}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-tournament-details",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments/{tournamentId}",
		Summary:     "View tournament details",
		Description: "Returns a published tournament with its registered teams and funding progress.",
		Tags:        []string{"Tournaments"},
	}, func(ctx context.Context, input *GetTournamentDetailsInput) (*GetTournamentDetailsOutput, error) {
		result, err := tournamentUseCase.GetDetails(ctx, input.TournamentID)
		if err != nil {
			if errors.Is(err, usecases.ErrTournamentNotFound) {
				return nil, huma.Error404NotFound("Tournament not found")
			}
			return nil, huma.Error500InternalServerError("Failed to load tournament", err)
		}

		teams := make([]TournamentTeamResponse, 0, len(result.Tournament.Teams))
		for _, team := range result.Tournament.Teams {
			teams = append(teams, TournamentTeamResponse{
				ID: team.ID, Name: team.Name, MemberCount: len(team.Members),
			})
		}

		return &GetTournamentDetailsOutput{Body: TournamentDetailsBody{
			Tournament: toTournamentResponse(result.Tournament, result.RegisteredCount),
			Teams:      teams,
			Funding: TournamentFundingResponse{
				GoalAmount:      result.Funding.GoalAmount,
				RaisedAmount:    result.Funding.RaisedAmount,
				RemainingAmount: result.Funding.RemainingAmount,
				SupporterCount:  result.Funding.SupporterCount,
				Percentage:      result.Funding.Percentage,
				Currency:        result.Tournament.Currency,
			},
		}}, nil
	})

	huma.Register(organizerGroup, huma.Operation{
		OperationID:   "create-tournament",
		Method:        http.MethodPost,
		Path:          "",
		Summary:       "Create tournament",
		Description:   "Creates a new tournament for the authenticated organizer. Requires an active organizer profile.\n\n" +
			"### Rules & Defaults:\n" +
			"- **Status & Visibility:** Defaults to `REGISTRATION_OPEN` and `published: true`.\n" +
			"- **Currency:** Defaulted to `THB`.\n" +
			"- **Registration Mode:** If `SOLO`, `minTeamSize` and `maxTeamSize` are automatically set to `1`. If `TEAM`, requires `1 <= minTeamSize <= maxTeamSize`.\n" +
			"- **Date Integrity (400):** Requires `registrationDeadline < startAt < endAt`.\n" +
			"- **Validation (400):** Requires non-empty `name`, `game`, `location`, `capacity >= 1`, and `entryFee >= 0`.",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"Tournaments"},
		Security:      []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *CreateTournamentInput) (*CreateTournamentOutput, error) {
		organizer, err := middlewares.GetOrganizerProfile(ctx)
		if err != nil {
			return nil, tournamentHTTPError(err)
		}

		tournament, err := tournamentUseCase.CreateTournament(ctx, organizer, usecases.CreateTournamentInput{
			Name:                 input.Body.Name,
			Description:          input.Body.Description,
			Game:                 input.Body.Game,
			Location:             input.Body.Location,
			StartsAt:             input.Body.StartAt,
			EndsAt:               input.Body.EndAt,
			RegistrationDeadline: input.Body.RegistrationDeadline,
			EntryFee:             input.Body.EntryFee,
			RegistrationMode:     input.Body.RegistrationMode,
			MinTeamSize:          input.Body.MinTeamSize,
			MaxTeamSize:          input.Body.MaxTeamSize,
			Capacity:             input.Body.Capacity,
		})
		if err != nil {
			return nil, tournamentHTTPError(err)
		}

		return &CreateTournamentOutput{
			Status: http.StatusCreated,
			Body:   toTournamentResponse(*tournament, 0),
		}, nil
	})

	huma.Register(organizerGroup, huma.Operation{
		OperationID: "update-tournament",
		Method:      http.MethodPatch,
		Path:        "/{tournamentId}",
		Summary:     "Update tournament",
		Description: "Updates an existing tournament configuration. Only permitted for the organizing owner with the following invariants:\n\n" +
			"### Restrictions & Invariants:\n" +
			"- **Ownership (403 Forbidden):** Only the organizer profile that created the tournament can modify it.\n" +
			"- **Lifecycle Immutability (409 Conflict):** Tournaments with status `ONGOING` or `COMPLETED` cannot have their configuration modified.\n" +
			"- **Capacity Floor (409 Conflict):** `capacity` cannot be reduced below the number of currently accepted teams/participants (`capacity >= acceptedCount`).\n" +
			"- **Roster Freezing (409 Conflict):** Once any team has locked or been accepted into the tournament (`lockedOrAcceptedCount > 0`), `registrationMode`, `minTeamSize`, and `maxTeamSize` cannot be altered.\n" +
			"- **Date Integrity (400 Bad Request):** Merged dates must maintain `registrationDeadline < startAt < endAt`.\n" +
			"- **Field Validation (400 Bad Request):** If provided, `name`, `game`, and `location` cannot be empty or whitespace; `entryFee` cannot be negative.",
		Tags:        []string{"Tournaments"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *UpdateTournamentInput) (*TournamentOutput, error) {
		organizer, err := middlewares.GetOrganizerProfile(ctx)
		if err != nil {
			return nil, tournamentHTTPError(err)
		}

		tournament, err := tournamentUseCase.UpdateTournament(ctx, organizer.ID, input.TournamentID, usecases.UpdateTournamentInput{
			Name:                 input.Body.Name,
			Description:          input.Body.Description,
			Game:                 input.Body.Game,
			Location:             input.Body.Location,
			StartsAt:             input.Body.StartAt,
			EndsAt:               input.Body.EndAt,
			RegistrationDeadline: input.Body.RegistrationDeadline,
			EntryFee:             input.Body.EntryFee,
			RegistrationMode:     input.Body.RegistrationMode,
			MinTeamSize:          input.Body.MinTeamSize,
			MaxTeamSize:          input.Body.MaxTeamSize,
			Capacity:             input.Body.Capacity,
		})
		if err != nil {
			return nil, tournamentHTTPError(err)
		}

		return &TournamentOutput{
			Body: toTournamentResponse(*tournament, 0),
		}, nil
	})
}

func tournamentHTTPError(err error) error {
	switch {
	case errors.Is(err, usecases.ErrTournamentNotFound):
		return huma.Error404NotFound("Tournament not found", err)
	case errors.Is(err, usecases.ErrNotTournamentOwner):
		return huma.Error403Forbidden("Only the organizing owner can modify this tournament", err)
	case errors.Is(err, middlewares.ErrOrganizerProfileRequired):
		return huma.Error403Forbidden("Organizer profile required to perform this action", err)
	case errors.Is(err, usecases.ErrCapacityBelowAcceptedTeams),
		errors.Is(err, usecases.ErrRosterRulesLocked),
		errors.Is(err, usecases.ErrTournamentCannotBeModified):
		return huma.Error409Conflict(err.Error(), err)
	case errors.Is(err, usecases.ErrInvalidTournamentDates),
		errors.Is(err, usecases.ErrInvalidTournamentName),
		errors.Is(err, usecases.ErrInvalidTournamentGame),
		errors.Is(err, usecases.ErrInvalidTournamentLocation),
		errors.Is(err, usecases.ErrInvalidTournamentFee),
		errors.Is(err, usecases.ErrInvalidTournamentCapacity),
		errors.Is(err, usecases.ErrInvalidRegistrationMode),
		errors.Is(err, usecases.ErrInvalidTeamSize):
		return huma.Error400BadRequest(err.Error(), err)
	default:
		return huma.Error500InternalServerError("Tournament operation failed", err)
	}
}

func optionalFee(value int64) *int64 {
	if value < 0 {
		return nil
	}
	return &value
}

func toTournamentResponse(tournament models.Tournament, registeredCount int) TournamentResponse {
	return TournamentResponse{
		ID: tournament.ID, Name: tournament.Name, Description: tournament.Description,
		Game: tournament.Game, Location: tournament.Location,
		OrganizerName: tournament.Organizer.OrganizerName,
		StartAt:       tournament.StartsAt, EndAt: tournament.EndsAt,
		RegistrationDeadline: tournament.RegistrationDeadline,
		EntryFee:             tournament.EntryFee, Currency: tournament.Currency,
		RegistrationMode: string(tournament.RegistrationMode),
		MinTeamSize:      tournament.MinTeamSize, MaxTeamSize: tournament.MaxTeamSize,
		Capacity: tournament.Capacity, RegisteredCount: registeredCount,
		Status: string(tournament.Status),
	}
}
