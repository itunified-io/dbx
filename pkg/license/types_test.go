package license

import (
	"errors"
	"testing"
	"time"
)

func TestLicense_IsValid(t *testing.T) {
	cases := []struct {
		name string
		lic  *License
		want error
	}{
		{"nil license", nil, ErrMissing},
		{"no expiry", &License{Tier: TierEnterprise}, nil},
		{"future expiry", &License{ExpiresAt: time.Now().Add(time.Hour).Unix()}, nil},
		{"past expiry", &License{ExpiresAt: time.Now().Add(-time.Hour).Unix()}, ErrExpired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.lic.IsValid()
			if !errors.Is(got, tc.want) && got != tc.want {
				t.Fatalf("IsValid() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLicense_HasBundle(t *testing.T) {
	l := &License{Bundles: []string{"provision", "dataguard"}}
	if !l.HasBundle("provision") {
		t.Error("expected HasBundle(provision) = true")
	}
	if l.HasBundle("audit") {
		t.Error("expected HasBundle(audit) = false")
	}
	var nilLic *License
	if nilLic.HasBundle("provision") {
		t.Error("nil license must not have any bundle")
	}
}

func TestLicense_HasTier_Ordering(t *testing.T) {
	cases := []struct {
		have, min Tier
		want      bool
	}{
		{TierCommunity, TierCommunity, true},
		{TierCommunity, TierBusiness, false},
		{TierCommunity, TierEnterprise, false},
		{TierBusiness, TierCommunity, true},
		{TierBusiness, TierBusiness, true},
		{TierBusiness, TierEnterprise, false},
		{TierEnterprise, TierCommunity, true},
		{TierEnterprise, TierBusiness, true},
		{TierEnterprise, TierEnterprise, true},
		{Tier("garbage"), TierCommunity, false},
		{TierEnterprise, Tier("garbage"), false},
	}
	for _, tc := range cases {
		l := &License{Tier: tc.have}
		got := l.HasTier(tc.min)
		if got != tc.want {
			t.Errorf("(have=%s,min=%s) HasTier=%v, want %v", tc.have, tc.min, got, tc.want)
		}
	}
}

func TestErrTierGate_Error(t *testing.T) {
	cases := []struct {
		e    *ErrTierGate
		want string
	}{
		{&ErrTierGate{Bundle: "provision", Tier: TierEnterprise}, "tier gate: provision bundle requires enterprise tier"},
		{&ErrTierGate{Bundle: "provision"}, "tier gate: provision bundle requires Enterprise tier"},
		{&ErrTierGate{Tier: TierBusiness}, "tier gate: requires business tier"},
		{&ErrTierGate{}, "tier gate: license required"},
	}
	for _, tc := range cases {
		if got := tc.e.Error(); got != tc.want {
			t.Errorf("Error() = %q, want %q", got, tc.want)
		}
	}
}

func TestErrTierGate_Unwrap(t *testing.T) {
	gate := &ErrTierGate{Bundle: "provision", Cause: ErrMissing}
	if !errors.Is(gate, ErrMissing) {
		t.Error("expected errors.Is(gate, ErrMissing) = true")
	}
}
