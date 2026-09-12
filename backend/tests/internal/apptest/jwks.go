package apptest

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const TestKeyID = "eventory-test-key-1"

// JWKSServer manages an in-memory RSA key pair and a mock HTTP server serving JWKS.
type JWKSServer struct {
	server     *httptest.Server
	privateKey *rsa.PrivateKey
}

var (
	defaultJWKSOnce sync.Once
	defaultJWKS     *JWKSServer
)

// StartMockJWKSServer spins up an in-memory HTTP server serving the JWKS endpoint.
func StartMockJWKSServer() *JWKSServer {
	defaultJWKSOnce.Do(func() {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(fmt.Sprintf("apptest: failed to generate RSA key pair: %v", err))
		}

		pub := &priv.PublicKey
		nStr := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
		eBytes := big.NewInt(int64(pub.E)).Bytes()
		eStr := base64.RawURLEncoding.EncodeToString(eBytes)

		jwksPayload := map[string]any{
			"keys": []map[string]any{
				{
					"kty": "RSA",
					"use": "sig",
					"alg": "RS256",
					"kid": TestKeyID,
					"n":   nStr,
					"e":   eStr,
				},
			},
		}
		data, err := json.Marshal(jwksPayload)
		if err != nil {
			panic(fmt.Sprintf("apptest: failed to marshal JWKS: %v", err))
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/auth/v1/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
		})

		ts := httptest.NewServer(mux)
		defaultJWKS = &JWKSServer{
			server:     ts,
			privateKey: priv,
		}
	})

	return defaultJWKS
}

// URL returns the root mock Supabase URL.
func (j *JWKSServer) URL() string {
	return j.server.URL
}

// SignJWT creates and signs a JWT claim containing sub, email, and exp with the test RSA private key.
func (j *JWKSServer) SignJWT(sub uuid.UUID, email string) string {
	claims := jwt.MapClaims{
		"sub":   sub.String(),
		"email": email,
		"exp":   time.Now().Add(2 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = TestKeyID
	tokenString, err := token.SignedString(j.privateKey)
	if err != nil {
		panic(fmt.Sprintf("apptest: failed to sign test JWT: %v", err))
	}
	return tokenString
}
