package license

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
)

func mustKeypair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey: %v", err)
	}
	return pub, priv
}

func TestSignVerify_Roundtrip(t *testing.T) {
	pub, priv := mustKeypair(t)
	claims := &License{
		Subject: "lab-dev",
		Tier:    TierEnterprise,
		Bundles: []string{"provision", "dataguard"},
	}
	tok, err := signJWT(claims, priv)
	if err != nil {
		t.Fatalf("signJWT: %v", err)
	}
	if strings.Count(tok, ".") != 2 {
		t.Fatalf("expected compact JWS with 2 dots, got %q", tok)
	}
	got, err := verifyJWT(tok, []ed25519.PublicKey{pub})
	if err != nil {
		t.Fatalf("verifyJWT: %v", err)
	}
	if got.Subject != "lab-dev" || got.Tier != TierEnterprise || len(got.Bundles) != 2 {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}

func TestVerify_RejectsUnknownSigner(t *testing.T) {
	_, priv := mustKeypair(t)
	otherPub, _ := mustKeypair(t)
	tok, err := signJWT(&License{Subject: "x", Tier: TierEnterprise}, priv)
	if err != nil {
		t.Fatalf("signJWT: %v", err)
	}
	_, err = verifyJWT(tok, []ed25519.PublicKey{otherPub})
	if !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestVerify_RejectsWrongAlg(t *testing.T) {
	pub, _ := mustKeypair(t)
	// Hand-crafted token with alg "none"
	tok := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ4In0." // base64url of {"alg":"none","typ":"JWT"}.{"sub":"x"}.empty
	_, err := verifyJWT(tok, []ed25519.PublicKey{pub})
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat (alg pin), got %v", err)
	}
}

func TestVerify_RejectsMalformed(t *testing.T) {
	pub, _ := mustKeypair(t)
	for _, bad := range []string{"", "abc", "a.b", "a.b.c.d", "!.@.#"} {
		if _, err := verifyJWT(bad, []ed25519.PublicKey{pub}); !errors.Is(err, ErrInvalidFormat) {
			t.Errorf("bad input %q: expected ErrInvalidFormat, got %v", bad, err)
		}
	}
}

func TestVerify_TrustedSubsetMatches(t *testing.T) {
	pubA, privA := mustKeypair(t)
	pubB, _ := mustKeypair(t)
	tok, _ := signJWT(&License{Subject: "x", Tier: TierEnterprise}, privA)
	// Trust list contains B first then A — must still find A.
	if _, err := verifyJWT(tok, []ed25519.PublicKey{pubB, pubA}); err != nil {
		t.Fatalf("expected verify success with subset, got %v", err)
	}
}
