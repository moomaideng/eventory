package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	// UserEmailContextKey stores the authenticated user's email in the context.
	UserEmailContextKey contextKey = "authenticated_user_email"
	// UserSubContextKey stores the user's UUID (sub claim) in the context.
	UserSubContextKey contextKey = "authenticated_user_sub"

	// DevToken represents the explicit token string used for local development.
	DevToken string = "dev-token"
	// DevEmail represents the mock email assigned when using DevToken.
	DevEmail string = "dev@eventory.gg"
	// DevSub represents the mock user ID assigned when using DevToken.
	DevSub string = "dev-user-001"
)

var (
	ErrUnauthorized = errors.New("unauthorized: valid bearer token required")
	ErrTokenExpired = errors.New("unauthorized: token has expired")
	ErrInvalidToken = errors.New("unauthorized: invalid cryptographic signature or claims")
)

// AuthMiddleware manages cryptographic verification of Supabase JWTs via JWKS.
type AuthMiddleware struct {
	api         huma.API
	jwks        keyfunc.Keyfunc
	supabaseURL string
	environment string
}

// NewAuthMiddleware initializes JWKS key fetching from Supabase and creates AuthMiddleware.
func NewAuthMiddleware(api huma.API, supabaseURL string, environment string) *AuthMiddleware {
	supabaseURL = strings.TrimRight(strings.TrimSpace(supabaseURL), "/")
	var jwks keyfunc.Keyfunc

	if supabaseURL != "" && !strings.Contains(supabaseURL, "your-project") {
		jwksURL := fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", supabaseURL)
		var err error
		jwks, err = keyfunc.NewDefault([]string{jwksURL})
		if err != nil {
			log.Printf("[WARN] Failed to initialize Supabase JWKS from %s: %v", jwksURL, err)
		} else {
			log.Printf("[INFO] Supabase JWKS initialized successfully from %s", jwksURL)
		}
	}

	return &AuthMiddleware{
		api:         api,
		jwks:        jwks,
		supabaseURL: supabaseURL,
		environment: strings.ToLower(strings.TrimSpace(environment)),
	}
}

// isDevMode checks if the current environment is development or local.
func (m *AuthMiddleware) isDevMode() bool {
	return m.environment == "development" || m.environment == "local" || m.environment == ""
}

// HumaMiddleware returns the Huma middleware handler.
func (m *AuthMiddleware) HumaMiddleware() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := ctx.Header("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Missing Authorization header", ErrUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" {
			_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "Bearer token is empty", ErrUnauthorized)
			return
		}

		// 1. Explicit Dev Token Check (Permitted ONLY in development/local environment)
		if m.isDevMode() && tokenString == DevToken {
			newCtx := context.WithValue(ctx.Context(), UserEmailContextKey, DevEmail)
			newCtx = context.WithValue(newCtx, UserSubContextKey, DevSub)
			ctx = huma.WithContext(ctx, newCtx)
			next(ctx)
			return
		}

		// 2. Cryptographic JWT Verification (ES256 / RS256 with Supabase JWKS)
		email, sub, err := m.verifyToken(tokenString)
		if err != nil {
			_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, err.Error(), err)
			return
		}

		// 3. Inject verified identity into context
		newCtx := context.WithValue(ctx.Context(), UserEmailContextKey, email)
		newCtx = context.WithValue(newCtx, UserSubContextKey, sub)
		ctx = huma.WithContext(ctx, newCtx)
		next(ctx)
	}
}

// verifyToken cryptographically verifies the token signature against Supabase JWKS public keys.
func (m *AuthMiddleware) verifyToken(tokenString string) (string, string, error) {
	var claims jwt.MapClaims

	// If JWKS is initialized from Supabase, verify signature cryptographically
	if m.jwks != nil {
		token, err := jwt.Parse(tokenString, m.jwks.Keyfunc)
		if err != nil || !token.Valid {
			return "", "", fmt.Errorf("cryptographic signature verification failed: %w", err)
		}
		var ok bool
		claims, ok = token.Claims.(jwt.MapClaims)
		if !ok {
			return "", "", ErrInvalidToken
		}
	} else {
		// Fallback when running offline without SUPABASE_URL configured in local dev
		parser := jwt.NewParser()
		_, _, err := parser.ParseUnverified(tokenString, &claims)
		if err != nil {
			return "", "", errors.New("invalid or malformed JWT token")
		}
	}

	// Verify expiration claim (exp)
	if expVal, ok := claims["exp"]; ok {
		var expUnix int64
		switch v := expVal.(type) {
		case float64:
			expUnix = int64(v)
		case json.Number:
			expUnix, _ = v.Int64()
		}
		if expUnix > 0 && time.Now().Unix() > expUnix {
			return "", "", ErrTokenExpired
		}
	}

	// Extract verified email claim
	email, _ := claims["email"].(string)
	if email == "" {
		if userMeta, ok := claims["user_metadata"].(map[string]interface{}); ok {
			email, _ = userMeta["email"].(string)
		}
	}

	if email == "" {
		return "", "", errors.New("email claim missing in token")
	}

	sub, _ := claims["sub"].(string)
	return strings.ToLower(strings.TrimSpace(email)), sub, nil
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

// GetAuthSub extracts the authenticated user's sub ID from request context.
func GetAuthSub(ctx context.Context) string {
	val := ctx.Value(UserSubContextKey)
	if val == nil {
		return ""
	}
	sub, ok := val.(string)
	if !ok {
		return ""
	}
	return sub
}
