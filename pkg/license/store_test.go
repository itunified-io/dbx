package license

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// withTempHome redirects Path() to a temp dir for the duration of the test.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	prev := pathOverride
	pathOverride = filepath.Join(dir, ".dbx")
	if err := os.MkdirAll(pathOverride, 0o700); err != nil {
		t.Fatalf("mkdir override: %v", err)
	}
	t.Cleanup(func() { pathOverride = prev })
	return dir
}

func TestPath_Override(t *testing.T) {
	withTempHome(t)
	p := Path()
	if filepath.Base(p) != "license.jwt" {
		t.Fatalf("Path() = %s, want suffix license.jwt", p)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	withTempHome(t)
	_, err := Load()
	if !errors.Is(err, ErrMissing) {
		t.Fatalf("expected ErrMissing, got %v", err)
	}
}

func TestSaveLoad_DevRoundtrip(t *testing.T) {
	withTempHome(t)
	tok, err := IssueDev(License{
		Subject:   "lab-dev",
		Tier:      TierEnterprise,
		Bundles:   []string{"provision"},
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("IssueDev: %v", err)
	}
	if err := Save(tok); err != nil {
		t.Fatalf("Save: %v", err)
	}
	lic, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !lic.Dev {
		t.Error("expected Dev=true on dev-issued license")
	}
	if !lic.HasBundle("provision") {
		t.Error("expected HasBundle(provision)")
	}
	if !lic.HasTier(TierEnterprise) {
		t.Error("expected Enterprise tier")
	}
	// File mode check
	info, err := os.Stat(Path())
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("license file mode = %o, want 0600", info.Mode().Perm())
	}
}

func TestIssueDev_KeyIsIdempotent(t *testing.T) {
	withTempHome(t)
	if _, err := IssueDev(License{Tier: TierEnterprise}); err != nil {
		t.Fatalf("first IssueDev: %v", err)
	}
	keyPath := devKeyPath()
	info1, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat dev key: %v", err)
	}
	if _, err := IssueDev(License{Tier: TierEnterprise}); err != nil {
		t.Fatalf("second IssueDev: %v", err)
	}
	info2, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("stat dev key 2: %v", err)
	}
	if info1.ModTime() != info2.ModTime() {
		t.Errorf("dev key was rewritten: %v -> %v", info1.ModTime(), info2.ModTime())
	}
}

func TestLoad_RejectsTamperedToken(t *testing.T) {
	withTempHome(t)
	tok, err := IssueDev(License{Tier: TierEnterprise, Bundles: []string{"provision"}})
	if err != nil {
		t.Fatalf("IssueDev: %v", err)
	}
	// Mutate one byte in the payload (middle segment).
	bad := []byte(tok)
	// Find dot positions
	first, second := -1, -1
	for i, b := range bad {
		if b == '.' {
			if first == -1 {
				first = i
			} else {
				second = i
				break
			}
		}
	}
	bad[first+1] ^= 0x01
	if err := Save(string(bad)); err != nil {
		t.Fatalf("Save: %v", err)
	}
	_, err = Load()
	if err == nil {
		t.Fatal("expected verify failure on tampered token")
	}
	_ = second
}

func TestRequireBundle_MissingLicense(t *testing.T) {
	withTempHome(t)
	err := RequireBundle("provision")
	if err == nil {
		t.Fatal("expected ErrTierGate, got nil")
	}
	gate, ok := err.(*ErrTierGate)
	if !ok {
		t.Fatalf("expected *ErrTierGate, got %T: %v", err, err)
	}
	if gate.Bundle != "provision" {
		t.Errorf("Bundle = %q, want provision", gate.Bundle)
	}
	if !errors.Is(err, ErrMissing) {
		t.Errorf("expected wrap of ErrMissing, got Cause=%v", gate.Cause)
	}
}

func TestRequireBundle_DevLicenseWithBundle(t *testing.T) {
	withTempHome(t)
	tok, _ := IssueDev(License{
		Tier:      TierEnterprise,
		Bundles:   []string{"provision"},
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	})
	if err := Save(tok); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := RequireBundle("provision"); err != nil {
		t.Fatalf("RequireBundle = %v, want nil", err)
	}
}

func TestRequireBundle_DevLicenseWithoutBundle(t *testing.T) {
	withTempHome(t)
	tok, _ := IssueDev(License{
		Tier:      TierEnterprise,
		Bundles:   []string{"audit"},
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	})
	_ = Save(tok)
	err := RequireBundle("provision")
	if err == nil {
		t.Fatal("expected ErrTierGate, got nil")
	}
	if _, ok := err.(*ErrTierGate); !ok {
		t.Fatalf("expected *ErrTierGate, got %T", err)
	}
}

func TestRequireBundle_Expired(t *testing.T) {
	withTempHome(t)
	tok, _ := IssueDev(License{
		Tier:      TierEnterprise,
		Bundles:   []string{"provision"},
		ExpiresAt: time.Now().Add(-time.Hour).Unix(),
	})
	_ = Save(tok)
	err := RequireBundle("provision")
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}

func TestRequireBundle_BusinessTierRejected(t *testing.T) {
	withTempHome(t)
	tok, _ := IssueDev(License{
		Tier:      TierBusiness,
		Bundles:   []string{"provision"},
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	})
	_ = Save(tok)
	if err := RequireBundle("provision"); err == nil {
		t.Fatal("expected gate failure on business tier")
	}
}

func TestRequireTier_CommunityWithNoLicense(t *testing.T) {
	withTempHome(t)
	if err := RequireTier(TierCommunity); err != nil {
		t.Errorf("Community gate must accept no-license: got %v", err)
	}
	if err := RequireTier(TierEnterprise); err == nil {
		t.Error("Enterprise gate must reject no-license")
	}
}

func TestDecodePubKey_Variants(t *testing.T) {
	// Empty
	if _, ok := decodePubKey(nil); ok {
		t.Error("empty input must return ok=false")
	}
	// Comment-only
	if _, ok := decodePubKey([]byte("# comment\n#another\n")); ok {
		t.Error("comment-only must return ok=false")
	}
	// Random non-32-byte raw
	if _, ok := decodePubKey([]byte{1, 2, 3}); ok {
		t.Error("3-byte input must return ok=false")
	}
}
