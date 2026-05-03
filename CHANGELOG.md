# Changelog

All notable changes to this project will be documented in this file.

## v2026.05.03.2 — 2026-05-03

### feat(oracle/sql): db sql exec-readwrite (#29) — privileged DDL/DML/PL-SQL via sqlplus / as sysdba

New cobra leaf `dbxcli db sql exec-readwrite` + new package
`pkg/oracle/sql/` for SSH-wrapped sqlplus execution. Required for
/lab-up Phase E.1 (FORCE LOGGING, FLASHBACK ON, ADD STANDBY LOGFILE)
and Phase E.2 (RECOVER MANAGED STANDBY DATABASE) — until now those
steps fell back to ad-hoc `ssh oracle@host 'sqlplus / as sysdba <<EOF'`
which broke the audit chain + OTEL story.

- New package `pkg/oracle/sql/`:
  - `exec_readwrite.go` — `ExecReadWrite(ctx, target, sql, opts)`
    shells out via `host.Executor` (SSH) to
    `sudo -u oracle bash -lc 'env ORACLE_SID=... ORACLE_HOME=... sqlplus -s / as sysdba <<HEREDOC ... HEREDOC'`.
    No read-only guard; accepts DDL/DML/anonymous PL/SQL. Returns
    `ExecResult{Statements, Stdout, Stderr, ExitCode, LogTail}`.
  - `sshexec.go` — package-private SSH executor mirroring
    `pkg/provision/install` (kept separate so executor evolution in
    each domain stays independent).
  - `exec_readwrite_test.go` — table-driven with `hosttest.MockExecutor`:
    happy path, exit-code propagation, ctx-cancel + `ErrCancelled`,
    empty SQL rejection, multi-statement audit splitting, heredoc-
    terminator injection guard (rejects reserved token AND bare `EOF`
    line), required-opts validation, log-tail bounding.
- New cobra leaf `dbxcli db sql exec-readwrite`
  (`cmd/dbxcli/root/db_sql.go`):
  - Flags: `--target`, `--sql | --sql-file`, `--oracle-sid`,
    `--oracle-home`, `--format json|table`, `--log-tail N`.
  - License gate: `license.RequireBundle("provision")` (Enterprise
    tier — same gate as install primitives, since FORCE LOGGING etc.
    are admin ops). Same pattern as `provision install` leaves.
  - Default output is human-readable (exit_code + stdout/stderr
    sections); `--format json` returns the structured `ExecResult`.
- **Deviation note**: The original task brief proposed adding
  `OracleSID` + `OracleHome` to `pkg/core/target.Target`. Deferred:
  the unified Target struct does not yet model per-database SID/Home
  pairs, and threading them through `ExecOptions` (caller-provided
  from the env.yaml) keeps this PR scoped. Target-struct extension
  can land as a follow-up when other commands also need SID/Home.

Closes #29.

## v2026.05.03.1 — 2026-05-03

### feat(license): pkg/license/ + RequireBundle gating on 8 provision install primitives (#27)

Slim Ed25519-JWT tier-gate facade for the dbxcli, plus enforcement
on every Oracle install primitive. Closes the TODO(#519) markers
introduced in PR #28.

- New package `pkg/license/`:
  - `types.go` — `Tier` ordering (community < business < enterprise),
    `License` with `HasBundle/HasTier/IsValid`, `ErrTierGate` sentinel
    that wraps the underlying `ErrMissing`/`ErrExpired`/...
  - `jwt.go` — minimal compact JWS (Ed25519/EdDSA only). Strict
    algorithm pinning rejects `none`, `HS256`, etc.
  - `store.go` — license at `~/.dbx/license.jwt` (mode 0600). Embedded
    production verification key at `pkg/license/keys/prod.pub` (empty
    placeholder until license CA is provisioned). Dev keys auto-trusted
    via `~/.dbx/.trust/*.pub`.
  - `issuer.go` — DEV-MODE: `IssueDev` self-signs with
    `~/.dbx/.signing-key.ed25519`, idempotent across calls. Stamps the
    `dev: true` claim so the verifier prints a warning on Load.
  - `require.go` — `RequireBundle(name)` and `RequireTier(min)`. Bundle
    gates implicitly require Enterprise tier.
- Wired `license.RequireBundle("provision")` at the top of every
  `dbxcli provision install <leaf>` RunE: grid, dbhome, root-sh, asmca,
  netca, asm-label, dbca, pdb. Replaces the placeholder
  `// TODO(#519): wire license.RequireBundle…` markers.
- New `cmd/dbxcli/root/license.go`:
  - `dbxcli license status` — table view of tier/bundles/expiry/source.
  - `dbxcli license activate <path>` — verifies + installs a JWT.
  - `dbxcli license issue …` — DEV-MODE self-sign with `--tier`,
    `--bundles`, `--subject`, `--expires`, `--out`.
- 22 new tests (sign/verify roundtrip, alg pinning, unknown-signer
  rejection, expired/wrong-tier/wrong-bundle gates, dev key idempotency,
  file mode 0600). Existing e2e test updated to accept the new
  tier-gate error path.

## v2026.05.02.1 — 2026-05-02

### feat(target): real YAML persistence at ~/.dbx/targets/<name>.yaml (#30)

`dbxcli target` is no longer a stub. Targets are persisted to disk and
the four lifecycle commands work end-to-end.

- `target add` — persists Target to `~/.dbx/targets/<entity_name>.yaml`
  (mode 0600, parent dir 0700). Previously a stub that printed params.
- `target list` — reads the store, optional `entity_type=<x>` filter,
  honours `--format table|json|yaml`.
- `target test` — loads target, runs `whoami` over SSH using existing
  `SSHConfig` (`-i <key_path>`, `BatchMode=yes`, 5s connect timeout).
- `target remove <name>` (new, also `rm`/`delete`) — idempotent file
  delete; accepts positional name or `entity_name=<name>`.

New `pkg/core/target/store.go`:
- `StoreDir()` resolves `~/.dbx/targets`
- `Save/Load/List/Remove` with `target store: …` error prefix
- Filesystem-safe name validation (regex
  `^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`) — rejects `../etc/passwd`,
  `foo/bar`, `foo bar`, empty, `.`, `..`, NUL byte
- 8 table-driven tests in `pkg/core/target/store_test.go`

Unblocks /lab-up Phase B.5 target registration shipped in
itunified-io/infrastructure PR #528.

## v2026.05.01.1 — 2026-05-01

### feat(provision): pkg/provision/install — 8 Oracle install primitives (#22)

New `dbxcli provision install …` subcommands wrapping Oracle silent
installs end-to-end. Plan 0a of the /lab-up Phase D + E master plan
(infrastructure issue #519).

Primitives (Cobra leaves):
- install grid       — runInstaller silent for Grid Infrastructure
- install dbhome     — runInstaller silent for Oracle DB home
- install root-sh    — root.sh execution + idempotency touchfile
- install asmca      — asmca silent diskgroup creation
- install netca      — netca silent listener creation
- install asm-label  — oracleasm/AFD disk labeling
- install dbca       — DBCA silent CDB creation
- install pdb        — DBCA silent PDB creation

All primitives:
- Idempotent (detect-and-skip via two-phase sentinel for non-idempotent
  ops; touchfile for idempotent root.sh)
- ctx-cancel safe (mid-run interrupt → DetectionStatePartial)
- Version-agnostic detection (no version-string substring matches)
- Shell-injection hardened (control-char + metachar Validate; shellEscape
  on every interpolated arg)
- `env ORACLE_HOME=<home> <home>/bin/<bin>` qualified probes
- `--reset` MVP non-destructive (prints recovery runbook to stderr)
- OTEL span per invocation with `dbx.*` attributes via the package-level
  `otel.GlobalExporter()` (Plan 0a Task 10)

E2E coverage:
- Per-primitive unit tests with `host/hosttest.MockExecutor`
- `pkg/provision/install/otel_test.go` — capture-exporter verifies span
  emission for all 8 primitives (StatusOK + StatusError paths)
- `cmd/dbxcli/root/provision_install_e2e_test.go` — leaf registration,
  --help renders, required-flag rejection (Plan 0a Task 12)

Deferred to follow-ups:
- Audit chain integration (#26 — dbx `pkg/audit/` doesn't exist yet)
- License gate enforcement (#27 — `pkg/license/` doesn't expose
  `RequireBundle` yet; TODO markers remain in `provision_install.go`)

## v2026.04.30.1

### feat: pkg/otel + Target.OTELAttrs (#19)

OTEL bus integration foundation per itunified-io/infrastructure ADR-0103a (Item 8 of agentic-AI hardening roadmap; Wave B item 3).

- `pkg/otel/attrs.go` — Attribute type + standard key constants (`dbx.entity_type`, `dbx.entity_name`, `dbx.db_unique_name`, `dbx.host`, `dbx.audit_hash`, `dbx.license_tier`, plus Plan-RAG `step_id`/`skill` and Item 1+2 cross-link `decision`/`deny_rule`)
- `pkg/otel/span.go` — Span + SpanBuilder + Status + Exporter interface + NoopExporter (audit dual-sink fallback)
- `pkg/core/target/otel.go` — `Target.OTELAttrs()` returns dbx.* attributes derived from EntityType + Name + endpoints; mirrors pkg/otel constants without import cycle
- 12 unit tests across both packages

Foundation only — OTLP HTTP exporter implementation lives in sibling pkg/otel/exporter (next PR). Default emitter is NoopExporter.

## v2026.04.11.1

### P22 — Managed Agents Transport

- feat(transport): streamable HTTP transport for remote MCP connections (#384)
- feat(transport): Ed25519 JWT authentication for agent-to-central communication (#384)

### P19 — Cloud Infrastructure Provisioning

- feat(cloud): CloudProvider interface — 16 methods for instance, volume, security group, load balancer, managed DB lifecycle (#384)
- feat(cloud): provider registry with thread-safe register/get/list (#384)
- feat(cloud): blueprint YAML parser and validator for infrastructure-as-code definitions (#384)
- feat(cloud): cost estimation engine with per-provider pricing data (#384)
- feat(cloud): workload profile recommendation engine for instance type selection (#384)

### P18 — Documentation Suite

- docs: quick-start guide — install, verify, connect Oracle/PostgreSQL, first commands, MCP setup (#384)
- docs: Oracle setup guide — client libraries, TNS, wallet auth, license declaration, tier tool counts (#384)
- docs: PostgreSQL setup guide — YAML profiles, CNPG optional, tier tool counts (#384)
- docs: host/OS setup guide — supported distros, SSH target config (#384)
- docs: monitoring setup guide — dbmon-agent, dbmon-central, VictoriaMetrics, Grafana, alerts (#384)
- docs: air-gap deployment guide — bundle creation, offline deploy, offline license activation (#384)
- docs: administration guide — target management, Vault, license gate, confirm gates, audit, RBAC (#384)
- docs: master tool reference — 735 tools across 25 domains (#384)
- docs: skill reference — 112 skills with slash commands, repos, license tiers (#384)
- docs: target YAML reference — all 8 entity types with full field specs (#384)
- docs: Vault integration guide — path layout, AppRole, credential rotation (#384)
- docs: Oracle license declaration guide — editions, options, enforcement modes (#384)
- docs: troubleshooting guide — top issues per engine with solutions (#384)
- docs: release notes template with CalVer format (#384)

### P17 — RAG Subsystem

- feat(rag): embedder, indexer, vector store, search, context builder (#384)
- feat(cli): `rag` CLI subcommand — search, context, index-status, index-refresh, sources (#384)

### P16 — Host/OS Monitoring (20 tools)

- feat(host): distro abstraction layer — RHEL, Ubuntu, SLES, OL detection (#384)
- feat(host): metric collectors — CPU, memory, disk I/O, network throughput, process, load avg (#384)
- feat(host): service, security, user, network parsers (#384)
- feat(host): package management — apt/dpkg/dnf/rpm (#384)
- feat(host): kernel params/modules/hugepages, service manager (#384)
- feat(host): filesystem mounts/LVM/inodes, network routes/DNS/NTP (#384)
- feat(host): user/group/sudoers audit, journald/auth log analysis (#384)
- feat(host): AppArmor/SSH security checks, patch/ksplice assessment (#384)
- feat(cli): `host` CLI subcommand — 20 commands across 15 groups (#384)

### P15 — Distribution Pipeline

- feat: GoReleaser multi-arch build pipeline (#384)
- chore: .dockerignore for multi-arch Docker builds (#384)

### P14 — Policy Engine

- feat(policy): policy engine core — types, YAML loader, rule evaluator, report generator (#384)
- feat(policy/os): OS check executors for CIS Linux and DISA STIG (#384)
- feat(policy): Oracle and PostgreSQL SQL-based policy executors (#384)
- feat(cli): `policy` CLI subcommand — scan, report, drift, status, fleet, remediate (#384)

## v2026.04.10.7

- feat: PostgreSQL connection management — connect, disconnect, pool status, connection info, test (#11)
- feat: PostgreSQL query execution — run, explain, prepared statements (#11)
- feat: PostgreSQL schema browser — list databases/schemas/tables/views/indexes/functions/triggers/sequences/extensions (#11)
- feat: PostgreSQL CRUD operations — insert, update, delete, upsert (#11)
- feat: PostgreSQL DBA operations — vacuum, analyze, reindex, bloat, locks, activity, config, tablespace, stats reset, kill, maintenance (#11)
- feat: PostgreSQL advanced DBA — pg_stat_statements, index advisor, table partitioning, connection pooler, custom GUC (#11)
- feat: PostgreSQL performance — slow queries, cache hit ratio (#11)
- feat: PostgreSQL health check — comprehensive cluster health (#11)
- feat: PostgreSQL security — SSL status, pg_hba rules, password policy, role audit (#11)
- feat: PostgreSQL audit logging — pgaudit status/config/log query (#11)
- feat: PostgreSQL compliance — CIS benchmarks, GDPR audit, data classification, retention policy, encryption status (#11)
- feat: PostgreSQL RBAC — role list/grant/revoke, privilege audit (#11)
- feat: PostgreSQL replication — streaming status, slots, lag monitor, switchover (#11)
- feat: PostgreSQL HA — Patroni status/switchover/reinit/restart, pgBouncer, HAProxy health, connection routing, failover test, split brain, witness, timeline (#11)
- feat: PostgreSQL backup — pg_basebackup, pgBackRest status/restore, WAL archive check (#11)
- feat: PostgreSQL migration — pg_dump/pg_restore, pg_upgrade, logical replication setup (#11)
- feat: PostgreSQL observability — pg_stat_activity, wait events, log tail, custom metrics (#11)
- feat: PostgreSQL multi-tenant — tenant create/list/isolate/resource limits/connection pool (#11)
- feat: PostgreSQL WAL management — WAL status, archive, retention, replay, size (#11)
- feat: PostgreSQL CNPG — cluster status/failover/backup/restore/hibernate/promote (#11)
- feat: PostgreSQL cross-cluster DR — 18 tools for S3 config, WAL shipping, base backup, PITR, promote, validate, monitor, failback, switchover, clone, retention, encryption, compression, bandwidth, parallel, cleanup, status, test (#11)
- feat: PostgreSQL RAG — pgvector operations, embedding, semantic search, collection, index, hybrid search, metadata filter (#11)
- feat: PostgreSQL Vault integration — credential rotate, dynamic secrets, lease management (#11)
- feat: PostgreSQL policy engine — OPA evaluate, policy sync (#11)
- feat: `pg` CLI command group with 24 subcommands (137 actions) (#11)
- feat: shared PostgreSQL query helpers and K8s/kubectl utilities (#11)

## v2026.04.10.6

- feat: `cmd/docgen` — Cobra doc generation for LLM-friendly CLI reference (#9)
- feat: `make docs` target generates `docs/cli/*.md` (70 files) + `llms.txt` (#9)
- feat: enrich all OSS commands with `Long` descriptions and `Example` strings (#9)
- docs: add Documentation Generation section to README.md (#9)

## v2026.04.10.5

- feat: add `rac` SSH domain with `srvctl` to default allowlist (#8)

## v2026.04.10.4

- feat: add `dataguard` SSH domain with `dgmgrl` to default allowlist (#7)

## v2026.04.10.3

- feat: Oracle Linux package management — rpm list/info, dnf install/update (#5)
- feat: Linux kernel parameter management — sysctl list/set, hugepages, OS info (#5)
- feat: Linux storage/LVM management — pv/vg/lv list, lv create, disk usage (#5)
- feat: Linux network diagnostics — NIC list, bond status, DNS check, NTP status (#5)
- feat: Linux security status — SELinux, firewall, service audit (#5)
- feat: Extended SSH allowlist for P4 commands (ip, nmcli, chronyc, sestatus, firewall-cmd, lsblk) (#5)
- feat: `linux` CLI command group with 5 subcommands (20 actions) (#5)

## v2026.04.10.2

- feat: Oracle read-only session operations — list, describe, top waiters (#3)
- feat: Oracle read-only tablespace operations — list, describe, usage summary (#3)
- feat: Oracle read-only user operations — list, describe, profiles (#3)
- feat: Oracle read-only schema browser — list, objects, describe (#3)
- feat: Read-only SQL execution with SELECT-only guard and EXPLAIN PLAN (#3)
- feat: Oracle redo log operations — list groups, switch history (#3)
- feat: Oracle undo/rollback operations — list, segment info (#3)
- feat: Oracle init parameter operations — list, describe, modified, hidden (#3)
- feat: Oracle advisor operations — segment advisor, SQL tuning list (#3)
- feat: `db` CLI command group with 9 Oracle subcommands (24 actions) (#3)
- feat: Shared query helpers (QueryRows, QueryRow) for map[string]any results (#3)

## v2026.04.10.1

- feat: Cobra CLI skeleton with exported root command for downstream repos (#1)
- feat: config package — YAML + env var loading with sensible defaults (#1)
- feat: target model — Oracle + PostgreSQL entity types with YAML parsing (#1)
- feat: target registry — load from ~/.dbx/targets/, resolve by entity_type (#1)
- feat: vault client — AppRole fetcher with credential caching and fallback (#1)
- feat: Ed25519 license validation with 14-day grace period (#1)
- feat: Oracle license gate — edition/option enforcement (strict/warn/audit-only) (#1)
- feat: OEM Management Pack gate (#1)
- feat: confirm gate — echo-back and double-confirm patterns (#1)
- feat: multi-sink audit trail with redaction (#1)
- feat: SSH execution with allowlist-based security model (#1)
- feat: 9-stage execution pipeline orchestrator (#1)
- feat: output formatter — table, JSON, YAML (#1)
- feat: REST API skeleton (net/http, /health, /api/v1/version, /api/v1/targets) (#1)
- feat: MCP adapter skeleton (JSON-RPC stdio, tool registry) (#1)
- feat: connection manager interface with engine stubs (#1)
