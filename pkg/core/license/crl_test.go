package license_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/core/license"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCRLCheckerNotConfigured(t *testing.T) {
	// When no URL is configured, CRL checker is a no-op
	checker := license.NewCRLChecker("", 4*time.Hour)
	assert.False(t, checker.IsRevoked("LIC-001"), "no URL = never revoked")
}

func TestCRLCheckerFindsRevokedJTI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a mock CRL response (simplified — real CRL is a signed JWT)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"revoked_jtis": []string{"LIC-REVOKED-001", "LIC-REVOKED-002"},
		})
	}))
	defer srv.Close()

	checker := license.NewCRLChecker(srv.URL, 4*time.Hour)
	err := checker.Refresh()
	require.NoError(t, err)

	assert.True(t, checker.IsRevoked("LIC-REVOKED-001"))
	assert.True(t, checker.IsRevoked("LIC-REVOKED-002"))
	assert.False(t, checker.IsRevoked("LIC-VALID-001"))
}

func TestCRLCheckerCacheTTL(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode(map[string]interface{}{
			"revoked_jtis": []string{"LIC-001"},
		})
	}))
	defer srv.Close()

	checker := license.NewCRLChecker(srv.URL, 4*time.Hour)

	// First refresh fetches from network
	checker.Refresh()
	assert.Equal(t, 1, callCount)

	// Second refresh within TTL uses cache
	checker.Refresh()
	assert.Equal(t, 1, callCount, "should not re-fetch within cache TTL")
}

func TestCRLCheckerNetworkFailureGraceful(t *testing.T) {
	// Point to a non-existent server
	checker := license.NewCRLChecker("http://127.0.0.1:1", 4*time.Hour)
	err := checker.Refresh()
	// Network failure is non-fatal: proceed with local JWT validation
	assert.Error(t, err)
	assert.False(t, checker.IsRevoked("LIC-001"), "network fail = not revoked (offline-first)")
}
