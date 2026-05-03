package license

import "errors"

// RequireBundle returns nil if the local license is present, valid
// (not expired), enterprise-tier, and includes the named bundle.
//
// Otherwise it returns an *ErrTierGate whose Cause indicates the
// underlying reason (ErrMissing, ErrExpired, etc).
//
// Bundle gates implicitly require Enterprise tier — Community and
// Business tiers never satisfy a bundle gate.
//
// Callers SHOULD wrap the returned error to add command context, e.g.:
//
//	if err := license.RequireBundle("provision"); err != nil {
//	    return fmt.Errorf("dbxcli provision install grid: %w", err)
//	}
func RequireBundle(bundle string) error {
	lic, err := Load()
	if err != nil {
		return &ErrTierGate{Bundle: bundle, Tier: TierEnterprise, Cause: err}
	}
	if vErr := lic.IsValid(); vErr != nil {
		return &ErrTierGate{Bundle: bundle, Tier: TierEnterprise, Cause: vErr}
	}
	if !lic.HasTier(TierEnterprise) {
		return &ErrTierGate{Bundle: bundle, Tier: TierEnterprise, Cause: errors.New("license tier is " + string(lic.Tier))}
	}
	if !lic.HasBundle(bundle) {
		return &ErrTierGate{Bundle: bundle, Tier: TierEnterprise, Cause: errors.New("license does not include bundle " + bundle)}
	}
	return nil
}

// RequireTier returns nil if the local license meets the requested
// minimum tier and is not expired.
//
// TierCommunity is satisfied by ANY valid license AND by ErrMissing
// (no license file = community user). Higher tiers require an
// explicit license.
func RequireTier(min Tier) error {
	lic, err := Load()
	if err != nil {
		if errors.Is(err, ErrMissing) && min == TierCommunity {
			return nil
		}
		return &ErrTierGate{Tier: min, Cause: err}
	}
	if vErr := lic.IsValid(); vErr != nil {
		return &ErrTierGate{Tier: min, Cause: vErr}
	}
	if !lic.HasTier(min) {
		return &ErrTierGate{Tier: min, Cause: errors.New("license tier is " + string(lic.Tier))}
	}
	return nil
}
