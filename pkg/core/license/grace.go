package license

import "fmt"

// LicenseState represents the current license enforcement state.
type LicenseState int

const (
	StateValid    LicenseState = iota // All features active
	StateGrace                        // Expired < 14 days, WARN banner, all features work
	StateReadOnly                     // Expired > 14 days, OSS-only, no EE mutations
	StateBlocked                      // Revoked or invalid signature, hard block
)

// String returns a human-readable state name.
func (s LicenseState) String() string {
	switch s {
	case StateValid:
		return "valid"
	case StateGrace:
		return "grace"
	case StateReadOnly:
		return "read-only"
	case StateBlocked:
		return "blocked"
	default:
		return "unknown"
	}
}

// AllowsBundle returns true if the state allows EE bundle access.
// ReadOnly and Blocked states block all EE bundles.
func (s LicenseState) AllowsBundle(bundle string) bool {
	switch s {
	case StateValid, StateGrace:
		return true
	case StateReadOnly, StateBlocked:
		return false
	default:
		return false
	}
}

// AllowsMutation returns true if the state allows write/mutate operations.
func (s LicenseState) AllowsMutation() bool {
	return s == StateValid || s == StateGrace
}

// WarningMessage returns a user-facing warning for non-valid states, or empty string.
func (s LicenseState) WarningMessage(graceDays int) string {
	switch s {
	case StateGrace:
		return fmt.Sprintf("WARNING: License expired. %d grace day(s) remaining. Renew at https://dbx.itunified.io/renew", graceDays)
	case StateReadOnly:
		return "WARNING: License expired beyond grace period. Running in read-only mode. Only OSS commands available."
	case StateBlocked:
		return "ERROR: License is invalid or revoked. All EE operations blocked."
	default:
		return ""
	}
}

// DetermineState resolves the final license state from local JWT validation
// and optional CRL check results.
// No phone-home dependency — purely local.
func DetermineState(localResult *JWTValidationResult, crl *CRLChecker) LicenseState {
	if localResult == nil {
		return StateBlocked
	}
	// CRL revocation overrides everything (if CRL is configured and checked)
	if crl != nil && crl.IsConfigured() && localResult.JTI != "" && crl.IsRevoked(localResult.JTI) {
		return StateBlocked
	}
	return localResult.State
}
