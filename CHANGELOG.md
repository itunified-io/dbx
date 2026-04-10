# Changelog

All notable changes to this project will be documented in this file.

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
