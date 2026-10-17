package rest

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type SponsorResponse struct {
	ID          uuid.UUID `json:"id"`
	SponsorName string    `json:"sponsorName"`
	TierName    string    `json:"tierName"`
	WebsiteURL  string    `json:"websiteUrl"`
	LogoPath    string    `json:"logoPath"`
}

type SponsorshipResponse struct {
	ID           uuid.UUID           `json:"id"`
	TournamentID uuid.UUID           `json:"tournamentId"`
	Status       models.PledgeStatus `json:"status" enum:"PENDING,CONFIRMED,CANCELLED,EXPIRED"`
	PaymentID    *uuid.UUID          `json:"paymentId,omitempty"`
	ExpiresAt    time.Time           `json:"expiresAt"`
}

type SponsorshipOutput struct{ Body SponsorshipResponse }
type OwnedSponsorshipResponse struct {
	SponsorshipResponse
	TournamentName string `json:"tournamentName"`
	TierName       string `json:"tierName"`
	AmountMinor    int64  `json:"amountMinor"`
	Currency       string `json:"currency"`
}
type OwnedSponsorshipsOutput struct {
	Body struct {
		Items []OwnedSponsorshipResponse `json:"items"`
		Total int64                      `json:"total"`
	}
}
type SponsorshipIDInput struct {
	ID uuid.UUID `path:"id"`
}
type SponsorshipUpload struct {
	RequestID  string        `form:"requestId" minLength:"36" maxLength:"36"`
	WebsiteURL string        `form:"websiteUrl" minLength:"1" maxLength:"2048"`
	Logo       huma.FormFile `form:"logo" contentType:"image/png,image/jpeg" required:"true"`
}

func RegisterSponsorshipRoutes(api huma.API, auth func(huma.Context, func(huma.Context)), sponsorships *usecases.SponsorshipUseCase) {
	huma.Register(api, huma.Operation{
		OperationID: "list-my-sponsorships", Method: http.MethodGet,
		Path: "/api/v1/tournaments/sponsorships/mine", Summary: "List the signed-in account's sponsorships", Tags: []string{"Sponsorship"},
		Middlewares: huma.Middlewares{auth}, Security: []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *struct {
		Page  int `query:"page" default:"1" minimum:"1" maximum:"1000000"`
		Limit int `query:"limit" default:"20" minimum:"1" maximum:"100"`
	}) (*OwnedSponsorshipsOutput, error) {
		payer, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		items, total, err := sponsorships.ListOwned(ctx, payer, input.Page, input.Limit)
		if err != nil {
			return nil, sponsorshipHTTPError(err)
		}
		out := &OwnedSponsorshipsOutput{}
		out.Body.Total = total
		out.Body.Items = make([]OwnedSponsorshipResponse, 0, len(items))
		for _, p := range items {
			out.Body.Items = append(out.Body.Items, OwnedSponsorshipResponse{
				SponsorshipResponse: sponsorshipResponse(&p.SponsorshipPledge), TournamentName: p.TournamentName,
				TierName: p.TierName, AmountMinor: p.AmountMinor, Currency: p.Currency,
			})
		}
		return out, nil
	})
	huma.Register(api, huma.Operation{
		OperationID: "start-sponsorship-checkout", Method: http.MethodPost,
		Path:    "/api/v1/tournaments/{tournamentId}/sponsorship-tiers/{tierId}/checkout",
		Summary: "Reserve a sponsorship slot with brand assets and start checkout", Tags: []string{"Sponsorship"},
		Middlewares: huma.Middlewares{auth}, Security: []map[string][]string{{"bearer": {}}},
		MaxBodyBytes: usecases.MaxSponsorLogoBytes + 16384,
	}, func(ctx context.Context, input *struct {
		TournamentID uuid.UUID `path:"tournamentId"`
		TierID       uuid.UUID `path:"tierId"`
		RawBody      huma.MultipartFormFiles[SponsorshipUpload]
	}) (*SponsorshipOutput, error) {
		payer, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}
		data := input.RawBody.Data()
		defer data.Logo.Close()
		defer input.RawBody.Form.RemoveAll()
		id, err := uuid.Parse(data.RequestID)
		if err != nil {
			return nil, huma.Error422UnprocessableEntity("Invalid checkout request ID")
		}
		logo, err := io.ReadAll(io.LimitReader(data.Logo, usecases.MaxSponsorLogoBytes+1))
		if err != nil {
			return nil, huma.Error400BadRequest("Unable to read logo")
		}
		pledge, err := sponsorships.Start(ctx, usecases.StartSponsorship{RequestID: id, TournamentID: input.TournamentID, TierID: input.TierID, PayerID: payer, WebsiteURL: data.WebsiteURL, Logo: logo})
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
		pledges, err := sponsorships.ListPublic(ctx, input.TournamentID)
		if err != nil {
			return nil, sponsorshipHTTPError(err)
		}
		out := &struct {
			Body struct {
				Items []SponsorResponse `json:"items"`
			}
		}{}
		out.Body.Items = make([]SponsorResponse, 0, len(pledges))
		for _, p := range pledges {
			out.Body.Items = append(out.Body.Items, SponsorResponse{ID: p.ID, SponsorName: p.SponsorName, TierName: p.TierName, WebsiteURL: p.WebsiteURL, LogoPath: "/api/v1/tournaments/sponsorships/" + p.ID.String() + "/logo"})
		}
		return out, nil
	})
	huma.Register(api, huma.Operation{
		OperationID: "get-sponsor-logo", Method: http.MethodGet,
		Path: "/api/v1/tournaments/sponsorships/{id}/logo", Summary: "Get a confirmed sponsor's logo", Tags: []string{"Sponsorship"},
		Responses: map[string]*huma.Response{"200": {Description: "Sponsor logo", Content: map[string]*huma.MediaType{"image/png": {Schema: &huma.Schema{Type: "string", Format: "binary"}}, "image/jpeg": {Schema: &huma.Schema{Type: "string", Format: "binary"}}}}},
	}, func(ctx context.Context, input *SponsorshipIDInput) (*struct {
		ContentType  string `header:"Content-Type"`
		CacheControl string `header:"Cache-Control"`
		NoSniff      string `header:"X-Content-Type-Options"`
		Body         []byte
	}, error) {
		logo, err := sponsorships.Logo(ctx, input.ID)
		if err != nil {
			return nil, sponsorshipHTTPError(err)
		}
		return &struct {
			ContentType  string `header:"Content-Type"`
			CacheControl string `header:"Cache-Control"`
			NoSniff      string `header:"X-Content-Type-Options"`
			Body         []byte
		}{logo.ContentType, "public, max-age=86400", "nosniff", logo.Data}, nil
	})
}

func sponsorshipResponse(p *models.SponsorshipPledge) SponsorshipResponse {
	return SponsorshipResponse{ID: p.ID, TournamentID: p.TournamentID, Status: p.Status, PaymentID: p.PaymentID, ExpiresAt: p.ExpiresAt}
}

func sponsorshipHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, repositories.ErrPledgeNotFound):
		return huma.Error404NotFound(err.Error())
	case errors.Is(err, repositories.ErrTournamentNotFound), errors.Is(err, usecases.ErrTournamentNotFound):
		return huma.Error404NotFound("Tournament not found")
	case errors.Is(err, repositories.ErrSponsorshipTierNotFound):
		return huma.Error404NotFound("Sponsorship package not found")
	case errors.Is(err, usecases.ErrSponsorRequired):
		return huma.Error403Forbidden(err.Error())
	case errors.Is(err, usecases.ErrTierSoldOut), errors.Is(err, usecases.ErrAlreadySponsored), errors.Is(err, usecases.ErrSponsorshipClosed):
		return huma.Error409Conflict(err.Error())
	case errors.Is(err, usecases.ErrInvalidSponsorBrand), errors.Is(err, usecases.ErrInvalidSponsorship):
		return huma.Error422UnprocessableEntity(err.Error())
	case errors.Is(err, usecases.ErrPaymentUnavailable):
		return huma.Error503ServiceUnavailable("Payment could not be verified. Retry to resume the same checkout.")
	default:
		slog.Error("sponsorship operation failed", "error", err)
		return huma.Error500InternalServerError("Unable to process sponsorship")
	}
}
