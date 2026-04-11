# Oracle Engine Installation Guide

**Audience:** Oracle DBA  
**Estimated setup time:** ~15 minutes  
**Applies to:** dbx Oracle Engine (OSS + licensed bundles)

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Add Oracle Target](#add-oracle-target)
3. [Configure Vault Credentials](#configure-vault-credentials)
4. [Declare Oracle License](#declare-oracle-license)
5. [Configure SSH for OS Tools](#configure-ssh-for-os-tools)
6. [MCP Adapter Setup](#mcp-adapter-setup)
7. [Verify Connection](#verify-connection)
8. [Tools Unlocked by Tier](#tools-unlocked-by-tier)
9. [Data Guard Target Setup](#data-guard-target-setup)
10. [RAC Target Setup](#rac-target-setup)
11. [OEM Target Setup](#oem-target-setup)
12. [GoldenGate Target Setup](#goldengate-target-setup)

---

## Prerequisites

Before adding an Oracle target, ensure the following are in place on the machine running `dbxcli`:

| Requirement | Minimum Version | Notes |
|---|---|---|
| Oracle Instant Client | 19c | Required for SQL*Net connectivity |
| `dbxcli` binary | Latest | See quick-start guide for download |
| TCP access to Oracle listener | Port 1521 (or custom) | Confirm with `nc -zv db.example.com 1521` |
| SSH key pair | Ed25519 recommended | Required for OS-level tools (ASM, alert log, host metrics) |
| Node.js | 18+ | Required for MCP adapter |

Verify Instant Client is discoverable:

```bash
# On Linux/macOS
export LD_LIBRARY_PATH=/opt/oracle/instantclient_19_22:$LD_LIBRARY_PATH

# Confirm sqlplus resolves (optional smoke test)
sqlplus -V
```

---

## Add Oracle Target

Use `dbxcli target add` to register the Oracle database. Parameters use `key=value` syntax.

```bash
dbxcli target add \
  entity_name=prod-oracle-01 \
  entity_type=oracle_database \
  host=db.example.com \
  port=1521 \
  service=ORCLPDB1
```

**Parameters:**

| Parameter | Required | Description |
|---|---|---|
| `entity_name` | Yes | Unique identifier for this target within dbx |
| `entity_type` | Yes | Must be `oracle_database` |
| `host` | Yes | Hostname or IP of the Oracle listener |
| `port` | Yes | Listener port (default: 1521) |
| `service` | Yes | Oracle service name (PDB or CDB service) |

To list registered targets:

```bash
dbxcli target list
```

---

## Configure Vault Credentials

Storing database credentials in HashiCorp Vault is strongly recommended over inline passwords. This ensures credentials are never persisted in dbx configuration files.

### Step 1: Configure the Vault backend

```bash
dbxcli vault configure \
  vault_addr=https://vault.example.com:8200 \
  vault_mount=secret \
  vault_auth_method=approle \
  role_id=<role-id> \
  secret_id=<secret-id>
```

### Step 2: Seed the Oracle credential in Vault

The Vault path must contain `username` and `password` keys:

```bash
vault kv put secret/dbx/prod-oracle-01 \
  username=dbx_monitor \
  password=<redacted>
```

### Step 3: Link the target to Vault

```bash
dbxcli target set prod-oracle-01 \
  credential=vault \
  vault_path=secret/dbx/prod-oracle-01
```

For environments without Vault, inline credentials are supported but not recommended:

```bash
dbxcli target set prod-oracle-01 \
  credential=inline \
  username=dbx_monitor \
  password=<your-password>
```

---

## Declare Oracle License

dbx enforces Oracle license declarations to prevent accidental use of licensed features. Set the edition and options that apply to this database installation.

```bash
dbxcli target set prod-oracle-01 \
  oracle_edition=enterprise \
  oracle_options=partitioning,advanced_compression,diagnostics_pack,tuning_pack,multitenant \
  license_mode=strict
```

**`--oracle-edition` values:**

| Value | Description |
|---|---|
| `standard_edition_2` | Oracle SE2 — limited feature set |
| `enterprise` | Oracle EE — full feature set |
| `enterprise_rac` | Oracle EE with Real Application Clusters |
| `personal` | Oracle PE — single user, development only |

**Common `--oracle-options` values:**

| Option Key | Description |
|---|---|
| `partitioning` | Partitioning option |
| `advanced_compression` | Advanced Compression option |
| `diagnostics_pack` | AWR, ADDM, ASH (included in Diagnostics Pack) |
| `tuning_pack` | SQL Tuning Advisor, SQL Access Advisor |
| `multitenant` | Full multitenant CDB/PDB management |
| `rac` | Real Application Clusters |
| `data_guard` | Active Data Guard (physical standby reads) |
| `goldengate` | Oracle GoldenGate integration |
| `label_security` | Oracle Label Security |
| `database_vault` | Oracle Database Vault |
| `spatial` | Oracle Spatial and Graph |
| `text` | Oracle Text |

**`--license-mode` values:**

| Value | Behavior |
|---|---|
| `strict` | Block any tool that requires an undeclared option |
| `warn` | Allow but log a warning for undeclared option usage |
| `permissive` | No enforcement — development/lab use only |

---

## Configure SSH for OS Tools

OS-level tools (alert log reading, ASM disk group inspection, host metrics, OPatch inventory) require SSH access to the database host.

```bash
dbxcli target set prod-oracle-01 \
  ssh_host=db.example.com \
  ssh_user=oracle \
  ssh_key=~/.ssh/id_ed25519_dbx \
  ssh_port=22
```

**Parameters:**

| Parameter | Description |
|---|---|
| `ssh_host` | SSH hostname (may differ from `host` if using a jump host) |
| `ssh_user` | OS user with read access to Oracle directories (typically `oracle`) |
| `ssh_key` | Path to private key — must correspond to an authorized key on the target |
| `ssh_port` | SSH port (default: 22) |

Verify SSH connectivity:

```bash
dbxcli target ssh-test prod-oracle-01
```

Expected output:

```
SSH connection to oracle@db.example.com:22 — OK
Oracle home detected: /u01/app/oracle/product/19.0.0/dbhome_1
```

---

## MCP Adapter Setup

The MCP adapter exposes dbx tools to Claude Code and Claude Desktop via the Model Context Protocol. Install the appropriate npm bundle for your licensed tier.

### Install

```bash
# Free (OSS) — Oracle tools only
npm install -g @itunified.io/mcp-oracle

# Core Bundle
npm install -g @itunified.io/mcp-oracle-core

# HA Bundle
npm install -g @itunified.io/mcp-oracle-ha

# Ops Bundle
npm install -g @itunified.io/mcp-oracle-ops

# Full Platform
npm install -g @itunified.io/mcp-oracle-full
```

### Claude Code Configuration

Add to `~/.claude/settings.json` (or project-level `.claude/settings.json`):

```json
{
  "mcpServers": {
    "dbx-oracle": {
      "command": "npx",
      "args": ["@itunified.io/mcp-oracle", "--config", "~/.dbx/oracle/prod-oracle-01.yaml"]
    }
  }
}
```

### Claude Desktop Configuration

Add to the Claude Desktop MCP config (typically `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "dbx-oracle": {
      "command": "npx",
      "args": ["@itunified.io/mcp-oracle", "--config", "/home/dba/.dbx/oracle/prod-oracle-01.yaml"]
    }
  }
}
```

---

## Verify Connection

After registration, run the following verification commands:

```bash
# List active sessions
dbxcli oracle session list entity=prod-oracle-01

# List tablespaces and free space
dbxcli oracle tablespace list entity=prod-oracle-01

# Show host and OS info
dbxcli oracle host info entity=prod-oracle-01
```

Expected output for `tablespace list`:

```
TABLESPACE_NAME   STATUS   SIZE_GB   USED_GB   FREE_GB   PCT_USED
SYSTEM            ONLINE   1.00      0.83      0.17      83%
SYSAUX            ONLINE   2.00      1.41      0.59      71%
UNDOTBS1          ONLINE   4.00      0.92      3.08      23%
TEMP              ONLINE   2.00      0.00      2.00       0%
USERS             ONLINE   50.00     18.23     31.77      36%
```

If connection fails:

```bash
dbxcli target diagnose prod-oracle-01
```

---

## Tools Unlocked by Tier

All tools available through the Oracle engine, organized by licensing tier.

| Tier | Total Tools | Skills | Key Capabilities |
|---|---|---|---|
| Free (OSS) | 48 | 0 | Session, Tablespace, User, Parameter, Schema, Redo, Undo, Alert, DataDict, SysInfo, OracleLinux |
| Core Bundle | 144 (+96) | 19 | All Free + AWR, ASH, SQL Tuning, Partitioning, RMAN basics, Multitenant, Scheduler |
| HA Bundle | 272 (+128) | 37 | All Core + Data Guard, RAC, ASM, GoldenGate, Flashback, PITR, Failover automation |
| Ops Bundle | 434 (+162) | 59 | All HA + OEM integration, Patch management, Database Vault, Audit Vault, Label Security, OPatch |
| Full Platform | 531 | 71 | All Ops + advanced analytics, compliance reporting, capacity forecasting, AI-assisted tuning |

**Free tier tool breakdown (48 tools):**

| Category | Count | Tools |
|---|---|---|
| Session | 3 | `session_list`, `session_kill`, `session_detail` |
| Tablespace | 3 | `tablespace_list`, `tablespace_detail`, `tablespace_free` |
| User | 3 | `user_list`, `user_detail`, `user_privs` |
| Parameter | 2 | `parameter_list`, `parameter_get` |
| Schema | 4 | `schema_list`, `object_list`, `index_list`, `constraint_list` |
| Redo | 2 | `redo_log_list`, `redo_switch_history` |
| Undo | 2 | `undo_advisor`, `undo_stat` |
| Alert Log | 2 | `alert_log_tail`, `alert_log_search` |
| Data Dictionary | 3 | `datadict_query`, `datadict_table_info`, `datadict_column_info` |
| SysInfo | 4 | `db_version`, `nls_settings`, `instance_info`, `feature_usage` |
| OracleLinux | 20 | CPU, memory, disk, network, process, OS version, kernel, patches, firewall, users, and more |

---

## Data Guard Target Setup

To enable Data Guard management tools, register both the primary and standby databases as targets and declare the Data Guard relationship.

### Register the standby

```bash
dbxcli target add \
  entity_name=prod-oracle-01-standby \
  entity_type=oracle_database \
  host=dr.example.com \
  port=1521 \
  service=ORCLPDB1_STBY
```

### Link primary to standby

```bash
dbxcli target set prod-oracle-01 \
  dg_role=primary \
  dg_standby=prod-oracle-01-standby

dbxcli target set prod-oracle-01-standby \
  dg_role=physical_standby \
  dg_primary=prod-oracle-01
```

### Enable Data Guard Broker (if in use)

```bash
dbxcli target set prod-oracle-01 \
  dg_broker=true \
  dg_broker_config=DGConfig1
```

Verify Data Guard status:

```bash
dbxcli oracle dg status entity=prod-oracle-01
```

---

## RAC Target Setup

For Oracle Real Application Clusters, register the cluster as a single logical target with multiple instance declarations.

### Register the RAC cluster

```bash
dbxcli target add \
  entity_name=prod-oracle-rac \
  entity_type=oracle_rac \
  scan_host=prod-scan.example.com \
  scan_port=1521 \
  service=ORCLPDB1 \
  asm_diskgroup=DATA,FRA
```

### Register individual instances

```bash
dbxcli target set prod-oracle-rac \
  rac_instances=prod-db01:10.0.1.11,prod-db02:10.0.1.12

dbxcli target set prod-oracle-rac \
  asm_nodes=prod-db01:10.0.1.11,prod-db02:10.0.1.12
```

Verify RAC status:

```bash
dbxcli oracle rac status entity=prod-oracle-rac
dbxcli oracle asm diskgroup list entity=prod-oracle-rac
```

---

## OEM Target Setup

Integration with Oracle Enterprise Manager Cloud Control allows dbx to pull performance data and job status via the OEM REST API.

```bash
dbxcli target add \
  entity_name=prod-oem \
  entity_type=oracle_oem \
  host=oem.example.com \
  port=7803 \
  oem_target_name=prod_oracle_db \
  oem_target_type=oracle_database
```

Configure OEM credentials in Vault:

```bash
vault kv put secret/dbx/prod-oem \
  username=oem_api_user \
  password=<redacted>

dbxcli target set prod-oem \
  credential=vault \
  vault_path=secret/dbx/prod-oem
```

---

## GoldenGate Target Setup

Register a GoldenGate Microservices deployment to manage extracts, replicats, and trails through dbx.

```bash
dbxcli target add \
  entity_name=prod-gg \
  entity_type=oracle_goldengate \
  host=gg.example.com \
  port=9011 \
  gg_deployment=MyDeployment \
  gg_service_manager_port=9010
```

Configure GoldenGate credentials:

```bash
vault kv put secret/dbx/prod-gg \
  username=ggadmin \
  password=<redacted>

dbxcli target set prod-gg \
  credential=vault \
  vault_path=secret/dbx/prod-gg
```

Link the source Oracle database to this GoldenGate deployment:

```bash
dbxcli target set prod-oracle-01 \
  goldengate_target=prod-gg
```

Verify GoldenGate status:

```bash
dbxcli oracle gg extract list entity=prod-gg
dbxcli oracle gg replicat list entity=prod-gg
```
