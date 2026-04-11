# PostgreSQL Engine Installation Guide

**Audience:** PostgreSQL DBA  
**Estimated setup time:** ~10 minutes  
**Applies to:** dbx PostgreSQL Engine (OSS + licensed tiers)

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Install MCP Adapter](#install-mcp-adapter)
3. [Configure Connection via YAML](#configure-connection-via-yaml)
4. [CNPG Configuration](#cnpg-configuration)
5. [Verify Connection](#verify-connection)
6. [Tools Unlocked by Tier](#tools-unlocked-by-tier)
7. [Vault Integration](#vault-integration)
8. [Multi-Profile Connection Switching](#multi-profile-connection-switching)

---

## Prerequisites

| Requirement | Minimum Version | Notes |
|---|---|---|
| Node.js | 18+ | Required for MCP adapter |
| PostgreSQL client libraries | 14+ | `libpq` required for native connections |
| `kubectl` | 1.26+ | Optional — required for CloudNativePG (CNPG) operations |
| HashiCorp Vault | 1.13+ | Optional — strongly recommended for credential management |
| dbxcli binary | Latest | Required for target registration and CLI operations |

Confirm Node.js version:

```bash
node --version
# v18.x.x or higher required
```

Confirm PostgreSQL client availability:

```bash
psql --version
# psql (PostgreSQL) 16.x or similar
```

---

## Install MCP Adapter

### Free Tier (OSS) — 27 tools

```bash
npm install -g @itunified.io/mcp-postgres
```

### PG Professional (Licensed) — 111 additional tools, 32 skills

Requires a valid dbx license key. The license is validated at startup.

```bash
npm install -g @itunified.io/mcp-postgres-enterprise
```

Set the license key (loaded from Vault or environment):

```bash
export DBX_LICENSE_KEY=<your-license-key>
```

Or reference from Vault:

```bash
dbxcli license configure \
  vault_path=secret/dbx/license \
  key_field=postgres_enterprise
```

Confirm installed version:

```bash
npx @itunified.io/mcp-postgres --version
npx @itunified.io/mcp-postgres-enterprise --version
```

---

## Configure Connection via YAML

Connection profiles are stored in `~/.dbx/postgres/` as named YAML files. Multiple profiles can coexist and are referenced by name at runtime.

### Profile file: `~/.dbx/postgres/prod-pg.yaml`

```yaml
# dbx PostgreSQL connection profile
# Profile name matches the filename (prod-pg)

profiles:
  primary:
    host: db.example.com
    port: 5432
    database: appdb
    sslmode: verify-full
    sslrootcert: ~/.dbx/certs/ca.crt
    credential: vault
    vault_path: secret/dbx/postgres/prod-pg/primary

  replica:
    host: replica.example.com
    port: 5432
    database: appdb
    sslmode: verify-full
    sslrootcert: ~/.dbx/certs/ca.crt
    credential: vault
    vault_path: secret/dbx/postgres/prod-pg/replica

  admin:
    host: db.example.com
    port: 5432
    database: postgres
    sslmode: verify-full
    sslrootcert: ~/.dbx/certs/ca.crt
    credential: vault
    vault_path: secret/dbx/postgres/prod-pg/admin

default_profile: primary
```

**Supported `sslmode` values:**

| Value | Description |
|---|---|
| `disable` | No SSL — development/lab only |
| `require` | SSL required, no certificate verification |
| `verify-ca` | SSL with CA certificate verification |
| `verify-full` | SSL with hostname + CA verification (recommended for production) |

**`credential` options:**

| Value | Description |
|---|---|
| `vault` | Fetch username and password from HashiCorp Vault at runtime |
| `inline` | Username and password stored directly in YAML (not recommended for production) |
| `pgpass` | Read from `~/.pgpass` file |
| `env` | Read from `PGUSER` / `PGPASSWORD` environment variables |

Register the profile as a dbx target:

```bash
dbxcli target add \
  entity_name=prod-pg \
  entity_type=postgres \
  config_file=~/.dbx/postgres/prod-pg.yaml \
  profile=primary
```

---

## CNPG Configuration

CloudNativePG (CNPG) configuration extends the base profile with cluster topology and disaster recovery (DR) settings. This is optional and only required when managing CNPG-deployed clusters.

### DR cluster config: `~/.dbx/postgres/cnpg-dr.yaml`

```yaml
# dbx CNPG DR cluster configuration

cnpg:
  primary_cluster:
    name: pg-primary
    namespace: database
    kubeconfig: ~/.kube/config
    context: production-cluster
    credential: vault
    vault_path: secret/dbx/postgres/cnpg/primary

  dr_cluster:
    name: pg-dr
    namespace: database
    kubeconfig: ~/.kube/config
    context: dr-cluster
    credential: vault
    vault_path: secret/dbx/postgres/cnpg/dr
    wal_archive:
      type: minio
      endpoint: https://minio.example.com
      bucket: pg-wal-archive
      credential: vault
      vault_path: secret/dbx/minio/wal-archive

  replication:
    mode: async
    recovery_target: latest
    pitr_enabled: true
```

Register the CNPG configuration:

```bash
dbxcli target add \
  entity_name=prod-cnpg \
  entity_type=postgres_cnpg \
  config_file=~/.dbx/postgres/cnpg-dr.yaml
```

---

## Verify Connection

After registration, verify connectivity with the following commands:

```bash
# Check connection status for the primary profile
dbxcli pg connection status entity=prod-pg profile=primary

# List all tables in the default database
dbxcli pg table list entity=prod-pg schema=public

# Report total database size
dbxcli pg database size entity=prod-pg
```

Expected output for `connection status`:

```
Profile      : primary
Host         : db.example.com:5432
Database     : appdb
Server       : PostgreSQL 16.2 on x86_64-pc-linux-gnu
SSL          : verify-full (TLSv1.3, cipher TLS_AES_256_GCM_SHA384)
Credential   : vault (secret/dbx/postgres/prod-pg/primary)
Status       : CONNECTED
```

If connection fails:

```bash
dbxcli target diagnose prod-pg
```

---

## Tools Unlocked by Tier

### Free Tier — 27 tools

| Category | Count | Tools |
|---|---|---|
| Connection | 5 | `pg_connect`, `pg_status`, `pg_version`, `pg_settings`, `pg_extensions` |
| Query | 3 | `pg_query`, `pg_explain`, `pg_query_stats` |
| Schema | 9 | `pg_table_list`, `pg_column_list`, `pg_index_list`, `pg_constraint_list`, `pg_view_list`, `pg_sequence_list`, `pg_function_list`, `pg_schema_list`, `pg_object_count` |
| CRUD | 4 | `pg_insert`, `pg_update`, `pg_delete`, `pg_select` |
| Server | 4 | `pg_process_list`, `pg_lock_list`, `pg_bloat_report`, `pg_vacuum_stats` |
| Database | 2 | `pg_database_list`, `pg_database_size` |

### PG Professional — +111 tools, +32 skills

| Category | Count | Tools |
|---|---|---|
| CNPG | 6 | Cluster status, failover, switchover, backup trigger, restore point, cluster describe |
| CNPG-DR | 18 | WAL archive status, DR lag, failover drill, PITR restore, cross-cluster replication, DR switchover, recovery window, backup list, WAL replay status, and more |
| HA | 10 | Patroni status, Repmgr topology, primary election, failover trigger, standby promote, quorum check, connection pool stats, HAProxy config, PgBouncer reload, HA health report |
| Backup | 4 | `pg_backup_create`, `pg_backup_list`, `pg_backup_restore`, `pg_backup_delete` |
| DBA | 16 | Index advisor, autovacuum tuning, table bloat, dead tuple cleanup, statistics reset, table freeze, pg_dump/restore wrappers, schema diff, sequence sync, and more |
| Security | 4 | `pg_role_audit`, `pg_privilege_matrix`, `pg_row_security_list`, `pg_ssl_config` |
| Audit | 3 | `pg_audit_log_search`, `pg_audit_config`, `pg_statement_log` |
| Compliance | 5 | CIS benchmark check, SOC 2 posture, GDPR field audit, encryption at rest check, access log summary |
| RBAC | 4 | `pg_role_create`, `pg_role_grant`, `pg_role_revoke`, `pg_role_tree` |
| Replication | 4 | `pg_replication_slots`, `pg_replication_lag`, `pg_wal_receiver_status`, `pg_replication_summary` |
| Migration | 4 | `pg_schema_export`, `pg_schema_import`, `pg_data_diff`, `pg_migration_plan` |
| Observability | 4 | `pg_query_heatmap`, `pg_wait_events`, `pg_table_access_stats`, `pg_io_stats` |
| Tenant | 5 | `pg_tenant_provision`, `pg_tenant_list`, `pg_tenant_isolate`, `pg_tenant_backup`, `pg_tenant_drop` |
| WAL | 5 | `pg_wal_stats`, `pg_wal_archive_status`, `pg_wal_size`, `pg_checkpoint_stats`, `pg_wal_switch` |
| RAG | 7 | `pg_vector_index_create`, `pg_vector_search`, `pg_embedding_insert`, `pg_embedding_list`, `pg_vector_stats`, `pg_index_refresh`, `pg_similarity_search` |
| Vault | 3 | `pg_vault_rotate`, `pg_vault_lease_renew`, `pg_vault_credential_audit` |
| Connection | 3 | `pg_pool_stats`, `pg_pool_reload`, `pg_connection_limit` |
| Policy | 2 | `pg_rls_create`, `pg_rls_test` |
| Health | 1 | `pg_health_report` |
| Capacity | 1 | `pg_capacity_forecast` |

---

## Vault Integration

When using Vault credentials, the Vault path must contain `username` and `password` keys.

Seed credentials for a profile:

```bash
vault kv put secret/dbx/postgres/prod-pg/primary \
  username=app_user \
  password=<redacted>
```

Configure Vault backend (if not already done):

```bash
dbxcli vault configure \
  vault_addr=https://vault.example.com:8200 \
  vault_mount=secret \
  vault_auth_method=approle \
  role_id=<role-id> \
  secret_id=<secret-id>
```

Test Vault connectivity:

```bash
dbxcli vault test
```

Vault credentials are fetched at connection time and never written to disk. Lease renewal is handled automatically by the PG Professional `pg_vault_lease_renew` tool.

---

## Multi-Profile Connection Switching

When multiple profiles are defined in a YAML config, switch between them at command invocation time:

```bash
# Query against the replica profile
dbxcli pg query entity=prod-pg profile=replica sql="SELECT count(*) FROM orders"

# Run DBA tasks against the admin profile
dbxcli pg bloat report entity=prod-pg profile=admin

# Switch the default profile for the active session
dbxcli target set prod-pg default_profile=replica
```

To list all configured profiles for a target:

```bash
dbxcli target profiles prod-pg
```

Output:

```
Profile    Host                  Database   SSL           Default
primary    db.example.com:5432   appdb      verify-full   yes
replica    replica.example.com   appdb      verify-full   no
admin      db.example.com:5432   postgres   verify-full   no
```

Profile selection order (highest to lowest priority):

1. Explicit `profile=` parameter on the command
2. `DBX_PG_PROFILE` environment variable
3. `default_profile` in the YAML config file
4. First declared profile in the file
