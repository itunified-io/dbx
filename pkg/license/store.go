package license

import (
	"crypto/ed25519"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed keys/prod.pub
var embeddedProdKey []byte

// pathOverride lets tests redirect Path() to a temp directory.
// Production callers MUST NOT touch this — they use Path() directly.
var pathOverride string

// Path returns the on-disk path of the dbx license file.
//
// Default: $HOME/.dbx/license.jwt. The directory is NOT created here;
// callers that intend to write (Save, IssueDev) are responsible for
// MkdirAll on the parent.
func Path() string {
	if pathOverride != "" {
		return filepath.Join(pathOverride, "license.jwt")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Best-effort fallback: cwd. Caller errors will surface clearly.
		return ".dbx/license.jwt"
	}
	return filepath.Join(home, ".dbx", "license.jwt")
}

// dbxDir returns the on-disk dbx config dir (parent of license.jwt).
func dbxDir() string {
	return filepath.Dir(Path())
}

// trustDir returns the directory of trusted dev public keys.
func trustDir() string {
	return filepath.Join(dbxDir(), ".trust")
}

// devKeyPath returns the path of the local dev signing private key.
func devKeyPath() string {
	return filepath.Join(dbxDir(), ".signing-key.ed25519")
}

// trustedKeys assembles the verifier trust list:
//   - the embedded production key (if non-empty)
//   - every *.pub file under ~/.dbx/.trust/ (dev keys)
func trustedKeys() []ed25519.PublicKey {
	var keys []ed25519.PublicKey
	if pk, ok := decodePubKey(embeddedProdKey); ok {
		keys = append(keys, pk)
	}
	matches, _ := filepath.Glob(filepath.Join(trustDir(), "*.pub"))
	for _, m := range matches {
		data, err := os.ReadFile(m)
		if err != nil {
			continue
		}
		if pk, ok := decodePubKey(data); ok {
			keys = append(keys, pk)
		}
	}
	return keys
}

// decodePubKey parses an Ed25519 public key from a file payload.
// It accepts:
//   - raw 32-byte binary
//   - base64 (std or raw URL), with #-comment lines stripped
//
// Empty / whitespace-only input returns ok=false.
func decodePubKey(data []byte) (ed25519.PublicKey, bool) {
	if len(data) == ed25519.PublicKeySize {
		return ed25519.PublicKey(append([]byte(nil), data...)), true
	}
	// Strip comments + whitespace.
	var sb strings.Builder
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		sb.WriteString(line)
	}
	s := sb.String()
	if s == "" {
		return nil, false
	}
	for _, dec := range []func(string) ([]byte, error){
		base64.StdEncoding.DecodeString,
		base64.RawStdEncoding.DecodeString,
		base64.URLEncoding.DecodeString,
		base64.RawURLEncoding.DecodeString,
	} {
		if raw, err := dec(s); err == nil && len(raw) == ed25519.PublicKeySize {
			return ed25519.PublicKey(raw), true
		}
	}
	return nil, false
}

// Load reads, parses, and verifies the dbx license at Path().
//
// Returns:
//   - (nil, ErrMissing)    — no file present; caller should treat as TierCommunity.
//   - (nil, ErrInvalidFormat | ErrInvalidSignature) — file present but unusable.
//   - (license, nil)       — license verified by a trusted key. Expiry NOT enforced
//     here; call lic.IsValid() / RequireBundle / RequireTier for that.
//
// Dev-issued licenses cause a one-line warning to be written to stderr.
func Load() (*License, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrMissing
		}
		return nil, fmt.Errorf("license: read %s: %w", Path(), err)
	}
	tok := strings.TrimSpace(string(data))
	if tok == "" {
		return nil, ErrMissing
	}
	keys := trustedKeys()
	if len(keys) == 0 {
		// No prod key embedded and no dev keys trusted. Any license
		// in this state must be treated as untrusted.
		return nil, fmt.Errorf("%w: no trusted verification keys (no prod.pub, no dev trust)", ErrInvalidSignature)
	}
	lic, err := verifyJWT(tok, keys)
	if err != nil {
		return nil, err
	}
	if lic.Dev {
		fmt.Fprintln(os.Stderr, "WARNING: dev-issued license loaded from "+Path()+" — not for production use")
	}
	return lic, nil
}

// Save writes the JWT string to Path() with mode 0600. The parent
// directory is created (mode 0700) if it does not exist.
func Save(jwt string) error {
	if err := os.MkdirAll(dbxDir(), 0o700); err != nil {
		return fmt.Errorf("license: mkdir %s: %w", dbxDir(), err)
	}
	if err := os.WriteFile(Path(), []byte(jwt), 0o600); err != nil {
		return fmt.Errorf("license: write %s: %w", Path(), err)
	}
	return nil
}
