package middlewares_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/middlewares"
)

func TestGetAuthEmail_Missing(t *testing.T) {
	_, err := middlewares.GetAuthEmail(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error when email is missing from context")
	}
}

func TestGetAuthEmail_Present(t *testing.T) {
	ctx := context.WithValue(context.Background(), middlewares.UserEmailContextKey, "player@eventory.gg")
	email, err := middlewares.GetAuthEmail(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if email != "player@eventory.gg" {
		t.Errorf("expected 'player@eventory.gg', got %v", email)
	}
}

func TestGetAuthUserID(t *testing.T) {
	expectedID := uuid.New()
	ctx := context.WithValue(context.Background(), middlewares.UserSubContextKey, expectedID.String())

	id, err := middlewares.GetAuthUserID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != expectedID {
		t.Errorf("expected %v, got %v", expectedID, id)
	}
}

func TestAuthMiddleware_CryptographicVerification(t *testing.T) {
	// 1. Generate RSA key pair for testing
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	keyID := "test-rsa-key-1"

	// 2. Start mock JWKS server
	pub := &privKey.PublicKey
	nStr := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	eStr := base64.RawURLEncoding.EncodeToString(eBytes)

	jwksPayload := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": keyID,
				"n":   nStr,
				"e":   eStr,
			},
		},
	}
	jwksBytes, _ := json.Marshal(jwksPayload)

	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksBytes)
	}))
	defer jwksServer.Close()

	// 3. Setup test Huma API & Middleware
	_, api := humatest.New(t)
	authMid := middlewares.NewAuthMiddleware(api, jwksServer.URL, "development")

	var capturedEmail, capturedSub string
	huma.Register(api, huma.Operation{
		OperationID: "test-auth",
		Method:      http.MethodGet,
		Path:        "/test-auth",
		Middlewares: huma.Middlewares{authMid.HumaMiddleware()},
	}, func(ctx context.Context, input *struct{}) (*struct{}, error) {
		capturedEmail, _ = middlewares.GetAuthEmail(ctx)
		capturedSub = middlewares.GetAuthSub(ctx)
		return nil, nil
	})

	signToken := func(key *rsa.PrivateKey, kid string, exp time.Duration, email, sub string) string {
		claims := jwt.MapClaims{
			"sub":   sub,
			"email": email,
			"exp":   time.Now().Add(exp).Unix(),
		}
		tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		tok.Header["kid"] = kid
		signed, err := tok.SignedString(key)
		if err != nil {
			t.Fatalf("failed to sign token: %v", err)
		}
		return signed
	}

	t.Run("valid token successfully authenticates", func(t *testing.T) {
		sub := uuid.New().String()
		email := "test@eventory.gg"
		tok := signToken(privKey, keyID, 1*time.Hour, email, sub)

		resp := api.Get("/test-auth", "Authorization: Bearer "+tok)
		if resp.Code != http.StatusOK && resp.Code != http.StatusNoContent {
			t.Fatalf("expected 200/204, got %d: %s", resp.Code, resp.Body.String())
		}
		if capturedEmail != email {
			t.Errorf("expected email %q, got %q", email, capturedEmail)
		}
		if capturedSub != sub {
			t.Errorf("expected sub %q, got %q", sub, capturedSub)
		}
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		tok := signToken(privKey, keyID, -1*time.Hour, "expired@eventory.gg", uuid.New().String())
		resp := api.Get("/test-auth", "Authorization: Bearer "+tok)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("tampered or foreign key signature returns 401", func(t *testing.T) {
		foreignKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		tok := signToken(foreignKey, keyID, 1*time.Hour, "tampered@eventory.gg", uuid.New().String())
		resp := api.Get("/test-auth", "Authorization: Bearer "+tok)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.Code)
		}
	})

	t.Run("dev-token bypass works in development mode", func(t *testing.T) {
		resp := api.Get("/test-auth", "Authorization: Bearer dev-token")
		if resp.Code != http.StatusOK && resp.Code != http.StatusNoContent {
			t.Fatalf("expected 200/204 for dev-token, got %d", resp.Code)
		}
		if capturedEmail != middlewares.DevEmail {
			t.Errorf("expected dev email, got %q", capturedEmail)
		}
	})
}
