package transport_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/core/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func generateTestKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	return pub, priv
}

func signTestJWT(t *testing.T, priv ed25519.PrivateKey, claims map[string]any) string {
	t.Helper()
	header := map[string]string{"alg": "EdDSA", "typ": "JWT"}
	hdrJSON, _ := json.Marshal(header)
	clJSON, _ := json.Marshal(claims)
	hdr64 := base64URLEncode(hdrJSON)
	cl64 := base64URLEncode(clJSON)
	msg := hdr64 + "." + cl64
	sig := ed25519.Sign(priv, []byte(msg))
	return msg + "." + base64URLEncode(sig)
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	pub, priv := generateTestKeyPair(t)
	token := signTestJWT(t, priv, map[string]any{
		"sub":  "CUST001",
		"tier": "professional",
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
		"entities": map[string]int{"included": 15, "max": 20},
	})

	handler := transport.JWTAuthMiddleware(pub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := transport.ClaimsFromContext(r.Context())
		assert.Equal(t, "CUST001", claims.Sub)
		assert.Equal(t, "professional", claims.Tier)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/mcp/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	pub, priv := generateTestKeyPair(t)
	token := signTestJWT(t, priv, map[string]any{
		"sub": "CUST001",
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	})

	handler := transport.JWTAuthMiddleware(pub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("POST", "/mcp/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuthMiddleware_MissingToken(t *testing.T) {
	pub, _ := generateTestKeyPair(t)

	handler := transport.JWTAuthMiddleware(pub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("POST", "/mcp/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuthMiddleware_InvalidSignature(t *testing.T) {
	pub, _ := generateTestKeyPair(t)
	_, otherPriv := generateTestKeyPair(t)
	token := signTestJWT(t, otherPriv, map[string]any{
		"sub": "CUST001",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	})

	handler := transport.JWTAuthMiddleware(pub)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("POST", "/mcp/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
