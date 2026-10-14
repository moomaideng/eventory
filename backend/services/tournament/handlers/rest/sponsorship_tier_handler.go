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

type SponsorshipTierResponse struct {
	ID           uuid.UUID `json:"id"`
	TournamentID uuid.UUID `json:"tournamentId"`
	Name         string    `json:"name"`
	Price        int64     `json:"price" doc:"Price in whole currency units"`
	Slots        int       `json:"slots" doc:"How many sponsors can buy this tier"`
	Perks        []string  `json:"perks"`
	SortOrder    int       `json:"sortOrder"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type SponsorshipTierBody struct {
	Name      string   `json:"name" minLength:"1" maxLength:"80" doc:"Tier name, e.g. Gold"`
	Price     int64    `json:"price" minimum:"1" doc:"Price in whole currency units"`
	Slots     int      `json:"slots" minimum:"1" doc:"How many sponsors can buy this tier"`
	Perks     []string `json:"perks" minItems:"1" maxItems:"10" doc:"What the sponsor receives"`
	SortOrder int      `json:"sortOrder,omitempty" default:"0" doc:"Display order, lowest first"`
}

type ListSponsorshipTiersInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
}

type ListSponsorshipTiersOutput struct {
	Body struct {
		Items []SponsorshipTierResponse `json:"items"`
	}
}

type CreateSponsorshipTierInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	Body         SponsorshipTierBody
}

type UpdateSponsorshipTierInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	TierID       uuid.UUID `path:"tierId" doc:"Sponsorship tier UUID"`
	Body         SponsorshipTierBody
}

type DeleteSponsorshipTierInput struct {
	TournamentID uuid.UUID `path:"tournamentId" doc:"Tournament UUID"`
	TierID       uuid.UUID `path:"tierId" doc:"Sponsorship tier UUID"`
}

type SponsorshipTierOutput struct {
	Body SponsorshipTierResponse
}

func RegisterSponsorshipTierRoutes(api huma.API, organizerTournamentGroup huma.API, tiers *usecases.SponsorshipTierUseCase) {
	huma.Register(api, huma.Operation{
		OperationID: "list-sponsorship-tiers",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments/{tournamentId}/sponsorship-tiers",
		Summary:     "List a tournament's sponsorship tiers",
		Description: "Returns the sponsorship packages of a published tournament in display order.",
		Tags:        []string{"Sponsorship"},
	}, func(ctx context.Context, input *ListSponsorshipTiersInput) (*ListSponsorshipTiersOutput, error) {
		items, err := tiers.ListPublic(ctx, input.TournamentID)
		if err != nil {
			return nil, sponsorshipTierHTTPError(err)
		}
		output := &ListSponsorshipTiersOutput{}
		output.Body.Items = make([]SponsorshipTierResponse, 0, len(items))
		for _, tier := range items {
			output.Body.Items = append(output.Body.Items, toSponsorshipTierResponse(tier))
		}
		return output, nil
	})

	huma.Register(organizerTournamentGroup, huma.Operation{
		OperationID:   "create-sponsorship-tier",
		Method:        http.MethodPost,
		Path:          "/{tournamentId}/sponsorship-tiers",
		Summary:       "Create a sponsorship tier",
		Description:   "Adds a sponsorship package to an owned tournament.",
		Tags:          []string{"Sponsorship"},
		Security:      []map[string][]string{{"bearer": {}}},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, input *CreateSponsorshipTierInput) (*SponsorshipTierOutput, error) {
		profile, err := middlewares.GetOrganizerProfile(ctx)
		if err != nil {
			return nil, sponsorshipTierHTTPError(err)
		}
		tier, err := tiers.Create(ctx, profile.AccountID, input.TournamentID, toSponsorshipTierInput(input.Body))
		if err != nil {
			return nil, sponsorshipTierHTTPError(err)
		}
		return &SponsorshipTierOutput{Body: toSponsorshipTierResponse(*tier)}, nil
	})

	huma.Register(organizerTournamentGroup, huma.Operation{
		OperationID: "update-sponsorship-tier",
		Method:      http.MethodPut,
		Path:        "/{tournamentId}/sponsorship-tiers/{tierId}",
		Summary:     "Update a sponsorship tier",
		Description: "Replaces the details of a sponsorship package on an owned tournament.",
		Tags:        []string{"Sponsorship"},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *UpdateSponsorshipTierInput) (*SponsorshipTierOutput, error) {
		profile, err := middlewares.GetOrganizerProfile(ctx)
		if err != nil {
			return nil, sponsorshipTierHTTPError(err)
		}
		tier, err := tiers.Update(ctx, profile.AccountID, input.TournamentID, input.TierID, toSponsorshipTierInput(input.Body))
		if err != nil {
			return nil, sponsorshipTierHTTPError(err)
		}
		return &SponsorshipTierOutput{Body: toSponsorshipTierResponse(*tier)}, nil
	})

	huma.Register(organizerTournamentGroup, huma.Operation{
		OperationID:   "delete-sponsorship-tier",
		Method:        http.MethodDelete,
		Path:          "/{tournamentId}/sponsorship-tiers/{tierId}",
		Summary:       "Delete a sponsorship tier",
		Description:   "Removes a sponsorship package from an owned tournament.",
		Tags:          []string{"Sponsorship"},
		Security:      []map[string][]string{{"bearer": {}}},
		DefaultStatus: http.StatusNoContent,
	}, func(ctx context.Context, input *DeleteSponsorshipTierInput) (*struct{}, error) {
		profile, err := middlewares.GetOrganizerProfile(ctx)
		if err != nil {
			return nil, sponsorshipTierHTTPError(err)
		}
		if err := tiers.Delete(ctx, profile.AccountID, input.TournamentID, input.TierID); err != nil {
			return nil, sponsorshipTierHTTPError(err)
		}
		return nil, nil
	})
}

func toSponsorshipTierInput(body SponsorshipTierBody) usecases.SponsorshipTierInput {
	return usecases.SponsorshipTierInput{
		Name:      body.Name,
		Price:     body.Price,
		Slots:     body.Slots,
		Perks:     body.Perks,
		SortOrder: body.SortOrder,
	}
}

func toSponsorshipTierResponse(tier models.SponsorshipTier) SponsorshipTierResponse {
	return SponsorshipTierResponse{
		ID:           tier.ID,
		TournamentID: tier.TournamentID,
		Name:         tier.Name,
		Price:        tier.Price,
		Slots:        tier.Slots,
		Perks:        tier.Perks,
		SortOrder:    tier.SortOrder,
		CreatedAt:    tier.CreatedAt,
		UpdatedAt:    tier.UpdatedAt,
	}
}

func sponsorshipTierHTTPError(err error) error {
	switch {
	case errors.Is(err, middlewares.ErrOrganizerProfileRequired):
		return huma.Error403Forbidden("Organizer profile required to perform this action", err)
	case errors.Is(err, usecases.ErrTournamentNotFound):
		// 404, not 403, so non-owners cannot confirm the tournament exists.
		return huma.Error404NotFound("Tournament not found or not owned by your account")
	case errors.Is(err, usecases.ErrSponsorshipTierNotFound):
		return huma.Error404NotFound("Sponsorship tier not found")
	case errors.Is(err, usecases.ErrSponsorshipTierNameDuplicate):
		return huma.Error409Conflict(err.Error())
	case errors.Is(err, usecases.ErrSponsorshipClosed):
		return huma.Error409Conflict(err.Error())
	case errors.Is(err, usecases.ErrInvalidSponsorshipTierName),
		errors.Is(err, usecases.ErrInvalidSponsorshipTierPrice),
		errors.Is(err, usecases.ErrInvalidSponsorshipTierSlots),
		errors.Is(err, usecases.ErrInvalidSponsorshipTierPerks):
		return huma.Error422UnprocessableEntity(err.Error())
	default:
		return huma.Error500InternalServerError("Unable to process sponsorship tier request")
	}
}
