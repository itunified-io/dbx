# dbx Administration Guide

This guide covers all operational aspects of managing a dbx deployment: target registration, Vault integration, Oracle license gate configuration, dbx license management, confirm gates, audit trail administration, RBAC, output formatting, and upgrade procedures.

**Audience:** IT administrators, DBAs, and DevOps engineers responsible for deploying and operating dbx in production environments.

---

## Table of Contents

1. [Target Management](#1-target-management)
2. [Vault Integration](#2-vault-integration)
3. [Oracle License Gate Administration](#3-oracle-license-gate-administration)
4. [dbx License Management](#4-dbx-license-management)
5. [Confirm Gates](#5-confirm-gates)
6. [Audit Trail Administration](#6-audit-trail-administration)
7. [User and RBAC Administration (EE)](#7-user-and-rbac-administration-ee)
8. [Output Formatting](#8-output-formatting)
9. [Upgrade and Migration](#9-upgrade-and-migration)

---

## 1. Target Management

Targets are the central concept in dbx. Every database, host, listener, ASM instance, or cluster that dbx manages must be registered as a named target. The target registry is stored on disk as YAML files at `~/.dbx/targets/`.

### 1.1 Entity Types

| Entity Type | Description |
|-------------|-------------|
| `oracle_database` | Single-instance Oracle database |
| `rac_database` | Oracle Real Application Clusters database |
| `oracle_listener` | Oracle TNS listener |
| `oracle_asm` | Oracle Automatic Storage Management instance |
| `oracle_host` | Host running Oracle software (SSH-based management) |
| `exadata` | Oracle Exadata Database Machine |
| `oda` | Oracle Database Appliance |
| `zdlra` | Zero Data Loss Recovery Appliance |
| `pg_database` | Single PostgreSQL database |
| `pg_cluster` | PostgreSQL cluster (CloudNativePG or Patroni) |
| `host` | Generic Linux/Unix host |

### 1.2 Adding Targets

Targets are registered with `dbxcli target add` using named `key=value` parameters.

**Oracle single-instance database:**

```bash
dbxcli target add \
  entity_name=prod-orcl \
  entity_type=oracle_database \
  host=db01.example.com \
  port=1521 \
  service=ORCL \
  user=dba_user \
  --ask-password
```

**Oracle RAC database:**

```bash
dbxcli target add \
  entity_name=prod-rac \
  entity_type=rac_database \
  host=rac-scan.example.com \
  port=1521 \
  service=RACDB \
  user=dba_user \
  --ask-password
```

**Oracle host (SSH-managed):**

```bash
dbxcli target add \
  entity_name=db-host-01 \
  entity_type=oracle_host \
  host=db01.example.com \
  ssh_user=oracle \
  ssh_key_path=~/.ssh/id_ed25519
```

**PostgreSQL database:**

```bash
dbxcli target add \
  entity_name=prod-pg \
  entity_type=pg_database \
  host=pg01.example.com \
  port=5432 \
  database=appdb \
  user=dba_user \
  sslmode=require \
  --ask-password
```

**Generic host:**

```bash
dbxcli target add \
  entity_name=app-server-01 \
  entity_type=host \
  host=app01.example.com \
  ssh_user=admin \
  ssh_key_path=~/.ssh/id_ed25519
```

### 1.3 Listing and Testing Targets

List all registered targets:

```bash
dbxcli target list
```

Filter by entity type:

```bash
dbxcli target list entity_type=oracle_database
```

Test connectivity to a specific target:

```bash
dbxcli target test entity_name=prod-orcl
```

Expected output:

```
Target:  prod-orcl
Type:    oracle_database
Host:    db01.example.com:1521/ORCL
Status:  connected
Version: Oracle Database 19c Enterprise Edition Release 19.0.0.0.0
Latency: 4ms
```

### 1.4 Removing and Editing Targets

Remove a target from the registry:

```bash
dbxcli target remove entity_name=prod-orcl
```

Edit a target by modifying its YAML file directly:

```bash
$EDITOR ~/.dbx/targets/prod-orcl.yaml
```

After editing, validate the file parses cleanly:

```bash
dbxcli target test entity_name=prod-orcl
```

### 1.5 Target YAML File Format

Each target is stored as a YAML file at `~/.dbx/targets/<entity-name>.yaml`. The file name must match the `name` field inside the file.

**Minimal Oracle target:**

```yaml
name: prod-orcl
type: oracle_database
description: "Production Oracle 19c — primary datacenter"

primary:
  host: db01.example.com
  port: 1521
  service: ORCL
  credential: prod-orcl-dba
```

**Full Oracle target with all endpoints:**

```yaml
name: prod-orcl
type: oracle_database
description: "Production Oracle 19c with Data Guard standby"

oracle_license:
  edition: enterprise
  options:
    - partitioning
    - advanced_security
    - diagnostics_pack
    - tuning_pack
  oem_packs:
    - diagnostics
    - tuning

primary:
  host: db01.example.com
  port: 1521
  service: ORCL
  vault_path: secret/data/oracle/prod-orcl

standby:
  host: db02.example.com
  port: 1521
  service: ORCLSTBY
  role: standby
  vault_path: secret/data/oracle/prod-orcl-standby

asm:
  host: db01.example.com
  port: 1521
  service: +ASM
  vault_path: secret/data/oracle/prod-orcl-asm

ssh:
  host: db01.example.com
  user: oracle
  vault_path: secret/data/oracle/prod-orcl-ssh

monitoring:
  agent_port: 9161

goldengate:
  rest_url: http://gg01.example.com:9011
  vault_path: secret/data/oracle/prod-orcl-gg

oem:
  rest_url: https://oem.example.com:7803/em
  vault_path: secret/data/oracle/prod-orcl-oem
```

**Full PostgreSQL target with CNPG and DR:**

```yaml
name: prod-pg
type: pg_cluster
description: "Production PostgreSQL 16 — CNPG cluster"

primary:
  host: pg01.example.com
  port: 5432
  database: appdb
  sslmode: require
  vault_path: secret/data/postgres/prod-pg

replica:
  host: pg02.example.com
  port: 5432
  database: appdb
  sslmode: require
  role: replica
  vault_path: secret/data/postgres/prod-pg-replica

pgbouncer:
  host: pgbouncer.example.com
  port: 6432

cnpg:
  cluster_name: prod-pg-cluster
  namespace: databases
  k8s_context: production

dr:
  remote_cluster: dr-pg-cluster
  remote_context: dr-production
  wal_archive: s3://pg-wal-archive/prod-pg

ssh:
  host: pg01.example.com
  user: postgres
  vault_path: secret/data/postgres/prod-pg-ssh

monitoring:
  agent_port: 9187
```

### 1.6 Target Groups and Fleet Organization

Targets can be logically grouped using a group configuration file at `~/.dbx/groups.yaml`. Groups are used for fleet-wide operations, reporting, and RBAC scoping.

```yaml
groups:
  production:
    - prod-orcl
    - prod-rac
    - prod-pg
    description: "All production databases"

  oracle-fleet:
    - prod-orcl
    - prod-rac
    - dr-orcl
    description: "All Oracle databases across all environments"

  pg-fleet:
    - prod-pg
    - dr-pg
    description: "All PostgreSQL databases"

  dr-tier:
    - dr-orcl
    - dr-pg
    description: "Disaster recovery targets"
```

Run an operation against all targets in a group:

```bash
dbxcli db session list group=production
dbxcli target test group=oracle-fleet
```

### 1.7 Multi-Endpoint Target Wiring

For complex Oracle environments, the endpoint fields map to specific management roles:

| Field | Purpose | Used By |
|-------|---------|---------|
| `primary` | Primary database connection | All `db` operations |
| `standby` | Data Guard standby connection | `dataguard` tools |
| `replica` | PostgreSQL read replica | PG read-only queries |
| `asm` | ASM instance connection | `asm` tools |
| `ssh` | OS-level access | `linux` and `oracle_host` tools |
| `monitoring` | Monitoring agent port | Metrics collection |
| `goldengate` | GoldenGate REST API | `goldengate` tools |
| `oem` | Oracle Enterprise Manager REST API | `oem` tools |
| `pgbouncer` | PgBouncer pooler endpoint | PG connection management |
| `cnpg` | CloudNativePG cluster metadata | CNPG tools |
| `dr` | Cross-cluster DR configuration | PG DR tools |

If a tool requires an endpoint that is not configured on the target, `dbxcli` reports a clear error rather than falling back silently.

---

## 2. Vault Integration

dbx integrates with HashiCorp Vault for all credential storage in production environments. Vault replaces the local credential store for targets that have `vault_path` configured.

### 2.1 AppRole Setup for dbx

Create a Vault policy granting dbx read access to database credentials:

```hcl
# ~/.dbx/vault-policy.hcl
path "secret/data/oracle/*" {
  capabilities = ["read"]
}

path "secret/data/postgres/*" {
  capabilities = ["read"]
}

path "secret/metadata/oracle/*" {
  capabilities = ["read", "list"]
}

path "secret/metadata/postgres/*" {
  capabilities = ["read", "list"]
}
```

Apply the policy:

```bash
vault policy write dbx-reader ~/.dbx/vault-policy.hcl
```

Create an AppRole for dbx:

```bash
vault auth enable approle
vault write auth/approle/role/dbx \
  token_policies="dbx-reader" \
  token_ttl=1h \
  token_max_ttl=4h \
  secret_id_ttl=720h
```

Retrieve the role ID and a secret ID:

```bash
vault read auth/approle/role/dbx/role-id
vault write -f auth/approle/role/dbx/secret-id
```

Configure dbx to use AppRole in `~/.dbx/config.yaml`:

```yaml
vault:
  address: https://vault.example.com:8200
  auth_method: approle
  role_id_env: VAULT_ROLE_ID
  secret_id_env: VAULT_SECRET_ID
```

Set the environment variables before running dbxcli:

```bash
export VAULT_ROLE_ID=<your-role-id>
export VAULT_SECRET_ID=<your-secret-id>
```

For production deployments, source these from a secrets manager or the system keychain rather than exporting them directly in shell profiles.

### 2.2 Per-Target Credential Paths

dbx follows a consistent Vault path convention. The `vault_path` field in a target YAML file specifies the exact KV v2 path for that endpoint's credentials.

**Oracle database credentials** (stored at `secret/data/oracle/<target-name>`):

```bash
vault kv put secret/oracle/prod-orcl \
  username=dba_user \
  password=<password>
```

Expected secret structure:

```json
{
  "username": "dba_user",
  "password": "..."
}
```

**PostgreSQL database credentials** (stored at `secret/data/postgres/<target-name>`):

```bash
vault kv put secret/postgres/prod-pg \
  username=dba_user \
  password=<password>
```

**SSH key credentials:**

```bash
vault kv put secret/oracle/prod-orcl-ssh \
  username=oracle \
  private_key="$(cat ~/.ssh/oracle_ed25519)"
```

When a target endpoint has `vault_path` set, dbx fetches credentials from Vault at connect time. The credentials are held in memory only and are never written to disk.

### 2.3 SSH Key Storage in Vault

SSH private keys for Oracle hosts and Linux hosts are stored in Vault alongside username information:

```bash
vault kv put secret/oracle/prod-orcl-ssh \
  username=oracle \
  private_key="$(cat /path/to/oracle_ed25519)" \
  passphrase=""
```

The `private_key` value must be the full PEM-encoded private key. dbx writes the key to a temporary file in memory (via `os.CreateTemp`) for the duration of the SSH session and removes it immediately after.

### 2.4 Dynamic Secrets and Lease Management

When using Vault database secrets engines (e.g., `database/creds/prod-orcl`), set `vault_path` to the dynamic credential path:

```yaml
primary:
  host: db01.example.com
  port: 1521
  service: ORCL
  vault_path: database/creds/prod-orcl
```

dbx fetches a new credential lease at connect time. The lease TTL is determined by the Vault role configuration. dbx does not automatically renew leases; connections opened with dynamic credentials will break if the lease expires during a long-running operation.

For long-running jobs (data pump exports, schema migrations), use static credentials with Vault or set the lease TTL high enough to cover the expected operation duration.

### 2.5 Credential Rotation Workflows

**Rotating Oracle DBA password:**

1. Generate a new password outside dbx.
2. Update the Oracle user password:

   ```bash
   dbxcli db user alter entity_name=prod-orcl user=dba_user new_password=<new> --confirm
   ```

3. Update the Vault secret:

   ```bash
   vault kv put secret/oracle/prod-orcl username=dba_user password=<new>
   ```

4. Verify connectivity:

   ```bash
   dbxcli target test entity_name=prod-orcl
   ```

**Rotating PostgreSQL role password:**

1. Update the PostgreSQL role:

   ```bash
   dbxcli pg user alter entity_name=prod-pg user=dba_user new_password=<new> --confirm
   ```

2. Update the Vault secret:

   ```bash
   vault kv put secret/postgres/prod-pg username=dba_user password=<new>
   ```

3. Verify connectivity:

   ```bash
   dbxcli target test entity_name=prod-pg
   ```

### 2.6 Offline Fallback Behavior

If Vault is unreachable when dbx starts:

- Targets with `vault_path` configured fail to connect. dbx logs a warning:

  ```
  WARN  vault unreachable — target prod-orcl credential fetch failed: connection refused
  ```

- Targets using the local credential store (no `vault_path`) continue to operate normally.
- dbx does not cache Vault credentials to disk. There is no offline grace period for Vault-backed credentials.

For environments with intermittent Vault availability, configure Vault agent (`vault agent`) as a local proxy. Vault agent caches credentials and can serve them from disk during outages.

---

## 3. Oracle License Gate Administration

The Oracle License Gate enforces that dbx only executes operations that are covered by the Oracle licenses declared on each target. This prevents accidental use of Oracle options (Partitioning, Advanced Security, Diagnostics Pack, etc.) that are not licensed.

### 3.1 Edition Declaration

Oracle edition is declared per-target in the `oracle_license` block:

```yaml
oracle_license:
  edition: enterprise   # or: standard2
```

Tools that require Enterprise Edition (e.g., Data Guard, partitioned tables, advanced compression) will be blocked or warned if the target declares `standard2`.

### 3.2 Oracle Option Declarations

Declare each Oracle option that is licensed for the target:

```yaml
oracle_license:
  edition: enterprise
  options:
    - partitioning
    - advanced_security
    - diagnostics_pack
    - tuning_pack
    - olap
    - spatial
    - label_security
    - database_vault
    - advanced_compression
    - active_data_guard
    - multitenant
    - goldengate
  oem_packs:
    - diagnostics
    - tuning
    - change_management
    - configuration
```

If an option is not listed, any tool that requires that option will be subject to gate enforcement (block, warn, or audit-only, depending on the configured mode).

### 3.3 Enforcement Modes

The gate mode is configured globally in `~/.dbx/config.yaml`:

```yaml
oracle_gate_mode: strict
```

Or per-session via environment variable:

```bash
export DBX_ORACLE_GATE_MODE=warn
```

| Mode | Behavior |
|------|----------|
| `strict` | Block operations that require undeclared options. The command exits with an error. |
| `warn` | Allow the operation but emit a warning to stderr and log to the audit trail. |
| `audit-only` | Allow the operation silently but record the license gap in the audit trail. |

**Default:** `strict`

In `strict` mode, attempting to run a Diagnostics Pack tool on a target that has not declared `diagnostics_pack` produces:

```
ERROR: oracle_license gate blocked
  Target:   prod-orcl
  Requires: diagnostics_pack
  Declared: partitioning, advanced_security
  Mode:     strict

Pass DBX_ORACLE_GATE_MODE=warn to proceed with a warning.
```

### 3.4 Per-Target Override vs Fleet-Wide

The gate mode applies fleet-wide from `~/.dbx/config.yaml`. Per-target overrides are not supported in the current release. If different targets require different enforcement strictness, use environment variables to switch mode before running a command against a specific target.

### 3.5 Fleet Audit Reporting

Generate a report of all Oracle license gate events across the fleet:

```bash
dbxcli audit query tool_prefix=oracle_gate --format table
```

The report shows all events where the gate fired (blocked, warned, or logged), including the target name, operation, missing option, gate decision, and timestamp.

Export to JSON for integration with compliance tooling:

```bash
dbxcli audit query tool_prefix=oracle_gate --format json > oracle-license-audit.json
```

---

## 4. dbx License Management

dbx uses Ed25519-signed JWT license files to enforce feature tiers. License files are issued by itunified.io and are tied to a customer ID and optional machine fingerprint.

### 4.1 License Activation

Place the license file at a known path and activate it:

```bash
dbxcli license activate --file /path/to/license.jwt
```

This validates the Ed25519 signature, copies the file to `~/.dbx/license.jwt`, and unlocks the features included in the license tier.

On success:

```
License activated.
  Customer:  ACME Corp (cust-12345)
  Tier:      core
  Bundles:   core
  Targets:   up to 25
  Expires:   2027-04-11
```

On signature failure:

```
ERROR: invalid license signature — the file may be corrupt or tampered with.
Contact support@itunified.io to obtain a replacement license.
```

### 4.2 License Status

View the current license status:

```bash
dbxcli license status
```

Example output:

```
License Status
  File:            ~/.dbx/license.jwt
  Customer:        ACME Corp (cust-12345)
  License ID:      lic-2026-abc123
  Tier:            core
  Bundles:         core
  Max Targets:     25
  Active Targets:  12
  Expires:         2027-04-11
  Grace Period:    not active
  Phone-Home:      last verified 2026-04-10 08:00 UTC
  Status:          VALID
```

### 4.3 License Tiers

| Tier | Bundle | Included Domains | Default Max Targets |
|------|--------|-----------------|---------------------|
| OSS (free) | `free` | db-read, linux, monitor-agent | Unlimited |
| Core Bundle | `core` | + db-mutate, performance, audit, partitioning, monitor-central | 25 |
| HA Bundle | `ha` | + dataguard, backup, rac, clusterware, asm | 50 |
| Ops Bundle | `ops` | + provision, patch, migration, datapump, goldengate, oem | 100 |
| PG Professional | `pg_professional` | pg-enterprise (mcp-postgres-enterprise) | 50 |

Bundles are additive. A customer may hold multiple bundle licenses simultaneously (e.g., `core` + `ha`). Each bundle license is issued as a separate JWT file. Activate each file individually:

```bash
dbxcli license activate --file /path/to/core.jwt
dbxcli license activate --file /path/to/ha.jwt
```

### 4.4 Phone-Home Verification

When a network connection is available, dbx verifies the license with `license.itunified.io` once per day. The verification result is cached locally for 72 hours.

Verification is passive: it does not block operations and does not transmit any database contents. The verification payload contains only the license ID, customer ID, and current target count.

**Offline grace period:** If phone-home verification fails for any reason (network outage, firewall restriction), dbx continues to operate normally for up to 14 days using the cached verification result. After 14 days without successful verification, EE features are locked until the next successful phone-home.

To check phone-home status:

```bash
dbxcli license status
```

The `Phone-Home` line shows the timestamp of the last successful verification and the number of days remaining in the grace period if active.

### 4.5 Target Count Enforcement

dbx counts all targets in `~/.dbx/targets/` against the license's `max_targets` limit. When the target count reaches the limit:

- Existing targets continue to operate normally.
- Adding a new target with `dbxcli target add` fails with:

  ```
  ERROR: target count limit reached (25/25) for license tier "core"
  Remove unused targets or upgrade your license to continue.
  ```

Target count is enforced at registration time, not at query time.

### 4.6 License File Location

The active license file is stored at `~/.dbx/license.jwt`. When running dbx as a service or in a container, set the `DBX_DATA_DIR` environment variable to point to a shared directory:

```bash
export DBX_DATA_DIR=/opt/dbx/data
dbxcli license activate --file /opt/dbx/license.jwt
```

All dbxcli, dbxctl, and REST API server processes sharing the same `DBX_DATA_DIR` use the same license.

---

## 5. Confirm Gates

dbx uses a four-level confirmation system to prevent accidental execution of destructive operations. Every tool declares its confirm level; the gate enforces it before execution begins.

### 5.1 Confirm Levels

| Level | Name | CLI Behavior | MCP Behavior |
|-------|------|-------------|-------------|
| 0 | None | Executes immediately | Executes immediately |
| 1 | Standard | Requires `--confirm` flag | Requires `confirm=true` parameter |
| 2 | Standard+Echo | Requires `--confirm` and typing an exact echo string | Requires echo-back token in response |
| 3 | Double-Confirm | Requires `--confirm`, echo string, then second confirmation | Two-step echo-back token exchange |

### 5.2 Standard Confirm (Level 1)

Operations that modify state but are reversible require `--confirm`:

```bash
dbxcli db session kill entity_name=prod-orcl session_id=1234 --confirm
```

Without `--confirm`:

```
ERROR: confirmation required: pass confirm=true or --confirm
  Operation: kill session 1234 on prod-orcl
  To proceed, rerun with --confirm
```

### 5.3 Standard+Echo Confirm (Level 2)

Significant or potentially data-affecting operations require the operator to type an exact string:

```bash
dbxcli db tablespace drop entity_name=prod-orcl tablespace=USERS_OLD --confirm
```

The CLI prompts:

```
To drop tablespace USERS_OLD on prod-orcl, type: DROP USERS_OLD
>
```

The operator must type `DROP USERS_OLD` exactly (case-sensitive) to proceed.

### 5.4 Double-Confirm (Level 3)

Catastrophic, irreversible operations (failover, disaster recovery activation, complete schema drop) require two sequential confirmations:

```bash
dbxcli pg cnpg promote entity_name=prod-pg --confirm --confirm-destructive
```

The CLI prompts twice:

```
To promote prod-pg to primary, type: PROMOTE prod-pg
> PROMOTE prod-pg

WARNING: This is irreversible. Type CONFIRM PROMOTE to proceed.
> CONFIRM PROMOTE
```

Both strings must match exactly. A mismatch at either step aborts the operation.

### 5.5 CLI Bypass Flags

| Flag | Effect |
|------|--------|
| `--confirm` | Satisfies Level 1 (Standard) without interactive prompt |
| `--confirm --confirm-destructive` | Satisfies Level 2 and Level 3 when combined with the echo string |

For use in automation and scripts, set both flags and pipe the echo string via stdin:

```bash
echo "DROP USERS_OLD" | dbxcli db tablespace drop \
  entity_name=prod-orcl \
  tablespace=USERS_OLD \
  --confirm
```

### 5.6 MCP Confirm Behavior

When dbx is invoked via the MCP adapter (AI assistant context), the confirm gate uses an echo-back token pattern rather than interactive prompts:

1. The tool call returns a `confirm_required` response containing an `echo_token`:

   ```json
   {
     "confirm_required": true,
     "confirm_level": "echo_back",
     "echo_token": "DROP USERS_OLD",
     "description": "Drop tablespace USERS_OLD on prod-orcl — this operation is irreversible"
   }
   ```

2. The AI assistant presents the echo token to the human operator and asks for confirmation.

3. The human explicitly approves. The AI assistant reissues the tool call with `confirm=true` and `confirm_echo=DROP USERS_OLD`.

4. The gate validates the echo matches and proceeds.

For Level 3 (Double-Confirm), steps 1-4 repeat with a second `echo_token`.

### 5.7 Operation Confirm Level Reference

| Operation | Level | Description |
|-----------|-------|-------------|
| `db session list` | None | Read-only |
| `db tablespace list` | None | Read-only |
| `pg table list` | None | Read-only |
| `db session kill` | Standard | Kill a session — reversible (session reconnects) |
| `db parameter set` | Standard | Modify a database parameter |
| `pg user alter` | Standard | Change user attributes or password |
| `db tablespace add` | Standard | Add datafile to tablespace |
| `db schema drop` | Standard+Echo | Drop a schema — potentially significant data loss |
| `db tablespace drop` | Standard+Echo | Drop a tablespace |
| `pg database drop` | Standard+Echo | Drop a PostgreSQL database |
| `dataguard switchover` | Standard+Echo | Planned role transition |
| `pg cnpg promote` | Double-Confirm | Emergency standalone promotion |
| `dataguard failover` | Double-Confirm | Unplanned failover — irreversible without reinstatement |
| `db datapump drop-job` | Standard+Echo | Delete a running Data Pump job |

---

## 6. Audit Trail Administration

Every tool invocation in dbx generates an audit event. Audit events are written to one or more configured sinks. The audit trail is immutable: events are append-only and include a SHA-256 chain for tamper detection.

### 6.1 Audit Event Structure

Each audit event contains:

| Field | Type | Description |
|-------|------|-------------|
| `event_id` | string | Unique 32-character hex ID |
| `timestamp` | RFC3339 | Event creation time |
| `interface` | string | `cli`, `mcp`, or `rest` |
| `user` | string | OS user or API key identifier |
| `tool` | string | Full tool name (e.g., `db.session.kill`) |
| `target` | string | Target entity name |
| `parameters` | object | Input parameters (sensitive fields redacted) |
| `confirm_type` | string | Confirm level applied, if any |
| `confirm_echo` | string | Echo token, if used |
| `result` | string | `success`, `error`, `blocked` |
| `error` | string | Error message on failure |
| `duration_ms` | int | Execution duration in milliseconds |
| `redacted_fields` | array | List of parameter fields that were redacted |

### 6.2 Audit Sink Configuration

Configure audit sinks in `~/.dbx/config.yaml`:

```yaml
audit_sink: file
```

Or set multiple sinks:

```yaml
audit:
  sinks:
    - type: file
      path: /var/log/dbx/audit.jsonl
      rotate_daily: true
      retain_days: 90

    - type: syslog
      facility: local0
      tag: dbx-audit

    - type: postgres
      dsn_env: DBX_AUDIT_PG_DSN
      table: dbx_audit_events

    - type: victoriametrics
      url: http://vm.example.com:8428/api/v1/import/jsonline
      labels:
        environment: production

    - type: slack
      webhook_env: DBX_AUDIT_SLACK_WH
      min_level: warn   # Only send warn/error events to Slack

    - type: stdout
```

The `file` sink writes newline-delimited JSON (NDJSON). Each line is one audit event.

### 6.3 Redaction Rules

Sensitive parameter values are automatically redacted in audit events. The following parameter names are always redacted:

- `password`
- `new_password`
- `old_password`
- `secret`
- `private_key`
- `passphrase`
- `token`
- `api_key`

Redacted fields appear as `"***REDACTED***"` in the audit event, and the field name is listed in `redacted_fields`.

To add custom redaction rules, configure in `~/.dbx/config.yaml`:

```yaml
audit:
  redact_fields:
    - my_secret_param
    - connection_string
```

### 6.4 Retention Policies

For the file sink, configure daily rotation and retention:

```yaml
audit:
  sinks:
    - type: file
      path: /var/log/dbx/audit.jsonl
      rotate_daily: true
      retain_days: 365
```

For the PostgreSQL sink, implement retention via a scheduled job:

```sql
DELETE FROM dbx_audit_events
WHERE timestamp < NOW() - INTERVAL '365 days';
```

For VictoriaMetrics, configure retention at the storage level using the `-retentionPeriod` flag on the VictoriaMetrics instance.

### 6.5 Tamper Detection

The file sink supports SHA-256 event chaining. Each event includes a `chain_hash` field that is the SHA-256 of the previous event's `chain_hash` concatenated with the current event's JSON. The genesis event uses a random seed written to `~/.dbx/audit-seed.hex`.

To verify the chain integrity:

```bash
dbxcli audit verify --file /var/log/dbx/audit.jsonl
```

Output:

```
Audit chain verification: PASS
  Events: 14823
  First:  2026-01-01T00:00:12Z (event_id: a1b2c3...)
  Last:   2026-04-11T09:22:45Z (event_id: f9e8d7...)
  Chain:  intact
```

If tampering is detected:

```
Audit chain verification: FAIL
  Events:  14823
  Break:   event_id c4d5e6... at 2026-03-15T14:22:01Z
  Reason:  chain_hash mismatch — event may have been modified or deleted
```

### 6.6 Querying the Audit Log

Query audit events with filters:

```bash
# All events for a specific target in the last 24 hours
dbxcli audit query target=prod-orcl since=24h

# All destructive operations (Standard+Echo and above)
dbxcli audit query confirm_type=echo_back

# All blocked operations
dbxcli audit query result=blocked

# Events by a specific user
dbxcli audit query user=alice since=7d

# All operations on a tool prefix
dbxcli audit query tool_prefix=db.session

# Output as JSON for external processing
dbxcli audit query since=7d --format json > audit-week.json
```

---

## 7. User and RBAC Administration (EE)

User and RBAC management is available in the Core Bundle and higher. It requires the REST API server (`dbxctl serve`) and is managed via `dbxcli` or the REST API.

### 7.1 Role Hierarchy

| Role | Description | Typical Use |
|------|-------------|-------------|
| `admin` | Full platform access — manage users, targets, config | Platform administrators |
| `operator` | Execute all read and write operations on assigned targets | DBAs, operations teams |
| `analyst` | Read-only access to all data — query, view, export | Developers, application teams |
| `viewer` | View target status and metadata only | Management, monitoring |
| `api-automation` | Restricted API key role for automation — scoped to specific tools | CI/CD pipelines, scheduled jobs |

Roles are hierarchical: `admin` inherits all `operator` permissions, `operator` inherits all `analyst` permissions, and so on.

### 7.2 Fleet Scoping

Users are granted access to specific target groups (fleet scopes), not all targets by default.

Create a user with fleet scope:

```bash
dbxcli user add \
  username=alice \
  role=operator \
  fleet_scope=production,oracle-fleet
```

List users and their scopes:

```bash
dbxcli user list
```

Update a user's fleet scope:

```bash
dbxcli user update username=alice fleet_scope=production
```

Remove a user:

```bash
dbxcli user remove username=alice --confirm
```

A user with `fleet_scope=production` can only see and operate on targets in the `production` group. Attempts to access targets outside their scope return:

```
ERROR: access denied — target dr-orcl is not in your fleet scope
```

### 7.3 OIDC Integration

dbx supports OIDC for user authentication. Configure in `~/.dbx/config.yaml`:

```yaml
auth:
  oidc:
    issuer: https://keycloak.example.com/realms/itunified
    client_id: dbx
    client_secret_env: DBX_OIDC_CLIENT_SECRET
    redirect_url: http://localhost:9090/callback
    scopes:
      - openid
      - profile
      - email
      - groups
    groups_claim: groups
    role_mapping:
      dbx-admins: admin
      dbx-operators: operator
      dbx-analysts: analyst
      dbx-viewers: viewer
```

The same configuration works for Azure AD, Okta, and Google Workspace OIDC providers. Replace `issuer` and `client_id` with the values from your identity provider.

For Azure AD:

```yaml
auth:
  oidc:
    issuer: https://login.microsoftonline.com/<tenant-id>/v2.0
    client_id: <application-id>
    client_secret_env: DBX_OIDC_CLIENT_SECRET
    scopes:
      - openid
      - profile
      - email
      - https://graph.microsoft.com/GroupMember.Read.All
```

### 7.4 API Key Management

API keys allow automation tools and CI/CD pipelines to authenticate without OIDC.

Create an API key:

```bash
dbxcli apikey create \
  name=ci-pipeline \
  role=api-automation \
  fleet_scope=production \
  expiry=90d
```

Output:

```
API Key created.
  Name:        ci-pipeline
  Key ID:      key-abc123
  Secret:      dbx_sk_...  (shown once — copy now)
  Role:        api-automation
  Fleet Scope: production
  Expires:     2026-07-10
```

The secret is shown exactly once. Store it in your secrets manager immediately.

List API keys:

```bash
dbxcli apikey list
```

Revoke an API key:

```bash
dbxcli apikey revoke key-id=key-abc123 --confirm
```

API keys are authenticated via the `Authorization: Bearer dbx_sk_...` header in REST API calls, or the `DBX_API_KEY` environment variable for `dbxcli`.

### 7.5 Session Management

List active sessions:

```bash
dbxcli session list
```

Revoke a specific session:

```bash
dbxcli session revoke session-id=sess-xyz789 --confirm
```

Revoke all sessions for a user (force re-authentication):

```bash
dbxcli session revoke-all username=alice --confirm
```

Session tokens expire after the TTL configured in `~/.dbx/config.yaml`:

```yaml
auth:
  session_ttl: 8h
  api_key_ttl: 1h
```

---

## 8. Output Formatting

All `dbxcli` commands support a `--format` flag that controls how results are rendered.

### 8.1 Format Options

| Format | Description | Use Case |
|--------|-------------|----------|
| `table` | Human-readable aligned table (default) | Interactive terminal use |
| `json` | Pretty-printed JSON | Debugging, REST API compatibility |
| `yaml` | YAML output | Config generation, GitOps pipelines |

Set the default format in `~/.dbx/config.yaml`:

```yaml
default_format: table
```

Override per-command:

```bash
dbxcli db session list --format json
dbxcli target list --format yaml
```

### 8.2 Table Format

The default table format uses aligned columns suitable for terminal display:

```
TARGET         TYPE              HOST                       STATUS
prod-orcl      oracle_database   db01.example.com:1521     connected
prod-pg        pg_database       pg01.example.com:5432     connected
db-host-01     oracle_host       db01.example.com          reachable
```

### 8.3 JSON Format

JSON output is an array of objects for list commands, or a single object for single-entity commands:

```bash
dbxcli target list --format json
```

```json
[
  {
    "name": "prod-orcl",
    "type": "oracle_database",
    "host": "db01.example.com",
    "port": 1521,
    "service": "ORCL",
    "status": "connected"
  },
  {
    "name": "prod-pg",
    "type": "pg_database",
    "host": "pg01.example.com",
    "port": 5432,
    "database": "appdb",
    "status": "connected"
  }
]
```

Error responses always include a top-level `error` field:

```json
{
  "error": "target not found: nonexistent-target"
}
```

### 8.4 MCP JSON Output Format

The MCP adapter wraps tool results in the standard MCP response envelope:

```json
{
  "content": [
    {
      "type": "text",
      "text": "[{\"name\":\"prod-orcl\",\"type\":\"oracle_database\",...}]"
    }
  ]
}
```

For confirm gate responses:

```json
{
  "content": [
    {
      "type": "text",
      "text": "{\"confirm_required\":true,\"confirm_level\":\"echo_back\",\"echo_token\":\"DROP USERS_OLD\",\"description\":\"Drop tablespace USERS_OLD on prod-orcl\"}"
    }
  ]
}
```

### 8.5 REST API JSON Format

The REST API (`dbxctl serve`, default port 8080) returns standard JSON responses:

```
GET /api/v1/targets
```

```json
{
  "targets": [
    {
      "name": "prod-orcl",
      "type": "oracle_database",
      ...
    }
  ],
  "count": 2
}
```

Error responses use RFC 7807 Problem Details:

```json
{
  "type": "https://dbx.itunified.io/errors/target-not-found",
  "title": "Target Not Found",
  "status": 404,
  "detail": "target 'nonexistent' is not registered",
  "instance": "/api/v1/targets/nonexistent"
}
```

### 8.6 Pipe-Friendly Output for Scripting

For shell scripting, use `--format json` combined with `jq`:

```bash
# Extract all oracle_database target names
dbxcli target list --format json | jq -r '.[] | select(.type=="oracle_database") | .name'

# Check if a specific target is connected
STATUS=$(dbxcli target test entity_name=prod-orcl --format json | jq -r '.status')
if [ "$STATUS" = "connected" ]; then
  echo "Target is reachable"
fi

# Extract session count
dbxcli db session list --format json | jq 'length'
```

For plain-text single-value output, use the `value` format (where supported):

```bash
dbxcli db parameter get entity_name=prod-orcl name=db_name --format value
# Output: ORCL
```

---

## 9. Upgrade and Migration

### 9.1 Binary Upgrade Procedure

**macOS (Homebrew):**

```bash
brew update && brew upgrade dbx
```

**Linux (install script):**

```bash
curl -fsSL https://get.itunified.de/dbx | sh
```

The install script is idempotent: it replaces the existing binary in `/usr/local/bin` with the new version.

**Manual binary replacement:**

1. Download the new binary for your platform from the GitHub release page.
2. Verify the SHA-256 checksum published in the release notes.
3. Replace the binary:

   ```bash
   chmod +x dbxcli-linux-amd64
   sudo mv dbxcli-linux-amd64 /usr/local/bin/dbxcli
   ```

4. Verify the installed version:

   ```bash
   dbxcli version
   ```

**Docker:**

```bash
docker pull ghcr.io/itunified-io/dbx:latest
```

Pin to a specific CalVer tag for production deployments:

```bash
docker pull ghcr.io/itunified-io/dbx:v2026.04.11.1
```

### 9.2 MCP Adapter Upgrade

The MCP adapters are distributed as npm packages. Upgrade globally:

```bash
npm update -g @itunified.io/mcp-oracle
npm update -g @itunified.io/mcp-postgres
```

For Claude Code, re-register the MCP server after upgrade to pick up the new binary:

```bash
claude mcp remove oracle
claude mcp add oracle -- npx -y @itunified.io/mcp-oracle
```

For Claude Desktop, no action is needed if you use `npx -y` in the server command — the `latest` tag is resolved at each startup. If you have pinned a specific version, update it in `claude_desktop_config.json`.

Verify the MCP adapter version:

```bash
npx @itunified.io/mcp-oracle --version
```

### 9.3 Rolling Upgrade for Monitoring Stack

When dbxcli is deployed alongside a monitoring agent (dbxctl in agent mode), upgrade the monitoring stack with a rolling strategy to avoid gaps in metric collection:

1. Upgrade the standby or secondary monitoring instance first.
2. Verify the upgraded instance is collecting metrics correctly.
3. Upgrade the primary monitoring instance.
4. Verify full metric continuity.

For Kubernetes deployments of the monitoring stack:

```bash
kubectl set image deployment/dbx-monitor dbx-monitor=ghcr.io/itunified-io/dbx:v2026.04.11.1 -n monitoring
kubectl rollout status deployment/dbx-monitor -n monitoring
```

### 9.4 Configuration Store Migration

dbx stores its runtime configuration in `~/.dbx/config.yaml` (or `$DBX_DATA_DIR/config.yaml`). The target YAML files in `~/.dbx/targets/` are forward-compatible: fields added in new versions are ignored by older versions, and missing optional fields default to zero values.

When a new version introduces a required field or schema change, a migration command is provided:

```bash
dbxcli config migrate --dry-run
dbxcli config migrate
```

The `--dry-run` flag prints the proposed changes without applying them. Review the output before running the migration.

Back up your configuration before migrating:

```bash
cp -r ~/.dbx ~/.dbx.backup-$(date +%Y%m%d)
```

### 9.5 License Compatibility Across Versions

License files are forward-compatible. A license issued for an older version of dbx is valid in newer versions as long as:

- The Ed25519 signature remains intact.
- The license has not expired.
- The license issuer URL (`license.itunified.io`) is reachable for phone-home verification.

If a new version of dbx introduces a new bundle or tier name not present in your existing license, the new feature domain will be treated as unlicensed until the license is reissued with the new bundle. Contact support@itunified.io for license reissuance.

To verify license compatibility after upgrade:

```bash
dbxcli license status
```

A status of `VALID` confirms the existing license works with the new binary version.

---

## Appendix: Configuration Reference

### ~/.dbx/config.yaml — Full Schema

```yaml
# Data directory (default: ~/.dbx)
data_dir: /opt/dbx/data

# Oracle license gate enforcement mode: strict | warn | audit-only
oracle_gate_mode: strict

# Default audit sink: file | syslog | postgres | victoriametrics | slack | stdout
audit_sink: file

# REST API server port (default: 8080)
rest_port: 8080

# Vault configuration
vault:
  address: https://vault.example.com:8200
  auth_method: approle
  role_id_env: VAULT_ROLE_ID
  secret_id_env: VAULT_SECRET_ID

# Audit configuration
audit:
  sinks:
    - type: file
      path: /var/log/dbx/audit.jsonl
      rotate_daily: true
      retain_days: 90
  redact_fields: []

# Authentication (EE)
auth:
  session_ttl: 8h
  api_key_ttl: 1h
  oidc:
    issuer: https://keycloak.example.com/realms/itunified
    client_id: dbx
    client_secret_env: DBX_OIDC_CLIENT_SECRET
    redirect_url: http://localhost:9090/callback
    scopes:
      - openid
      - profile
      - email
      - groups
    groups_claim: groups
    role_mapping:
      dbx-admins: admin
      dbx-operators: operator
      dbx-analysts: analyst
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DBX_DATA_DIR` | `~/.dbx` | Override the data directory |
| `DBX_ORACLE_GATE_MODE` | `strict` | Override Oracle gate enforcement mode |
| `DBX_AUDIT_SINK` | `file` | Override the audit sink type |
| `DBX_REST_PORT` | `8080` | Override the REST API port |
| `DBX_VAULT_ADDRESS` | — | Override the Vault address |
| `VAULT_ROLE_ID` | — | AppRole role ID |
| `VAULT_SECRET_ID` | — | AppRole secret ID |
| `DBX_API_KEY` | — | API key for non-interactive authentication |
| `DBX_TARGET` | — | Default target entity name for MCP adapters |
