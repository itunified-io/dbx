# pkg/license/keys/

This directory holds public Ed25519 verification keys for dbx licenses.

## prod.pub

`prod.pub` is the production license verification key.

**Status: PLACEHOLDER**. This file is intentionally empty until the dbx
license CA is provisioned. While empty:

- Production-issued licenses CANNOT verify (there is no key to verify
  against).
- Only DEV-MODE licenses (self-signed locally and trusted via
  `~/.dbx/.trust/`) verify successfully.
- All tier-gated commands therefore require a dev-issued license,
  printed with a `WARNING: dev-issued license` banner.

When the production CA is ready, replace this file with the
PEM-encoded Ed25519 public key. The verifier reads it via Go's
`embed.FS` at compile time, so a clean rebuild ships the new key.

## Format

Two formats are supported:

1. **Raw 32-byte** Ed25519 public key (`ed25519.PublicKeySize`).
2. **Base64-encoded** (standard or raw URL) 32-byte key, with optional
   `# ...` comment lines stripped before decoding.

The verifier auto-detects format. Empty / whitespace-only files are
treated as "no production key" and silently skipped (so the binary
still builds and dev licenses still work).
