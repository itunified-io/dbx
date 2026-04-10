# dbx — CLAUDE.md

## Project Overview

dbx is a multi-database lifecycle management framework written in Go. It provides CLI (`dbxcli`), service controller (`dbxctl`), REST API, and Go library for managing Oracle, PostgreSQL, Host/OS, and Engineered Systems databases.

- **Language:** Go 1.24+
- **License:** AGPL-3.0 (public, OSS)
- **Module:** `github.com/itunified-io/dbx`

## Architecture

```
cmd/
  dbxcli/          # CLI binary (Cobra)
  dbxctl/          # Service controller binary
pkg/
  config/          # Configuration loading (YAML + env + flags)
  target/          # Target model (connection registry)
  engine/          # Engine interface + implementations
  pipeline/        # 9-stage execution pipeline
  audit/           # Structured audit trail
  license/         # JWT license validation (Ed25519)
  vault/           # HashiCorp Vault client
  api/             # REST API (net/http)
  mcp/             # MCP JSON-RPC adapter
internal/
  version/         # Build version injection
```

## Git Workflow

### Branching — NEVER work on main
- `main` branch = current production state, **protected**
- **ALL changes** via feature branches + PR
- Branching: `feature/<issue-nr>-<description>`, `fix/<issue-nr>-<description>`

### GitHub Issues — mandatory
- **Every change must have a GitHub issue**
- Commit messages reference the issue: `feat: add target model (#3)`

### Versioning (CalVer)
- Schema: `YYYY.MM.DD.TS` (e.g., `v2026.04.10.1`)

### Release Workflow — MANDATORY after every PR merge
1. Update CHANGELOG.md
2. Create annotated git tag
3. Push tag
4. Create GitHub release

## Build & Test

```bash
make build     # Build dbxcli + dbxctl
make test      # Run tests with race detector
make lint      # golangci-lint
make clean     # Remove binaries
```

## Conventions

- **Language:** English only (code, docs, commits)
- **Public repo** — no real hostnames, IPs, or secrets (use placeholders)
- Go standard project layout
- Table-driven tests
- Errors: `fmt.Errorf("operation: %w", err)` wrapping
- No global state — pass dependencies via constructors
