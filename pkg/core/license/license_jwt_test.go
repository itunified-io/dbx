package license_test

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/itunified-io/dbx/pkg/core/license"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testEdKeyPair() (ed25519.PublicKey, ed25519.PrivateKey) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	return pub, priv
}

// signTestJWT creates a signed JWT license for testing.
func signTestJWT(t *testing.T, priv ed25519.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(license.SigningMethodEdDSA, claims)
	token.Header["kid"] = kid
	token.Header["alg"] = "EdDSA"
	signed, err := token.SignedString(priv)
	require.NoError(t, err)
	return signed
}

func TestValidateValidJWTLicense(t *testing.T) {
	pub, priv := testEdKeyPair()
	kid := "itu-lic-2026-01"
	tokenStr := signTestJWT(t, priv, kid, jwt.MapClaims{
		"jti": "LIC-2026-00042",
		"sub": "ORG-00015",
		"iss": "license.itunified.io",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(360 * 24 * time.Hour).Unix(),
		"customer": "Acme Corp",
		"tier":     "production",
		"bundles":  []string{"core", "ha", "ops"},
		"entities": map[string]interface{}{
			"oracle_database": map[string]interface{}{"included": 10, "max": 15},
			"pg_database":     map[string]interface{}{"included": 5, "max": 10},
			"host":            map[string]interface{}{"included": 10, "max": 20},
		},
		"grace": map[string]interface{}{"overage_entities": 2, "overage_days": 30},
		"trial": false,
	})

	v := license.NewJWTValidator(license.EdKeySet{kid: pub}, nil)
	result, err := v.ValidateJWT(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, license.StateValid, result.State)
	assert.Equal(t, "production", result.Tier)
}

func TestValidateExpiredGracePeriod(t *testing.T) {
	pub, priv := testEdKeyPair()
	kid := "itu-lic-2026-01"
	tokenStr := signTestJWT(t, priv, kid, jwt.MapClaims{
		"jti": "LIC-2026-00043",
		"sub": "ORG-002",
		"iss": "license.itunified.io",
		"iat": time.Now().Add(-400 * 24 * time.Hour).Unix(),
		"exp": time.Now().Add(-5 * 24 * time.Hour).Unix(), // expired 5 days ago
		"tier":    "production",
		"bundles": []string{"core"},
		"entities": map[string]interface{}{
			"oracle_database": map[string]interface{}{"included": 5, "max": 5},
		},
		"grace": map[string]interface{}{"overage_entities": 0, "overage_days": 0},
		"trial": false,
	})

	v := license.NewJWTValidator(license.EdKeySet{kid: pub}, nil)
	result, err := v.ValidateJWT(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, license.StateGrace, result.State)
	assert.Equal(t, 9, result.GraceDaysRemaining) // 14 - 5
}

func TestValidateExpiredPastGrace(t *testing.T) {
	pub, priv := testEdKeyPair()
	kid := "itu-lic-2026-01"
	tokenStr := signTestJWT(t, priv, kid, jwt.MapClaims{
		"jti": "LIC-2026-00044",
		"sub": "ORG-003",
		"iss": "license.itunified.io",
		"iat": time.Now().Add(-400 * 24 * time.Hour).Unix(),
		"exp": time.Now().Add(-20 * 24 * time.Hour).Unix(), // expired 20 days ago
		"tier":    "core",
		"bundles": []string{"core"},
		"entities": map[string]interface{}{
			"oracle_database": map[string]interface{}{"included": 5, "max": 5},
		},
		"grace": map[string]interface{}{"overage_entities": 0, "overage_days": 0},
		"trial": false,
	})

	v := license.NewJWTValidator(license.EdKeySet{kid: pub}, nil)
	result, err := v.ValidateJWT(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, license.StateReadOnly, result.State)
}

func TestValidateRejectsHS256Algorithm(t *testing.T) {
	pub, _ := testEdKeyPair()
	kid := "itu-lic-2026-01"
	// Craft a token with alg=HS256 (algorithm confusion attack — must be rejected)
	hmacToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti": "LIC-HACK", "sub": "ORG-EVIL", "iss": "license.itunified.io",
		"exp": time.Now().Add(360 * 24 * time.Hour).Unix(),
	})
	hmacToken.Header["kid"] = kid
	tokenStr, _ := hmacToken.SignedString([]byte("attacker-secret"))

	v := license.NewJWTValidator(license.EdKeySet{kid: pub}, nil)
	_, err := v.ValidateJWT(tokenStr)
	assert.Error(t, err)
}

func TestValidateRejectsUnknownKid(t *testing.T) {
	pub, priv := testEdKeyPair()
	tokenStr := signTestJWT(t, priv, "unknown-kid", jwt.MapClaims{
		"jti": "LIC-001", "sub": "ORG-001", "iss": "license.itunified.io",
		"exp": time.Now().Add(360 * 24 * time.Hour).Unix(),
		"trial": false,
	})

	v := license.NewJWTValidator(license.EdKeySet{"itu-lic-2026-01": pub}, nil)
	_, err := v.ValidateJWT(tokenStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown kid")
}

func TestCheckEntityLimit(t *testing.T) {
	entities := license.EntityClaims{
		"oracle_database": {Included: 10, Max: 15},
		"pg_database":     {Included: 5, Max: 10},
	}
	grace := license.GraceClaim{OverageEntities: 2, OverageDays: 30}

	// Within max: OK
	assert.NoError(t, license.CheckEntityLimit(entities, grace, "oracle_database", 14))
	// At max: OK
	assert.NoError(t, license.CheckEntityLimit(entities, grace, "oracle_database", 15))
	// Over max but within overage grace: OK (warning)
	assert.NoError(t, license.CheckEntityLimit(entities, grace, "oracle_database", 17))
	// Over max + overage: REJECT
	assert.Error(t, license.CheckEntityLimit(entities, grace, "oracle_database", 18))
	// Unknown entity type: REJECT
	assert.Error(t, license.CheckEntityLimit(entities, grace, "mysql_database", 1))
}

func TestCheckBundleEntitlement(t *testing.T) {
	bundles := []string{"core", "ha"}
	assert.NoError(t, license.CheckBundleEntitlement(bundles, "core"))
	assert.NoError(t, license.CheckBundleEntitlement(bundles, "ha"))
	assert.Error(t, license.CheckBundleEntitlement(bundles, "ops"))
}

func TestRegisterTargetWithinLimit(t *testing.T) {
	entities := license.EntityClaims{
		"oracle_database": {Included: 10, Max: 15},
		"pg_database":     {Included: 5, Max: 10},
	}
	grace := license.GraceClaim{OverageEntities: 2, OverageDays: 30}
	registered := map[string]int{"oracle_database": 3, "pg_database": 2}

	result, err := license.CheckRegistration(entities, grace, registered, "oracle_database")
	require.NoError(t, err)
	assert.Equal(t, 11, result.Remaining) // 15 - 3 - 1 (the new one)
}

func TestRegisterTargetAtMax(t *testing.T) {
	entities := license.EntityClaims{
		"oracle_database": {Included: 5, Max: 5},
	}
	grace := license.GraceClaim{OverageEntities: 2, OverageDays: 30}
	registered := map[string]int{"oracle_database": 5}

	// Within overage grace
	result, err := license.CheckRegistration(entities, grace, registered, "oracle_database")
	require.NoError(t, err)
	assert.True(t, result.InOverage)
	assert.Equal(t, 1, result.OverageRemaining)
}

func TestRegisterTargetExceedsOverage(t *testing.T) {
	entities := license.EntityClaims{
		"oracle_database": {Included: 5, Max: 5},
	}
	grace := license.GraceClaim{OverageEntities: 2, OverageDays: 30}
	registered := map[string]int{"oracle_database": 7} // 5 max + 2 overage already used

	_, err := license.CheckRegistration(entities, grace, registered, "oracle_database")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "limit exceeded")
}

func TestRegistrationResponseIncludesCapacity(t *testing.T) {
	entities := license.EntityClaims{
		"oracle_database": {Included: 10, Max: 15},
		"pg_database":     {Included: 5, Max: 10},
		"host":            {Included: 10, Max: 20},
	}
	registered := map[string]int{"oracle_database": 8, "pg_database": 3, "host": 5}

	capacity := license.EntityCapacity(entities, registered)
	assert.Equal(t, 7, capacity["oracle_database"])
	assert.Equal(t, 7, capacity["pg_database"])
	assert.Equal(t, 15, capacity["host"])
}
