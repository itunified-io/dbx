# dbx Tool Reference

**Platform Total: 735 tools across 25 repos**

## Conventions

- **Tool Name**: MCP tool identifier (e.g., `oracle_session_list`)
- **CLI Command**: Equivalent dbxcli command (e.g., `dbxcli db session list`)
- **Confirm Level**: None (read-only) | Standard (reversible) | Standard+Echo (significant) | Double-Confirm (catastrophic)
- **License**: OSS | Core | HA | Ops | PG-Pro | Host-Pro
- **Oracle Gate**: Required Oracle edition/options (if any)

All tools accept `entity_name` (target name) and `format` (table/json/yaml) parameters unless noted otherwise.

---

## Tool Count Summary

| Domain | Repo | OSS | Licensed | Total |
|--------|------|-----|----------|-------|
| Oracle Core DB (read) | mcp-oracle | 28 | --- | 28 |
| Oracle Core DB (mutate) | mcp-oracle-ee | --- | 38 | 38 |
| Oracle Linux | mcp-oracle-ol | 20 | --- | 20 |
| Performance (ASH/AWR/ADDM) | mcp-oracle-ee-performance | --- | 32 | 32 |
| Unified Audit | mcp-oracle-ee-audit | --- | 22 | 22 |
| Partitioning | mcp-oracle-ee-partitioning | --- | 26 | 26 |
| Data Guard | mcp-oracle-ee-dataguard | --- | 28 | 28 |
| Backup (RMAN) | mcp-oracle-ee-backup | --- | 34 | 34 |
| RAC | mcp-oracle-ee-rac | --- | 24 | 24 |
| Clusterware | mcp-oracle-ee-clusterware | --- | 22 | 22 |
| ASM | mcp-oracle-ee-asm | --- | 20 | 20 |
| Provisioning | mcp-oracle-ee-provision | --- | 22 | 22 |
| Patching | mcp-oracle-ee-patch | --- | 30 | 30 |
| Migration | mcp-oracle-ee-migration | --- | 20 | 20 |
| Data Pump | mcp-oracle-ee-datapump | --- | 18 | 18 |
| GoldenGate | mcp-oracle-ee-goldengate | --- | 24 | 24 |
| OEM Cloud Control | mcp-oracle-ee-oem | --- | 28 | 28 |
| PostgreSQL Core | mcp-postgres | 27 | --- | 27 |
| PostgreSQL Enterprise | mcp-postgres-enterprise | --- | 111 | 111 |
| Host/OS Core | mcp-host | 20 | --- | 20 |
| Host/OS Enterprise | mcp-host-enterprise | --- | 40 | 40 |
| Monitoring Agent | mcp-dbmonitor | 16 | --- | 16 |
| Monitoring Central | mcp-dbmonitor-ee | --- | 35 | 35 |
| Policy Engine | dbx/dbx-ee | 3 | 5 | 8 |
| RAG | dbx/dbx-ee | 2 | 3 | 5 |
| **TOTAL** | | **116** | **619** | **735** |

---

## Oracle Core Database --- Read-Only (OSS, 28 tools)

Transport: SQL*Net. Oracle Gate: None.

### Session Management (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_session_list` | `dbxcli db session list` | None | List active sessions with status/username filter |
| `oracle_session_describe` | `dbxcli db session describe` | None | Detailed session info by SID,SERIAL# |
| `oracle_session_top_waiters` | `dbxcli db session top-waiters` | None | Top N sessions by cumulative wait time |

### Tablespace Management (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_tablespace_list` | `dbxcli db tablespace list` | None | List all tablespaces with status |
| `oracle_tablespace_describe` | `dbxcli db tablespace describe` | None | Tablespace detail: datafiles, autoextend, contents |
| `oracle_tablespace_usage` | `dbxcli db tablespace usage` | None | Usage summary: size, used, free, pct_used |

### User Management (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_user_list` | `dbxcli db user list` | None | List database users with status and profile |
| `oracle_user_describe` | `dbxcli db user describe` | None | User detail: roles, privileges, quotas |
| `oracle_user_profile_list` | `dbxcli db user profile-list` | None | List profiles with resource limits |

### Parameter Management (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_parameter_list` | `dbxcli db parameter list` | None | List all init parameters with values |
| `oracle_parameter_search` | `dbxcli db parameter search` | None | Search parameters by name pattern |

### Schema Introspection (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_schema_tables` | `dbxcli db schema tables` | None | List tables with row count, size |
| `oracle_schema_indexes` | `dbxcli db schema indexes` | None | List indexes with type, uniqueness, status |
| `oracle_schema_views` | `dbxcli db schema views` | None | List views with text preview |
| `oracle_schema_constraints` | `dbxcli db schema constraints` | None | List constraints by type |

### Redo Log (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_redo_list` | `dbxcli db redo list` | None | List redo log groups and members |
| `oracle_redo_status` | `dbxcli db redo status` | None | Current group, switch frequency, archive status |

### Undo (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_undo_status` | `dbxcli db undo status` | None | Undo tablespace status and retention |
| `oracle_undo_usage` | `dbxcli db undo usage` | None | Active/unexpired/expired extents breakdown |

### Alert Log (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_alert_tail` | `dbxcli db alert tail` | None | Tail last N lines of alert log |
| `oracle_alert_search` | `dbxcli db alert search` | None | Search alert log for ORA-/pattern |

### Data Dictionary (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_dict_views` | `dbxcli db dict views` | None | List data dictionary views |
| `oracle_dict_synonyms` | `dbxcli db dict synonyms` | None | List public/private synonyms |
| `oracle_dict_dependencies` | `dbxcli db dict dependencies` | None | Object dependency tree |

### System Info (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_sys_version` | `dbxcli db sys version` | None | Database version, edition, patch level |
| `oracle_sys_instance` | `dbxcli db sys instance` | None | Instance name, status, startup time, host |
| `oracle_sys_sga` | `dbxcli db sys sga` | None | SGA component sizes and advisory |
| `oracle_sys_pga` | `dbxcli db sys pga` | None | PGA target, allocated, advisory |

---

## Oracle Core Database --- Mutating (Core, 38 tools)

Transport: SQL*Net. Oracle Gate: None (EE features gated separately).

### Session (1 tool)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_session_kill` | `dbxcli db session kill` | Standard | Kill session by SID,SERIAL# |

### Tablespace CRUD (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_tablespace_create` | `dbxcli db tablespace create` | Standard+Echo | Create tablespace with datafile |
| `oracle_tablespace_resize` | `dbxcli db tablespace resize` | Standard+Echo | Resize datafile or add datafile |
| `oracle_tablespace_drop` | `dbxcli db tablespace drop` | Standard+Echo | Drop tablespace including contents |
| `oracle_tablespace_autoextend` | `dbxcli db tablespace autoextend` | Standard | Set autoextend on/off |
| `oracle_tablespace_coalesce` | `dbxcli db tablespace coalesce` | Standard | Coalesce free extents |

### User CRUD (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_user_create` | `dbxcli db user create` | Standard+Echo | Create database user |
| `oracle_user_alter` | `dbxcli db user alter` | Standard | Alter user profile/quota/password |
| `oracle_user_drop` | `dbxcli db user drop` | Standard+Echo | Drop user cascade |
| `oracle_user_lock` | `dbxcli db user lock` | Standard | Lock user account |
| `oracle_user_unlock` | `dbxcli db user unlock` | Standard | Unlock user account |
| `oracle_user_grant` | `dbxcli db user grant` | Standard | Grant role or privilege |

### Schema DDL (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_schema_create_table` | `dbxcli db schema create-table` | Standard+Echo | Create table with columns/constraints |
| `oracle_schema_alter_table` | `dbxcli db schema alter-table` | Standard+Echo | Alter table (add/modify/drop column) |
| `oracle_schema_drop_table` | `dbxcli db schema drop-table` | Standard+Echo | Drop table purge |
| `oracle_schema_create_index` | `dbxcli db schema create-index` | Standard+Echo | Create index (B-tree/bitmap/function) |
| `oracle_schema_drop_index` | `dbxcli db schema drop-index` | Standard+Echo | Drop index |
| `oracle_schema_create_view` | `dbxcli db schema create-view` | Standard+Echo | Create or replace view |
| `oracle_schema_drop_view` | `dbxcli db schema drop-view` | Standard+Echo | Drop view |
| `oracle_schema_analyze` | `dbxcli db schema analyze` | Standard | Gather table/index statistics |

### Parameter (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_parameter_set_memory` | `dbxcli db parameter set-memory` | Standard | Set parameter in memory (scope=memory) |
| `oracle_parameter_set_spfile` | `dbxcli db parameter set-spfile` | Standard+Echo | Set parameter in spfile (requires restart) |
| `oracle_parameter_set_both` | `dbxcli db parameter set-both` | Standard+Echo | Set parameter scope=both |

### Redo Management (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_redo_switch` | `dbxcli db redo switch` | Standard | Force log switch |
| `oracle_redo_add` | `dbxcli db redo add` | Standard+Echo | Add redo log group |
| `oracle_redo_drop` | `dbxcli db redo drop` | Standard+Echo | Drop inactive redo log group |
| `oracle_redo_resize` | `dbxcli db redo resize` | Standard+Echo | Resize redo logs (add new, drop old) |
| `oracle_redo_archive_current` | `dbxcli db redo archive-current` | Standard | Archive current online redo |

### Undo (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_undo_guarantee_retention` | `dbxcli db undo guarantee-retention` | Standard | Enable/disable guaranteed retention |
| `oracle_undo_flashback_enable` | `dbxcli db undo flashback-enable` | Standard+Echo | Enable flashback database |

### Recovery (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_recovery_flashback_query` | `dbxcli db recovery flashback-query` | None | AS OF query for point-in-time data |
| `oracle_recovery_flashback_table` | `dbxcli db recovery flashback-table` | Standard+Echo | Flashback table to timestamp/SCN |
| `oracle_recovery_flashback_database` | `dbxcli db recovery flashback-database` | Double-Confirm | Flashback entire database |

### System (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_sys_flush_shared_pool` | `dbxcli db sys flush-shared-pool` | Standard | Flush shared pool |
| `oracle_sys_flush_buffer_cache` | `dbxcli db sys flush-buffer-cache` | Standard+Echo | Flush buffer cache |
| `oracle_sys_checkpoint` | `dbxcli db sys checkpoint` | Standard | Force checkpoint |
| `oracle_sys_restricted_session` | `dbxcli db sys restricted-session` | Standard+Echo | Enable/disable restricted session |
| `oracle_sys_kill_system_sessions` | `dbxcli db sys kill-system-sessions` | Standard+Echo | Kill all non-SYS sessions |

---

## Oracle Linux (OSS, 20 tools)

Transport: SSH. Oracle Gate: None.

### Package Management (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_linux_package_list` | `dbxcli linux package list` | None | List installed packages |
| `oracle_linux_package_install` | `dbxcli linux package install` | Standard | Install package (dnf/apt/zypper) |
| `oracle_linux_package_remove` | `dbxcli linux package remove` | Standard+Echo | Remove package |
| `oracle_linux_package_update` | `dbxcli linux package update` | Standard | Update package or all |

### Kernel (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_linux_kernel_params` | `dbxcli linux kernel params` | None | List sysctl parameters |
| `oracle_linux_kernel_modules` | `dbxcli linux kernel modules` | None | List loaded kernel modules |

### Storage (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_linux_storage_lvm` | `dbxcli linux storage lvm` | None | LVM volume groups, logical volumes |
| `oracle_linux_storage_fs` | `dbxcli linux storage fs` | None | Filesystem list with types |
| `oracle_linux_storage_df` | `dbxcli linux storage df` | None | Disk free space per mount |
| `oracle_linux_storage_mount` | `dbxcli linux storage mount` | None | Active mount points |

### Network (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_linux_network_interfaces` | `dbxcli linux network interfaces` | None | Network interface status |
| `oracle_linux_network_routes` | `dbxcli linux network routes` | None | Routing table |
| `oracle_linux_network_bonding` | `dbxcli linux network bonding` | None | Bond interface status |
| `oracle_linux_network_dns` | `dbxcli linux network dns` | None | DNS resolver configuration |

### Security (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_linux_security_firewall` | `dbxcli linux security firewall` | None | Firewall rules (firewalld/ufw) |
| `oracle_linux_security_selinux` | `dbxcli linux security selinux` | None | SELinux mode and status |
| `oracle_linux_security_users` | `dbxcli linux security users` | None | OS user accounts |
| `oracle_linux_security_sudoers` | `dbxcli linux security sudoers` | None | Sudoers entries |

### System (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_linux_service_list` | `dbxcli linux service list` | None | Systemd unit list |
| `oracle_linux_service_status` | `dbxcli linux service status` | None | Single service detail |

---

## Performance (Core, 32 tools)

Transport: SQL*Net. Oracle Gate: diagnostics_pack (ASH/AWR/ADDM), tuning_pack (SQL Tuning).
Statspack tools have no Oracle Gate (SE2 alternative).

### ASH --- Active Session History (6 tools)

| Tool | CLI | Confirm | Oracle Gate | Description |
|------|-----|---------|-------------|-------------|
| `oracle_ash_top` | `dbxcli perf ash top` | None | diagnostics_pack | Top sessions by wait class |
| `oracle_ash_history` | `dbxcli perf ash history` | None | diagnostics_pack | Session activity over time range |
| `oracle_ash_sql` | `dbxcli perf ash sql` | None | diagnostics_pack | SQL-level ASH analysis |
| `oracle_ash_wait_chains` | `dbxcli perf ash wait-chains` | None | diagnostics_pack | Blocking session chains |
| `oracle_ash_compare` | `dbxcli perf ash compare` | None | diagnostics_pack | Compare two time periods |
| `oracle_ash_dimensions` | `dbxcli perf ash dimensions` | None | diagnostics_pack | Pivot by any ASH dimension |

### AWR --- Automatic Workload Repository (7 tools)

| Tool | CLI | Confirm | Oracle Gate | Description |
|------|-----|---------|-------------|-------------|
| `oracle_awr_report` | `dbxcli perf awr report` | None | diagnostics_pack | Generate AWR report (text/HTML) |
| `oracle_awr_snapshots` | `dbxcli perf awr snapshots` | None | diagnostics_pack | List AWR snapshots |
| `oracle_awr_create_snapshot` | `dbxcli perf awr create-snapshot` | Standard | diagnostics_pack | Create manual snapshot |
| `oracle_awr_drop_snapshot` | `dbxcli perf awr drop-snapshot` | Standard | diagnostics_pack | Drop snapshot range |
| `oracle_awr_baseline_create` | `dbxcli perf awr baseline-create` | Standard | diagnostics_pack | Create AWR baseline |
| `oracle_awr_baseline_list` | `dbxcli perf awr baseline-list` | None | diagnostics_pack | List baselines |
| `oracle_awr_diff` | `dbxcli perf awr diff` | None | diagnostics_pack | Compare two AWR periods |

### ADDM (5 tools)

| Tool | CLI | Confirm | Oracle Gate | Description |
|------|-----|---------|-------------|-------------|
| `oracle_addm_report` | `dbxcli perf addm report` | None | diagnostics_pack | Generate ADDM report |
| `oracle_addm_findings` | `dbxcli perf addm findings` | None | diagnostics_pack | List ADDM findings by impact |
| `oracle_addm_recommendations` | `dbxcli perf addm recommendations` | None | diagnostics_pack | ADDM recommendations |
| `oracle_addm_history` | `dbxcli perf addm history` | None | diagnostics_pack | Historical ADDM results |
| `oracle_addm_task_create` | `dbxcli perf addm task-create` | Standard | diagnostics_pack | Create manual ADDM task |

### SQL Tuning Advisor (6 tools)

| Tool | CLI | Confirm | Oracle Gate | Description |
|------|-----|---------|-------------|-------------|
| `oracle_sqltune_analyze` | `dbxcli perf sqltune analyze` | Standard | tuning_pack | Analyze SQL statement |
| `oracle_sqltune_report` | `dbxcli perf sqltune report` | None | tuning_pack | SQL Tuning report |
| `oracle_sqltune_accept_profile` | `dbxcli perf sqltune accept-profile` | Standard | tuning_pack | Accept SQL profile |
| `oracle_sqltune_drop_profile` | `dbxcli perf sqltune drop-profile` | Standard | tuning_pack | Drop SQL profile |
| `oracle_sqltune_auto_task` | `dbxcli perf sqltune auto-task` | None | tuning_pack | Auto SQL Tuning task status |
| `oracle_sqltune_sql_monitor` | `dbxcli perf sqltune sql-monitor` | None | tuning_pack | Real-time SQL monitoring |

### Wait Events (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_wait_top` | `dbxcli perf wait top` | None | Top wait events by time |
| `oracle_wait_history` | `dbxcli perf wait history` | None | Wait event history |
| `oracle_wait_class_summary` | `dbxcli perf wait class-summary` | None | Summary by wait class |
| `oracle_wait_system_events` | `dbxcli perf wait system-events` | None | System-level wait statistics |

### Statspack (4 tools, no Oracle Gate)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_statspack_snap` | `dbxcli perf statspack snap` | Standard | Create Statspack snapshot |
| `oracle_statspack_report` | `dbxcli perf statspack report` | None | Generate Statspack report |
| `oracle_statspack_purge` | `dbxcli perf statspack purge` | Standard | Purge old snapshots |
| `oracle_statspack_baseline` | `dbxcli perf statspack baseline` | Standard | Create Statspack baseline |

---

## Unified Audit (Core, 22 tools)

Transport: SQL*Net. Oracle Gate: None (Unified Audit is standard in 12c+).

### Audit Policies (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_audit_policy_list` | `dbxcli audit policy list` | None | List audit policies |
| `oracle_audit_policy_create` | `dbxcli audit policy create` | Standard | Create unified audit policy |
| `oracle_audit_policy_enable` | `dbxcli audit policy enable` | Standard | Enable policy for users |
| `oracle_audit_policy_disable` | `dbxcli audit policy disable` | Standard | Disable audit policy |
| `oracle_audit_policy_drop` | `dbxcli audit policy drop` | Standard+Echo | Drop audit policy |
| `oracle_audit_policy_describe` | `dbxcli audit policy describe` | None | Policy definition detail |
| `oracle_audit_top_activities` | `dbxcli audit top-activities` | None | Top audited activities |
| `oracle_audit_settings` | `dbxcli audit settings` | None | Audit configuration settings |

### Fine-Grained Auditing (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_fga_add_policy` | `dbxcli audit fga add-policy` | Standard | Add FGA policy |
| `oracle_fga_drop_policy` | `dbxcli audit fga drop-policy` | Standard+Echo | Drop FGA policy |
| `oracle_fga_list` | `dbxcli audit fga list` | None | List FGA policies |
| `oracle_fga_enable` | `dbxcli audit fga enable` | Standard | Enable FGA policy |
| `oracle_fga_disable` | `dbxcli audit fga disable` | Standard | Disable FGA policy |
| `oracle_fga_trail` | `dbxcli audit fga trail` | None | Query FGA audit trail |

### Audit Trail Analysis (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_audit_query` | `dbxcli audit query` | None | Query unified audit trail |
| `oracle_audit_report` | `dbxcli audit report` | None | Generate audit report |
| `oracle_audit_archive` | `dbxcli audit archive` | Standard | Archive old audit records |
| `oracle_audit_export` | `dbxcli audit export` | None | Export audit data (CSV/JSON) |
| `oracle_audit_analyze` | `dbxcli audit analyze` | None | Analyze audit patterns |
| `oracle_audit_clean` | `dbxcli audit clean` | Standard+Echo | Purge old audit records |
| `oracle_audit_count` | `dbxcli audit count` | None | Audit record counts by policy |
| `oracle_audit_dashboard` | `dbxcli audit dashboard` | None | Audit activity dashboard |

---

## Partitioning (Core, 26 tools)

Transport: SQL*Net. Oracle Gate: partitioning.

### Partition Info (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_partition_list` | `dbxcli partition list` | None | List partitioned tables |
| `oracle_partition_describe` | `dbxcli partition describe` | None | Partition detail for a table |
| `oracle_partition_stats` | `dbxcli partition stats` | None | Partition statistics |
| `oracle_partition_keys` | `dbxcli partition keys` | None | Partition key columns |
| `oracle_partition_columns` | `dbxcli partition columns` | None | All columns in partitioned table |
| `oracle_partition_subpartitions` | `dbxcli partition subpartitions` | None | List subpartitions |
| `oracle_partition_indexes` | `dbxcli partition indexes` | None | Local/global partition indexes |
| `oracle_partition_ddl` | `dbxcli partition ddl` | None | Generate partition DDL |

### Create Partitioned Tables (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_partition_create_range` | `dbxcli partition create-range` | Standard+Echo | Create range-partitioned table |
| `oracle_partition_create_list` | `dbxcli partition create-list` | Standard+Echo | Create list-partitioned table |
| `oracle_partition_create_hash` | `dbxcli partition create-hash` | Standard+Echo | Create hash-partitioned table |
| `oracle_partition_create_interval` | `dbxcli partition create-interval` | Standard+Echo | Create interval-partitioned table |
| `oracle_partition_create_reference` | `dbxcli partition create-reference` | Standard+Echo | Create reference-partitioned table |
| `oracle_partition_convert` | `dbxcli partition convert` | Standard+Echo | Convert non-partitioned to partitioned |
| `oracle_partition_coalesce` | `dbxcli partition coalesce` | Standard | Coalesce hash partitions |
| `oracle_partition_set_interval` | `dbxcli partition set-interval` | Standard | Set interval value |

### Composite Partitioning (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_partition_create_composite` | `dbxcli partition create-composite` | Standard+Echo | Create composite partitioned table |
| `oracle_partition_add_subpartition` | `dbxcli partition add-subpartition` | Standard | Add subpartition |
| `oracle_partition_template` | `dbxcli partition template` | Standard | Set subpartition template |
| `oracle_partition_modify_composite` | `dbxcli partition modify-composite` | Standard | Modify composite partition |

### Partition Maintenance (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_partition_split` | `dbxcli partition split` | Standard+Echo | Split partition |
| `oracle_partition_merge` | `dbxcli partition merge` | Standard+Echo | Merge partitions |
| `oracle_partition_exchange` | `dbxcli partition exchange` | Standard+Echo | Exchange partition with table |
| `oracle_partition_truncate` | `dbxcli partition truncate` | Standard+Echo | Truncate partition |
| `oracle_partition_drop` | `dbxcli partition drop` | Standard+Echo | Drop partition |
| `oracle_partition_move` | `dbxcli partition move` | Standard+Echo | Move partition to tablespace |

---

## Data Guard (HA, 28 tools)

Transport: SQL*Net + SSH (dgmgrl).

### Broker Configuration (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_dg_show_config` | `dbxcli dg show-config` | None | Show DG broker configuration |
| `oracle_dg_create_config` | `dbxcli dg create-config` | Standard+Echo | Create DG broker config |
| `oracle_dg_enable` | `dbxcli dg enable` | Standard | Enable DG configuration |
| `oracle_dg_disable` | `dbxcli dg disable` | Standard | Disable DG configuration |
| `oracle_dg_remove` | `dbxcli dg remove` | Standard+Echo | Remove DG configuration |
| `oracle_dg_validate` | `dbxcli dg validate` | None | Validate DG configuration |

### Switchover (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_dg_switchover_prepare` | `dbxcli dg switchover prepare` | None | Pre-switchover validation |
| `oracle_dg_switchover_execute` | `dbxcli dg switchover execute` | Standard+Echo | Execute role switchover |
| `oracle_dg_switchover_rollback` | `dbxcli dg switchover rollback` | Standard+Echo | Rollback failed switchover |

### Failover (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_dg_failover_prepare` | `dbxcli dg failover prepare` | None | Pre-failover validation |
| `oracle_dg_failover_execute` | `dbxcli dg failover execute` | Double-Confirm | Execute failover (data loss possible) |
| `oracle_dg_failover_reinstate` | `dbxcli dg failover reinstate` | Standard+Echo | Reinstate old primary as standby |

### Validate (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_dg_validate_database` | `dbxcli dg validate database` | None | Validate database readiness |
| `oracle_dg_validate_transport` | `dbxcli dg validate transport` | None | Validate redo transport |
| `oracle_dg_validate_apply` | `dbxcli dg validate apply` | None | Validate apply configuration |
| `oracle_dg_validate_gap` | `dbxcli dg validate gap` | None | Check for redo gaps |

### Transport (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_dg_transport_status` | `dbxcli dg transport status` | None | Transport lag and status |
| `oracle_dg_transport_lag` | `dbxcli dg transport lag` | None | Current transport lag |
| `oracle_dg_transport_mode` | `dbxcli dg transport mode` | None | SYNC/ASYNC mode |
| `oracle_dg_transport_configure` | `dbxcli dg transport configure` | Standard | Configure transport mode |

### Apply (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_dg_apply_status` | `dbxcli dg apply status` | None | Apply process status |
| `oracle_dg_apply_lag` | `dbxcli dg apply lag` | None | Current apply lag |
| `oracle_dg_apply_rate` | `dbxcli dg apply rate` | None | Apply rate (MB/s) |
| `oracle_dg_apply_configure` | `dbxcli dg apply configure` | Standard | Configure apply settings |

### GAP Resolution (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_dg_gap_detect` | `dbxcli dg gap detect` | None | Detect redo gaps |
| `oracle_dg_gap_resolve` | `dbxcli dg gap resolve` | Standard | Resolve gap (ship missing logs) |
| `oracle_dg_gap_archive` | `dbxcli dg gap archive` | Standard | Archive gap logs |
| `oracle_dg_gap_status` | `dbxcli dg gap status` | None | Gap resolution status |

---

## Backup --- RMAN (HA, 34 tools)

Transport: SSH (rman). Oracle Gate: None (RMAN is standard).

### Backup (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rman_backup_full` | `dbxcli backup full` | Standard | Full database backup |
| `oracle_rman_backup_incr_l0` | `dbxcli backup incr-l0` | Standard | Incremental level 0 |
| `oracle_rman_backup_incr_l1` | `dbxcli backup incr-l1` | Standard | Incremental level 1 |
| `oracle_rman_backup_archivelog` | `dbxcli backup archivelog` | Standard | Archive log backup |
| `oracle_rman_backup_controlfile` | `dbxcli backup controlfile` | Standard | Controlfile backup |
| `oracle_rman_backup_spfile` | `dbxcli backup spfile` | Standard | SPFILE backup |
| `oracle_rman_backup_validate` | `dbxcli backup validate` | None | Validate backup integrity |
| `oracle_rman_backup_policy` | `dbxcli backup policy` | None | Show backup retention policy |

### Restore (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rman_restore_database` | `dbxcli restore database` | Double-Confirm | Restore full database |
| `oracle_rman_restore_tablespace` | `dbxcli restore tablespace` | Standard+Echo | Restore tablespace |
| `oracle_rman_restore_datafile` | `dbxcli restore datafile` | Standard+Echo | Restore specific datafile |
| `oracle_rman_restore_controlfile` | `dbxcli restore controlfile` | Double-Confirm | Restore controlfile |
| `oracle_rman_restore_spfile` | `dbxcli restore spfile` | Standard+Echo | Restore SPFILE |
| `oracle_rman_restore_archivelog` | `dbxcli restore archivelog` | Standard | Restore archive logs |

### Recover (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rman_recover_database` | `dbxcli recover database` | Double-Confirm | Recover database to current |
| `oracle_rman_recover_tablespace` | `dbxcli recover tablespace` | Standard+Echo | Recover tablespace |
| `oracle_rman_recover_datafile` | `dbxcli recover datafile` | Standard+Echo | Recover datafile |
| `oracle_rman_recover_pitr` | `dbxcli recover pitr` | Double-Confirm | Point-in-time recovery |

### Catalog (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rman_catalog_crosscheck` | `dbxcli catalog crosscheck` | Standard | Crosscheck backup pieces |
| `oracle_rman_catalog_delete_obsolete` | `dbxcli catalog delete-obsolete` | Standard+Echo | Delete obsolete backups |
| `oracle_rman_catalog_list` | `dbxcli catalog list` | None | List backup sets |
| `oracle_rman_catalog_report` | `dbxcli catalog report` | None | Backup report |
| `oracle_rman_catalog_resync` | `dbxcli catalog resync` | Standard | Resync catalog |
| `oracle_rman_catalog_register` | `dbxcli catalog register` | Standard | Register database in catalog |

### SBT (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rman_sbt_channel_config` | `dbxcli sbt channel-config` | Standard | Configure SBT channel |
| `oracle_rman_sbt_backup_piece` | `dbxcli sbt backup-piece` | None | List SBT backup pieces |
| `oracle_rman_sbt_validate` | `dbxcli sbt validate` | None | Validate SBT pieces |
| `oracle_rman_sbt_delete` | `dbxcli sbt delete` | Standard+Echo | Delete SBT pieces |

### Job Management (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rman_job_status` | `dbxcli backup job status` | None | Running backup job status |
| `oracle_rman_job_history` | `dbxcli backup job history` | None | Backup job history |
| `oracle_rman_job_cancel` | `dbxcli backup job cancel` | Standard | Cancel running backup |
| `oracle_rman_job_schedule` | `dbxcli backup job schedule` | Standard | Schedule backup job |
| `oracle_rman_job_report` | `dbxcli backup job report` | None | Job execution report |
| `oracle_rman_job_clean` | `dbxcli backup job clean` | Standard | Clean old job records |

---

## RAC (HA, 24 tools)

Transport: SQL*Net + SSH (srvctl).

### Instance Management (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rac_instance_list` | `dbxcli rac instance list` | None | List RAC instances |
| `oracle_rac_instance_start` | `dbxcli rac instance start` | Standard+Echo | Start instance |
| `oracle_rac_instance_stop` | `dbxcli rac instance stop` | Standard+Echo | Stop instance |
| `oracle_rac_instance_status` | `dbxcli rac instance status` | None | Instance status detail |
| `oracle_rac_instance_alter` | `dbxcli rac instance alter` | Standard | Alter instance parameters |
| `oracle_rac_instance_shutdown` | `dbxcli rac instance shutdown` | Double-Confirm | Shutdown instance (abort) |

### Services (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rac_service_list` | `dbxcli rac service list` | None | List RAC services |
| `oracle_rac_service_add` | `dbxcli rac service add` | Standard | Add service |
| `oracle_rac_service_remove` | `dbxcli rac service remove` | Standard+Echo | Remove service |
| `oracle_rac_service_relocate` | `dbxcli rac service relocate` | Standard | Relocate service to node |
| `oracle_rac_service_start` | `dbxcli rac service start` | Standard | Start service |
| `oracle_rac_service_stop` | `dbxcli rac service stop` | Standard | Stop service |
| `oracle_rac_service_status` | `dbxcli rac service status` | None | Service status |
| `oracle_rac_service_modify` | `dbxcli rac service modify` | Standard | Modify service properties |

### Interconnect (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rac_interconnect_status` | `dbxcli rac interconnect status` | None | Interconnect link status |
| `oracle_rac_interconnect_stats` | `dbxcli rac interconnect stats` | None | Interconnect throughput stats |
| `oracle_rac_interconnect_latency` | `dbxcli rac interconnect latency` | None | Interconnect latency |
| `oracle_rac_interconnect_errors` | `dbxcli rac interconnect errors` | None | Interconnect error counts |

### Cache Fusion (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rac_cache_stats` | `dbxcli rac cache stats` | None | Cache fusion statistics |
| `oracle_rac_cache_gc_blocks` | `dbxcli rac cache gc-blocks` | None | Global cache block transfers |
| `oracle_rac_cache_cr_blocks` | `dbxcli rac cache cr-blocks` | None | CR block statistics |
| `oracle_rac_cache_current_blocks` | `dbxcli rac cache current-blocks` | None | Current block statistics |

### GES/GCS (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_rac_ges_enqueue_stats` | `dbxcli rac ges enqueue-stats` | None | Global enqueue stats |
| `oracle_rac_gcs_lock_stats` | `dbxcli rac gcs lock-stats` | None | GCS lock statistics |

---

## Clusterware (HA, 22 tools)

Transport: SSH (crsctl, srvctl).

### CRS Resources (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_crs_resource_list` | `dbxcli crs resource list` | None | List CRS resources |
| `oracle_crs_resource_start` | `dbxcli crs resource start` | Standard | Start resource |
| `oracle_crs_resource_stop` | `dbxcli crs resource stop` | Standard | Stop resource |
| `oracle_crs_resource_relocate` | `dbxcli crs resource relocate` | Standard | Relocate resource |
| `oracle_crs_resource_status` | `dbxcli crs resource status` | None | Resource status detail |
| `oracle_crs_resource_check` | `dbxcli crs resource check` | None | Resource health check |

### Node Management (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_crs_node_list` | `dbxcli crs node list` | None | List cluster nodes |
| `oracle_crs_node_add` | `dbxcli crs node add` | Standard+Echo | Add node to cluster |
| `oracle_crs_node_remove` | `dbxcli crs node remove` | Double-Confirm | Remove node from cluster |
| `oracle_crs_node_status` | `dbxcli crs node status` | None | Node status |

### VIP (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_crs_vip_list` | `dbxcli crs vip list` | None | List VIPs |
| `oracle_crs_vip_start` | `dbxcli crs vip start` | Standard | Start VIP |
| `oracle_crs_vip_stop` | `dbxcli crs vip stop` | Standard | Stop VIP |

### SCAN (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_crs_scan_list` | `dbxcli crs scan list` | None | List SCAN listeners |
| `oracle_crs_scan_status` | `dbxcli crs scan status` | None | SCAN listener status |
| `oracle_crs_scan_add` | `dbxcli crs scan add` | Standard+Echo | Add SCAN listener |

### Voting Disk (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_crs_voting_list` | `dbxcli crs voting list` | None | List voting disks |
| `oracle_crs_voting_add` | `dbxcli crs voting add` | Standard+Echo | Add voting disk |
| `oracle_crs_voting_remove` | `dbxcli crs voting remove` | Double-Confirm | Remove voting disk |

### OCR (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_crs_ocr_backup` | `dbxcli crs ocr backup` | Standard | Backup OCR |
| `oracle_crs_ocr_restore` | `dbxcli crs ocr restore` | Double-Confirm | Restore OCR |
| `oracle_crs_ocr_export` | `dbxcli crs ocr export` | Standard | Export OCR to file |

---

## ASM (HA, 20 tools)

Transport: SQL*Net (ASM instance) + SSH (asmcmd).

### Diskgroup (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_asm_diskgroup_list` | `dbxcli asm diskgroup list` | None | List diskgroups |
| `oracle_asm_diskgroup_create` | `dbxcli asm diskgroup create` | Standard+Echo | Create diskgroup |
| `oracle_asm_diskgroup_resize` | `dbxcli asm diskgroup resize` | Standard+Echo | Resize diskgroup |
| `oracle_asm_diskgroup_drop` | `dbxcli asm diskgroup drop` | Double-Confirm | Drop diskgroup |
| `oracle_asm_diskgroup_mount` | `dbxcli asm diskgroup mount` | Standard | Mount diskgroup |
| `oracle_asm_diskgroup_dismount` | `dbxcli asm diskgroup dismount` | Standard | Dismount diskgroup |

### Disk (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_asm_disk_list` | `dbxcli asm disk list` | None | List ASM disks |
| `oracle_asm_disk_add` | `dbxcli asm disk add` | Standard+Echo | Add disk to diskgroup |
| `oracle_asm_disk_remove` | `dbxcli asm disk remove` | Standard+Echo | Remove disk (rebalance) |
| `oracle_asm_disk_replace` | `dbxcli asm disk replace` | Standard+Echo | Replace failed disk |
| `oracle_asm_disk_online` | `dbxcli asm disk online` | Standard | Online disk |
| `oracle_asm_disk_offline` | `dbxcli asm disk offline` | Standard | Offline disk |

### ACFS (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_asm_acfs_list` | `dbxcli asm acfs list` | None | List ACFS filesystems |
| `oracle_asm_acfs_create` | `dbxcli asm acfs create` | Standard+Echo | Create ACFS volume |
| `oracle_asm_acfs_resize` | `dbxcli asm acfs resize` | Standard+Echo | Resize ACFS volume |
| `oracle_asm_acfs_snapshot` | `dbxcli asm acfs snapshot` | Standard | Create ACFS snapshot |

### Rebalance (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_asm_rebalance_status` | `dbxcli asm rebalance status` | None | Rebalance progress |
| `oracle_asm_rebalance_start` | `dbxcli asm rebalance start` | Standard | Start manual rebalance |
| `oracle_asm_rebalance_stop` | `dbxcli asm rebalance stop` | Standard | Stop rebalance |
| `oracle_asm_rebalance_power` | `dbxcli asm rebalance power` | Standard | Set rebalance power |

---

## Provisioning (Ops, 22 tools)

Transport: SSH (dbca, PL/SQL).

### DBCA (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_provision_create_db` | `dbxcli provision create-db` | Standard+Echo | Create database (DBCA) |
| `oracle_provision_create_cdb` | `dbxcli provision create-cdb` | Standard+Echo | Create CDB |
| `oracle_provision_clone_hot` | `dbxcli provision clone-hot` | Standard+Echo | Hot clone database |
| `oracle_provision_clone_cold` | `dbxcli provision clone-cold` | Standard+Echo | Cold clone database |
| `oracle_provision_delete` | `dbxcli provision delete` | Double-Confirm | Delete database |
| `oracle_provision_template_list` | `dbxcli provision template-list` | None | List DBCA templates |

### PDB Lifecycle (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_provision_pdb_create` | `dbxcli provision pdb create` | Standard+Echo | Create PDB |
| `oracle_provision_pdb_clone` | `dbxcli provision pdb clone` | Standard+Echo | Clone PDB |
| `oracle_provision_pdb_plug` | `dbxcli provision pdb plug` | Standard+Echo | Plug in PDB |
| `oracle_provision_pdb_unplug` | `dbxcli provision pdb unplug` | Standard+Echo | Unplug PDB |
| `oracle_provision_pdb_drop` | `dbxcli provision pdb drop` | Double-Confirm | Drop PDB |
| `oracle_provision_pdb_open` | `dbxcli provision pdb open` | Standard | Open PDB |
| `oracle_provision_pdb_close` | `dbxcli provision pdb close` | Standard | Close PDB |
| `oracle_provision_pdb_relocate` | `dbxcli provision pdb relocate` | Standard+Echo | Relocate PDB to CDB |

### Template (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_provision_template_list2` | `dbxcli provision template list` | None | List templates |
| `oracle_provision_template_create` | `dbxcli provision template create` | Standard | Create template from DB |
| `oracle_provision_template_import` | `dbxcli provision template import` | Standard | Import template |
| `oracle_provision_template_export` | `dbxcli provision template export` | Standard | Export template |

### Silent Mode (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_provision_silent_generate` | `dbxcli provision silent generate` | None | Generate response file |
| `oracle_provision_silent_validate` | `dbxcli provision silent validate` | None | Validate response file |
| `oracle_provision_silent_execute` | `dbxcli provision silent execute` | Standard+Echo | Execute silent install |
| `oracle_provision_silent_log` | `dbxcli provision silent log` | None | View silent install log |

---

## Patching (Ops, 30 tools)

Transport: SSH (opatch, datapatch) + SQL*Net.

### OPatch (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_patch_opatch_apply` | `dbxcli patch apply` | Standard+Echo | Apply patch |
| `oracle_patch_opatch_rollback` | `dbxcli patch rollback` | Standard+Echo | Rollback patch |
| `oracle_patch_opatch_lsinventory` | `dbxcli patch lsinventory` | None | List installed patches |
| `oracle_patch_opatch_conflict` | `dbxcli patch conflict` | None | Check patch conflicts |
| `oracle_patch_opatch_prerequisite` | `dbxcli patch prerequisite` | None | Run prerequisite checks |
| `oracle_patch_opatch_verify` | `dbxcli patch verify` | None | Verify patch integrity |
| `oracle_patch_opatch_version` | `dbxcli patch version` | None | OPatch version |
| `oracle_patch_opatch_history` | `dbxcli patch history` | None | Patch history |

### Datapatch (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_patch_datapatch_apply` | `dbxcli patch datapatch apply` | Standard | Apply datapatch |
| `oracle_patch_datapatch_rollback` | `dbxcli patch datapatch rollback` | Standard+Echo | Rollback datapatch |
| `oracle_patch_datapatch_status` | `dbxcli patch datapatch status` | None | Datapatch status |
| `oracle_patch_datapatch_history` | `dbxcli patch datapatch history` | None | Datapatch history |

### Gold Image (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_patch_gold_build` | `dbxcli patch gold build` | Standard+Echo | Build gold image |
| `oracle_patch_gold_deploy` | `dbxcli patch gold deploy` | Standard+Echo | Deploy gold image to target |
| `oracle_patch_gold_list` | `dbxcli patch gold list` | None | List gold images |
| `oracle_patch_gold_compare` | `dbxcli patch gold compare` | None | Compare target to gold image |
| `oracle_patch_gold_validate` | `dbxcli patch gold validate` | None | Validate gold image |
| `oracle_patch_gold_clean` | `dbxcli patch gold clean` | Standard | Clean old gold images |

### Patch Analysis (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_patch_analyze` | `dbxcli patch analyze` | None | Analyze available patches |
| `oracle_patch_mos_advisory` | `dbxcli patch mos-advisory` | None | MOS patch advisories |
| `oracle_patch_download` | `dbxcli patch download` | Standard | Download patch from MOS |
| `oracle_patch_plan` | `dbxcli patch plan` | None | Create patching plan |
| `oracle_patch_verify_plan` | `dbxcli patch verify-plan` | None | Verify patching plan |
| `oracle_patch_report` | `dbxcli patch report` | None | Patch compliance report |

### Fleet Patching (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_patch_fleet_status` | `dbxcli patch fleet status` | None | Fleet patch levels |
| `oracle_patch_fleet_apply` | `dbxcli patch fleet apply` | Standard+Echo | Fleet-wide patch apply |
| `oracle_patch_fleet_rollback` | `dbxcli patch fleet rollback` | Standard+Echo | Fleet-wide rollback |
| `oracle_patch_fleet_report` | `dbxcli patch fleet report` | None | Fleet compliance report |
| `oracle_patch_fleet_plan` | `dbxcli patch fleet plan` | None | Fleet patching plan |
| `oracle_patch_fleet_verify` | `dbxcli patch fleet verify` | None | Verify fleet readiness |

---

## Migration (Ops, 20 tools)

Transport: SSH + SQL*Net.

### AutoUpgrade (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_migrate_autoupgrade_precheck` | `dbxcli migrate precheck` | None | Pre-upgrade checks |
| `oracle_migrate_autoupgrade_fixup` | `dbxcli migrate fixup` | Standard | Apply fixups |
| `oracle_migrate_autoupgrade_upgrade` | `dbxcli migrate upgrade` | Standard+Echo | Execute upgrade |
| `oracle_migrate_autoupgrade_rollback` | `dbxcli migrate rollback` | Standard+Echo | Rollback upgrade |
| `oracle_migrate_autoupgrade_status` | `dbxcli migrate status` | None | Upgrade status |
| `oracle_migrate_autoupgrade_log` | `dbxcli migrate log` | None | Upgrade log viewer |

### PDB Convert (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_migrate_pdb_analyze` | `dbxcli migrate pdb analyze` | None | Analyze non-CDB for conversion |
| `oracle_migrate_pdb_execute` | `dbxcli migrate pdb execute` | Standard+Echo | Convert non-CDB to PDB |
| `oracle_migrate_pdb_validate` | `dbxcli migrate pdb validate` | None | Validate conversion |
| `oracle_migrate_pdb_rollback` | `dbxcli migrate pdb rollback` | Standard+Echo | Rollback conversion |

### DBUA (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_migrate_dbua_precheck` | `dbxcli migrate dbua precheck` | None | DBUA precheck |
| `oracle_migrate_dbua_upgrade` | `dbxcli migrate dbua upgrade` | Standard+Echo | DBUA upgrade |
| `oracle_migrate_dbua_status` | `dbxcli migrate dbua status` | None | DBUA status |
| `oracle_migrate_dbua_log` | `dbxcli migrate dbua log` | None | DBUA log |

### Compatibility (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_migrate_compat_check` | `dbxcli migrate compat check` | None | Compatibility check |
| `oracle_migrate_compat_matrix` | `dbxcli migrate compat matrix` | None | Version compatibility matrix |
| `oracle_migrate_compat_report` | `dbxcli migrate compat report` | None | Compatibility report |
| `oracle_migrate_compat_advisor` | `dbxcli migrate compat advisor` | None | Upgrade path advisor |

### Transport (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_migrate_transport_export` | `dbxcli migrate transport export` | Standard | Transportable tablespace export |
| `oracle_migrate_transport_import` | `dbxcli migrate transport import` | Standard+Echo | Transportable tablespace import |

---

## Data Pump (Ops, 18 tools)

Transport: SQL*Net (DBMS_DATAPUMP).

### Export (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_datapump_export_full` | `dbxcli datapump export full` | Standard | Full database export |
| `oracle_datapump_export_schema` | `dbxcli datapump export schema` | Standard | Schema-level export |
| `oracle_datapump_export_table` | `dbxcli datapump export table` | Standard | Table-level export |
| `oracle_datapump_export_query` | `dbxcli datapump export query` | Standard | Query-filtered export |
| `oracle_datapump_export_network` | `dbxcli datapump export network` | Standard | Network link export |

### Import (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_datapump_import_full` | `dbxcli datapump import full` | Standard+Echo | Full database import |
| `oracle_datapump_import_schema` | `dbxcli datapump import schema` | Standard+Echo | Schema-level import |
| `oracle_datapump_import_table` | `dbxcli datapump import table` | Standard+Echo | Table-level import |
| `oracle_datapump_import_remap` | `dbxcli datapump import remap` | Standard+Echo | Import with schema/tablespace remap |
| `oracle_datapump_import_network` | `dbxcli datapump import network` | Standard+Echo | Network link import |

### Job Management (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_datapump_job_status` | `dbxcli datapump job status` | None | Running job status |
| `oracle_datapump_job_kill` | `dbxcli datapump job kill` | Standard | Kill running job |
| `oracle_datapump_job_parallel` | `dbxcli datapump job parallel` | Standard | Set parallelism |
| `oracle_datapump_job_estimate` | `dbxcli datapump job estimate` | None | Estimate export size |

### Filters (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_datapump_filter_include` | `dbxcli datapump filter include` | None | Set include filter |
| `oracle_datapump_filter_exclude` | `dbxcli datapump filter exclude` | None | Set exclude filter |
| `oracle_datapump_filter_content` | `dbxcli datapump filter content` | None | Data/metadata only |
| `oracle_datapump_filter_query` | `dbxcli datapump filter query` | None | Set query filter |

---

## GoldenGate (Ops, 24 tools)

Transport: REST (GG Microservices) or SSH (GGSCI).

### Extract (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_gg_extract_add` | `dbxcli gg extract add` | Standard+Echo | Add extract process |
| `oracle_gg_extract_start` | `dbxcli gg extract start` | Standard | Start extract |
| `oracle_gg_extract_stop` | `dbxcli gg extract stop` | Standard | Stop extract |
| `oracle_gg_extract_stats` | `dbxcli gg extract stats` | None | Extract statistics |
| `oracle_gg_extract_delete` | `dbxcli gg extract delete` | Standard+Echo | Delete extract |

### Replicat (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_gg_replicat_add` | `dbxcli gg replicat add` | Standard+Echo | Add replicat process |
| `oracle_gg_replicat_start` | `dbxcli gg replicat start` | Standard | Start replicat |
| `oracle_gg_replicat_stop` | `dbxcli gg replicat stop` | Standard | Stop replicat |
| `oracle_gg_replicat_stats` | `dbxcli gg replicat stats` | None | Replicat statistics |
| `oracle_gg_replicat_delete` | `dbxcli gg replicat delete` | Standard+Echo | Delete replicat |

### Pump (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_gg_pump_add` | `dbxcli gg pump add` | Standard+Echo | Add pump process |
| `oracle_gg_pump_start` | `dbxcli gg pump start` | Standard | Start pump |
| `oracle_gg_pump_stop` | `dbxcli gg pump stop` | Standard | Stop pump |
| `oracle_gg_pump_stats` | `dbxcli gg pump stats` | None | Pump statistics |

### Trail (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_gg_trail_list` | `dbxcli gg trail list` | None | List trail files |
| `oracle_gg_trail_purge` | `dbxcli gg trail purge` | Standard | Purge old trails |
| `oracle_gg_trail_repair` | `dbxcli gg trail repair` | Standard | Repair corrupted trail |
| `oracle_gg_trail_info` | `dbxcli gg trail info` | None | Trail file info |

### Heartbeat (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_gg_heartbeat_status` | `dbxcli gg heartbeat status` | None | Heartbeat status |
| `oracle_gg_heartbeat_lag` | `dbxcli gg heartbeat lag` | None | End-to-end lag |
| `oracle_gg_heartbeat_history` | `dbxcli gg heartbeat history` | None | Lag history |

### Process (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_gg_process_status` | `dbxcli gg process status` | None | All process status |
| `oracle_gg_process_all` | `dbxcli gg process all` | None | All processes detail |
| `oracle_gg_process_restart` | `dbxcli gg process restart` | Standard | Restart process |

---

## OEM Cloud Control (Ops, 28 tools)

Transport: REST (OEM Cloud Control API).

### Target Management (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_oem_target_discover` | `dbxcli oem target discover` | Standard | Auto-discover targets |
| `oracle_oem_target_add` | `dbxcli oem target add` | Standard | Add target to OEM |
| `oracle_oem_target_remove` | `dbxcli oem target remove` | Standard+Echo | Remove target |
| `oracle_oem_target_status` | `dbxcli oem target status` | None | Target status |
| `oracle_oem_target_promote` | `dbxcli oem target promote` | Standard | Promote to production |

### Job Management (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_oem_job_create` | `dbxcli oem job create` | Standard | Create OEM job |
| `oracle_oem_job_status` | `dbxcli oem job status` | None | Job execution status |
| `oracle_oem_job_history` | `dbxcli oem job history` | None | Job history |
| `oracle_oem_job_cancel` | `dbxcli oem job cancel` | Standard | Cancel running job |
| `oracle_oem_job_schedule` | `dbxcli oem job schedule` | Standard | Schedule job |
| `oracle_oem_job_delete` | `dbxcli oem job delete` | Standard+Echo | Delete job |

### Metric (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_oem_metric_query` | `dbxcli oem metric query` | None | Query metric values |
| `oracle_oem_metric_threshold` | `dbxcli oem metric threshold` | None | Metric thresholds |
| `oracle_oem_metric_alert` | `dbxcli oem metric alert` | None | Active metric alerts |
| `oracle_oem_metric_history` | `dbxcli oem metric history` | None | Metric history |
| `oracle_oem_metric_comparison` | `dbxcli oem metric comparison` | None | Compare metrics across targets |

### Compliance (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_oem_compliance_frameworks` | `dbxcli oem compliance frameworks` | None | List compliance frameworks |
| `oracle_oem_compliance_scan` | `dbxcli oem compliance scan` | Standard | Run compliance scan |
| `oracle_oem_compliance_report` | `dbxcli oem compliance report` | None | Compliance report |
| `oracle_oem_compliance_violations` | `dbxcli oem compliance violations` | None | List violations |
| `oracle_oem_compliance_baseline` | `dbxcli oem compliance baseline` | Standard | Create compliance baseline |

### Patch (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_oem_patch_recommendation` | `dbxcli oem patch recommendation` | None | Patch recommendations |
| `oracle_oem_patch_plan` | `dbxcli oem patch plan` | None | Create patch plan |
| `oracle_oem_patch_apply` | `dbxcli oem patch apply` | Standard+Echo | Apply patch via OEM |
| `oracle_oem_patch_status` | `dbxcli oem patch status` | None | Patch job status |

### Configuration (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `oracle_oem_config_search` | `dbxcli oem config search` | None | Search configurations |
| `oracle_oem_config_compare` | `dbxcli oem config compare` | None | Compare target configs |
| `oracle_oem_config_history` | `dbxcli oem config history` | None | Configuration change history |

---

## PostgreSQL Core (OSS, 27 tools)

### Connection (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_connection_connect` | `dbxcli pg connection connect` | None | Connect to database |
| `pg_connection_disconnect` | `dbxcli pg connection disconnect` | None | Disconnect |
| `pg_connection_status` | `dbxcli pg connection status` | None | Connection status |
| `pg_connection_list` | `dbxcli pg connection list` | None | List connections |
| `pg_connection_switch` | `dbxcli pg connection switch` | None | Switch active connection |

### Query (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_query` | `dbxcli pg query` | None | Execute read-only query |
| `pg_query_explain` | `dbxcli pg query explain` | None | EXPLAIN ANALYZE |
| `pg_query_prepared` | `dbxcli pg query prepared` | None | List prepared statements |

### Schema (9 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_schema_tables` | `dbxcli pg schema tables` | None | List tables |
| `pg_schema_indexes` | `dbxcli pg schema indexes` | None | List indexes |
| `pg_schema_views` | `dbxcli pg schema views` | None | List views |
| `pg_schema_functions` | `dbxcli pg schema functions` | None | List functions |
| `pg_schema_enums` | `dbxcli pg schema enums` | None | List enum types |
| `pg_schema_types` | `dbxcli pg schema types` | None | List custom types |
| `pg_schema_sequences` | `dbxcli pg schema sequences` | None | List sequences |
| `pg_schema_triggers` | `dbxcli pg schema triggers` | None | List triggers |
| `pg_schema_constraints` | `dbxcli pg schema constraints` | None | List constraints |

### CRUD (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_crud_insert` | `dbxcli pg crud insert` | Standard | Insert rows |
| `pg_crud_update` | `dbxcli pg crud update` | Standard | Update rows |
| `pg_crud_delete` | `dbxcli pg crud delete` | Standard+Echo | Delete rows |
| `pg_crud_upsert` | `dbxcli pg crud upsert` | Standard | Insert or update |

### Server (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_server_version` | `dbxcli pg server version` | None | PostgreSQL version |
| `pg_server_settings` | `dbxcli pg server settings` | None | Server settings (pg_settings) |
| `pg_server_reload` | `dbxcli pg server reload` | Standard | Reload configuration |
| `pg_server_uptime` | `dbxcli pg server uptime` | None | Server uptime |

### Database (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_database_size` | `dbxcli pg database size` | None | Database size |
| `pg_table_sizes` | `dbxcli pg table sizes` | None | Table sizes with bloat |

---

## PostgreSQL Enterprise (PG-Pro, 111 tools)

### CNPG (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_cnpg_cluster_status` | `dbxcli pg cnpg cluster-status` | None | CNPG cluster status |
| `pg_cnpg_cluster_list` | `dbxcli pg cnpg cluster-list` | None | List CNPG clusters |
| `pg_cnpg_backup_list` | `dbxcli pg cnpg backup-list` | None | List backups |
| `pg_cnpg_backup_create` | `dbxcli pg cnpg backup-create` | Standard | Create on-demand backup |
| `pg_cnpg_failover` | `dbxcli pg cnpg failover` | Standard+Echo | Trigger CNPG failover |
| `pg_cnpg_switchover` | `dbxcli pg cnpg switchover` | Standard+Echo | CNPG switchover |

### CNPG-DR (18 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_cnpg_dr_status` | `dbxcli pg cnpg-dr status` | None | DR cluster status |
| `pg_cnpg_dr_promote` | `dbxcli pg cnpg-dr promote` | Double-Confirm | Promote DR to primary |
| `pg_cnpg_dr_demote` | `dbxcli pg cnpg-dr demote` | Standard+Echo | Demote primary to replica |
| `pg_cnpg_dr_switchover` | `dbxcli pg cnpg-dr switchover` | Standard+Echo | Cross-cluster switchover |
| `pg_cnpg_dr_validate` | `dbxcli pg cnpg-dr validate` | None | Validate DR readiness |
| `pg_cnpg_dr_lag` | `dbxcli pg cnpg-dr lag` | None | Cross-cluster replication lag |
| `pg_cnpg_dr_wal_status` | `dbxcli pg cnpg-dr wal-status` | None | WAL archive status |
| `pg_cnpg_dr_wal_verify` | `dbxcli pg cnpg-dr wal-verify` | None | Verify WAL continuity |
| `pg_cnpg_dr_restore` | `dbxcli pg cnpg-dr restore` | Double-Confirm | Restore from DR backup |
| `pg_cnpg_dr_pitr` | `dbxcli pg cnpg-dr pitr` | Double-Confirm | Point-in-time recovery |
| `pg_cnpg_dr_backup_status` | `dbxcli pg cnpg-dr backup-status` | None | DR backup status |
| `pg_cnpg_dr_backup_list` | `dbxcli pg cnpg-dr backup-list` | None | List DR backups |
| `pg_cnpg_dr_minio_status` | `dbxcli pg cnpg-dr minio-status` | None | MinIO WAL archive status |
| `pg_cnpg_dr_test_failover` | `dbxcli pg cnpg-dr test-failover` | Standard+Echo | DR failover drill |
| `pg_cnpg_dr_network_check` | `dbxcli pg cnpg-dr network-check` | None | Cross-cluster connectivity |
| `pg_cnpg_dr_config` | `dbxcli pg cnpg-dr config` | None | DR configuration |
| `pg_cnpg_dr_history` | `dbxcli pg cnpg-dr history` | None | DR event history |
| `pg_cnpg_dr_runbook` | `dbxcli pg cnpg-dr runbook` | None | DR runbook generator |

### HA (10 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_ha_status` | `dbxcli pg ha status` | None | HA cluster status |
| `pg_ha_failover` | `dbxcli pg ha failover` | Double-Confirm | Manual failover |
| `pg_ha_switchover` | `dbxcli pg ha switchover` | Standard+Echo | Planned switchover |
| `pg_ha_readiness` | `dbxcli pg ha readiness` | None | HA readiness check |
| `pg_ha_timeline` | `dbxcli pg ha timeline` | None | Timeline history |
| `pg_ha_replica_list` | `dbxcli pg ha replica-list` | None | List replicas |
| `pg_ha_replica_lag` | `dbxcli pg ha replica-lag` | None | Per-replica lag |
| `pg_ha_slot_list` | `dbxcli pg ha slot-list` | None | Replication slots |
| `pg_ha_slot_create` | `dbxcli pg ha slot-create` | Standard | Create replication slot |
| `pg_ha_slot_drop` | `dbxcli pg ha slot-drop` | Standard | Drop replication slot |

### Backup (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_backup_dump` | `dbxcli pg backup dump` | Standard | pg_dump |
| `pg_backup_restore` | `dbxcli pg backup restore` | Standard+Echo | pg_restore |
| `pg_backup_pitr` | `dbxcli pg backup pitr` | Double-Confirm | Point-in-time recovery |
| `pg_backup_status` | `dbxcli pg backup status` | None | Backup status and history |

### DBA (16 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_dba_vacuum` | `dbxcli pg dba vacuum` | Standard | VACUUM (optionally FULL) |
| `pg_dba_vacuum_status` | `dbxcli pg dba vacuum-status` | None | Autovacuum status |
| `pg_dba_reindex` | `dbxcli pg dba reindex` | Standard | REINDEX table/index |
| `pg_dba_bloat` | `dbxcli pg dba bloat` | None | Table/index bloat analysis |
| `pg_dba_sessions` | `dbxcli pg dba sessions` | None | Active sessions detail |
| `pg_dba_locks` | `dbxcli pg dba locks` | None | Lock analysis |
| `pg_dba_cancel_query` | `dbxcli pg dba cancel-query` | Standard | Cancel running query |
| `pg_dba_terminate` | `dbxcli pg dba terminate` | Standard | Terminate backend |
| `pg_dba_partitions` | `dbxcli pg dba partitions` | None | Partition summary |
| `pg_dba_extensions` | `dbxcli pg dba extensions` | None | List extensions |
| `pg_dba_toast` | `dbxcli pg dba toast` | None | TOAST table analysis |
| `pg_dba_buffer_usage` | `dbxcli pg dba buffer-usage` | None | Shared buffer usage |
| `pg_dba_checkpoint` | `dbxcli pg dba checkpoint` | Standard | Force checkpoint |
| `pg_dba_stats_reset` | `dbxcli pg dba stats-reset` | Standard+Echo | Reset statistics |
| `pg_dba_analyze` | `dbxcli pg dba analyze` | Standard | ANALYZE table |
| `pg_dba_autovacuum_check` | `dbxcli pg dba autovacuum-check` | None | Autovacuum health check |

### Security (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_security_ssl_status` | `dbxcli pg security ssl-status` | None | SSL connection status |
| `pg_security_hba_check` | `dbxcli pg security hba-check` | None | pg_hba.conf analysis |
| `pg_security_password_policy` | `dbxcli pg security password-policy` | None | Password policy status |
| `pg_security_privileges` | `dbxcli pg security privileges` | None | Privilege audit |

### Audit (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_audit_query_log` | `dbxcli pg audit query-log` | None | Query audit log |
| `pg_audit_connections` | `dbxcli pg audit connections` | None | Connection audit |
| `pg_audit_permissions` | `dbxcli pg audit permissions` | None | Permission changes audit |

### Compliance (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_compliance_cis_scan` | `dbxcli pg compliance cis-scan` | None | CIS benchmark scan |
| `pg_compliance_ssl_check` | `dbxcli pg compliance ssl-check` | None | SSL compliance check |
| `pg_compliance_gdpr` | `dbxcli pg compliance gdpr` | None | GDPR data classification |
| `pg_compliance_retention` | `dbxcli pg compliance retention` | None | Data retention audit |
| `pg_compliance_report` | `dbxcli pg compliance report` | None | Compliance report |

### RBAC (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_rbac_roles` | `dbxcli pg rbac roles` | None | List roles with memberships |
| `pg_rbac_grants` | `dbxcli pg rbac grants` | None | Object grants matrix |
| `pg_rbac_rls_policies` | `dbxcli pg rbac rls-policies` | None | Row-level security policies |
| `pg_rbac_default_privileges` | `dbxcli pg rbac default-privileges` | None | Default privilege settings |

### Replication (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_replication_status` | `dbxcli pg replication status` | None | Streaming replication status |
| `pg_replication_slots` | `dbxcli pg replication slots` | None | Replication slot details |
| `pg_replication_lag` | `dbxcli pg replication lag` | None | Replication lag |
| `pg_replication_conflicts` | `dbxcli pg replication conflicts` | None | Replication conflicts |

### Migration (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_migration_schema_diff` | `dbxcli pg migration schema-diff` | None | Schema diff between databases |
| `pg_migration_data_compare` | `dbxcli pg migration data-compare` | None | Data comparison |
| `pg_migration_execute` | `dbxcli pg migration execute` | Standard+Echo | Execute migration |
| `pg_migration_rollback` | `dbxcli pg migration rollback` | Standard+Echo | Rollback migration |

### Observability (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_observe_wait_events` | `dbxcli pg observe wait-events` | None | Wait event analysis |
| `pg_observe_checkpoints` | `dbxcli pg observe checkpoints` | None | Checkpoint statistics |
| `pg_observe_io` | `dbxcli pg observe io` | None | I/O statistics |
| `pg_observe_slow_queries` | `dbxcli pg observe slow-queries` | None | Slow query analysis |

### Tenant (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_tenant_list` | `dbxcli pg tenant list` | None | List tenants |
| `pg_tenant_create` | `dbxcli pg tenant create` | Standard+Echo | Create tenant schema |
| `pg_tenant_drop` | `dbxcli pg tenant drop` | Double-Confirm | Drop tenant |
| `pg_tenant_quota` | `dbxcli pg tenant quota` | None | Tenant resource quotas |
| `pg_tenant_drift` | `dbxcli pg tenant drift` | None | Schema drift detection |

### WAL (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_wal_status` | `dbxcli pg wal status` | None | WAL accumulation status |
| `pg_wal_archive_status` | `dbxcli pg wal archive-status` | None | WAL archive status |
| `pg_wal_archive_lag` | `dbxcli pg wal archive-lag` | None | Archive lag |
| `pg_wal_growth` | `dbxcli pg wal growth` | None | WAL growth rate |
| `pg_wal_clean` | `dbxcli pg wal clean` | Standard | Clean old WAL segments |

### RAG (7 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_rag_index_file` | `dbxcli pg rag index-file` | Standard | Index document file |
| `pg_rag_index_status` | `dbxcli pg rag index-status` | None | Index status |
| `pg_rag_search` | `dbxcli pg rag search` | None | Semantic search |
| `pg_rag_context` | `dbxcli pg rag context` | None | Build LLM context |
| `pg_rag_vector_store_status` | `dbxcli pg rag vector-store-status` | None | pgvector store status |
| `pg_rag_exceptions` | `dbxcli pg rag exceptions` | None | RAG exceptions |
| `pg_rag_sources` | `dbxcli pg rag sources` | None | Document sources |

### Vault (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_vault_connect` | `dbxcli pg vault connect` | None | Connect via Vault credentials |
| `pg_vault_status` | `dbxcli pg vault status` | None | Vault credential status |
| `pg_vault_rotate` | `dbxcli pg vault rotate` | Standard | Rotate credentials |

### Connection (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_ee_connection_connect` | `dbxcli pg ee-connect` | None | Connect (enterprise) |
| `pg_ee_connection_disconnect` | `dbxcli pg ee-disconnect` | None | Disconnect |
| `pg_ee_connection_status` | `dbxcli pg ee-status` | None | Connection status |

### Policy (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_policy_status` | `dbxcli pg policy status` | None | Policy engine status |
| `pg_policy_reload` | `dbxcli pg policy reload` | None | Reload policies |

### Health (1 tool)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_health_autovacuum` | `dbxcli pg health autovacuum` | None | Autovacuum health check |

### Capacity (1 tool)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_capacity_forecast` | `dbxcli pg capacity forecast` | None | Storage growth forecast |

### Performance (2 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `pg_perf_slow_queries` | `dbxcli pg perf slow-queries` | None | Current slow queries |
| `pg_perf_index_usage` | `dbxcli pg perf index-usage` | None | Index usage analysis |

---

## Host/OS Core (OSS, 20 tools)

Transport: SSH.

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `host_info` | `dbxcli host info` | None | OS info, distro, kernel, uptime |
| `host_cpu` | `dbxcli host cpu` | None | CPU usage, cores, load average |
| `host_memory` | `dbxcli host memory` | None | RAM, swap, hugepages |
| `host_disk_io` | `dbxcli host disk-io` | None | IOPS, throughput, latency per device |
| `host_disk_space` | `dbxcli host disk-space` | None | Mount points, usage, inodes |
| `host_network` | `dbxcli host network` | None | Interface stats, rx/tx, errors |
| `host_process_top` | `dbxcli host process-top` | None | Top N processes by CPU or memory |
| `host_process_list` | `dbxcli host process-list` | None | All processes, filter by user/name |
| `host_filesystem` | `dbxcli host filesystem` | None | Mount types, LVM, NFS |
| `host_kernel_params` | `dbxcli host kernel-params` | None | Sysctl values |
| `host_service_list` | `dbxcli host service-list` | None | Systemd units, status |
| `host_service_status` | `dbxcli host service-status` | None | Single service detail |
| `host_package_list` | `dbxcli host package-list` | None | Installed packages |
| `host_package_updates` | `dbxcli host package-updates` | None | Available security/bug updates |
| `host_user_list` | `dbxcli host user-list` | None | OS users, groups |
| `host_user_sessions` | `dbxcli host user-sessions` | None | Active SSH/console sessions |
| `host_uptime` | `dbxcli host uptime` | None | Boot time, uptime, reboot history |
| `host_load_history` | `dbxcli host load-history` | None | Load average over time |
| `host_ntp_status` | `dbxcli host ntp-status` | None | Time sync status |
| `host_dns_config` | `dbxcli host dns-config` | None | DNS resolver configuration |

---

## Host/OS Enterprise (Host-Pro, 40 tools)

Transport: SSH.

### CIS/STIG Hardening (8 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `host_cis_l1_scan` | `dbxcli host cis l1-scan` | None | CIS Level 1 benchmark scan |
| `host_cis_l2_scan` | `dbxcli host cis l2-scan` | None | CIS Level 2 benchmark scan |
| `host_stig_scan` | `dbxcli host stig scan` | None | DISA STIG compliance scan |
| `host_stig_report` | `dbxcli host stig report` | None | STIG compliance report |
| `host_harden_plan` | `dbxcli host harden plan` | None | Generate hardening plan |
| `host_harden_apply` | `dbxcli host harden apply` | Standard+Echo | Apply hardening changes |
| `host_harden_rollback` | `dbxcli host harden rollback` | Standard+Echo | Rollback hardening |
| `host_harden_status` | `dbxcli host harden status` | None | Hardening status |

### Patch Compliance (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `host_patch_scan` | `dbxcli host patch scan` | None | Scan for missing patches |
| `host_patch_cve` | `dbxcli host patch cve` | None | CVE correlation |
| `host_patch_plan` | `dbxcli host patch plan` | None | Create patch plan |
| `host_patch_apply` | `dbxcli host patch apply` | Standard+Echo | Apply patches |
| `host_patch_rollback` | `dbxcli host patch rollback` | Standard+Echo | Rollback patches |
| `host_patch_status` | `dbxcli host patch status` | None | Patch compliance status |

### Policy Scans (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `host_policy_scan` | `dbxcli host policy scan` | None | Policy compliance scan |
| `host_policy_report` | `dbxcli host policy report` | None | Policy report |
| `host_policy_drift` | `dbxcli host policy drift` | None | Configuration drift |
| `host_policy_fleet_scan` | `dbxcli host policy fleet-scan` | None | Fleet-wide scan |
| `host_policy_fleet_report` | `dbxcli host policy fleet-report` | None | Fleet compliance report |

### Security (6 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `host_security_firewall_audit` | `dbxcli host security firewall-audit` | None | Firewall rule audit |
| `host_security_selinux_audit` | `dbxcli host security selinux-audit` | None | SELinux audit |
| `host_security_pam_audit` | `dbxcli host security pam-audit` | None | PAM configuration audit |
| `host_security_sudoers_audit` | `dbxcli host security sudoers-audit` | None | Sudoers audit |
| `host_security_key_rotation` | `dbxcli host security key-rotation` | Standard | SSH key rotation |
| `host_security_ssh_audit` | `dbxcli host security ssh-audit` | None | SSH configuration audit |

### Capacity (4 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `host_capacity_forecast` | `dbxcli host capacity forecast` | None | Growth forecast |
| `host_capacity_sizing` | `dbxcli host capacity sizing` | None | Resource sizing advisor |
| `host_capacity_alerts` | `dbxcli host capacity alerts` | None | Resource threshold alerts |
| `host_capacity_trends` | `dbxcli host capacity trends` | None | Historical trend analysis |

### Log Analysis (5 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `host_log_journal` | `dbxcli host log journal` | None | journalctl analysis |
| `host_log_auth` | `dbxcli host log auth` | None | Auth log analysis |
| `host_log_syslog` | `dbxcli host log syslog` | None | Syslog analysis |
| `host_log_pattern` | `dbxcli host log pattern` | None | Pattern detection |
| `host_log_anomaly` | `dbxcli host log anomaly` | None | Anomaly detection |

### User Audit (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `host_user_privilege_audit` | `dbxcli host user privilege-audit` | None | Privilege audit |
| `host_user_dormant` | `dbxcli host user dormant` | None | Dormant account detection |
| `host_user_key_schedule` | `dbxcli host user key-schedule` | None | Key rotation schedule |

### Advanced (3 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `host_advanced_numa` | `dbxcli host advanced numa` | None | NUMA topology |
| `host_advanced_irq` | `dbxcli host advanced irq` | None | IRQ balance analysis |
| `host_advanced_custom` | `dbxcli host advanced custom` | None | Custom collector output |

---

## Monitoring Agent (OSS, 16 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `dbmon_agent_deploy` | `dbxcli dbmon agent deploy` | Standard | Deploy agent to host |
| `dbmon_agent_configure` | `dbxcli dbmon agent configure` | Standard | Configure agent |
| `dbmon_agent_start` | `dbxcli dbmon agent start` | Standard | Start agent |
| `dbmon_agent_stop` | `dbxcli dbmon agent stop` | Standard | Stop agent |
| `dbmon_agent_status` | `dbxcli dbmon agent status` | None | Agent status |
| `dbmon_agent_restart` | `dbxcli dbmon agent restart` | Standard | Restart agent |
| `dbmon_agent_upgrade` | `dbxcli dbmon agent upgrade` | Standard+Echo | Upgrade agent binary |
| `dbmon_agent_logs` | `dbxcli dbmon agent logs` | None | View agent logs |
| `dbmon_agent_config_reload` | `dbxcli dbmon agent config-reload` | Standard | Hot-reload config |
| `dbmon_agent_metric_query` | `dbxcli dbmon agent metric-query` | None | Query collected metrics |
| `dbmon_agent_metric_list` | `dbxcli dbmon agent metric-list` | None | List available metrics |
| `dbmon_agent_collector_list` | `dbxcli dbmon agent collector-list` | None | List collectors |
| `dbmon_agent_collector_enable` | `dbxcli dbmon agent collector-enable` | Standard | Enable collector |
| `dbmon_agent_collector_disable` | `dbxcli dbmon agent collector-disable` | Standard | Disable collector |
| `dbmon_agent_health` | `dbxcli dbmon agent health` | None | Agent health check |
| `dbmon_agent_version` | `dbxcli dbmon agent version` | None | Agent version |

---

## Monitoring Central (Core, 35 tools)

| Tool | CLI | Confirm | Description |
|------|-----|---------|-------------|
| `dbmon_fleet_list` | `dbxcli dbmon fleet list` | None | List managed fleet |
| `dbmon_fleet_status` | `dbxcli dbmon fleet status` | None | Fleet-wide status |
| `dbmon_fleet_health` | `dbxcli dbmon fleet health` | None | Fleet health summary |
| `dbmon_agent_list` | `dbxcli dbmon agent-list` | None | List registered agents |
| `dbmon_agent_register` | `dbxcli dbmon agent register` | Standard | Register agent |
| `dbmon_agent_deregister` | `dbxcli dbmon agent deregister` | Standard+Echo | Deregister agent |
| `dbmon_agent_status2` | `dbxcli dbmon agent-status` | None | Agent connection status |
| `dbmon_config_push` | `dbxcli dbmon config push` | Standard | Push config to agents |
| `dbmon_config_pull` | `dbxcli dbmon config pull` | None | Pull config from agent |
| `dbmon_config_diff` | `dbxcli dbmon config diff` | None | Config diff (expected vs actual) |
| `dbmon_drift_scan` | `dbxcli dbmon drift scan` | None | Configuration drift scan |
| `dbmon_drift_report` | `dbxcli dbmon drift report` | None | Drift report |
| `dbmon_drift_baseline` | `dbxcli dbmon drift baseline` | Standard | Create drift baseline |
| `dbmon_compliance_scan` | `dbxcli dbmon compliance scan` | None | Compliance scan |
| `dbmon_compliance_report` | `dbxcli dbmon compliance report` | None | Compliance report |
| `dbmon_alert_list` | `dbxcli dbmon alert list` | None | List active alerts |
| `dbmon_alert_ack` | `dbxcli dbmon alert ack` | Standard | Acknowledge alert |
| `dbmon_alert_resolve` | `dbxcli dbmon alert resolve` | Standard | Resolve alert |
| `dbmon_alert_rule_list` | `dbxcli dbmon alert rule-list` | None | List alert rules |
| `dbmon_alert_rule_create` | `dbxcli dbmon alert rule-create` | Standard | Create alert rule |
| `dbmon_alert_rule_delete` | `dbxcli dbmon alert rule-delete` | Standard+Echo | Delete alert rule |
| `dbmon_job_list` | `dbxcli dbmon job list` | None | List scheduled jobs |
| `dbmon_job_create` | `dbxcli dbmon job create` | Standard | Create job |
| `dbmon_job_status` | `dbxcli dbmon job status` | None | Job execution status |
| `dbmon_job_cancel` | `dbxcli dbmon job cancel` | Standard | Cancel job |
| `dbmon_job_history` | `dbxcli dbmon job history` | None | Job execution history |
| `dbmon_report_generate` | `dbxcli dbmon report generate` | None | Generate fleet report |
| `dbmon_report_schedule` | `dbxcli dbmon report schedule` | Standard | Schedule periodic report |
| `dbmon_report_list` | `dbxcli dbmon report list` | None | List reports |
| `dbmon_dashboard_list` | `dbxcli dbmon dashboard list` | None | List dashboards |
| `dbmon_dashboard_data` | `dbxcli dbmon dashboard data` | None | Dashboard data query |
| `dbmon_upgrade_plan` | `dbxcli dbmon upgrade plan` | None | Fleet upgrade plan |
| `dbmon_upgrade_execute` | `dbxcli dbmon upgrade execute` | Standard+Echo | Execute fleet upgrade |
| `dbmon_upgrade_status` | `dbxcli dbmon upgrade status` | None | Upgrade status |
| `dbmon_upgrade_rollback` | `dbxcli dbmon upgrade rollback` | Standard+Echo | Rollback upgrade |

---

## Policy Engine (8 tools)

| Tool | CLI | Confirm | License | Description |
|------|-----|---------|---------|-------------|
| `policy_scan` | `dbxcli policy scan` | None | OSS (OS) / Core (Oracle) / PG-Pro | Scan target against policy framework |
| `policy_report` | `dbxcli policy report` | None | OSS / Core / PG-Pro | Generate compliance report (JSON/HTML/CSV) |
| `policy_drift` | `dbxcli policy drift` | None | Core / PG-Pro | Compare current state vs baseline |
| `policy_status` | `dbxcli policy status` | None | OSS | Show loaded policies and versions |
| `policy_reload` | `dbxcli policy reload` | None | OSS | Reload policy files from disk |
| `policy_fleet_scan` | `dbxcli policy fleet-scan` | None | Core / PG-Pro | Scan all targets in fleet group |
| `policy_fleet_report` | `dbxcli policy fleet-report` | None | Core / PG-Pro | Aggregate compliance across fleet |
| `policy_remediate` | `dbxcli policy remediate` | Standard+Echo | Core / PG-Pro | Apply remediation for failing rules |

---

## RAG --- Retrieval-Augmented Generation (5 tools)

| Tool | CLI | Confirm | License | Description |
|------|-----|---------|---------|-------------|
| `rag_search` | `dbxcli rag search` | None | OSS (basic) / Core (full) | Semantic search across indexed documents |
| `rag_context` | `dbxcli rag context` | None | Core | Build token-limited LLM context |
| `rag_index_status` | `dbxcli rag index-status` | None | OSS | Show index statistics |
| `rag_index_refresh` | `dbxcli rag index-refresh` | Standard | Core | Re-index documents from sources |
| `rag_sources` | `dbxcli rag sources` | None | OSS | List configured document sources |

---

## Verification

| Category | Count |
|----------|-------|
| Oracle Core Read | 28 |
| Oracle Core Mutate | 38 |
| Oracle Linux | 20 |
| Performance | 32 |
| Unified Audit | 22 |
| Partitioning | 26 |
| Data Guard | 28 |
| Backup (RMAN) | 34 |
| RAC | 24 |
| Clusterware | 22 |
| ASM | 20 |
| Provisioning | 22 |
| Patching | 30 |
| Migration | 20 |
| Data Pump | 18 |
| GoldenGate | 24 |
| OEM | 28 |
| PostgreSQL Core | 27 |
| PostgreSQL Enterprise | 111 |
| Host Core | 20 |
| Host Enterprise | 40 |
| Monitoring Agent | 16 |
| Monitoring Central | 35 |
| Policy Engine | 8 |
| RAG | 5 |
| **TOTAL** | **735** |
