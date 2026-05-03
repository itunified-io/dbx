package license

// DEV-MODE ONLY
//
// This file implements local self-signed license issuance for lab
// and development use. Production licenses MUST be issued by the
// dbx license CA (TBD), which holds an offline Ed25519 signing key.
//
// On first invocation, IssueDev:
//   1. Generates an Ed25519 keypair into ~/.dbx/.signing-key.ed25519 (mode 0600).
//   2. Writes the matching public key to ~/.dbx/.trust/dev-<fingerprint>.pub (mode 0644).
//   3. Signs a License with the freshly-generated private key.
//   4. Returns the compact JWS.
//
// Subsequent invocations reuse the same private key.

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// IssueDev signs a License with a locally-generated Ed25519 key,
// auto-trusting it via ~/.dbx/.trust/.
//
// The supplied claims have their Dev flag forced to true so the
// verifier prints a warning when the resulting JWT is later loaded.
//
// IssueDev returns the compact JWS string. Callers wishing to
// persist it to ~/.dbx/license.jwt should pass the result to Save.
func IssueDev(claims License) (string, error) {
	if claims.Tier == "" {
		return "", errors.New("license: IssueDev: tier required")
	}
	priv, err := loadOrCreateDevKey()
	if err != nil {
		return "", err
	}
	claims.Dev = true // DEV-MODE: stamp every dev-issued license
	return signJWT(&claims, priv)
}

// loadOrCreateDevKey returns the dev signing private key, creating
// it (and its trusted public counterpart) on first call.
func loadOrCreateDevKey() (ed25519.PrivateKey, error) {
	if err := os.MkdirAll(dbxDir(), 0o700); err != nil {
		return nil, fmt.Errorf("license: mkdir %s: %w", dbxDir(), err)
	}
	if err := os.MkdirAll(trustDir(), 0o700); err != nil {
		return nil, fmt.Errorf("license: mkdir %s: %w", trustDir(), err)
	}

	if data, err := os.ReadFile(devKeyPath()); err == nil {
		if len(data) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("license: dev key %s: unexpected size %d", devKeyPath(), len(data))
		}
		return ed25519.PrivateKey(append([]byte(nil), data...)), nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("license: read dev key: %w", err)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("license: generate dev key: %w", err)
	}
	if err := os.WriteFile(devKeyPath(), priv, 0o600); err != nil {
		return nil, fmt.Errorf("license: write dev key: %w", err)
	}

	// Public key fingerprint (first 8 bytes of sha256, hex-encoded)
	// keeps trust filenames short + collision-resistant for human eyes.
	sum := sha256.Sum256(pub)
	fp := hex.EncodeToString(sum[:8])
	pubFile := filepath.Join(trustDir(), "dev-"+fp+".pub")
	pubB64 := base64.StdEncoding.EncodeToString(pub)
	contents := "# DEV-MODE Ed25519 dev license verification key\n" +
		"# Fingerprint: " + fp + "\n" +
		pubB64 + "\n"
	if err := os.WriteFile(pubFile, []byte(contents), 0o644); err != nil {
		return nil, fmt.Errorf("license: write trust pub: %w", err)
	}
	return priv, nil
}
