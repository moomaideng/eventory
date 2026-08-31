package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

type contextKey string

const (
	// UserEmailContextKey stores the authenticated user's email address in the context.
	UserEmailContextKey contextKey = "authenticated_user_email"
	// DefaultDevEmail is used as explicit fallback for zero-friction local offline development.
	DefaultDevEmail string = "dev@eventory.gg"
)

var (
	ErrUnauthorized = errors.New("unauthorized: valid bearer token required")
)

// AuthMiddleware manages authentication verification with Supabase Auth API and explicit local dev fallbacks.
type AuthMiddleware struct {
	api         huma.API
	supabaseURL string
	environment string
}

// NewAuthMiddleware creates a new instance of AuthMiddleware with Huma API reference.
func NewAuthMiddleware(api huma.API, supabaseURL string, environment string) *AuthMiddleware {
	return &AuthMiddleware{
		api:         api,
		supabaseURL: strings.TrimSpace(supabaseURL),
		environment: strings.ToLower(strings.TrimSpace(environment)),
	}
}

// HumaMiddleware returns a Huma middleware function that enforces authentication.
func (m *AuthMiddleware) HumaMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := ctx.Header("Authorization")
		devHeader := ctx.Header("X-Dev-Email")

		// 1. Explicit Local Offline Development Fallback
		// If running in development without configured Supabase URL, or using dev token / dev header
		if m.isDevMode() && (authHeader == "Bearer dev-token" || devHeader != "" || m.supabaseURL == "") {
			devEmail := DefaultDevEmail
			if devHeader != "" {
				devEmail = strings.ToLower(strings.TrimSpace(devHeader))
			}
			// Explicitly inject dev email into context
			newCtx := context.WithValue(ctx.Context(), UserEmailContextKey, devEmail)
			ctx = huma.WithContext(ctx, newCtx)
			next(ctx)
			return
		}

		// 2. Validate Bearer Token Header presence
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Missing or invalid Authorization header", ErrUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)

		if tokenString == "" {
			_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Bearer token is empty", ErrUnauthorized)
			return
		}

		// 3. Verify Token via Supabase Auth API Endpoint (/auth/v1/user)
		if m.supabaseURL != "" {
			email, err := m.verifyWithSupabaseAPI(tokenString)
			if err != nil {
				_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Failed to authenticate with Supabase Auth API", err)
				return
			}
			newCtx := context.WithValue(ctx.Context(), UserEmailContextKey, email)
			ctx = huma.WithContext(ctx, newCtx)
			next(ctx)
			return
		}

		// Fallback: If no verification method configured and not in dev mode, deny access
		_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Authentication service unconfigured", ErrUnauthorized)
	}
}

// verifyWithSupabaseAPI queries the Supabase Auth /auth/v1/user endpoint to validate token.
func (m *AuthMiddleware) verifyWithSupabaseAPI(tokenString string) (string, error) {
	reqURL := fmt.Sprintf("%s/auth/v1/user", strings.TrimRight(m.supabaseURL, "/"))
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+tokenString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("supabase returned status: %d", resp.StatusCode)
	}

	var userResponse struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return "", err
	}

	if userResponse.Email == "" {
		return "", errors.New("email not found in Supabase user response")
	}

	return strings.ToLower(strings.TrimSpace(userResponse.Email)), nil
}

// isDevMode checks if the current environment is local development.
func (m *AuthMiddleware) isDevMode() bool {
	return m.environment == "development" || m.environment == "local" || m.environment == ""
}

// GetAuthEmail extracts the authenticated user's email from request context.
func GetAuthEmail(ctx context.Context) (string, error) {
	val := ctx.Value(UserEmailContextKey)
	if val == nil {
		return "", ErrUnauthorized
	}
	email, ok := val.(string)
	if !ok || email == "" {
		return "", ErrUnauthorized
	}
	return email, nil
}
