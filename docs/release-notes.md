# Release Notes

dbx follows CalVer versioning: `YYYY.MM.DD.TS` where `TS` is an incrementing sequence number for multiple releases on the same calendar day.

---

## Release Template

Use this template for each new release entry. Place new entries at the top of the file, above all previous entries.

```markdown
## v<YYYY.MM.DD.TS>

**Released:** <YYYY-MM-DD>

### New Features

- Brief description of what was added and why it matters. Reference the issue or plan number where applicable.

### Improvements

- Enhancement to an existing feature, performance improvement, or UX change.

### Bug Fixes

- Description of the bug and what was fixed. Include error messages or symptoms where helpful.

### Breaking Changes

- Description of the breaking change, what it affects, and the migration path.

### Security Fixes

- Brief description (do not disclose exploit details). Reference CVE if assigned.

### Deprecations

- Feature or flag being deprecated, the replacement, and the planned removal version.
```

Sections with no entries for a given release are omitted.

---

## v2026.04.10.1

**Released:** 2026-04-10

This is the initial public release of dbx, shipping the foundational platform components from implementation plans P14 through P17. The release establishes the CLI binary (`dbxcli`), licensing infrastructure, host monitoring, compliance policy engine, and the RAG subsystem.

### New Features

#### Auth and Licensing (P14)

- **Ed25519 JWT licensing**: dbx license keys are Ed25519-signed JWTs. The CLI binary verifies the signature locally against the embedded public key without requiring network access for the initial check. License metadata (plan tier, seat count, expiry) is embedded in the token claims.
- **Phone-home validation**: the CLI performs a daily HTTPS check-in to the license validation endpoint. On success, the local cache is updated and the grace period clock is reset. The check is non-blocking and does not delay command execution.
- **Offline grace period**: when phone-home fails, licensed tools remain available for 72 hours. Grace period status and remaining time are reported by `dbxcli license status`. After expiry, licensed tools refuse to execute until connectivity is restored or an offline token is applied.
- **Offline activation**: air-gapped deployments generate a machine fingerprint with `dbxcli license offline-fingerprint`. The fingerprint is exchanged for a signed offline token through the customer portal. Offline tokens encode validity windows and are renewed via the same exchange process.
- **Plan tiers**: Free, Standard, and Enterprise tiers gate access to tool groups. Tool documentation indicates the minimum required plan. `dbxcli license status` shows the active plan and available bundles.
- **Fleet licensing**: Enterprise fleet keys activate multiple machines under a single key. Seat consumption is reported via `dbxcli license status --fleet`.

#### Distribution (P14)

- **Homebrew tap**: macOS and Linux users can install the `dbxcli` binary via `brew tap itunified-io/dbx && brew install dbx`. The tap is updated on each release.
- **Docker image**: `ghcr.io/itunified-io/dbx:latest` and `ghcr.io/itunified-io/dbx:v2026.04.10.1` are published to GitHub Container Registry. Multi-architecture images are available for `linux/amd64` and `linux/arm64`.
- **npm MCP adapters**: `@itunified.io/mcp-oracle` and `@itunified.io/mcp-postgres` are published to npm. These packages expose the dbx tool inventory as MCP servers for use with Claude and other MCP-compatible AI assistants. The npm packages do not include the `dbxcli` binary.
- **Air-gap bundle**: a self-contained tar archive containing the `dbxcli` binary, Oracle Instant Client (Basic, 21c), required shared libraries, and offline activation tooling is available for download from the customer portal. Supported platforms: `linux/amd64`, `linux/arm64`.

#### Host Monitoring (P15)

- **SSH-based host runner**: all host operations execute over SSH using the `oracle_host` or `host` target type. No agent binary needs to be deployed on monitored hosts. The SSH key is retrieved from Vault at connection time.
- **20 OSS host tools** across five categories:

  | Category | Tools |
  |----------|-------|
  | `linux kernel` | `info`, `hugepages`, `param-list`, `param-set` |
  | `linux network` | `nic-list`, `bond-status`, `ntp-status`, `dns-check` |
  | `linux package` | `list`, `info`, `install`, `update` |
  | `linux security` | `service-status`, `selinux-status`, `firewall-list` |
  | `linux storage` | `disk-usage`, `pv-list`, `vg-list`, `lv-list`, `lv-create` |

- **Oracle environment detection**: on `oracle_host` targets, `linux kernel info` additionally reports `ORACLE_HOME`, `ORACLE_SID`, and inventory registration status.
- **Huge page validation**: `linux kernel hugepages` reports current huge page allocation and the recommended value based on the Oracle SGA configuration of all databases on the host. A `--fix` flag applies the recommended value via `sysctl`.
- **NTP drift alerting**: `linux network ntp-status` reports NTP sync status, stratum, and offset. Offsets exceeding 100 ms generate a warning; offsets exceeding 1000 ms return an error exit code, enabling use in monitoring scripts and CI pipelines.

#### Policy Engine (P16)

- **CIS Benchmark policies**: built-in policy bundles for Oracle Database 19c CIS Benchmark Level 1 and Level 2, and PostgreSQL 15/16 CIS Benchmark. Policies are versioned YAML files bundled with the binary.
- **STIG policies**: DISA STIG for Oracle Database 19c (V2R2) policy bundle included. STIG IDs are mapped to dbx check identifiers for cross-reference.
- **Custom policies**: operators can author custom policies in YAML and place them in `~/.dbx/policies/`. The policy schema supports `check_id`, `description`, `severity`, `engine`, `query`, `expected_result`, and `remediation` fields.
- **`dbxcli policy check`**: runs a named policy or all policies matching a filter against the active target. Returns pass/fail per check with evidence (actual vs. expected value).
- **`dbxcli policy list`**: lists available built-in and custom policies with metadata.
- **`dbxcli policy report`**: generates a structured compliance report (table, JSON, or Markdown) summarizing pass/fail counts by severity and category. Suitable for inclusion in audit packages.
- **Drift detection**: policies that declare `baseline: true` record the expected state on first pass. Subsequent runs compare against the recorded baseline and flag deviations. Useful for detecting configuration drift between scheduled audits.
- **PostgreSQL policy integration**: `dbxcli pg policy check` and `dbxcli pg policy list` expose the same policy engine for PostgreSQL targets.

#### RAG Subsystem (P17)

- **Multi-provider embedding support**: the RAG subsystem integrates with OpenAI (`text-embedding-3-small`, `text-embedding-3-large`), Cohere (`embed-english-v3.0`), and a local Ollama-compatible endpoint for air-gapped deployments. The provider is configured per collection.
- **pgvector storage**: embeddings are stored in a PostgreSQL table with a `vector` column backed by `pgvector`. The target database must have the `pgvector` extension installed (`CREATE EXTENSION vector`).
- **Semantic search**: `dbxcli pg rag search` performs approximate nearest-neighbour search using `pgvector`'s HNSW or IVFFlat indexes.
- **Hybrid search**: combines vector similarity with full-text search (`tsvector`/`tsquery`) using Reciprocal Rank Fusion (RRF). The `--hybrid` flag enables this mode. Hybrid search improves recall for queries that mix semantic intent with precise technical terms.
- **Ingest pipeline**: `dbxcli pg rag ingest` accepts plain text files, Markdown files, and SQL scripts. Text is chunked with configurable `chunk_size` and `chunk_overlap` parameters, embedded, and written to the collection table.
- **Collection management**: `dbxcli pg rag collection-create`, `collection-drop`, and `collections` manage named vector collections within a PostgreSQL database. Each collection stores its embedding provider, model, dimension, and index configuration as metadata.
- **Index status and rebuild**: `dbxcli pg rag index-status` reports the index type, build progress, and approximate entry count. `dbxcli pg rag index-rebuild` drops and recreates the index (useful after bulk ingest).
- **Configuration**: `dbxcli pg rag config` shows and sets the active provider, model, and endpoint URL for the current target.

### New CLI Commands

The following top-level command groups are available in this release:

| Command | Description |
|---------|-------------|
| `dbxcli target` | Register, list, test, and manage connection targets |
| `dbxcli db` | Oracle database operations (sessions, tablespaces, redo, undo, schema, advisors, SQL execution) |
| `dbxcli linux` | Linux host operations over SSH (kernel, network, package, security, storage) |
| `dbxcli host` | Host information and diagnostics (maps to `linux` subcommands for Linux targets) |
| `dbxcli pg` | PostgreSQL operations (connection, DBA, schema, replication, HA, backup, CNPG, RAG, policy, audit, RBAC, security, WAL, DR, migration) |
| `dbxcli serve` | Start the MCP server in stdio or HTTP/SSE transport mode |
| `dbxcli mcp` | MCP server management (list tools, validate config, show transport status) |
| `dbxcli license` | License activation, status, phone-home, and fleet management |
| `dbxcli policy` | Policy check, list, and report (Oracle; `pg policy` for PostgreSQL) |
| `dbxcli pg rag` | RAG subsystem: collections, ingest, search, index, and configuration |

Full command reference: [CLI Reference](cli/dbxcli.md)

### Known Limitations

- Oracle RAC-aware operations (`db rac`) are planned for a subsequent release. The `rac_database` target type is registered and validated but RAC-specific tooling is not yet included.
- The GoldenGate subsystem (`db gg`) is planned and the `oracle_host` target `gg_endpoint` field is reserved but not yet wired.
- OEM integration tools are planned for Enterprise tier. The `oracle_host` target `oem_endpoint` field is reserved.
- Policy drift detection baseline comparison requires two runs against the same target. First-run baselines are stored in `~/.dbx/policy-baselines/` and are not synchronized between dbx installations.
- Air-gap bundle is available for `linux/amd64` and `linux/arm64`. Windows air-gap bundle is planned.

---

## SEE ALSO

- [Quick Start](quick-start.md) — Get started in 5 minutes
- [Target YAML Reference](config/targets.md) — Full target schema
- [Vault Integration](config/vault.md) — Credential management
- [Oracle License Declaration](config/oracle-license.md) — License gating and compliance
- [Troubleshooting Guide](troubleshooting.md) — Common issues and resolutions
