package middlewares_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func TestJWTTokenParsing(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":   "user-123",
		"email": "56312tanwa@gmail.com",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	parser := jwt.NewParser()
	var parsedClaims jwt.MapClaims
	_, _, err = parser.ParseUnverified(tokenString, &parsedClaims)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	email, ok := parsedClaims["email"].(string)
	if !ok || email != "56312tanwa@gmail.com" {
		t.Errorf("expected email '56312tanwa@gmail.com', got %v", email)
	}
}
