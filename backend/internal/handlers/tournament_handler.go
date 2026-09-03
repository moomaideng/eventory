package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/usecases"
)

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
	Capacity             int       `json:"capacity"`
	RegisteredCount      int       `json:"registeredCount"`
	Status               string    `json:"status"`
}

// query input
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

func RegisterTournamentRoutes(api huma.API, tournamentUseCase *usecases.TournamentUseCase) {
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
		for _, tournament := range result.Items {
			items = append(items, toTournamentResponse(tournament))
		}
		return &SearchTournamentsOutput{Body: TournamentListBody{
			Items: items, Total: result.Total, Page: result.Page, PageSize: result.PageSize,
		}}, nil
	})
}

func optionalFee(value int64) *int64 {
	if value < 0 {
		return nil
	}
	return &value
}

// response
func toTournamentResponse(tournament models.Tournament) TournamentResponse {
	return TournamentResponse{
		ID: tournament.ID, Name: tournament.Name, Description: tournament.Description,
		Game: tournament.Game, Location: tournament.Location,
		OrganizerName: tournament.Organizer.OrganizerName,
		StartAt:       tournament.StartAt, EndAt: tournament.EndAt,
		RegistrationDeadline: tournament.RegistrationDeadline,
		EntryFee:             tournament.EntryFee, Currency: tournament.Currency,
		Capacity: tournament.Capacity, RegisteredCount: tournament.RegisteredCount,
		Status: tournament.Status,
	}
}
