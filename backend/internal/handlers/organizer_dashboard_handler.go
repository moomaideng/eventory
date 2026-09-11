package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/usecases"
)

type OrganizerDashboardMetricsResponse struct {
	TotalEntries          int `json:"totalEntries"`
	FormingEntries        int `json:"formingEntries"`
	LockedEntries         int `json:"lockedEntries"`
	AcceptedEntries       int `json:"acceptedEntries" doc:"Accepted entries consuming tournament capacity"`
	RejectedEntries       int `json:"rejectedEntries"`
	ConfirmedParticipants int `json:"confirmedParticipants" doc:"Distinct accounts on accepted entries"`
	AvailableSpots        int `json:"availableSpots"`
}

type OrganizerTournamentSummary struct {
	Tournament TournamentResponse                `json:"tournament"`
	Published  bool                              `json:"published"`
	Metrics    OrganizerDashboardMetricsResponse `json:"metrics"`
	Funding    TournamentFundingResponse         `json:"funding"`
}

type OrganizerTournamentListInput struct {
	Page     int `query:"page" default:"1" minimum:"1" maximum:"1000000"`
	PageSize int `query:"pageSize" default:"12" minimum:"1" maximum:"100"`
}

type OrganizerTournamentListOutput struct {
	Body struct {
		Items    []OrganizerTournamentSummary `json:"items"`
		Total    int64                        `json:"total"`
		Page     int                          `json:"page"`
		PageSize int                          `json:"pageSize"`
	}
}

type OrganizerDashboardInput struct {
	TournamentID uuid.UUID `path:"id"`
}

type OrganizerDashboardEntry struct {
	ID        uuid.UUID                 `json:"id"`
	Name      string                    `json:"name"`
	Status    string                    `json:"status"`
	CreatedAt time.Time                 `json:"createdAt"`
	Members   []TeamLobbyMemberResponse `json:"members"`
}

type OrganizerDashboardOutput struct {
	Body struct {
		Summary OrganizerTournamentSummary `json:"summary"`
		Entries []OrganizerDashboardEntry  `json:"entries"`
	}
}

func RegisterOrganizerDashboardRoutes(api huma.API, dashboard *usecases.OrganizerDashboardUseCase, accounts *usecases.AccountUseCase) {
	huma.Register(api, huma.Operation{
		OperationID: "list-my-tournaments", Method: http.MethodGet, Path: "/accounts/me/tournaments",
		Summary: "List the current organizer's tournaments", Tags: []string{"Organizer dashboard"},
		Security: []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *OrganizerTournamentListInput) (*OrganizerTournamentListOutput, error) {
		account, err := authenticatedAccount(ctx, accounts)
		if err != nil {
			return nil, err
		}
		result, err := dashboard.List(ctx, account.ID, input.Page, input.PageSize)
		if err != nil {
			return nil, organizerDashboardHTTPError(err)
		}
		output := &OrganizerTournamentListOutput{}
		output.Body.Items = make([]OrganizerTournamentSummary, 0, len(result.Items))
		for _, item := range result.Items {
			output.Body.Items = append(output.Body.Items, toOrganizerSummary(item))
		}
		output.Body.Total, output.Body.Page, output.Body.PageSize = result.Total, result.Page, result.PageSize
		return output, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-organizer-tournament-dashboard", Method: http.MethodGet, Path: "/tournaments/{id}/dashboard",
		Summary: "View an owned tournament's dashboard and registrations", Tags: []string{"Organizer dashboard"},
		Security: []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *OrganizerDashboardInput) (*OrganizerDashboardOutput, error) {
		account, err := authenticatedAccount(ctx, accounts)
		if err != nil {
			return nil, err
		}
		result, err := dashboard.Get(ctx, account.ID, input.TournamentID)
		if err != nil {
			return nil, organizerDashboardHTTPError(err)
		}
		output := &OrganizerDashboardOutput{}
		output.Body.Summary = toOrganizerSummary(*result)
		output.Body.Entries = make([]OrganizerDashboardEntry, 0, len(result.Tournament.Teams))
		for _, team := range result.Tournament.Teams {
			members := make([]TeamLobbyMemberResponse, 0, len(team.Members))
			for _, member := range team.Members {
				members = append(members, TeamLobbyMemberResponse{
					ID: member.ID, AccountID: member.AccountID, Handle: member.Account.Handle, DisplayName: member.Account.DisplayName,
					Role: string(member.Role), JoinedAt: member.JoinedAt,
				})
			}
			output.Body.Entries = append(output.Body.Entries, OrganizerDashboardEntry{
				ID: team.ID, Name: team.Name, Status: string(team.Status), CreatedAt: team.CreatedAt, Members: members,
			})
		}
		return output, nil
	})
}

func organizerDashboardHTTPError(err error) error {
	switch {
	case errors.Is(err, usecases.ErrTournamentNotFound):
		return huma.Error404NotFound("Tournament not found or not owned by your account")
	case errors.Is(err, usecases.ErrInvalidDashboardPage):
		return huma.Error400BadRequest("Invalid dashboard page")
	default:
		return huma.Error500InternalServerError("Unable to load organizer dashboard")
	}
}

func toOrganizerSummary(result usecases.OrganizerDashboard) OrganizerTournamentSummary {
	m := result.Metrics
	return OrganizerTournamentSummary{
		Tournament: toTournamentResponse(result.Tournament, m.AcceptedEntries), Published: result.Tournament.Published,
		Metrics: OrganizerDashboardMetricsResponse{
			TotalEntries: m.TotalEntries, FormingEntries: m.FormingEntries, LockedEntries: m.LockedEntries,
			AcceptedEntries: m.AcceptedEntries, RejectedEntries: m.RejectedEntries,
			ConfirmedParticipants: m.ConfirmedParticipants, AvailableSpots: m.AvailableSpots,
		},
		Funding: TournamentFundingResponse{
			GoalAmount: result.Funding.GoalAmount, RaisedAmount: result.Funding.RaisedAmount,
			RemainingAmount: result.Funding.RemainingAmount, SupporterCount: result.Funding.SupporterCount,
			Percentage: result.Funding.Percentage, Currency: result.Tournament.Currency,
		},
	}
}
