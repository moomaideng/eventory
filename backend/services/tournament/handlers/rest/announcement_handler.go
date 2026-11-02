package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/usecases"
	"github.com/moomaideng/eventory/services/tournament/models"
)

type CreateAnnouncementInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
	Body         struct {
		Title string `json:"title" minLength:"1" maxLength:"150" doc:"Announcement title"`
		Body  string `json:"body" minLength:"1" maxLength:"5000" doc:"Announcement message (plain text)"`
	}
}

type ListAnnouncementsInput struct {
	TournamentID uuid.UUID `path:"id" doc:"Tournament ID"`
}

// AnnouncementResponse is what the public sees. The author's account ID is
// deliberately left out: only the display name is needed to show who posted.
type AnnouncementResponse struct {
	ID           uuid.UUID `json:"id"`
	TournamentID uuid.UUID `json:"tournamentId"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	AuthorName   string    `json:"authorName"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CreateAnnouncementOutput struct {
	Body AnnouncementResponse
}

type ListAnnouncementsOutput struct {
	Body struct {
		Items []AnnouncementResponse `json:"items"`
	}
}

func RegisterAnnouncementRoutes(
	api huma.API,
	auth func(huma.Context, func(huma.Context)),
	announcements *usecases.AnnouncementUseCase,
) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-announcement",
		Method:        http.MethodPost,
		Path:          "/api/v1/tournaments/{id}/announcements",
		DefaultStatus: http.StatusCreated,
		Summary:       "Post a tournament announcement",
		Description: "Posts an announcement (for example a venue change) to the " +
			"tournament's announcement list. Only the tournament's owner or an admin may post.",
		Tags:        []string{"Announcements"},
		Middlewares: huma.Middlewares{auth},
		Security:    []map[string][]string{{"bearer": {}}},
	}, func(ctx context.Context, input *CreateAnnouncementInput) (*CreateAnnouncementOutput, error) {
		accountID, err := authenticatedAccountID(ctx)
		if err != nil {
			return nil, err
		}

		announcement, err := announcements.Create(ctx, accountID, input.TournamentID,
			input.Body.Title, input.Body.Body)
		if err != nil {
			return nil, announcementHTTPError(err)
		}

		return &CreateAnnouncementOutput{Body: toAnnouncementResponse(*announcement)}, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "list-announcements",
		Method:      http.MethodGet,
		Path:        "/api/v1/tournaments/{id}/announcements",
		Summary:     "List tournament announcements",
		Description: "Returns a published tournament's announcements, newest first. " +
			"No sign-in is required. A draft or missing tournament returns 404.",
		Tags: []string{"Announcements"},
	}, func(ctx context.Context, input *ListAnnouncementsInput) (*ListAnnouncementsOutput, error) {
		items, err := announcements.List(ctx, input.TournamentID)
		if err != nil {
			return nil, announcementHTTPError(err)
		}

		output := &ListAnnouncementsOutput{}
		output.Body.Items = make([]AnnouncementResponse, 0, len(items))
		for _, item := range items {
			output.Body.Items = append(output.Body.Items, toAnnouncementResponse(item))
		}
		return output, nil
	})
}

func toAnnouncementResponse(a models.TournamentAnnouncement) AnnouncementResponse {
	return AnnouncementResponse{
		ID:           a.ID,
		TournamentID: a.TournamentID,
		Title:        a.Title,
		Body:         a.Body,
		AuthorName:   a.AuthorDisplayName,
		CreatedAt:    a.CreatedAt,
	}
}

func announcementHTTPError(err error) error {
	if mapped := mapAccountPortError(err); mapped != nil {
		return mapped
	}
	switch {
	case errors.Is(err, usecases.ErrTournamentNotFound):
		// A tournament the caller cannot post to, a draft, and one that does
		// not exist all look the same: 404, never 403.
		return huma.Error404NotFound("Tournament not found")
	case errors.Is(err, usecases.ErrAnnouncementTitleRequired):
		return huma.Error422UnprocessableEntity("Title is required")
	case errors.Is(err, usecases.ErrAnnouncementTitleTooLong):
		return huma.Error422UnprocessableEntity("Title must be at most 150 characters")
	case errors.Is(err, usecases.ErrAnnouncementBodyRequired):
		return huma.Error422UnprocessableEntity("Message is required")
	case errors.Is(err, usecases.ErrAnnouncementBodyTooLong):
		return huma.Error422UnprocessableEntity("Message must be at most 5000 characters")
	default:
		return huma.Error500InternalServerError("Unable to process the announcement request", err)
	}
}
