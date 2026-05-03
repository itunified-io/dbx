// Package license provides a tier-gated license facade for the dbxcli.
//
// This package implements ADR-0094 (per-tool JWT licensing) for the dbx
// monolith CLI: an Ed25519-signed JWT stored at ~/.dbx/license.jwt,
// loaded once per command, and consulted via RequireBundle/RequireTier
// at the top of gated leaf RunE functions.
//
// Licensing model:
//   - Tiers: community < business < enterprise (ordered).
//   - Bundles: capability tags within a license, e.g. "provision",
//     "dataguard", "audit". Enterprise tier with the matching bundle
//     unlocks gated commands.
//   - Production licenses verify against an embedded Ed25519 public key
//     at pkg/license/keys/prod.pub (placeholder until the dbx license
//     CA is provisioned).
//   - DEV-MODE licenses are self-signed by the local issuer and
//     verify against locally-trusted keys under ~/.dbx/.trust/.
//
// This package is intentionally independent of pkg/core/license: that
// package implements the broader JWT validation infrastructure used
// by the pipeline runtime (kid-based key sets, CRL, grace, entity
// counts). pkg/license is the slim tier-gate the dbxcli leaf
// commands call into. The two are aligned semantically (same Ed25519
// + JWT primitives) but kept decoupled to avoid cyclic dependencies
// and to keep this gate trivially auditable.
package license

import (
	"errors"
	"time"
)

// Tier is the license tier level. Tiers are ordered:
//
//	community < business < enterprise
//
// HasTier checks whether a license meets a minimum requested tier.
type Tier string

const (
	// TierCommunity is the default tier for users with no license file.
	// OSS / read-only / non-gated tools work at this tier.
	TierCommunity Tier = "community"

	// TierBusiness unlocks the standard paid feature set.
	TierBusiness Tier = "business"

	// TierEnterprise unlocks every paid feature including ops/provision,
	// data guard, audit, and other Enterprise-only bundles.
	TierEnterprise Tier = "enterprise"
)

// tierRank returns the ordering rank of a tier. Higher = more privileged.
// Unknown tiers map to -1 (less privileged than community).
func tierRank(t Tier) int {
	switch t {
	case TierCommunity:
		return 0
	case TierBusiness:
		return 1
	case TierEnterprise:
		return 2
	default:
		return -1
	}
}

// License is the decoded JWT payload of a dbx license.
//
// Field tags match the JWT claim names defined in ADR-0094.
type License struct {
	Subject   string   `json:"sub"`
	Tier      Tier     `json:"tier"`
	Bundles   []string `json:"bundles"`
	IssuedAt  int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`

	// Dev marks a license as dev-issued (self-signed). When true,
	// the verifier emits a warning on Load() and downstream gates
	// MAY treat the license as untrusted in production builds.
	Dev bool `json:"dev,omitempty"`
}

// IsValid returns nil if the license is currently valid (not expired)
// and an error otherwise.
//
// Expiry is checked against the wall clock. Licenses with ExpiresAt
// of zero are treated as non-expiring (used by long-lived dev keys).
func (l *License) IsValid() error {
	if l == nil {
		return ErrMissing
	}
	if l.ExpiresAt == 0 {
		return nil
	}
	if time.Now().Unix() >= l.ExpiresAt {
		return ErrExpired
	}
	return nil
}

// HasBundle reports whether the license includes the named bundle.
func (l *License) HasBundle(name string) bool {
	if l == nil {
		return false
	}
	for _, b := range l.Bundles {
		if b == name {
			return true
		}
	}
	return false
}

// HasTier reports whether the license meets or exceeds the requested
// minimum tier. Unknown tiers always fail.
func (l *License) HasTier(min Tier) bool {
	if l == nil {
		return false
	}
	have := tierRank(l.Tier)
	want := tierRank(min)
	if have < 0 || want < 0 {
		return false
	}
	return have >= want
}

// Sentinel errors. Callers SHOULD use errors.Is to test for these.
var (
	// ErrMissing indicates no license file is present at the default path.
	// Callers MAY treat this as TierCommunity.
	ErrMissing = errors.New("license: no license file")

	// ErrExpired indicates the license file is well-formed and
	// signed by a trusted key, but past its exp claim.
	ErrExpired = errors.New("license: expired")

	// ErrInvalidSignature indicates the JWT signature did not verify
	// against any trusted key.
	ErrInvalidSignature = errors.New("license: invalid signature")

	// ErrInvalidFormat indicates the file at the license path is
	// not a parseable EdDSA-signed JWT.
	ErrInvalidFormat = errors.New("license: invalid format")
)

// ErrTierGate is returned by RequireTier/RequireBundle when a license
// gate is not satisfied. It implements Unwrap so callers can match
// the sentinel via errors.Is.
type ErrTierGate struct {
	Bundle string
	Tier   Tier
	Cause  error
}

// Error renders a human-readable gate message.
func (e *ErrTierGate) Error() string {
	switch {
	case e.Bundle != "" && e.Tier != "":
		return "tier gate: " + e.Bundle + " bundle requires " + string(e.Tier) + " tier"
	case e.Bundle != "":
		return "tier gate: " + e.Bundle + " bundle requires Enterprise tier"
	case e.Tier != "":
		return "tier gate: requires " + string(e.Tier) + " tier"
	default:
		return "tier gate: license required"
	}
}

// Unwrap surfaces the underlying cause (e.g. ErrMissing, ErrExpired).
func (e *ErrTierGate) Unwrap() error { return e.Cause }
