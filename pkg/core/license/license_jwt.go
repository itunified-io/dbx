package license

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	GracePeriodDays    = 14
	DefaultLicensePath = ".dbx/license.jwt"
)

// EdKeySet maps kid values to Ed25519 public keys.
// Multiple keys are embedded in the binary for seamless rotation.
type EdKeySet map[string]ed25519.PublicKey

// SigningMethodEdDSA is a custom jwt.SigningMethod for Ed25519.
// This bridges golang-jwt/jwt/v5 with Ed25519 signing.
var SigningMethodEdDSA = &signingMethodEdDSA{}

type signingMethodEdDSA struct{}

func (m *signingMethodEdDSA) Verify(signingString string, sig []byte, key interface{}) error {
	pubKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("EdDSA verify: invalid key type %T", key)
	}
	if !ed25519.Verify(pubKey, []byte(signingString), sig) {
		return fmt.Errorf("EdDSA signature verification failed")
	}
	return nil
}

func (m *signingMethodEdDSA) Sign(signingString string, key interface{}) ([]byte, error) {
	privKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("EdDSA sign: invalid key type %T", key)
	}
	return ed25519.Sign(privKey, []byte(signingString)), nil
}

func (m *signingMethodEdDSA) Alg() string { return "EdDSA" }

func init() {
	jwt.RegisterSigningMethod("EdDSA", func() jwt.SigningMethod { return SigningMethodEdDSA })
}

// LicenseClaims are the decoded JWT claims from a license file.
type LicenseClaims struct {
	jwt.RegisteredClaims
	Customer string       `json:"customer"`
	Email    string       `json:"email,omitempty"`
	Edition  string       `json:"edition"`
	Tier     string       `json:"tier"`
	Bundles  []string     `json:"bundles"`
	Entities EntityClaims `json:"entities"`
	LicGrace GraceClaim   `json:"grace"`
	Users    int          `json:"users,omitempty"`
	Features Features     `json:"features,omitempty"`
	Trial    bool         `json:"trial"`
}

// Features are boolean feature flags beyond bundles.
type Features struct {
	DBMonitorEE      bool `json:"dbmonitor_ee,omitempty"`
	OEMIntegration   bool `json:"oem_integration,omitempty"`
	FleetCompliance  bool `json:"fleet_compliance,omitempty"`
	CustomCollectors bool `json:"custom_collectors,omitempty"`
}

// JWTValidationResult is the outcome of JWT license validation.
type JWTValidationResult struct {
	State              LicenseState
	GraceDaysRemaining int
	Warning            string
	Tier               string
	Bundles            []string
	Entities           EntityClaims
	Grace              GraceClaim
	JTI                string
}

// JWTValidator performs JWT EdDSA license validation with kid-based key selection.
type JWTValidator struct {
	keys EdKeySet
	crl  *CRLChecker // optional, nil if not configured
}

// NewJWTValidator creates a license validator with embedded Ed25519 public keys.
func NewJWTValidator(keys EdKeySet, crl *CRLChecker) *JWTValidator {
	return &JWTValidator{keys: keys, crl: crl}
}

// ValidateJWT runs the full license validation flow per spec 14.2:
// 1. Decode JWT header — verify alg = "EdDSA"
// 2. Select public key by kid from embedded key set
// 3. Verify EdDSA signature (local, instant, zero network)
// 4. Check exp claim (standard JWT expiry + 14-day grace)
// 5. [Optional] CRL check if configured
func (v *JWTValidator) ValidateJWT(tokenStr string) (*JWTValidationResult, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &LicenseClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Algorithm pinning: EdDSA only. Reject none, HS256, RS256.
		if t.Method.Alg() != "EdDSA" {
			return nil, fmt.Errorf("algorithm pinning: only EdDSA accepted, got %s", t.Method.Alg())
		}
		// kid-based key selection
		kid, ok := t.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, fmt.Errorf("missing kid in JWT header")
		}
		pubKey, found := v.keys[kid]
		if !found {
			return nil, fmt.Errorf("unknown kid: %s (license signed by unknown key)", kid)
		}
		return pubKey, nil
	}, jwt.WithValidMethods([]string{"EdDSA"}))
	if err != nil {
		// Only handle expiry errors gracefully — all other errors are hard failures.
		// This prevents algorithm confusion attacks from falling through to checkExpiry.
		if errors.Is(err, jwt.ErrTokenExpired) && token != nil && token.Claims != nil {
			claims, ok := token.Claims.(*LicenseClaims)
			if ok && claims.ExpiresAt != nil {
				return v.checkExpiry(claims)
			}
		}
		return nil, fmt.Errorf("license JWT validation failed: %w", err)
	}

	claims, ok := token.Claims.(*LicenseClaims)
	if !ok {
		return nil, fmt.Errorf("invalid license JWT claims")
	}

	// Optional CRL check
	if v.crl != nil && v.crl.IsConfigured() {
		v.crl.Refresh() // async-safe, respects cache TTL
		if claims.ID != "" && v.crl.IsRevoked(claims.ID) {
			return &JWTValidationResult{State: StateBlocked, JTI: claims.ID}, nil
		}
	}

	return v.checkExpiry(claims)
}

func (v *JWTValidator) checkExpiry(claims *LicenseClaims) (*JWTValidationResult, error) {
	result := &JWTValidationResult{
		Tier:     claims.Tier,
		Bundles:  claims.Bundles,
		Entities: claims.Entities,
		Grace:    claims.LicGrace,
		JTI:      claims.ID,
	}

	if claims.ExpiresAt == nil {
		result.State = StateValid
		return result, nil
	}

	now := time.Now()
	expires := claims.ExpiresAt.Time
	if now.Before(expires) {
		result.State = StateValid
		return result, nil
	}

	daysSinceExpiry := int(now.Sub(expires).Hours() / 24)
	if daysSinceExpiry <= GracePeriodDays {
		remaining := GracePeriodDays - daysSinceExpiry
		result.State = StateGrace
		result.GraceDaysRemaining = remaining
		result.Warning = fmt.Sprintf("License expired %d day(s) ago. %d grace day(s) remaining.", daysSinceExpiry, remaining)
		return result, nil
	}

	result.State = StateReadOnly
	result.Warning = "License expired beyond grace period. Read-only mode enforced."
	return result, nil
}

// LoadLicenseJWT loads a JWT license file from the given path, or from defaults:
// 1. --license-file flag
// 2. DBX_LICENSE_FILE env var
// 3. ~/.dbx/license.jwt (default)
func LoadLicenseJWT(path string) (string, error) {
	if path == "" {
		path = os.Getenv("DBX_LICENSE_FILE")
	}
	if path == "" {
		home, _ := os.UserHomeDir()
		path = home + "/" + DefaultLicensePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read license file %s: %w", path, err)
	}
	return string(data), nil
}
