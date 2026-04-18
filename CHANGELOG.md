# Changelog

All notable changes to this project will be documented in this file.

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
