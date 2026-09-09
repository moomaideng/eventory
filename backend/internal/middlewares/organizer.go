package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
)

type organizerContextKey string

const (
	// OrganizerProfileContextKey stores the authenticated organizer's profile in the context.
	OrganizerProfileContextKey organizerContextKey = "authenticated_organizer_profile"
)

var (
	ErrOrganizerProfileRequired = errors.New("forbidden: organizer profile required")
)

// OrganizerMiddleware enforces that the authenticated user possesses an organizer profile.
type OrganizerMiddleware struct {
	api         huma.API
	accountRepo repositories.AccountRepository
}

// NewOrganizerMiddleware creates a new OrganizerMiddleware instance.
func NewOrganizerMiddleware(api huma.API, accountRepo repositories.AccountRepository) *OrganizerMiddleware {
	return &OrganizerMiddleware{
		api:         api,
		accountRepo: accountRepo,
	}
}

// HumaMiddleware returns the Huma middleware handler.
func (m *OrganizerMiddleware) HumaMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		var profile *models.OrganizerProfile
		var err error

		if userID, uErr := GetAuthUserID(ctx.Context()); uErr == nil && userID != uuid.Nil {
			profile, err = m.accountRepo.FindOrganizerProfileByAccountID(ctx.Context(), userID)
			if err != nil {
				_ = huma.WriteErr(m.api, ctx, http.StatusInternalServerError, "Failed to verify organizer status", err)
				return
			}
		}

		if profile == nil {
			if email, eErr := GetAuthEmail(ctx.Context()); eErr == nil && email != "" {
				if acc, aErr := m.accountRepo.FindByEmail(ctx.Context(), email); aErr == nil && acc != nil {
					profile, err = m.accountRepo.FindOrganizerProfileByAccountID(ctx.Context(), acc.ID)
					if err != nil {
						_ = huma.WriteErr(m.api, ctx, http.StatusInternalServerError, "Failed to verify organizer status", err)
						return
					}
				}
			}
		}

		if profile == nil {
			_ = huma.WriteErr(m.api, ctx, http.StatusForbidden, "Organizer profile required to perform this action", ErrOrganizerProfileRequired)
			return
		}

		newCtx := context.WithValue(ctx.Context(), OrganizerProfileContextKey, profile)
		ctx = huma.WithContext(ctx, newCtx)
		next(ctx)
	}
}

// GetOrganizerProfile extracts the authenticated organizer profile from the context.
func GetOrganizerProfile(ctx context.Context) (*models.OrganizerProfile, error) {
	val := ctx.Value(OrganizerProfileContextKey)
	if val == nil {
		return nil, ErrOrganizerProfileRequired
	}
	profile, ok := val.(*models.OrganizerProfile)
	if !ok || profile == nil {
		return nil, ErrOrganizerProfileRequired
	}
	return profile, nil
}
