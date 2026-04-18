// Package transport provides MCP streamable HTTP transport with JWT authentication.
package transport

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const claimsKey contextKey = "mcp_claims"

// MCPClaims are the JWT claims from an MCP access token.
type MCPClaims struct {
	Sub      string    `json:"sub"`
	Tier     string    `json:"tier"`
	Exp      int64     `json:"exp"`
	Entities *struct {
		Included int `json:"included"`
		Max      int `json:"max"`
	} `json:"entities,omitempty"`
}

// ClaimsFromContext extracts MCPClaims from the request context.
func ClaimsFromContext(ctx context.Context) *MCPClaims {
	c, _ := ctx.Value(claimsKey).(*MCPClaims)
	return c
}

// JWTAuthMiddleware validates Ed25519-signed JWT Bearer tokens on incoming MCP requests.
func JWTAuthMiddleware(publicKey ed25519.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(auth, "Bearer ")

			claims, err := verifyEdDSAJWT(token, publicKey)
			if err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
				return
			}

			if time.Now().Unix() > claims.Exp {
				http.Error(w, `{"error":"token expired"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func verifyEdDSAJWT(token string, pub ed25519.PublicKey) (*MCPClaims, error) {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	sig, err := base64URLDecode(parts[2])
	if err != nil {
		return nil, errInvalidToken
	}

	msg := parts[0] + "." + parts[1]
	if !ed25519.Verify(pub, []byte(msg), sig) {
		return nil, errInvalidSignature
	}

	payload, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}

	var claims MCPClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errInvalidToken
	}

	return &claims, nil
}

func base64URLDecode(s string) ([]byte, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

var (
	errInvalidToken     = &authError{"invalid token format"}
	errInvalidSignature = &authError{"invalid signature"}
)

type authError struct{ msg string }

func (e *authError) Error() string { return e.msg }
