package apptest

import (
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
)

var userCounter atomic.Uint64

// UniqueSuffix returns a monotonically increasing string to guarantee isolation across scenarios.
func UniqueSuffix() string {
	return fmt.Sprintf("%d", userCounter.Add(1))
}

// TestUser represents an isolated user identity with test JWT credentials.
type TestUser struct {
	ID          uuid.UUID
	Email       string
	Handle      string
	DisplayName string
	Token       string
}

// AuthHeaders returns HTTP headers with Bearer authentication token.
func (u *TestUser) AuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + u.Token,
	}
}

// NewTestUser generates a unique user with pre-signed JWT claims.
func NewTestUser() *TestUser {
	suffix := UniqueSuffix()
	id := uuid.New()
	email := fmt.Sprintf("player_%s@test.eventory.gg", suffix)
	handle := fmt.Sprintf("player_%s", suffix)
	displayName := fmt.Sprintf("Player %s", suffix)
	token := GenerateTestJWT(id, email)

	return &TestUser{
		ID:          id,
		Email:       email,
		Handle:      handle,
		DisplayName: displayName,
		Token:       token,
	}
}

// GenerateTestJWT signs a JWT claim containing sub, email, and exp using the test RSA private key.
func GenerateTestJWT(sub uuid.UUID, email string) string {
	return StartMockJWKSServer().SignJWT(sub, email)
}
