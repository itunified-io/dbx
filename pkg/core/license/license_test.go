package license_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/core/license"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestLicense(t *testing.T, claims license.Claims, expiry time.Time) ([]byte, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	claims.ExpiresAt = expiry.Unix()
	payload, err := json.Marshal(claims)
	require.NoError(t, err)

	sig := ed25519.Sign(priv, payload)
	lic := license.RawLicense{Payload: payload, Signature: sig}
	data, err := json.Marshal(lic)
	require.NoError(t, err)
	return data, pub
}

func TestValidLicense(t *testing.T) {
	claims := license.Claims{
		LicenseID:  "LIC-2026-00042",
		CustomerID: "CUST-001",
		Bundles:    []string{"core", "ha"},
		MaxTargets: 10,
	}
	data, pub := generateTestLicense(t, claims, time.Now().Add(365*24*time.Hour))

	v := license.NewValidator(pub)
	result, err := v.Validate(data)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Equal(t, "LIC-2026-00042", result.Claims.LicenseID)
	assert.Contains(t, result.Claims.Bundles, "core")
	assert.Contains(t, result.Claims.Bundles, "ha")
}

func TestExpiredLicenseWithGrace(t *testing.T) {
	claims := license.Claims{
		LicenseID:  "LIC-EXPIRED",
		Bundles:    []string{"core"},
		MaxTargets: 5,
	}
	data, pub := generateTestLicense(t, claims, time.Now().Add(-5*24*time.Hour))

	v := license.NewValidator(pub)
	result, err := v.Validate(data)
	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.True(t, result.InGrace)
	assert.Equal(t, 9, result.GraceDaysRemaining)
}

func TestExpiredLicenseBeyondGrace(t *testing.T) {
	claims := license.Claims{
		LicenseID:  "LIC-DEAD",
		Bundles:    []string{"core"},
		MaxTargets: 5,
	}
	data, pub := generateTestLicense(t, claims, time.Now().Add(-20*24*time.Hour))

	v := license.NewValidator(pub)
	result, err := v.Validate(data)
	require.NoError(t, err)
	assert.False(t, result.Valid)
	assert.False(t, result.InGrace)
}

func TestInvalidSignature(t *testing.T) {
	claims := license.Claims{LicenseID: "LIC-TAMPERED", Bundles: []string{"core"}}
	data, _ := generateTestLicense(t, claims, time.Now().Add(365*24*time.Hour))

	otherPub, _, _ := ed25519.GenerateKey(rand.Reader)
	v := license.NewValidator(otherPub)
	_, err := v.Validate(data)
	assert.Error(t, err)
}

func TestBundleCheck(t *testing.T) {
	claims := license.Claims{Bundles: []string{"core", "ha"}}
	data, pub := generateTestLicense(t, claims, time.Now().Add(365*24*time.Hour))

	v := license.NewValidator(pub)
	result, _ := v.Validate(data)
	assert.True(t, result.HasBundle("core"))
	assert.True(t, result.HasBundle("ha"))
	assert.False(t, result.HasBundle("ops"))
}
