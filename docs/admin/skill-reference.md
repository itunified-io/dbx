# dbx Skill Reference

Complete reference for all 112 Claude Code skills provided by the dbx multi-database management platform. Skills are slash commands that orchestrate one or more MCP tools through a structured workflow to accomplish a goal in a single invocation.

## Contents

- [Conventions](#conventions)
- [Skill Count Summary](#skill-count-summary)
- [Oracle OSS Skills](#oracle-oss-skills-mcp-oracle)
- [Oracle Linux Skills](#oracle-linux-skills-mcp-oracle-ol)
- [Oracle EE Base Skills](#oracle-ee-base-skills-mcp-oracle-ee)
- [Performance Skills](#performance-skills-mcp-oracle-ee-performance)
- [Audit Skills](#audit-skills-mcp-oracle-ee-audit)
- [Partitioning Skills](#partitioning-skills-mcp-oracle-ee-partitioning)
- [Data Guard Skills](#data-guard-skills-mcp-oracle-ee-dataguard)
- [Backup Skills](#backup-skills-mcp-oracle-ee-backup)
- [RAC Skills](#rac-skills-mcp-oracle-ee-rac)
- [Clusterware Skills](#clusterware-skills-mcp-oracle-ee-clusterware)
- [ASM Skills](#asm-skills-mcp-oracle-ee-asm)
- [Provisioning Skills](#provisioning-skills-mcp-oracle-ee-provision)
- [Patching Skills](#patching-skills-mcp-oracle-ee-patch)
- [Migration Skills](#migration-skills-mcp-oracle-ee-migration)
- [Data Pump Skills](#data-pump-skills-mcp-oracle-ee-datapump)
- [GoldenGate Skills](#goldengate-skills-mcp-oracle-ee-goldengate)
- [OEM Skills](#oem-skills-mcp-oracle-ee-oem)
- [PostgreSQL Enterprise Skills](#postgresql-enterprise-skills-mcp-postgres-enterprise)
- [Host OSS Skills](#host-oss-skills-mcp-host)
- [Host Enterprise Skills](#host-enterprise-skills-mcp-host-enterprise)
- [Monitoring Agent Skills](#monitoring-agent-skills-mcp-dbmonitor)
- [Monitoring Central Skills](#monitoring-central-skills-mcp-dbmonitor-ee)
- [Alphabetical Skill Index](#alphabetical-skill-index)

---

## Conventions

Each skill entry documents the following attributes:

| Attribute | Description |
|-----------|-------------|
| **Slash Command** | The trigger used in Claude Code (e.g., `/ora-health`) |
| **Repo** | The MCP adapter repo that provides the underlying tools |
| **License** | Required license tier: `OSS` (free), `EE` (Enterprise Edition), `EE+` (advanced EE module) |
| **Duration** | Typical wall-clock execution time |
| **Tools Used** | The MCP tools the skill orchestrates, in invocation order |
| **Workflow** | The named phases the skill executes |

### License Tiers

| Tier | Description |
|------|-------------|
| `OSS` | Open-source, no license key required |
| `EE` | Enterprise Edition — requires valid JWT license |
| `EE+` | Advanced EE module — requires EE license with specific feature flag |

### Confirm Gates

Skills that perform destructive or irreversible operations enforce one of two confirm gates before proceeding:

- **Standard confirm** — Single acknowledgement prompt. Used for controlled-risk operations (e.g., switchover).
- **Double-Confirm** — Two separate acknowledgement prompts with an explicit entity name echo. Used for data-destructive operations (e.g., failover, point-in-time recovery, destructive restores).

---

## Skill Count Summary

| Engine | Repo(s) | Skills |
|--------|---------|--------|
| Oracle OSS | mcp-oracle | 4 |
| Oracle Linux | mcp-oracle-ol | 3 |
| Oracle EE Base | mcp-oracle-ee | 5 |
| Performance | mcp-oracle-ee-performance | 4 |
| Audit | mcp-oracle-ee-audit | 3 |
| Partitioning | mcp-oracle-ee-partitioning | 3 |
| Data Guard | mcp-oracle-ee-dataguard | 4 |
| Backup | mcp-oracle-ee-backup | 5 |
| RAC | mcp-oracle-ee-rac | 3 |
| Clusterware | mcp-oracle-ee-clusterware | 3 |
| ASM | mcp-oracle-ee-asm | 3 |
| Provisioning | mcp-oracle-ee-provision | 3 |
| Patching | mcp-oracle-ee-patch | 4 |
| Migration | mcp-oracle-ee-migration | 3 |
| Data Pump | mcp-oracle-ee-datapump | 3 |
| GoldenGate | mcp-oracle-ee-goldengate | 3 |
| OEM | mcp-oracle-ee-oem | 4 |
| PostgreSQL Enterprise | mcp-postgres-enterprise | 32 |
| Host OSS | mcp-host | 3 |
| Host Enterprise | mcp-host-enterprise | 6 |
| Monitoring Agent | mcp-dbmonitor | 4 |
| Monitoring Central | mcp-dbmonitor-ee | 6 |
| **TOTAL** | | **112** |

---

## Oracle OSS Skills (mcp-oracle)

### /ora-health

Quick Oracle database health check — the primary first-look skill for any Oracle target.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle |
| License | OSS |
| Duration | ~10 seconds |

**Tools Used:**
1. `oracle_sys_instance` — Check instance status, open mode, and uptime
2. `oracle_sys_sga` — Verify SGA allocation and component sizes
3. `oracle_tablespace_usage` — Check tablespace fill levels across all tablespaces
4. `oracle_session_top_waiters` — Identify current wait bottlenecks by wait class
5. `oracle_alert_tail` — Scan the last 200 lines of the alert log for ORA- errors
6. `oracle_redo_status` — Check redo log group status and recent switch frequency

**Workflow:**
1. **Collect** — Run all 6 tools in parallel against the target
2. **Analyze** — Apply thresholds: tablespace >85% = WARN, >95% = CRIT; >10 ORA- errors in last hour = CRIT; redo switches >4/hour = WARN
3. **Report** — Traffic-light summary (GREEN / YELLOW / RED) per category with actionable next steps

**Example:**
```
/ora-health entity_name=prod-orcl
```

---

**Remaining Oracle OSS skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-info | mcp-oracle | OSS | ~5s | oracle_sys_version, oracle_sys_instance, oracle_parameter_list | Full database inventory: version, components, NLS settings, init parameters |
| /ora-test | mcp-oracle | OSS | ~30s | oracle_sys_instance, oracle_tablespace_usage, oracle_session_list, oracle_alert_tail | Live test suite validating MCP-to-Oracle connectivity and core read paths; results posted to Slack |
| /ora-explore | mcp-oracle | OSS | ~8s | oracle_schema_list, oracle_schema_objects, oracle_tablespace_usage, oracle_session_list | Interactive schema and object browser for discovery and onboarding |

---

## Oracle Linux Skills (mcp-oracle-ol)

### /ol-health

Oracle Linux host health check covering OS-level prerequisites for database operation.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ol |
| License | OSS |
| Duration | ~15 seconds |

**Tools Used:**
1. `ol_kernel_hugepages` — Verify HugePage allocation matches SGA requirements
2. `ol_kernel_param_list` — Check Oracle-required kernel parameters (shmmax, semaphores, etc.)
3. `ol_storage_disk_usage` — Review filesystem usage on Oracle data, redo, and archive destinations
4. `ol_network_ntp_status` — Validate NTP synchronisation (critical for RAC and Data Guard)
5. `ol_security_selinux_status` — Check SELinux mode and Oracle-relevant policy

**Workflow:**
1. **Collect** — Run all tools in parallel
2. **Analyze** — Flag deviations from Oracle Linux best practices; cross-reference HugePage allocation against reported SGA size
3. **Report** — Categorised findings with remediation commands

**Example:**
```
/ol-health entity_name=db-host-01
```

---

**Remaining Oracle Linux skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ol-harden | mcp-oracle-ol | OSS | ~20s | ol_kernel_param_set, ol_security_firewall_list, ol_security_service_status | Apply Oracle Linux hardening baseline: kernel parameters, unnecessary services, firewall rules |
| /ol-test | mcp-oracle-ol | OSS | ~30s | ol_kernel_info, ol_network_nic_list, ol_storage_lv_list | Live test suite for Oracle Linux MCP adapter; validates read and write tool paths |

---

## Oracle EE Base Skills (mcp-oracle-ee)

### /ora-session

Oracle session diagnostics and management — inspect, trace, and selectively terminate problem sessions.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee |
| License | EE |
| Duration | ~12 seconds |

**Tools Used:**
1. `oracle_session_list` — List all active sessions with module, action, SQL ID
2. `oracle_session_top_waiters` — Rank sessions by wait time and wait class
3. `oracle_session_describe` — Fetch full session detail for flagged sessions
4. `oracle_sql_explain` — Show execution plan for the current SQL of blocked sessions
5. `oracle_session_kill` — Terminate session after confirm gate (Standard confirm)

**Workflow:**
1. **Inventory** — List all sessions; identify long-running, blocked, or high-waiter sessions
2. **Drill-down** — Describe flagged sessions and fetch execution plans for their active SQL
3. **Recommend** — Suggest kill candidates with justification
4. **Act** — Execute kills after Standard confirm gate

**Example:**
```
/ora-session entity_name=prod-orcl filter=blocked
```

---

**Remaining Oracle EE Base skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-tablespace | mcp-oracle-ee | EE | ~8s | oracle_tablespace_list, oracle_tablespace_describe, oracle_tablespace_usage | Tablespace capacity analysis with autoextend status, segment growth trends, and resize recommendations |
| /ora-user | mcp-oracle-ee | EE | ~10s | oracle_user_list, oracle_user_describe, oracle_user_profiles | User account review: lock status, default tablespace, profile, privilege summary |
| /ora-param | mcp-oracle-ee | EE | ~8s | oracle_parameter_list, oracle_parameter_describe, oracle_parameter_modified | Parameter audit: list non-default parameters, describe selected parameter, flag hidden parameter deviations |
| /ora-sql | mcp-oracle-ee | EE | ~15s | oracle_sql_exec, oracle_sql_explain, oracle_advisor_sql_tuning | SQL execution, explain plan, and optional SQL Tuning Advisor invocation for a given SQL ID or text |

---

## Performance Skills (mcp-oracle-ee-performance)

### /ora-perf

Real-time Oracle performance overview — combines ASH top activity, AWR load differential, and wait event analysis into a single actionable report.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-performance |
| License | EE+ |
| Duration | ~25 seconds |

**Tools Used:**
1. `oracle_ash_top_activity` — Top SQL and sessions from ASH for the last 15 minutes
2. `oracle_ash_wait_classes` — Wait class breakdown from ASH
3. `oracle_awr_diff` — AWR snapshot delta comparing current period to a prior baseline snapshot
4. `oracle_session_top_waiters` — Live wait event leader-board
5. `oracle_advisor_segment` — Segment advisor scan for row-chaining and fragmentation on hot objects

**Workflow:**
1. **Baseline** — Fetch ASH top activity and AWR diff in parallel
2. **Live view** — Collect current top waiters
3. **Correlate** — Cross-reference ASH hot objects against AWR regression and current live waits
4. **Report** — Ranked findings: top wait events, top SQL by elapsed time delta, segment health for hot objects

**Example:**
```
/ora-perf entity_name=prod-orcl lookback_minutes=30
```

---

**Remaining Performance skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-ash | mcp-oracle-ee-performance | EE+ | ~20s | oracle_ash_top_activity, oracle_ash_wait_classes, oracle_ash_time_range | Deep ASH analysis with configurable time range; compare two windows to isolate regression |
| /ora-awr | mcp-oracle-ee-performance | EE+ | ~30s | oracle_awr_report, oracle_awr_diff, oracle_awr_baseline_list | AWR report generation for a named snapshot range plus baseline comparison; outputs text or HTML |
| /ora-tune | mcp-oracle-ee-performance | EE+ | ~40s | oracle_advisor_sql_tuning, oracle_advisor_segment, oracle_sql_explain | SQL Tuning Advisor task creation, execution, and recommendation extraction; includes index advice |

---

## Audit Skills (mcp-oracle-ee-audit)

### /ora-audit

Unified audit policy management workflow — review existing policies, create or modify policies, and verify coverage.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-audit |
| License | EE+ |
| Duration | ~15 seconds |

**Tools Used:**
1. `oracle_audit_policy_list` — List all unified audit policies and their enabled status
2. `oracle_audit_policy_describe` — Show conditions and actions for a specific policy
3. `oracle_audit_policy_create` — Create a new unified audit policy (Standard confirm)
4. `oracle_audit_policy_enable` — Enable policy for one or more users or roles

**Workflow:**
1. **Inventory** — List current policies grouped by enabled/disabled
2. **Gap analysis** — Check coverage against standard action categories (DDL, DML, privilege use, logon)
3. **Remediate** — Create or enable missing policies after Standard confirm
4. **Verify** — Re-list policies to confirm state

**Example:**
```
/ora-audit entity_name=prod-orcl
```

---

**Remaining Audit skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-audit-report | mcp-oracle-ee-audit | EE+ | ~20s | oracle_audit_trail_query, oracle_audit_policy_list, oracle_user_list | Compliance audit report: recent audit trail entries, policy coverage summary, privileged user activity |
| /ora-fga | mcp-oracle-ee-audit | EE+ | ~15s | oracle_fga_policy_list, oracle_fga_policy_create, oracle_fga_audit_trail | Fine-grained auditing setup and review: column-level audit policies, FGA trail query, handler configuration |

---

## Partitioning Skills (mcp-oracle-ee-partitioning)

### /ora-partition

Partition inventory and health check — survey partition layout, segment sizes, and identify candidates for maintenance.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-partitioning |
| License | EE+ |
| Duration | ~15 seconds |

**Tools Used:**
1. `oracle_partition_list` — List all partitioned tables and indexes with partition counts
2. `oracle_partition_describe` — Show partition key, strategy, and interval for selected objects
3. `oracle_tablespace_usage` — Cross-reference partition segments against tablespace fill
4. `oracle_advisor_segment` — Identify partition-level segment fragmentation

**Workflow:**
1. **Survey** — List all partitioned objects sorted by segment size
2. **Detail** — Describe partition strategy for objects above size threshold
3. **Analyse** — Flag maxvalue partitions requiring split, stale statistics, and fragmented segments
4. **Report** — Sorted findings with recommended maintenance actions

**Example:**
```
/ora-partition entity_name=prod-orcl schema=sales
```

---

**Remaining Partitioning skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-partition-maintain | mcp-oracle-ee-partitioning | EE+ | ~20s | oracle_partition_split, oracle_partition_merge, oracle_partition_drop, oracle_partition_exchange | Execute partition maintenance operations (split, merge, drop, exchange) with Standard confirm gate |
| /ora-partition-archive | mcp-oracle-ee-partitioning | EE+ | ~25s | oracle_partition_exchange, oracle_partition_drop, oracle_tablespace_usage | Partition archival workflow: exchange partition into archive table, verify, then drop source partition |

---

## Data Guard Skills (mcp-oracle-ee-dataguard)

### /ora-dg

Data Guard configuration overview and lag monitoring.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-dataguard |
| License | EE+ |
| Duration | ~15 seconds |

**Tools Used:**
1. `oracle_dg_config` — Show Data Guard configuration: primary, standbys, protection mode
2. `oracle_dg_status` — Per-database status, apply lag, transport lag
3. `oracle_dg_gap` — Archive gap detection between primary and each standby
4. `oracle_redo_status` — Redo log status on primary (switch frequency)

**Workflow:**
1. **Config** — Retrieve full Data Guard configuration
2. **Health** — Check all database roles, apply lag, and transport lag
3. **Gap check** — Detect archive gaps; flag if gap > 0
4. **Report** — Status summary per standby with lag values and gap indicator

**Example:**
```
/ora-dg entity_name=prod-orcl
```

---

**Remaining Data Guard skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-dg-switch | mcp-oracle-ee-dataguard | EE+ | ~60s | oracle_dg_status, oracle_dg_switchover, oracle_dg_config | Planned switchover workflow: pre-checks, lag drain wait, switchover execution (Standard confirm), post-switch verify |
| /ora-dg-failover | mcp-oracle-ee-dataguard | EE+ | ~45s | oracle_dg_status, oracle_dg_failover, oracle_dg_config | Emergency failover to standby; pre-checks, confirm gap acceptance, execute (Double-Confirm: entity name + "FAILOVER" keyword) |
| /ora-dg-validate | mcp-oracle-ee-dataguard | EE+ | ~20s | oracle_dg_config, oracle_dg_status, oracle_dg_gap, oracle_sys_instance | End-to-end Data Guard validation: configuration correctness, lag thresholds, redo transport, apply service, observer status |

> **Note:** `/ora-dg-switch` requires Standard confirm. `/ora-dg-failover` requires Double-Confirm (entity name echo + explicit "FAILOVER" acknowledgement) because it is a potentially data-losing operation.

---

## Backup Skills (mcp-oracle-ee-backup)

### /ora-backup

RMAN backup execution workflow — validate prerequisites, run the backup job, and confirm completion.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-backup |
| License | EE+ |
| Duration | ~varies (monitoring loop) |

**Tools Used:**
1. `oracle_rman_status` — Check last backup status and retention policy
2. `oracle_tablespace_usage` — Verify sufficient space in backup destination
3. `oracle_rman_backup` — Execute RMAN backup (incremental level 0 or 1, archivelog, or full)
4. `oracle_alert_tail` — Monitor alert log for backup-related messages during execution
5. `oracle_rman_status` — Re-check after completion to confirm success

**Workflow:**
1. **Pre-check** — Validate last backup currency, destination space, and archive mode
2. **Execute** — Run RMAN backup command with requested type and retention
3. **Monitor** — Poll job status and tail alert log until completion
4. **Verify** — Confirm backup registered and crosscheck passes

**Example:**
```
/ora-backup entity_name=prod-orcl type=incremental level=1
```

---

**Remaining Backup skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-restore | mcp-oracle-ee-backup | EE+ | ~varies | oracle_rman_status, oracle_rman_restore, oracle_sys_instance | Datafile or tablespace restore from RMAN catalog; Standard confirm gate before execution |
| /ora-pitr | mcp-oracle-ee-backup | EE+ | ~varies | oracle_rman_status, oracle_rman_pitr, oracle_sys_instance | Point-in-time recovery to a target SCN or timestamp; Double-Confirm gate (entity name + target time echo) |
| /ora-rman-validate | mcp-oracle-ee-backup | EE+ | ~20s | oracle_rman_validate, oracle_rman_status, oracle_alert_tail | RMAN validate all backups in catalog; report corrupt or expired pieces without restoring |
| /ora-rman-policy | mcp-oracle-ee-backup | EE+ | ~10s | oracle_rman_status, oracle_rman_policy_set, oracle_rman_retention | Review and update RMAN retention policy, archivelog deletion policy, and backup optimization settings |

> **Note:** `/ora-pitr` is a data-destructive operation and enforces Double-Confirm. All data changes after the recovery target time are lost.

---

## RAC Skills (mcp-oracle-ee-rac)

### /ora-rac

Oracle RAC cluster health overview — instances, interconnect, and service distribution.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-rac |
| License | EE+ |
| Duration | ~15 seconds |

**Tools Used:**
1. `oracle_rac_instances` — List all RAC instances with status and uptime
2. `oracle_rac_interconnect` — Interconnect latency and throughput statistics
3. `oracle_rac_services` — Service preferred/available instance mapping
4. `oracle_session_list` — Session distribution across instances

**Workflow:**
1. **Instance survey** — List all instances and flag any not in OPEN state
2. **Interconnect** — Check latency and retransmit rates; flag if latency > 1ms average
3. **Services** — Verify services are balanced; flag services running on non-preferred instance
4. **Report** — Per-instance status with interconnect and service health summary

**Example:**
```
/ora-rac entity_name=prod-rac
```

---

**Remaining RAC skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-rac-service | mcp-oracle-ee-rac | EE+ | ~12s | oracle_rac_services, oracle_rac_service_relocate, oracle_rac_service_stop | RAC service management: relocate or stop a service on a specific instance (Standard confirm) |
| /ora-rac-diag | mcp-oracle-ee-rac | EE+ | ~20s | oracle_rac_instances, oracle_rac_interconnect, oracle_alert_tail, oracle_rac_gc_stats | RAC diagnostics: Global Cache statistics, GES lock analysis, interconnect packet loss, alert log ORA-600 scan |

---

## Clusterware Skills (mcp-oracle-ee-clusterware)

### /ora-crs

Oracle Clusterware (CRS) health check — resource status, voting disks, and OCR integrity.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-clusterware |
| License | EE+ |
| Duration | ~15 seconds |

**Tools Used:**
1. `oracle_crs_resource_status` — Status of all CRS-managed resources (databases, listeners, scan, VIPs)
2. `oracle_crs_voting_disk` — Voting disk count and accessibility
3. `oracle_crs_ocr_check` — OCR integrity check output
4. `oracle_crs_alert_tail` — Scan CRS alert log for recent errors

**Workflow:**
1. **Resources** — List all CRS resources; flag any in OFFLINE or INTERMEDIATE state
2. **Quorum** — Verify voting disk count is odd and all disks are accessible
3. **OCR** — Check OCR integrity; flag any mirror degradation
4. **Report** — Clusterware health summary with remediation guidance for any failures

**Example:**
```
/ora-crs entity_name=prod-rac
```

---

**Remaining Clusterware skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-crs-node | mcp-oracle-ee-clusterware | EE+ | ~15s | oracle_crs_node_status, oracle_crs_resource_status, oracle_crs_eviction_log | Node-level CRS diagnostics: node role, resource pinning, eviction history from CRSD logs |
| /ora-crs-scan | mcp-oracle-ee-clusterware | EE+ | ~10s | oracle_crs_scan_status, oracle_crs_listener_status, oracle_crs_vip_status | SCAN listener, SCAN VIP, and local VIP status check across all cluster nodes |

---

## ASM Skills (mcp-oracle-ee-asm)

### /ora-asm

ASM disk group health check — capacity, redundancy, and rebalance status.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-asm |
| License | EE+ |
| Duration | ~12 seconds |

**Tools Used:**
1. `oracle_asm_diskgroup_list` — List all disk groups with state, type, and usage
2. `oracle_asm_diskgroup_describe` — Per-disk-group detail: redundancy, allocation unit, compatibility
3. `oracle_asm_disk_list` — Disk status within each group (header status, path)
4. `oracle_asm_rebalance_status` — Active rebalance operation status and estimated completion

**Workflow:**
1. **Groups** — List disk groups; flag any not in MOUNTED state or above 85% usage
2. **Disks** — Check each disk for NORMAL header status; flag any MISSING or DROPPING
3. **Rebalance** — Report active rebalance progress if any
4. **Report** — Disk group health summary with capacity alerts

**Example:**
```
/ora-asm entity_name=prod-asm
```

---

**Remaining ASM skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-asm-rebalance | mcp-oracle-ee-asm | EE+ | ~10s | oracle_asm_rebalance_status, oracle_asm_rebalance_start, oracle_asm_rebalance_cancel | Monitor, start, or cancel an ASM rebalance operation with power setting (Standard confirm for start/cancel) |
| /ora-asm-disk | mcp-oracle-ee-asm | EE+ | ~15s | oracle_asm_disk_list, oracle_asm_disk_add, oracle_asm_disk_drop | ASM disk management: list disks, add new disks to a group, or drop disks with rebalance trigger (Standard confirm) |

---

## Provisioning Skills (mcp-oracle-ee-provision)

### /ora-create-db

New Oracle database provisioning workflow — create a database from scratch using DBCA parameters.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-provision |
| License | EE+ |
| Duration | ~15–30 minutes |

**Tools Used:**
1. `oracle_provision_prereq_check` — Validate host prerequisites (memory, disk, kernel parameters)
2. `oracle_provision_dbca_generate` — Generate DBCA response file from provided parameters
3. `oracle_provision_create_db` — Execute DBCA database creation (Standard confirm)
4. `oracle_sys_instance` — Verify new instance opens successfully
5. `oracle_tablespace_usage` — Confirm initial tablespace creation

**Workflow:**
1. **Pre-flight** — Check host prerequisites and resolve any blockers
2. **Configure** — Generate DBCA response file; present configuration summary for review
3. **Create** — Execute DBCA after Standard confirm gate
4. **Verify** — Connect to new instance, confirm OPEN status and tablespace layout

**Example:**
```
/ora-create-db entity_name=new-orcl db_name=ORCL db_unique_name=ORCL_P1 charset=AL32UTF8
```

---

**Remaining Provisioning skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-clone-db | mcp-oracle-ee-provision | EE+ | ~varies | oracle_provision_prereq_check, oracle_provision_clone, oracle_sys_instance | Clone an existing database to a new target host using RMAN duplicate or Data Pump full export |
| /ora-pdb | mcp-oracle-ee-provision | EE+ | ~5–10 min | oracle_provision_pdb_create, oracle_provision_pdb_list, oracle_sys_instance | Pluggable database lifecycle management: create, clone, open, close, drop PDB (Standard confirm for drop) |

---

## Patching Skills (mcp-oracle-ee-patch)

### /ora-patch

Oracle patch apply workflow — stage, pre-check, apply, and post-validate a patch.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-patch |
| License | EE+ |
| Duration | ~30–90 minutes |

**Tools Used:**
1. `oracle_patch_prereq` — OPatch prerequisites and conflict analysis for the target patch
2. `oracle_patch_apply` — Apply patch via OPatch (Standard confirm)
3. `oracle_patch_datapatch` — Run datapatch for SQL component updates
4. `oracle_sys_version` — Verify installed patch level post-apply
5. `oracle_alert_tail` — Scan alert log for patch-related errors

**Workflow:**
1. **Pre-check** — Run OPatch prereq; display conflict and space analysis
2. **Stage** — Confirm patch staging location and Oracle Home target
3. **Apply** — Execute OPatch apply after Standard confirm
4. **Datapatch** — Run datapatch if required by the patch bundle
5. **Verify** — Confirm patch registered in dba_registry_sqlpatch and alert log is clean

**Example:**
```
/ora-patch entity_name=prod-orcl patch_id=34133642 oracle_home=/u01/app/oracle/product/19c/dbhome_1
```

---

**Remaining Patching skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-patch-gold | mcp-oracle-ee-patch | EE+ | ~20s | oracle_patch_gold_image_list, oracle_patch_gold_image_apply, oracle_sys_version | Apply a pre-validated gold image patch set across Oracle Homes using fleet-style patching |
| /ora-patch-analyze | mcp-oracle-ee-patch | EE+ | ~15s | oracle_patch_prereq, oracle_patch_conflicts, oracle_patch_space_check | Dry-run patch analysis: conflict detection, space requirements, prerequisite validation — no changes applied |
| /ora-patch-rollback | mcp-oracle-ee-patch | EE+ | ~30–60 min | oracle_patch_rollback, oracle_sys_version, oracle_alert_tail | Roll back a previously applied patch via OPatch rollback (Standard confirm); verify rollback and run datapatch if needed |

---

## Migration Skills (mcp-oracle-ee-migration)

### /ora-migrate

Oracle database migration orchestration — full migration workflow from source analysis through cutover.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-migration |
| License | EE+ |
| Duration | ~varies by method |

**Tools Used:**
1. `oracle_migration_precheck` — Source database analysis: size, characterset, incompatibilities
2. `oracle_migration_method_select` — Recommend migration method (Data Pump, RMAN, GoldenGate, TTS)
3. `oracle_migration_execute` — Execute selected migration method (Standard confirm)
4. `oracle_sys_instance` — Verify target database post-migration
5. `oracle_migration_validate` — Row count and object count comparison between source and target

**Workflow:**
1. **Analyse** — Collect source metadata; estimate migration window
2. **Plan** — Present method recommendation with estimated downtime
3. **Execute** — Run migration after Standard confirm
4. **Validate** — Compare object and row counts; flag discrepancies
5. **Report** — Migration summary with validation results and cutover readiness

**Example:**
```
/ora-migrate source=source-orcl target=target-orcl method=datapump
```

---

**Remaining Migration skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-migrate-precheck | mcp-oracle-ee-migration | EE+ | ~15s | oracle_migration_precheck, oracle_sys_version, oracle_tablespace_usage | Standalone pre-migration analysis: source characterset, invalid objects, unsupported features, estimated size |
| /ora-migrate-rollback | mcp-oracle-ee-migration | EE+ | ~varies | oracle_migration_rollback, oracle_sys_instance, oracle_alert_tail | Rollback a failed migration to restore source database to service (Standard confirm) |

---

## Data Pump Skills (mcp-oracle-ee-datapump)

### /ora-pump-export

Data Pump export job creation and monitoring.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-datapump |
| License | EE+ |
| Duration | ~varies by export size |

**Tools Used:**
1. `oracle_datapump_dir_check` — Verify Oracle directory object exists and has sufficient space
2. `oracle_datapump_export` — Create and start expdp job with provided parameters (Standard confirm)
3. `oracle_datapump_status` — Poll job status and log file tail
4. `oracle_tablespace_usage` — Re-check destination space after completion

**Workflow:**
1. **Pre-check** — Validate directory object, permissions, and destination space
2. **Configure** — Present expdp parameter summary for review
3. **Execute** — Submit export job after Standard confirm
4. **Monitor** — Poll status at configured interval; tail log for errors
5. **Report** — Export summary: dump file set, row counts, elapsed time

**Example:**
```
/ora-pump-export entity_name=prod-orcl schemas=HR,SALES directory=EXPDP_DIR
```

---

**Remaining Data Pump skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-pump-import | mcp-oracle-ee-datapump | EE+ | ~varies | oracle_datapump_dir_check, oracle_datapump_import, oracle_datapump_status | Data Pump import workflow: remap schemas/tablespaces, submit impdp, monitor progress (Standard confirm) |
| /ora-pump-status | mcp-oracle-ee-datapump | EE+ | ~5s | oracle_datapump_status, oracle_datapump_list | List all active and recent Data Pump jobs with status, progress percentage, and log file reference |

---

## GoldenGate Skills (mcp-oracle-ee-goldengate)

### /ora-gg

GoldenGate replication status overview — process health, lag, and trail file status.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-goldengate |
| License | EE+ |
| Duration | ~15 seconds |

**Tools Used:**
1. `oracle_gg_manager_status` — GoldenGate Manager process status on source and target
2. `oracle_gg_process_list` — List all Extract, Replicat, and Pump processes with status
3. `oracle_gg_lag` — Per-process lag in seconds and trail file positions
4. `oracle_gg_trail_status` — Trail file usage and sequence numbers

**Workflow:**
1. **Manager** — Confirm Manager is running on both source and target
2. **Processes** — List all processes; flag any in STOPPED, ABENDED, or RBS state
3. **Lag** — Report per-process lag; flag Extract > 60s, Replicat > 300s
4. **Report** — Replication topology summary with lag traffic light

**Example:**
```
/ora-gg entity_name=prod-orcl gg_home=/u01/app/oracle/gg
```

---

**Remaining GoldenGate skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-gg-lag | mcp-oracle-ee-goldengate | EE+ | ~10s | oracle_gg_lag, oracle_gg_process_list, oracle_gg_statistics | Focused lag analysis: per-process lag trend, statistics comparison, trail throughput metrics |
| /ora-gg-manage | mcp-oracle-ee-goldengate | EE+ | ~10s | oracle_gg_process_start, oracle_gg_process_stop, oracle_gg_process_list | Start, stop, or restart individual GoldenGate processes (Standard confirm); verify status after operation |

---

## OEM Skills (mcp-oracle-ee-oem)

### /ora-oem

Oracle Enterprise Manager target overview — monitored target status, open incidents, and metric compliance.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-oracle-ee-oem |
| License | EE+ |
| Duration | ~15 seconds |

**Tools Used:**
1. `oracle_oem_target_status` — OEM agent status and target availability
2. `oracle_oem_incident_list` — Open incidents and alerts for the target
3. `oracle_oem_metric_violations` — Current metric threshold violations
4. `oracle_oem_blackout_list` — Active maintenance blackouts for the target

**Workflow:**
1. **Target** — Confirm OEM agent connectivity and target availability percentage
2. **Incidents** — List open incidents by severity (CRITICAL, WARNING, ADVISORY)
3. **Metrics** — Show current threshold violations with metric values and thresholds
4. **Blackouts** — Report any active blackouts that may suppress alerting
5. **Report** — OEM health dashboard summary

**Example:**
```
/ora-oem entity_name=prod-orcl oem_host=oem.example.com
```

---

**Remaining OEM skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /ora-oem-job | mcp-oracle-ee-oem | EE+ | ~10s | oracle_oem_job_list, oracle_oem_job_submit, oracle_oem_job_status | List, submit, and monitor OEM deployment procedure or job execution (Standard confirm for submission) |
| /ora-oem-compliance | mcp-oracle-ee-oem | EE+ | ~20s | oracle_oem_compliance_results, oracle_oem_compliance_standards | OEM Compliance Framework: fetch results for assigned compliance standards; flag failed rules |
| /ora-oem-patch | mcp-oracle-ee-oem | EE+ | ~15s | oracle_oem_patch_plan_list, oracle_oem_patch_plan_deploy, oracle_oem_job_status | Manage OEM patch plans: list available patch plans, initiate deployment procedure (Standard confirm) |

---

## PostgreSQL Enterprise Skills (mcp-postgres-enterprise)

### /pg-health

Quick PostgreSQL database health check — the primary first-look skill for any PostgreSQL target.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-postgres-enterprise |
| License | EE |
| Duration | ~10 seconds |

**Tools Used:**
1. `pg_connection_status` — Verify connectivity and server version
2. `pg_server_activity` — Active connections count vs. max_connections
3. `pg_tablespace_usage` — Tablespace fill levels
4. `pg_bgwriter_stats` — Buffer hit ratio and checkpoint frequency
5. `pg_replication_status` — Replication slot lag and standby sync status
6. `pg_log_errors` — Scan recent log for ERROR and FATAL entries

**Workflow:**
1. **Collect** — Run all 6 tools in parallel
2. **Analyze** — Flag: connection saturation >80%, buffer hit ratio <95%, replication lag >30s, FATAL errors in log
3. **Report** — Traffic-light summary per category with recommended actions

**Example:**
```
/pg-health entity_name=prod-pg
```

---

**Remaining PostgreSQL Enterprise skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /pg-perf | mcp-postgres-enterprise | EE | ~20s | pg_stat_statements_top, pg_wait_events, pg_bgwriter_stats, pg_vacuum_stats | Real-time performance overview: top SQL by total time, wait event breakdown, vacuum and autovacuum status |
| /pg-io | mcp-postgres-enterprise | EE | ~15s | pg_io_stats, pg_tablespace_usage, pg_bgwriter_stats | I/O analysis: per-relation block reads, heap and index fetch ratios, shared_buffers hit rate |
| /pg-slowlog | mcp-postgres-enterprise | EE | ~15s | pg_log_slow_queries, pg_stat_statements_top | Slow query analysis: parse pg_log for slow queries, correlate with pg_stat_statements for execution counts |
| /pg-tune | mcp-postgres-enterprise | EE | ~25s | pg_parameter_list, pg_parameter_recommend, pg_stat_statements_top | Parameter tuning recommendations based on server hardware profile, workload type, and current statistics |
| /pg-capacity | mcp-postgres-enterprise | EE | ~15s | pg_tablespace_usage, pg_table_sizes, pg_index_sizes | Capacity planning: database size breakdown by schema, largest tables and indexes, growth trend from statistics age |
| /pg-security | mcp-postgres-enterprise | EE | ~15s | pg_user_list, pg_privilege_list, pg_hba_check | Security posture: superuser count, public schema privileges, pg_hba.conf rule review, idle connection check |
| /pg-ssl | mcp-postgres-enterprise | EE | ~10s | pg_ssl_status, pg_parameter_list | SSL/TLS configuration: verify ssl=on, check ssl_cert_file expiry, report cipher and protocol settings |
| /pg-harden | mcp-postgres-enterprise | EE | ~20s | pg_parameter_list, pg_hba_check, pg_user_list, pg_privilege_list | Hardening review: non-default trust entries in pg_hba, excessive privileges, pg_password_policy compliance |
| /pg-comply | mcp-postgres-enterprise | EE | ~20s | pg_comply_cis_benchmark, pg_comply_password_policy, pg_comply_encryption_status | Compliance check against CIS PostgreSQL benchmark; password policy status; encryption at rest verification |
| /pg-audit | mcp-postgres-enterprise | EE | ~10s | pg_audit_pgaudit_status, pg_audit_log_analysis | pgaudit extension status, current audit settings, recent audit trail sample |
| /pg-audit-full | mcp-postgres-enterprise | EE | ~25s | pg_audit_pgaudit_status, pg_audit_ddl_history, pg_audit_log_analysis, pg_comply_privilege_audit | Full audit report: DDL history, DML audit trail, privilege change log, pgaudit coverage gap analysis |
| /pg-ha | mcp-postgres-enterprise | EE | ~15s | pg_replication_status, pg_cnpg_cluster_status, pg_cnpg_backup_status | High availability health: replication slots, CNPG cluster status, WAL archiving lag, switchover readiness |
| /pg-dr | mcp-postgres-enterprise | EE | ~15s | pg_replication_status, pg_pgbackrest_status, pg_cnpg_backup_status | Disaster recovery posture: cross-cluster replication status, pgBackRest WAL archive continuity, last backup age |
| /pg-backup | mcp-postgres-enterprise | EE | ~10s | pg_pgbackrest_status, pg_pgbackrest_backup, pg_barman_status | Backup status across all configured tools (pgBackRest, Barman, CNPG); flag expired, missing, or degraded backups |
| /pg-dba | mcp-postgres-enterprise | EE | ~15s | pg_vacuum_stats, pg_bloat_check, pg_index_health | DBA maintenance overview: autovacuum status, table and index bloat analysis, dead tuple counts, stale statistics |
| /pg-partition | mcp-postgres-enterprise | EE | ~12s | pg_partition_list, pg_table_sizes, pg_partition_detach | Partition layout survey: declarative partition trees, partition sizes, detached or orphaned partition detection |
| /pg-observe | mcp-postgres-enterprise | EE | ~15s | pg_stat_statements_top, pg_wait_events, pg_lock_analysis | Observability deep-dive: top SQL, wait event tree, lock wait chain analysis for blocked sessions |
| /pg-alert-setup | mcp-postgres-enterprise | EE | ~10s | pg_parameter_list, pg_alert_rule_create | Configure alerting rules in the monitoring agent for a target: connection saturation, replication lag, WAL volume |
| /pg-wal | mcp-postgres-enterprise | EE | ~10s | pg_replication_status, pg_wal_stats, pg_pgbackrest_status | WAL health: generation rate, archive status, replication slot retention, WAL receiver lag |
| /pg-rbac | mcp-postgres-enterprise | EE | ~15s | pg_user_list, pg_privilege_list, pg_role_membership | RBAC audit: role hierarchy, privilege grants, membership chains, superuser and createdb flag review |
| /pg-rls-audit | mcp-postgres-enterprise | EE | ~12s | pg_rls_policy_list, pg_privilege_list | Row-level security audit: RLS policies per table, enabled/forced status, privilege exceptions |
| /pg-tenant | mcp-postgres-enterprise | EE | ~10s | pg_user_list, pg_schema_list, pg_privilege_list | Multi-tenant schema inventory: per-tenant schema owner, object count, connection quota |
| /pg-tenant-onboard | mcp-postgres-enterprise | EE | ~15s | pg_user_create, pg_schema_create, pg_privilege_grant, pg_rls_policy_create | New tenant provisioning: create database user, schema, set search_path, apply RLS policies (Standard confirm) |
| /pg-tenant-drift | mcp-postgres-enterprise | EE | ~15s | pg_user_list, pg_schema_list, pg_privilege_list, pg_rls_policy_list | Tenant configuration drift detection: compare each tenant schema against the expected baseline |
| /pg-incident | mcp-postgres-enterprise | EE | ~15s | pg_log_errors, pg_stat_activity, pg_lock_analysis, pg_replication_status | Active incident triage: correlate current errors, blocked sessions, locks, and replication status into a timeline |
| /pg-incident-review | mcp-postgres-enterprise | EE | ~20s | pg_log_errors, pg_stat_statements_top, pg_vacuum_stats | Post-incident review: parse log window around incident time, correlate with slow queries and autovacuum activity |
| /pg-context | mcp-postgres-enterprise | EE | ~10s | pg_connection_status, pg_parameter_list, pg_schema_list | Snapshot current database context for sharing with a support engagement: version, parameters, schema summary |
| /pg-evidence | mcp-postgres-enterprise | EE | ~20s | pg_log_errors, pg_stat_statements_top, pg_audit_log_analysis | Evidence collection for a defined time window: filtered log, query stats, audit trail — output as structured report |
| /pg-migrate | mcp-postgres-enterprise | EE | ~varies | pg_migrate_precheck, pg_migrate_execute, pg_connection_status | PostgreSQL version upgrade or schema migration workflow: precheck, execute, validate (Standard confirm) |
| /pg-full-report | mcp-postgres-enterprise | EE | ~60s | pg_health (all tools), pg_perf (all tools), pg_comply (all tools), pg_backup (all tools) | Comprehensive database report combining health, performance, security, compliance, and backup into a single document |
| /pg-test | mcp-postgres-enterprise | EE | ~45s | (all 32 tool categories sampled) | Live test suite validating all mcp-postgres-enterprise tool paths against a test database; results posted to Slack |

---

## Host OSS Skills (mcp-host)

### /host-health

Generic host health check — OS-level vitals for any database server.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-host |
| License | OSS |
| Duration | ~10 seconds |

**Tools Used:**
1. `host_kernel_info` — OS version, kernel version, uptime
2. `host_storage_disk_usage` — Filesystem usage across all mounts
3. `host_network_ntp_status` — NTP synchronisation status
4. `host_security_service_status` — Critical service status (sshd, firewalld, etc.)

**Workflow:**
1. **Collect** — Run all tools in parallel
2. **Analyze** — Flag filesystem >85%, NTP drift >1s, critical service not running
3. **Report** — Host health summary with severity indicators

**Example:**
```
/host-health entity_name=db-host-01
```

---

**Remaining Host OSS skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /host-info | mcp-host | OSS | ~5s | host_kernel_info, host_network_nic_list, host_storage_lv_list | Full host inventory: hardware profile, network interfaces, LVM layout, package list summary |
| /host-test | mcp-host | OSS | ~20s | host_kernel_info, host_storage_disk_usage, host_network_dns_check | Live test suite for mcp-host adapter; validates connectivity and core read paths |

---

## Host Enterprise Skills (mcp-host-enterprise)

### /host-harden

Host hardening workflow — apply and verify security hardening baseline.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-host-enterprise |
| License | EE |
| Duration | ~30 seconds |

**Tools Used:**
1. `host_security_selinux_status` — Check SELinux enforcement mode
2. `host_security_firewall_list` — Review firewall rules
3. `host_kernel_param_list` — Check security-relevant kernel parameters
4. `host_kernel_param_set` — Apply remediation settings (Standard confirm)
5. `host_security_service_status` — Verify unnecessary services are disabled

**Workflow:**
1. **Assess** — Collect current security posture across all categories
2. **Gap analysis** — Compare against hardening baseline (CIS or custom profile)
3. **Remediate** — Apply recommended parameter and service changes after Standard confirm
4. **Verify** — Re-collect and confirm all gaps closed

**Example:**
```
/host-harden entity_name=db-host-01 profile=cis-oracle-linux
```

---

**Remaining Host Enterprise skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /host-comply | mcp-host-enterprise | EE | ~20s | host_security_selinux_status, host_kernel_param_list, host_security_firewall_list | Compliance check against a named profile (CIS Level 1/2); report pass/fail per control |
| /host-patch | mcp-host-enterprise | EE | ~varies | host_package_list, host_package_update, host_kernel_info | OS patch workflow: list available updates, filter by severity, apply selected packages (Standard confirm) |
| /host-security | mcp-host-enterprise | EE | ~15s | host_security_firewall_list, host_security_selinux_status, host_security_service_status | Security posture snapshot: firewall rule audit, SELinux policy, listening services inventory |
| /host-capacity | mcp-host-enterprise | EE | ~15s | host_storage_disk_usage, host_storage_lv_list, host_storage_vg_list | Capacity planning: filesystem utilisation trends, LVM free space, VG expansion opportunities |
| /host-audit | mcp-host-enterprise | EE | ~20s | host_package_list, host_security_service_status, host_kernel_param_list | Host audit report: installed packages, running services, kernel parameters, user accounts — structured output |

---

## Monitoring Agent Skills (mcp-dbmonitor)

### /dbmon-deploy

Deploy the dbx monitoring agent to a new target host.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-dbmonitor |
| License | OSS |
| Duration | ~5 minutes |

**Tools Used:**
1. `dbmon_agent_prereq` — Check target host prerequisites (OS, ports, permissions)
2. `dbmon_agent_install` — Install agent binary and configuration (Standard confirm)
3. `dbmon_agent_configure` — Apply monitoring target configuration and credential profile
4. `dbmon_agent_status` — Verify agent starts and reports metrics to central

**Workflow:**
1. **Pre-flight** — Validate target host meets agent prerequisites
2. **Install** — Deploy agent package and configuration files after Standard confirm
3. **Configure** — Write target credentials and scrape intervals
4. **Verify** — Confirm agent status is RUNNING and first metrics received by central

**Example:**
```
/dbmon-deploy entity_name=db-host-01 target_type=oracle
```

---

**Remaining Monitoring Agent skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /dbmon-status | mcp-dbmonitor | OSS | ~5s | dbmon_agent_status, dbmon_agent_config | Show agent status, version, last heartbeat, configured targets and scrape intervals |
| /dbmon-metrics | mcp-dbmonitor | OSS | ~10s | dbmon_metric_query, dbmon_agent_status | Query and display current metric values from the local agent for a specific target |
| /dbmon-test | mcp-dbmonitor | OSS | ~30s | dbmon_agent_status, dbmon_metric_query, dbmon_agent_config | Live test: verify agent connectivity, metric collection, and central reporting |

---

## Monitoring Central Skills (mcp-dbmonitor-ee)

### /dbmon-fleet

Fleet-wide monitoring dashboard — status of all monitored targets from the central controller.

| Attribute | Value |
|-----------|-------|
| Repo | mcp-dbmonitor-ee |
| License | EE |
| Duration | ~15 seconds |

**Tools Used:**
1. `dbmon_fleet_status` — All registered targets with agent status and last heartbeat
2. `dbmon_fleet_alerts` — Open alerts across the fleet grouped by severity
3. `dbmon_fleet_compliance` — Compliance score per target
4. `dbmon_fleet_jobs` — Active and recent job execution summary

**Workflow:**
1. **Inventory** — List all targets; flag any with agent DOWN or missed heartbeat >5 minutes
2. **Alerts** — Aggregate open alerts by severity; surface CRITICAL count
3. **Compliance** — Fleet compliance score overview
4. **Report** — Fleet health dashboard with per-target drill-down links

**Example:**
```
/dbmon-fleet
```

---

**Remaining Monitoring Central skills (compact):**

| Skill | Repo | License | Duration | Key Tools | Description |
|-------|------|---------|----------|-----------|-------------|
| /dbmon-drift | mcp-dbmonitor-ee | EE | ~20s | dbmon_fleet_status, dbmon_config_baseline, dbmon_fleet_compliance | Configuration drift detection: compare each target's monitoring config against the registered baseline |
| /dbmon-comply | mcp-dbmonitor-ee | EE | ~20s | dbmon_fleet_compliance, dbmon_comply_report | Fleet compliance report: per-target compliance scores, failed control summary, trend over last 30 days |
| /dbmon-alert | mcp-dbmonitor-ee | EE | ~10s | dbmon_fleet_alerts, dbmon_alert_rule_list, dbmon_alert_rule_create | Alert management: list current rules and open alerts, create or modify alert rules (Standard confirm for changes) |
| /dbmon-job | mcp-dbmonitor-ee | EE | ~10s | dbmon_fleet_jobs, dbmon_job_submit, dbmon_job_status | Fleet job management: list scheduled and ad-hoc jobs, submit new job to one or more targets (Standard confirm) |
| /dbmon-upgrade | mcp-dbmonitor-ee | EE | ~varies | dbmon_fleet_status, dbmon_agent_upgrade, dbmon_agent_status | Coordinated fleet agent upgrade: pre-check versions, apply upgrades in rolling order (Standard confirm per batch) |

---

## Alphabetical Skill Index

| Skill | Engine | Repo | License | Description |
|-------|--------|------|---------|-------------|
| /dbmon-alert | Monitoring | mcp-dbmonitor-ee | EE | Alert rule management and open alert review |
| /dbmon-comply | Monitoring | mcp-dbmonitor-ee | EE | Fleet compliance report with trend |
| /dbmon-deploy | Monitoring | mcp-dbmonitor | OSS | Deploy monitoring agent to a new host |
| /dbmon-drift | Monitoring | mcp-dbmonitor-ee | EE | Configuration drift detection across fleet |
| /dbmon-fleet | Monitoring | mcp-dbmonitor-ee | EE | Fleet-wide status dashboard |
| /dbmon-job | Monitoring | mcp-dbmonitor-ee | EE | Fleet job submission and status |
| /dbmon-metrics | Monitoring | mcp-dbmonitor | OSS | Query current metrics from local agent |
| /dbmon-status | Monitoring | mcp-dbmonitor | OSS | Agent status and configuration |
| /dbmon-test | Monitoring | mcp-dbmonitor | OSS | Live test for monitoring agent |
| /dbmon-upgrade | Monitoring | mcp-dbmonitor-ee | EE | Rolling fleet agent upgrade |
| /host-audit | Host | mcp-host-enterprise | EE | Structured host audit report |
| /host-capacity | Host | mcp-host-enterprise | EE | Filesystem and LVM capacity planning |
| /host-comply | Host | mcp-host-enterprise | EE | CIS compliance check for host |
| /host-harden | Host | mcp-host-enterprise | EE | Apply and verify host hardening baseline |
| /host-health | Host | mcp-host | OSS | OS-level host health check |
| /host-info | Host | mcp-host | OSS | Full host inventory |
| /host-patch | Host | mcp-host-enterprise | EE | OS patch workflow |
| /host-security | Host | mcp-host-enterprise | EE | Host security posture snapshot |
| /host-test | Host | mcp-host | OSS | Live test for mcp-host adapter |
| /ol-harden | Oracle Linux | mcp-oracle-ol | OSS | Apply Oracle Linux hardening baseline |
| /ol-health | Oracle Linux | mcp-oracle-ol | OSS | Oracle Linux host health check |
| /ol-test | Oracle Linux | mcp-oracle-ol | OSS | Live test for Oracle Linux adapter |
| /ora-ash | Oracle | mcp-oracle-ee-performance | EE+ | Deep ASH analysis with time range comparison |
| /ora-asm | Oracle | mcp-oracle-ee-asm | EE+ | ASM disk group health check |
| /ora-asm-disk | Oracle | mcp-oracle-ee-asm | EE+ | ASM disk management |
| /ora-asm-rebalance | Oracle | mcp-oracle-ee-asm | EE+ | Monitor or control ASM rebalance |
| /ora-audit | Oracle | mcp-oracle-ee-audit | EE+ | Unified audit policy management |
| /ora-audit-report | Oracle | mcp-oracle-ee-audit | EE+ | Compliance audit report generation |
| /ora-awr | Oracle | mcp-oracle-ee-performance | EE+ | AWR report and baseline comparison |
| /ora-backup | Oracle | mcp-oracle-ee-backup | EE+ | RMAN backup execution workflow |
| /ora-clone-db | Oracle | mcp-oracle-ee-provision | EE+ | Clone database to new target host |
| /ora-create-db | Oracle | mcp-oracle-ee-provision | EE+ | New database provisioning via DBCA |
| /ora-crs | Oracle | mcp-oracle-ee-clusterware | EE+ | CRS health check |
| /ora-crs-node | Oracle | mcp-oracle-ee-clusterware | EE+ | Node-level CRS diagnostics |
| /ora-crs-scan | Oracle | mcp-oracle-ee-clusterware | EE+ | SCAN listener and VIP status |
| /ora-dg | Oracle | mcp-oracle-ee-dataguard | EE+ | Data Guard overview and lag monitoring |
| /ora-dg-failover | Oracle | mcp-oracle-ee-dataguard | EE+ | Emergency failover (Double-Confirm) |
| /ora-dg-switch | Oracle | mcp-oracle-ee-dataguard | EE+ | Planned switchover (Standard confirm) |
| /ora-dg-validate | Oracle | mcp-oracle-ee-dataguard | EE+ | End-to-end Data Guard validation |
| /ora-explore | Oracle | mcp-oracle | OSS | Schema and object browser |
| /ora-fga | Oracle | mcp-oracle-ee-audit | EE+ | Fine-grained auditing setup and review |
| /ora-gg | Oracle | mcp-oracle-ee-goldengate | EE+ | GoldenGate replication status overview |
| /ora-gg-lag | Oracle | mcp-oracle-ee-goldengate | EE+ | Focused GoldenGate lag analysis |
| /ora-gg-manage | Oracle | mcp-oracle-ee-goldengate | EE+ | Start, stop, or restart GoldenGate processes |
| /ora-health | Oracle | mcp-oracle | OSS | Quick Oracle database health check |
| /ora-info | Oracle | mcp-oracle | OSS | Full database inventory |
| /ora-migrate | Oracle | mcp-oracle-ee-migration | EE+ | Full migration orchestration workflow |
| /ora-migrate-precheck | Oracle | mcp-oracle-ee-migration | EE+ | Standalone pre-migration analysis |
| /ora-migrate-rollback | Oracle | mcp-oracle-ee-migration | EE+ | Roll back a failed migration |
| /ora-oem | Oracle | mcp-oracle-ee-oem | EE+ | OEM target overview and incidents |
| /ora-oem-compliance | Oracle | mcp-oracle-ee-oem | EE+ | OEM compliance framework results |
| /ora-oem-job | Oracle | mcp-oracle-ee-oem | EE+ | OEM job list and submission |
| /ora-oem-patch | Oracle | mcp-oracle-ee-oem | EE+ | OEM patch plan management |
| /ora-param | Oracle | mcp-oracle-ee | EE | Parameter audit: non-default and hidden |
| /ora-partition | Oracle | mcp-oracle-ee-partitioning | EE+ | Partition inventory and health check |
| /ora-partition-archive | Oracle | mcp-oracle-ee-partitioning | EE+ | Partition archival workflow |
| /ora-partition-maintain | Oracle | mcp-oracle-ee-partitioning | EE+ | Partition maintenance operations |
| /ora-patch | Oracle | mcp-oracle-ee-patch | EE+ | Oracle patch apply workflow |
| /ora-patch-analyze | Oracle | mcp-oracle-ee-patch | EE+ | Dry-run patch conflict and space analysis |
| /ora-patch-gold | Oracle | mcp-oracle-ee-patch | EE+ | Apply gold image patch set |
| /ora-patch-rollback | Oracle | mcp-oracle-ee-patch | EE+ | Roll back an applied patch |
| /ora-pdb | Oracle | mcp-oracle-ee-provision | EE+ | Pluggable database lifecycle management |
| /ora-perf | Oracle | mcp-oracle-ee-performance | EE+ | Real-time performance overview |
| /ora-pitr | Oracle | mcp-oracle-ee-backup | EE+ | Point-in-time recovery (Double-Confirm) |
| /ora-pump-export | Oracle | mcp-oracle-ee-datapump | EE+ | Data Pump export job |
| /ora-pump-import | Oracle | mcp-oracle-ee-datapump | EE+ | Data Pump import job |
| /ora-pump-status | Oracle | mcp-oracle-ee-datapump | EE+ | Active and recent Data Pump job status |
| /ora-rac | Oracle | mcp-oracle-ee-rac | EE+ | RAC cluster health overview |
| /ora-rac-diag | Oracle | mcp-oracle-ee-rac | EE+ | RAC diagnostics: GC stats, interconnect |
| /ora-rac-service | Oracle | mcp-oracle-ee-rac | EE+ | RAC service management |
| /ora-restore | Oracle | mcp-oracle-ee-backup | EE+ | Datafile or tablespace restore |
| /ora-rman-policy | Oracle | mcp-oracle-ee-backup | EE+ | Review and update RMAN retention policy |
| /ora-rman-validate | Oracle | mcp-oracle-ee-backup | EE+ | Validate all RMAN backup catalog entries |
| /ora-session | Oracle | mcp-oracle-ee | EE | Session diagnostics and management |
| /ora-sql | Oracle | mcp-oracle-ee | EE | SQL execution, explain plan, tuning |
| /ora-tablespace | Oracle | mcp-oracle-ee | EE | Tablespace capacity analysis |
| /ora-test | Oracle | mcp-oracle | OSS | Live test suite for mcp-oracle adapter |
| /ora-tune | Oracle | mcp-oracle-ee-performance | EE+ | SQL Tuning Advisor and index recommendations |
| /ora-user | Oracle | mcp-oracle-ee | EE | User account review |
| /pg-alert-setup | PostgreSQL | mcp-postgres-enterprise | EE | Configure alerting rules for a PG target |
| /pg-audit | PostgreSQL | mcp-postgres-enterprise | EE | pgaudit status and log analysis |
| /pg-audit-full | PostgreSQL | mcp-postgres-enterprise | EE | Full audit report with DDL history |
| /pg-backup | PostgreSQL | mcp-postgres-enterprise | EE | Backup status across all configured tools |
| /pg-capacity | PostgreSQL | mcp-postgres-enterprise | EE | Capacity planning and size breakdown |
| /pg-comply | PostgreSQL | mcp-postgres-enterprise | EE | CIS benchmark and compliance check |
| /pg-context | PostgreSQL | mcp-postgres-enterprise | EE | Snapshot database context for support |
| /pg-dba | PostgreSQL | mcp-postgres-enterprise | EE | DBA maintenance overview: vacuum, bloat |
| /pg-dr | PostgreSQL | mcp-postgres-enterprise | EE | Disaster recovery posture review |
| /pg-evidence | PostgreSQL | mcp-postgres-enterprise | EE | Evidence collection for a time window |
| /pg-full-report | PostgreSQL | mcp-postgres-enterprise | EE | Comprehensive database report |
| /pg-ha | PostgreSQL | mcp-postgres-enterprise | EE | High availability health check |
| /pg-harden | PostgreSQL | mcp-postgres-enterprise | EE | PostgreSQL hardening review |
| /pg-health | PostgreSQL | mcp-postgres-enterprise | EE | Quick PostgreSQL health check |
| /pg-incident | PostgreSQL | mcp-postgres-enterprise | EE | Active incident triage |
| /pg-incident-review | PostgreSQL | mcp-postgres-enterprise | EE | Post-incident log and query review |
| /pg-io | PostgreSQL | mcp-postgres-enterprise | EE | I/O analysis and buffer hit rates |
| /pg-migrate | PostgreSQL | mcp-postgres-enterprise | EE | Version upgrade or schema migration |
| /pg-observe | PostgreSQL | mcp-postgres-enterprise | EE | Observability deep-dive: SQL, waits, locks |
| /pg-partition | PostgreSQL | mcp-postgres-enterprise | EE | Partition layout survey |
| /pg-perf | PostgreSQL | mcp-postgres-enterprise | EE | Real-time performance overview |
| /pg-rbac | PostgreSQL | mcp-postgres-enterprise | EE | RBAC audit: roles, privileges, membership |
| /pg-rls-audit | PostgreSQL | mcp-postgres-enterprise | EE | Row-level security policy audit |
| /pg-security | PostgreSQL | mcp-postgres-enterprise | EE | Security posture: users, privileges, hba |
| /pg-slowlog | PostgreSQL | mcp-postgres-enterprise | EE | Slow query analysis from log |
| /pg-ssl | PostgreSQL | mcp-postgres-enterprise | EE | SSL/TLS configuration review |
| /pg-tenant | PostgreSQL | mcp-postgres-enterprise | EE | Multi-tenant schema inventory |
| /pg-tenant-drift | PostgreSQL | mcp-postgres-enterprise | EE | Tenant configuration drift detection |
| /pg-tenant-onboard | PostgreSQL | mcp-postgres-enterprise | EE | New tenant provisioning |
| /pg-test | PostgreSQL | mcp-postgres-enterprise | EE | Live test suite for all EE tool paths |
| /pg-tune | PostgreSQL | mcp-postgres-enterprise | EE | Parameter tuning recommendations |
| /pg-wal | PostgreSQL | mcp-postgres-enterprise | EE | WAL health and archiving status |

---

*Last updated: 2026-04-11 — 112 skills across 22 repos.*
