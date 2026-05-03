package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// jwtHeader is the fixed header for dbx licenses. Algorithm is pinned
// to EdDSA — verifiers MUST reject any other alg to prevent confusion
// attacks (e.g., none / HS256 with the public key as HMAC secret).
type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// signJWT produces a compact JWS (header.payload.sig, base64url) for
// the given license claims using the supplied Ed25519 private key.
//
// The output is a single self-contained string suitable for writing
// to ~/.dbx/license.jwt.
func signJWT(claims *License, priv ed25519.PrivateKey) (string, error) {
	if priv == nil {
		return "", errors.New("license: signJWT: nil private key")
	}
	hdrJSON, err := json.Marshal(jwtHeader{Alg: "EdDSA", Typ: "JWT"})
	if err != nil {
		return "", fmt.Errorf("license: marshal header: %w", err)
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("license: marshal claims: %w", err)
	}
	hdr := base64.RawURLEncoding.EncodeToString(hdrJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := hdr + "." + payload
	sig := ed25519.Sign(priv, []byte(signingInput))
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// verifyJWT parses and verifies a compact JWS produced by signJWT
// against any of the supplied trusted public keys.
//
// On success it returns the decoded License. On failure it returns
// ErrInvalidFormat, ErrInvalidSignature, or a wrapping error.
//
// The function is deliberately strict: it rejects unknown alg values
// (alg pinning) and requires every key in trusted to be 32 bytes.
func verifyJWT(token string, trusted []ed25519.PublicKey) (*License, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidFormat
	}
	hdrBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: header b64: %v", ErrInvalidFormat, err)
	}
	var hdr jwtHeader
	if err := json.Unmarshal(hdrBytes, &hdr); err != nil {
		return nil, fmt.Errorf("%w: header json: %v", ErrInvalidFormat, err)
	}
	// Algorithm pinning. Reject anything other than EdDSA outright.
	if hdr.Alg != "EdDSA" {
		return nil, fmt.Errorf("%w: unexpected alg %q", ErrInvalidFormat, hdr.Alg)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: payload b64: %v", ErrInvalidFormat, err)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: sig b64: %v", ErrInvalidFormat, err)
	}

	signingInput := []byte(parts[0] + "." + parts[1])

	verified := false
	for _, pub := range trusted {
		if len(pub) != ed25519.PublicKeySize {
			continue
		}
		if ed25519.Verify(pub, signingInput, sig) {
			verified = true
			break
		}
	}
	if !verified {
		return nil, ErrInvalidSignature
	}

	var lic License
	if err := json.Unmarshal(payloadBytes, &lic); err != nil {
		return nil, fmt.Errorf("%w: claims json: %v", ErrInvalidFormat, err)
	}
	return &lic, nil
}
