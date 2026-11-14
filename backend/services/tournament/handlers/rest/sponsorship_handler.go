package rest

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type SponsorResponse struct {
	ID          uuid.UUID `json:"id"`
	SponsorName string    `json:"sponsorName"`
	TierName    string    `json:"tierName"`
	WebsiteURL  string    `json:"websiteUrl"`
	LogoURL     string    `json:"logoUrl"`
}

type SponsorshipResponse struct {
	ID           uuid.UUID           `json:"id"`
	TournamentID uuid.UUID           `json:"tournamentId"`
	Status       models.PledgeStatus `json:"status" enum:"PENDING,CONFIRMED,CANCELLED,EXPIRED"`
	PaymentID    *uuid.UUID          `json:"paymentId,omitempty"`
}

type OwnedSponsorshipResponse struct {
	SponsorshipResponse
	TournamentName string `json:"tournamentName"`
	TierName       string `json:"tierName"`
	AmountMinor    int64  `json:"amountMinor" doc:"Pledge amount in minor currency units (satang; 100 satang = 1 THB)"`
	Currency       string `json:"currency"`
}

type StartSponsorshipInput struct {
	TournamentID uuid.UUID `path:"tournamentId"`
	TierID       uuid.UUID `path:"tierId"`
	Body         struct {
		RequestID uuid.UUID `json:"requestId"`
	}
}

type SponsorshipIDInput struct {
	ID uuid.UUID `path:"id"`
}

type ListMySponsorshipsInput struct {
	Page  int `query:"page" default:"1" minimum:"1"`
	Limit int `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type SponsorshipOutput struct {
	Body SponsorshipResponse
}

func RegisterSponsorshipRoutes(api huma.API, auth func(huma.Context, func(huma.Context)), sponsorships *usecases.SponsorshipUseCase) {
	huma.Register(api, huma.Operation{
		OperationID: "list-my-sponsorships", Method: http.MethodGet,
		Path: "/api/v1/tournaments/sponsorships/mine", Summary: "List sponsorships owned by the caller", Tags: []string{"Sponsorship"},
		Middlewares: huma.Middlewares{auth}, Security: []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *ListMySponsorshipsInput) (*struct {
		Body struct {
			Items []OwnedSponsorshipResponse `json:"items"`
			Total int64                      `json:"total"`
			Page  int                        `json:"page"`
			Limit int                        `json:"limit"`
		}
	}, error) {
		payer, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		items, total, err := sponsorships.ListOwned(ctx, payer, input.Page, input.Limit)
		if err != nil {
			return nil, sponsorshipHTTPError(err)
		}
		out := &struct {
			Body struct {
				Items []OwnedSponsorshipResponse `json:"items"`
				Total int64                      `json:"total"`
				Page  int                        `json:"page"`
				Limit int                        `json:"limit"`
			}
		}{}
		out.Body.Items = make([]OwnedSponsorshipResponse, 0, len(items))
		for _, p := range items {
			out.Body.Items = append(out.Body.Items, OwnedSponsorshipResponse{
				SponsorshipResponse: SponsorshipResponse{ID: p.ID, TournamentID: p.TournamentID, Status: p.Status, PaymentID: p.PaymentID},
				TournamentName:      p.TournamentName, TierName: p.TierName, AmountMinor: p.AmountMinor, Currency: p.Currency,
			})
		}
		out.Body.Total, out.Body.Page, out.Body.Limit = total, input.Page, input.Limit
		return out, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "start-sponsorship", Method: http.MethodPost,
		Path: "/api/v1/tournaments/{tournamentId}/sponsorship-tiers/{tierId}/checkout",
		Summary: "Start or resume a sponsorship checkout", Tags: []string{"Sponsorship"},
		Middlewares: huma.Middlewares{auth}, Security: []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *StartSponsorshipInput) (*SponsorshipOutput, error) {
		payer, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		pledge, err := sponsorships.Start(ctx, usecases.StartSponsorship{
			RequestID: input.Body.RequestID, TournamentID: input.TournamentID, TierID: input.TierID, PayerID: payer,
		})
		if err != nil {
			return nil, sponsorshipHTTPError(err)
		}
		return &SponsorshipOutput{Body: sponsorshipResponse(pledge)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "refresh-my-sponsorship", Method: http.MethodPost,
		Path: "/api/v1/tournaments/sponsorships/{id}/refresh", Summary: "Verify an owned sponsorship payment", Tags: []string{"Sponsorship"},
		Middlewares: huma.Middlewares{auth}, Security: []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *SponsorshipIDInput) (*SponsorshipOutput, error) {
		payer, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		pledge, err := sponsorships.GetOwned(ctx, payer, input.ID)
		if err != nil {
			return nil, sponsorshipHTTPError(err)
		}
		return &SponsorshipOutput{Body: sponsorshipResponse(pledge)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-tournament-sponsors", Method: http.MethodGet,
		Path: "/api/v1/tournaments/{tournamentId}/sponsors", Summary: "List confirmed sponsors of a published tournament", Tags: []string{"Sponsorship"},
	}, func(ctx context.Context, input *ListSponsorshipTiersInput) (*struct {
		Body struct {
			Items []SponsorResponse `json:"items"`
		}
	}, error) {
		sponsors, err := sponsorships.ListPublic(ctx, input.TournamentID)
		if err != nil {
			return nil, sponsorshipHTTPError(err)
		}
		out := &struct {
			Body struct {
				Items []SponsorResponse `json:"items"`
			}
		}{}
		out.Body.Items = make([]SponsorResponse, 0, len(sponsors))
		for _, s := range sponsors {
			out.Body.Items = append(out.Body.Items, SponsorResponse{
				ID: s.ID, SponsorName: s.SponsorName, TierName: s.TierName, WebsiteURL: s.WebsiteURL, LogoURL: s.LogoURL,
			})
		}
		return out, nil
	})
}

func sponsorshipResponse(p *models.SponsorshipPledge) SponsorshipResponse {
	return SponsorshipResponse{ID: p.ID, TournamentID: p.TournamentID, Status: p.Status, PaymentID: p.PaymentID}
}

func sponsorshipHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, usecases.ErrSponsorshipNotFound):
		return huma.Error404NotFound(err.Error())
	case errors.Is(err, usecases.ErrTournamentNotFound):
		return huma.Error404NotFound("Tournament not found")
	case errors.Is(err, usecases.ErrSponsorshipTierNotFound):
		return huma.Error404NotFound("Sponsorship package not found")
	case errors.Is(err, usecases.ErrSponsorRequired):
		return huma.Error403Forbidden(err.Error())
	case errors.Is(err, usecases.ErrTierSoldOut), errors.Is(err, usecases.ErrAlreadySponsored), errors.Is(err, usecases.ErrSponsorshipClosed):
		return huma.Error409Conflict(err.Error())
	case errors.Is(err, usecases.ErrInvalidSponsorship):
		return huma.Error422UnprocessableEntity(err.Error())
	case errors.Is(err, usecases.ErrPaymentUnavailable):
		return huma.Error503ServiceUnavailable("Payment could not be verified. Retry to resume the same checkout.")
	default:
		slog.Error("sponsorship operation failed", "error", err)
		return huma.Error500InternalServerError("Unable to process sponsorship")
	}
}
