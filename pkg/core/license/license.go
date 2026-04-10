// Package license provides Ed25519 license validation with grace period support.
package license

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"time"
)

const gracePeriodDays = 14

// Claims is the payload inside a license file.
type Claims struct {
	LicenseID  string   `json:"license_id"`
	CustomerID string   `json:"customer_id"`
	Tier       string   `json:"tier"`
	Bundles    []string `json:"bundles"`
	MaxTargets int      `json:"max_targets"`
	ExpiresAt  int64    `json:"expires_at"`
	Entities   *Entities `json:"entities,omitempty"`
	Grace      *Grace    `json:"grace,omitempty"`
}

// Entities tracks included and max entity counts.
type Entities struct {
	Included int `json:"included"`
	Max      int `json:"max"`
}

// Grace tracks overage allowance.
type Grace struct {
	OverageEntities int `json:"overage_entities"`
	OverageDays     int `json:"overage_days"`
}

// RawLicense is the on-disk format: JSON payload + Ed25519 signature.
type RawLicense struct {
	Payload   json.RawMessage `json:"payload"`
	Signature []byte          `json:"signature"`
}

// ValidationResult is the outcome of license validation.
type ValidationResult struct {
	Valid              bool
	InGrace            bool
	GraceDaysRemaining int
	Claims             Claims
}

// HasBundle returns true if the license includes the named bundle.
func (r *ValidationResult) HasBundle(name string) bool {
	for _, b := range r.Claims.Bundles {
		if b == name {
			return true
		}
	}
	return false
}

// Validator checks Ed25519-signed license files.
type Validator struct {
	publicKey ed25519.PublicKey
}

// NewValidator creates a license validator with the given public key.
func NewValidator(pub ed25519.PublicKey) *Validator {
	return &Validator{publicKey: pub}
}

// Validate checks signature, parses claims, and evaluates expiry + grace.
func (v *Validator) Validate(data []byte) (*ValidationResult, error) {
	var raw RawLicense
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, errors.New("invalid license format")
	}

	if !ed25519.Verify(v.publicKey, raw.Payload, raw.Signature) {
		return nil, errors.New("invalid license signature")
	}

	var claims Claims
	if err := json.Unmarshal(raw.Payload, &claims); err != nil {
		return nil, errors.New("invalid license claims")
	}

	expiry := time.Unix(claims.ExpiresAt, 0)
	now := time.Now()

	result := &ValidationResult{Claims: claims}

	if now.Before(expiry) {
		result.Valid = true
		return result, nil
	}

	daysSinceExpiry := int(now.Sub(expiry).Hours() / 24)
	if daysSinceExpiry <= gracePeriodDays {
		result.Valid = true
		result.InGrace = true
		result.GraceDaysRemaining = gracePeriodDays - daysSinceExpiry
		return result, nil
	}

	result.Valid = false
	return result, nil
}
