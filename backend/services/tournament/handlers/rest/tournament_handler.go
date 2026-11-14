package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type TournamentOutput struct {
	Body TournamentResponse
}

type CreateTournamentRequest struct {
	Name                 string     `json:"name" minLength:"1" maxLength:"160" doc:"Tournament name"`
	Description          string     `json:"description" minLength:"1" maxLength:"500" doc:"Short summary description"`
	Detail               string     `json:"detail,omitempty" maxLength:"10000" doc:"Tournament rules and full details (Markdown supported)"`
	Game                 string     `json:"game" minLength:"1" maxLength:"80" doc:"Game title"`
	Location             string     `json:"location" minLength:"1" maxLength:"160" doc:"Location or 'Online'"`
	StartAt              *time.Time `json:"startAt,omitempty" doc:"Tournament start timestamp (RFC 3339)"`
	EndAt                *time.Time `json:"endAt,omitempty" doc:"Tournament end timestamp (RFC 3339)"`
	RegistrationDeadline *time.Time `json:"registrationDeadline,omitempty" doc:"Registration deadline timestamp (RFC 3339)"`
	EntryFeeMinor        *int64     `json:"entryFeeMinor,omitempty" minimum:"0" maximum:"9007199254740991" doc:"Entry fee in minor currency units (satang; 100 satang = 1 THB)"`
	RegistrationMode     *string    `json:"registrationMode,omitempty" enum:"SOLO,TEAM" doc:"Registration mode: SOLO or TEAM"`
	MinTeamSize          *int       `json:"minTeamSize,omitempty" minimum:"1" doc:"Minimum team size (coerced to 1 for SOLO)"`
	MaxTeamSize          *int       `json:"maxTeamSize,omitempty" minimum:"1" doc:"Maximum team size (coerced to 1 for SOLO)"`
	Capacity             *int       `json:"capacity,omitempty" minimum:"1" doc:"Maximum participant or team capacity"`
	FundingGoalMinor     *int64     `json:"fundingGoalMinor,omitempty" minimum:"1" maximum:"9007199254740991" doc:"Crowdfunding goal in minor currency units (satang; 100 satang = 1 THB); requires fundingDeadline"`
	FundingDeadline      *time.Time `json:"fundingDeadline,omitempty" doc:"Crowdfunding deadline before registrationDeadline; requires fundingGoalMinor"`
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
	Description          *string    `json:"description,omitempty" maxLength:"500" doc:"New short summary"`
	Detail               *string    `json:"detail,omitempty" maxLength:"10000" doc:"New tournament rules and full details (Markdown supported)"`
	Game                 *string    `json:"game,omitempty" minLength:"1" maxLength:"80" doc:"New game title"`
	Location             *string    `json:"location,omitempty" minLength:"1" maxLength:"160" doc:"New location"`
	StartAt              *time.Time `json:"startAt,omitempty" doc:"New start timestamp"`
	EndAt                *time.Time `json:"endAt,omitempty" doc:"New end timestamp"`
	RegistrationDeadline *time.Time `json:"registrationDeadline,omitempty" doc:"New registration deadline timestamp"`
	EntryFeeMinor        *int64     `json:"entryFeeMinor,omitempty" minimum:"0" maximum:"9007199254740991" doc:"New entry fee in minor currency units (satang; 100 satang = 1 THB)"`
	RegistrationMode     *string    `json:"registrationMode,omitempty" enum:"SOLO,TEAM" doc:"New registration mode"`
	MinTeamSize          *int       `json:"minTeamSize,omitempty" minimum:"1" doc:"New minimum team size"`
	MaxTeamSize          *int       `json:"maxTeamSize,omitempty" minimum:"1" doc:"New maximum team size"`
	Capacity             *int       `json:"capacity,omitempty" minimum:"1" doc:"New capacity"`
	FundingGoalMinor     *int64     `json:"fundingGoalMinor,omitempty" minimum:"1" maximum:"9007199254740991" doc:"New crowdfunding goal in minor currency units (satang) while the campaign is active"`
	FundingDeadline      *time.Time `json:"fundingDeadline,omitempty" doc:"New crowdfunding deadline while the campaign is active"`
}

type UpdateTournamentInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	Body         UpdateTournamentRequest
}

type TournamentResponse struct {
	ID                   uuid.UUID  `json:"id"`
	Name                 string     `json:"name"`
	Description          string     `json:"description" doc:"Short summary description"`
	Detail               string     `json:"detail" doc:"Tournament rules and full details (Markdown supported)"`
	Game                 string     `json:"game"`
	Location             string     `json:"location"`
	OrganizerID          uuid.UUID  `json:"organizerId"`
	OrganizerName        string     `json:"organizerName"`
	StartAt              *time.Time `json:"startAt,omitempty"`
	EndAt                *time.Time `json:"endAt,omitempty"`
	RegistrationDeadline *time.Time `json:"registrationDeadline,omitempty"`
	EntryFeeMinor        int64      `json:"entryFeeMinor" doc:"Entry fee in minor currency units (satang; 100 satang = 1 THB)"`
	Currency             string     `json:"currency"`
	RegistrationMode     string     `json:"registrationMode" enum:"SOLO,TEAM"`
	MinTeamSize          int        `json:"minTeamSize"`
	MaxTeamSize          int        `json:"maxTeamSize"`
	Capacity             int        `json:"capacity"`
	RegisteredCount      int        `json:"registeredCount"`
	Status               string     `json:"status"`
}

type SearchTournamentsInput struct {
	Q                string `query:"q" maxLength:"100" doc:"Case-insensitive name, game, or description search"`
	OrganizerID      string `query:"organizerId" doc:"Filter published tournaments owned by an organizer account UUID"`
	StartFrom        string `query:"startFrom" doc:"Earliest tournament start date (YYYY-MM-DD)"`
	StartTo          string `query:"startTo" doc:"Latest tournament start date (YYYY-MM-DD)"`
	MinEntryFeeMinor int64 `query:"minEntryFeeMinor" default:"-1" minimum:"-1" doc:"Minimum entry fee in minor units (satang); omit to disable"`
	MaxEntryFeeMinor int64 `query:"maxEntryFeeMinor" default:"-1" minimum:"-1" doc:"Maximum entry fee in minor units (satang); omit to disable"`
	Status      string `query:"status" enum:"DRAFT,CROWDFUNDING,REGISTRATION,ONGOING,COMPLETED"`
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
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
}

type TournamentFundingResponse struct {
	GoalAmountMinor      int64      `json:"goalAmountMinor" doc:"Funding goal in minor currency units (satang; 100 satang = 1 THB)"`
	Deadline             *time.Time `json:"deadline,omitempty" doc:"Crowdfunding deadline when this tournament is crowdfunding"`
	RaisedAmountMinor    int64      `json:"raisedAmountMinor" doc:"Amount raised in minor currency units (satang; 100 satang = 1 THB)"`
	RemainingAmountMinor int64      `json:"remainingAmountMinor" doc:"Amount remaining to reach the goal in minor currency units (satang)"`
	SupporterCount       int        `json:"supporterCount"`
	Percentage           float64    `json:"percentage" doc:"Percentage of the funding goal raised"`
	Currency             string     `json:"currency"`
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

func RegisterTournamentRoutes(
	api huma.API,
	auth, organizer func(huma.Context, func(huma.Context)),
	tournamentUseCase *usecases.TournamentUseCase,
) {
	huma.Register(api, huma.Operation{
		OperationID: "search-tournaments",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments",
		Summary:     "Browse and filter tournaments",
		Description: "Returns published tournaments matching schedule, budget, status, and text filters.",
		Tags:        []string{"Tournaments"},
	}, func(ctx context.Context, input *SearchTournamentsInput) (*SearchTournamentsOutput, error) {
		result, err := tournamentUseCase.Search(ctx, usecases.SearchTournamentsInput{
			Query: input.Q, OrganizerID: input.OrganizerID, StartFrom: input.StartFrom, StartTo: input.StartTo,
			MinEntryFeeMinor: optionalFee(input.MinEntryFeeMinor), MaxEntryFeeMinor: optionalFee(input.MaxEntryFeeMinor),
			Status: input.Status, Sort: input.Sort, Page: input.Page, PageSize: input.PageSize,
		})
		if err != nil {
			if errors.Is(err, usecases.ErrInvalidTournamentFilters) {
				return nil, huma.Error400BadRequest("Invalid tournament filters", err)
			}
			if mapped := mapAccountPortError(err); mapped != nil {
				return nil, mapped
			}
			return nil, huma.Error500InternalServerError("Failed to search tournaments", err)
		}

		items := make([]TournamentResponse, 0, len(result.Items))
		for _, item := range result.Items {
			items = append(items, toTournamentResponse(item.Tournament, item.OrganizerName, item.RegisteredCount))
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
			if mapped := mapAccountPortError(err); mapped != nil {
				return nil, mapped
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
			Tournament: toTournamentResponse(result.Tournament, result.OrganizerName, result.RegisteredCount),
			Teams:      teams,
			Funding: TournamentFundingResponse{
				GoalAmountMinor:      result.Funding.GoalAmountMinor,
				Deadline:             result.Funding.Deadline,
				RaisedAmountMinor:    result.Funding.RaisedAmountMinor,
				RemainingAmountMinor: result.Funding.RemainingAmountMinor,
				SupporterCount:       result.Funding.SupporterCount,
				Percentage:           result.Funding.Percentage,
				Currency:             result.Tournament.Currency,
			},
		}}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-tournament",
		Method:      http.MethodPost,
		Path:        "/api/v1/tournaments",
		Summary:     "Create tournament",
		Description: "Creates a new tournament for the authenticated organizer. Requires an active organizer profile.\n\n" +
			"### Rules & Defaults:\n" +
			"- **Status & Visibility:** Defaults to `REGISTRATION`; valid fundingGoalMinor and fundingDeadline start in `CROWDFUNDING`.\n" +
			"- **Currency:** Defaulted to `THB`.\n" +
			"- **Registration Mode:** If `SOLO`, `minTeamSize` and `maxTeamSize` are automatically set to `1`. If `TEAM`, requires `1 <= minTeamSize <= maxTeamSize`.\n" +
			"- **Date Integrity (400):** Requires `registrationDeadline < startAt < endAt`.\n" +
			"- **Validation (400):** Requires non-empty `name`, `game`, `location`, `capacity >= 1`, and `entryFeeMinor >= 0`.",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"Tournaments"},
		Middlewares:   huma.Middlewares{auth, organizer},
		Security:      []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *CreateTournamentInput) (*CreateTournamentOutput, error) {
		organizer, err := middlewares.GetOrganizerProfile(ctx)
		if err != nil {
			return nil, tournamentHTTPError(err)
		}

		tournament, err := tournamentUseCase.CreateTournament(ctx, organizer.AccountID, usecases.CreateTournamentInput{
			Name:                 input.Body.Name,
			Description:          input.Body.Description,
			Detail:               input.Body.Detail,
			Game:                 input.Body.Game,
			Location:             input.Body.Location,
			StartsAt:             input.Body.StartAt,
			EndsAt:               input.Body.EndAt,
			RegistrationDeadline: input.Body.RegistrationDeadline,
			EntryFeeMinor:        input.Body.EntryFeeMinor,
			RegistrationMode:     input.Body.RegistrationMode,
			MinTeamSize:          input.Body.MinTeamSize,
			MaxTeamSize:          input.Body.MaxTeamSize,
			Capacity:             input.Body.Capacity,
			FundingGoalMinor:     input.Body.FundingGoalMinor,
			FundingDeadline:      input.Body.FundingDeadline,
		})
		if err != nil {
			return nil, tournamentHTTPError(err)
		}

		return &CreateTournamentOutput{
			Status: http.StatusCreated,
			Body:   toTournamentResponse(*tournament, organizer.OrganizerName, 0),
		}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "update-tournament",
		Method:      http.MethodPatch,
		Path:        "/api/v1/tournaments/{tournamentId}",
		Summary:     "Update tournament",
		Description: "Updates an existing tournament configuration. Only permitted for the organizing owner with the following invariants:\n\n" +
			"### Restrictions & Invariants:\n" +
			"- **Ownership (403 Forbidden):** Only the organizer profile that created the tournament can modify it.\n" +
			"- **Lifecycle Immutability (409 Conflict):** Tournaments with status `ONGOING` or `COMPLETED` cannot have their configuration modified.\n" +
			"- **Capacity Floor (409 Conflict):** `capacity` cannot be reduced below the number of currently accepted teams/participants (`capacity >= acceptedCount`).\n" +
			"- **Roster Freezing (409 Conflict):** Once any team has locked or been accepted into the tournament (`lockedOrAcceptedCount > 0`), `registrationMode`, `minTeamSize`, and `maxTeamSize` cannot be altered.\n" +
			"- **Date Integrity (400 Bad Request):** Merged dates must maintain `registrationDeadline < startAt < endAt`.\n" +
			"- **Crowdfunding:** The owner may edit the goal and funding deadline only while the campaign is active; the funding deadline must remain before registration closes. Planned event dates and location can be updated during this stage.\n" +
			"- **Field Validation (400 Bad Request):** If provided, `name`, `game`, and `location` cannot be empty or whitespace; `entryFeeMinor` cannot be negative.",
		Tags:        []string{"Tournaments"},
		Middlewares: huma.Middlewares{auth, organizer},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *UpdateTournamentInput) (*TournamentOutput, error) {
		organizer, err := middlewares.GetOrganizerProfile(ctx)
		if err != nil {
			return nil, tournamentHTTPError(err)
		}

		tournament, err := tournamentUseCase.UpdateTournament(ctx, organizer.AccountID, input.TournamentID, usecases.UpdateTournamentInput{
			Name:                 input.Body.Name,
			Description:          input.Body.Description,
			Detail:               input.Body.Detail,
			Game:                 input.Body.Game,
			Location:             input.Body.Location,
			StartsAt:             input.Body.StartAt,
			EndsAt:               input.Body.EndAt,
			RegistrationDeadline: input.Body.RegistrationDeadline,
			EntryFeeMinor:        input.Body.EntryFeeMinor,
			RegistrationMode:     input.Body.RegistrationMode,
			MinTeamSize:          input.Body.MinTeamSize,
			MaxTeamSize:          input.Body.MaxTeamSize,
			Capacity:             input.Body.Capacity,
			FundingGoalMinor:     input.Body.FundingGoalMinor,
			FundingDeadline:      input.Body.FundingDeadline,
		})
		if err != nil {
			return nil, tournamentHTTPError(err)
		}

		return &TournamentOutput{
			Body: toTournamentResponse(*tournament, organizer.OrganizerName, 0),
		}, nil
	})
}

func tournamentHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, usecases.ErrTournamentNotFound):
		return huma.Error404NotFound("Tournament not found", err)
	case errors.Is(err, usecases.ErrNotTournamentOwner):
		return huma.Error403Forbidden("Only the organizing owner can modify this tournament", err)
	case errors.Is(err, middlewares.ErrOrganizerProfileRequired):
		return huma.Error403Forbidden("Organizer profile required to perform this action", err)
	case errors.Is(err, usecases.ErrCapacityBelowAcceptedTeams),
		errors.Is(err, usecases.ErrRosterRulesLocked),
		errors.Is(err, usecases.ErrTournamentCannotBeModified),
		errors.Is(err, usecases.ErrTournamentStateChanged),
		errors.Is(err, usecases.ErrCrowdfundingClosed):
		return huma.Error409Conflict(err.Error(), err)
	case errors.Is(err, usecases.ErrInvalidTournamentDates),
		errors.Is(err, usecases.ErrInvalidTournamentName),
		errors.Is(err, usecases.ErrInvalidTournamentDescription),
		errors.Is(err, usecases.ErrInvalidTournamentDetail),
		errors.Is(err, usecases.ErrInvalidTournamentGame),
		errors.Is(err, usecases.ErrInvalidTournamentLocation),
		errors.Is(err, usecases.ErrInvalidTournamentFee),
		errors.Is(err, usecases.ErrInvalidTournamentCapacity),
		errors.Is(err, usecases.ErrInvalidRegistrationMode),
		errors.Is(err, usecases.ErrInvalidTeamSize),
		errors.Is(err, usecases.ErrInvalidCrowdfunding):
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

func toTournamentResponse(tournament models.Tournament, organizerName string, registeredCount int) TournamentResponse {
	var startAt, endAt, regDeadline *time.Time
	if !tournament.StartsAt.IsZero() {
		startAt = &tournament.StartsAt
	}
	if !tournament.EndsAt.IsZero() {
		endAt = &tournament.EndsAt
	}
	if !tournament.RegistrationDeadline.IsZero() {
		regDeadline = &tournament.RegistrationDeadline
	}
	return TournamentResponse{
		ID: tournament.ID, Name: tournament.Name, Description: tournament.Description,
		Detail: tournament.Detail,
		Game: tournament.Game, Location: tournament.Location,
		OrganizerID:   tournament.OwnerID(),
		OrganizerName: organizerName,
		StartAt:       startAt, EndAt: endAt,
		RegistrationDeadline: regDeadline,
		EntryFeeMinor:        tournament.EntryFeeMinor, Currency: tournament.Currency,
		RegistrationMode: string(tournament.RegistrationMode),
		MinTeamSize:      tournament.MinTeamSize, MaxTeamSize: tournament.MaxTeamSize,
		Capacity: tournament.Capacity, RegisteredCount: registeredCount,
		Status: string(tournament.Status),
	}
}
