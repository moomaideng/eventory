package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	sharedmiddlewares "github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
)

type organizerContextKey string

const (
	// OrganizerProfileContextKey stores the authenticated organizer's profile in the context.
	OrganizerProfileContextKey organizerContextKey = "authenticated_organizer_profile"
)

var ErrOrganizerProfileRequired = errors.New("forbidden: organizer profile required")

// OrganizerMiddleware enforces that the authenticated user possesses an organizer profile.
type OrganizerMiddleware struct {
	api        huma.API
	accountSvc ports.AccountService
}

// NewOrganizerMiddleware creates a new OrganizerMiddleware instance.
func NewOrganizerMiddleware(api huma.API, accountSvc ports.AccountService) *OrganizerMiddleware {
	return &OrganizerMiddleware{api: api, accountSvc: accountSvc}
}

// HumaMiddleware returns the Huma middleware handler.
func (m *OrganizerMiddleware) HumaMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		accountID, err := sharedmiddlewares.GetAuthUserID(ctx.Context())
		if err != nil {
			_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Missing Authorization header", err)
			return
		}

		profile, err := m.accountSvc.GetOrganizerProfileByAccountID(ctx.Context(), accountID)
		if err != nil {
			if errors.Is(err, ports.ErrOrganizerNotFound) {
				_ = huma.WriteErr(m.api, ctx, http.StatusForbidden, "Organizer profile required to perform this action", ErrOrganizerProfileRequired)
				return
			}
			_ = huma.WriteErr(m.api, ctx, http.StatusInternalServerError, "Failed to verify organizer status", err)
			return
		}

		newCtx := context.WithValue(ctx.Context(), OrganizerProfileContextKey, profile)
		ctx = huma.WithContext(ctx, newCtx)
		next(ctx)
	}
}

// GetOrganizerProfile extracts the authenticated organizer profile from the context.
func GetOrganizerProfile(ctx context.Context) (*ports.OrganizerInfo, error) {
	val := ctx.Value(OrganizerProfileContextKey)
	if val == nil {
		return nil, ErrOrganizerProfileRequired
	}
	profile, ok := val.(*ports.OrganizerInfo)
	if !ok || profile == nil {
		return nil, ErrOrganizerProfileRequired
	}
	return profile, nil
}
