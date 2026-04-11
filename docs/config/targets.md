# Target YAML Reference

Targets are the named connection endpoints that `dbxcli` and the MCP adapters operate against. Each target is stored as a YAML file at `~/.dbx/targets/<name>.yaml`. The target name corresponds to the file stem and must match the `entity_name` field inside the file.

---

## Storage Location

```
~/.dbx/targets/
  prod-orcl.yaml
  prod-pg.yaml
  rac-cluster.yaml
  web01.yaml
```

All files in this directory are loaded at startup. Symlinks are followed. Files that fail YAML parse are skipped with a warning; they do not abort startup.

---

## Common Fields

The following fields are valid on every target type unless noted otherwise.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `entity_name` | string | — | Unique name for this target. Must match the filename stem. Required. |
| `entity_type` | string | — | Target type (see types below). Required. |
| `description` | string | `""` | Free-text description shown in `target list`. |
| `tags` | list[string] | `[]` | Arbitrary labels for grouping and filtering. |
| `group` | string | `""` | Fleet group membership (see Groups section). |
| `enabled` | bool | `true` | Set to `false` to disable without removing the file. |
| `timeout_seconds` | int | `30` | Connection and query timeout. Range: 1–3600. |
| `vault` | object | — | Vault credential source (see Vault Integration). |
| `env_override` | object | — | Environment variable overrides (see Env Overrides). |

---

## Target Types

### `oracle_database`

A single-instance Oracle database or a PDB within a CDB.

```yaml
entity_name: prod-orcl
entity_type: oracle_database
description: Production Oracle 19c database

host: db01.example.com
port: 1521                    # default: 1521
service: ORCL                 # service name or SID
protocol: tcp                 # tcp | tcps (TLS). default: tcp
tns_alias: ""                 # optional: use TNS alias instead of host/port/service
wallet_path: ""               # optional: Oracle Wallet directory (for tcps)
sysdba: false                 # connect AS SYSDBA. default: false

user: dba_user
password: ""                  # plaintext only for development. use vault in production
vault:
  path: secret/data/oracle/prod-orcl
  user_key: username
  password_key: password

nls_lang: "AMERICAN_AMERICA.AL32UTF8"  # default: empty (inherits environment)
instant_client_path: ""       # override auto-detected Instant Client path

oracle_edition: enterprise    # enterprise | standard2. required for licensed tools
oracle_options:               # licensed options declared for this target
  - diagnostics_pack
  - tuning_pack

tags: [production, oracle, tier1]
group: prod-fleet
timeout_seconds: 30
```

**Validation rules:**
- One of `host`/`port`/`service` or `tns_alias` must be provided.
- When `protocol: tcps`, either `wallet_path` or a system trust store must be available.
- `oracle_edition` and `oracle_options` are required to use licensed tools; the CLI enforces this at tool execution time.
- `user` and `password` are mutually exclusive with `vault` for the same credential. If both are set, `vault` takes precedence.

---

### `rac_database`

An Oracle Real Application Clusters database with SCAN-based connectivity.

```yaml
entity_name: rac-prod
entity_type: rac_database
description: Production 3-node RAC cluster

scan_host: rac-scan.example.com
scan_port: 1521               # default: 1521
service: RACPRD               # service name registered on the SCAN listener
protocol: tcp

user: dba_user
vault:
  path: secret/data/oracle/rac-prod
  user_key: username
  password_key: password

rac_nodes:                    # individual node connection details for direct access
  - name: rac1
    host: rac1.example.com
    port: 1521
  - name: rac2
    host: rac2.example.com
    port: 1521
  - name: rac3
    host: rac3.example.com
    port: 1521

oracle_edition: enterprise
oracle_options:
  - diagnostics_pack
  - tuning_pack
  - real_application_testing

tags: [production, rac, oracle]
group: prod-fleet
```

**Validation rules:**
- `scan_host` and `service` are required.
- Individual `rac_nodes` entries are optional but required for node-specific operations (`db rac node-status`).

---

### `oracle_listener`

An Oracle Net listener for health checks, service inspection, and restart operations.

```yaml
entity_name: prod-listener
entity_type: oracle_listener
description: Production Oracle listener on db01

host: db01.example.com
port: 1521
listener_name: LISTENER       # default: LISTENER

ssh_target: db01-ssh          # oracle_host target name for lsnrctl operations via SSH

tags: [production, listener]
group: prod-fleet
```

**Validation rules:**
- `host` and `port` are required.
- `ssh_target` must reference a registered `oracle_host` target when SSH-based lsnrctl operations are used.

---

### `oracle_asm`

An Oracle Automatic Storage Management instance.

```yaml
entity_name: prod-asm
entity_type: oracle_asm
description: Production ASM instance

host: db01.example.com
port: 1521
service: +ASM                 # ASM service name, typically +ASM

user: asmsnmp
vault:
  path: secret/data/oracle/prod-asm
  user_key: username
  password_key: password

ssh_target: db01-ssh          # oracle_host target for asmcmd operations

tags: [production, asm]
group: prod-fleet
```

---

### `oracle_host`

An Oracle database server accessible via SSH for OS-level and ASM operations.

```yaml
entity_name: db01-ssh
entity_type: oracle_host
description: Oracle DB server db01 — SSH access

host: db01.example.com
port: 22                      # default: 22
user: oracle                  # OS user for SSH
ssh_key_path: ~/.ssh/id_ed25519
ssh_jump_host: ""             # optional: bastion/jump host (user@host format)
ssh_known_hosts_file: ~/.ssh/known_hosts
strict_host_key_checking: true  # default: true. set false only in controlled environments

oracle_home: /u01/app/oracle/product/19.0.0/db_1
oracle_base: /u01/app/oracle
oracle_sid: ORCL

vault:
  path: secret/data/ssh/db01
  ssh_key: oracle_ssh_key     # vault key containing PEM-encoded private key

tags: [production, ssh, oracle]
group: prod-fleet
```

**Validation rules:**
- `host` and `user` are required.
- One of `ssh_key_path` or `vault.ssh_key` must be provided.
- `strict_host_key_checking: false` emits a warning in strict license enforcement mode.

---

### `pg_database`

A single PostgreSQL database.

```yaml
entity_name: prod-pg
entity_type: pg_database
description: Production PostgreSQL 16 database

host: pg01.example.com
port: 5432                    # default: 5432
database: appdb               # default: postgres
user: dba_user
password: ""                  # use vault in production

sslmode: require              # disable | allow | prefer | require | verify-ca | verify-full
sslcert: ""                   # path to client certificate
sslkey: ""                    # path to client private key
sslrootcert: ""               # path to CA certificate

vault:
  path: secret/data/postgres/prod-pg
  user_key: username
  password_key: password

connect_timeout: 10           # seconds. default: 10
application_name: dbxcli      # default: dbxcli

tags: [production, postgres, tier1]
group: pg-fleet
timeout_seconds: 30
```

---

### `pg_cluster`

A PostgreSQL cluster with primary and one or more standbys. Used for HA operations, replication monitoring, and switchover.

```yaml
entity_name: pg-ha-cluster
entity_type: pg_cluster
description: 3-node Patroni cluster

primary:
  host: pg-primary.example.com
  port: 5432
  database: appdb
  user: dba_user
  vault:
    path: secret/data/postgres/pg-ha-cluster
    user_key: username
    password_key: password

standbys:
  - name: standby1
    host: pg-standby1.example.com
    port: 5432
  - name: standby2
    host: pg-standby2.example.com
    port: 5432

patroni_api:
  host: pg-primary.example.com
  port: 8008
  vault:
    path: secret/data/postgres/pg-ha-cluster-patroni
    token_key: api_token

pgbouncer:
  host: pgbouncer.example.com
  port: 6432
  admin_database: pgbouncer
  vault:
    path: secret/data/postgres/pgbouncer
    user_key: username
    password_key: password

sslmode: verify-full
sslrootcert: /etc/ssl/certs/pg-ca.crt

tags: [production, postgres, ha]
group: pg-fleet
```

---

### `host`

A generic Linux/Unix host for OS-level diagnostics, kernel checks, and package management. Not specific to Oracle or PostgreSQL.

```yaml
entity_name: web01
entity_type: host
description: Web application server

host: web01.example.com
port: 22
user: sysadmin
ssh_key_path: ~/.ssh/id_ed25519
ssh_jump_host: bastion@bastion.example.com
strict_host_key_checking: true

vault:
  path: secret/data/ssh/web01
  ssh_key: ssh_private_key

os_family: linux              # linux | aix | solaris. default: linux
sudo_required: true           # default: false
sudo_nopasswd: true           # default: false

tags: [web, linux]
group: web-fleet
```

---

## Multi-Endpoint Target Example

A single production Oracle deployment often requires credentials and connectivity for multiple subsystems. The recommended pattern is to define one target per endpoint type, then link them via reference fields.

### Primary database

```yaml
entity_name: prod-db
entity_type: oracle_database
description: Production Oracle 19c primary

host: db01.example.com
port: 1521
service: PRODDB
protocol: tcps
wallet_path: /etc/oracle/wallets/prod

user: dba_user
vault:
  path: secret/data/oracle/prod-db
  user_key: username
  password_key: password

oracle_edition: enterprise
oracle_options:
  - diagnostics_pack
  - tuning_pack
  - advanced_security

tags: [production, oracle, primary]
group: prod-fleet
timeout_seconds: 30
```

### Physical standby database

```yaml
entity_name: prod-db-standby
entity_type: oracle_database
description: Production Oracle 19c physical standby

host: db02.example.com
port: 1521
service: PRODDB_STB
protocol: tcps
wallet_path: /etc/oracle/wallets/prod

user: dba_user
vault:
  path: secret/data/oracle/prod-db-standby
  user_key: username
  password_key: password

oracle_edition: enterprise
oracle_options:
  - diagnostics_pack

tags: [production, oracle, standby]
group: prod-fleet
```

### ASM instance (primary node)

```yaml
entity_name: prod-asm
entity_type: oracle_asm
description: ASM on primary node

host: db01.example.com
port: 1521
service: +ASM

user: asmsnmp
vault:
  path: secret/data/oracle/prod-asm
  user_key: username
  password_key: password

ssh_target: prod-db01-ssh
tags: [production, asm]
group: prod-fleet
```

### SSH host — primary node

```yaml
entity_name: prod-db01-ssh
entity_type: oracle_host
description: SSH access to primary DB node

host: db01.example.com
port: 22
user: oracle
vault:
  path: secret/data/ssh/db01
  ssh_key: oracle_ssh_key

oracle_home: /u01/app/oracle/product/19.0.0/db_1
oracle_base: /u01/app/oracle
oracle_sid: PRODDB

tags: [production, ssh, oracle]
group: prod-fleet
```

### SSH host — standby node

```yaml
entity_name: prod-db02-ssh
entity_type: oracle_host
description: SSH access to standby DB node

host: db02.example.com
port: 22
user: oracle
vault:
  path: secret/data/ssh/db02
  ssh_key: oracle_ssh_key

oracle_home: /u01/app/oracle/product/19.0.0/db_1
oracle_base: /u01/app/oracle
oracle_sid: PRODDB_STB

tags: [production, ssh, oracle, standby]
group: prod-fleet
```

### Oracle Enterprise Manager (OEM) monitoring endpoint

```yaml
entity_name: prod-oem
entity_type: oracle_host
description: OEM management server

host: oem.example.com
port: 22
user: oracle
vault:
  path: secret/data/ssh/oem
  ssh_key: oem_ssh_key

tags: [production, oem, monitoring]
group: prod-fleet
```

### GoldenGate hub

```yaml
entity_name: prod-gg
entity_type: oracle_host
description: GoldenGate Microservices hub

host: gg.example.com
port: 22
user: oracle
vault:
  path: secret/data/ssh/gg
  ssh_key: gg_ssh_key

tags: [production, goldengate, replication]
group: prod-fleet
```

---

## Target Groups and Fleet Organization

Groups enable bulk operations and scoped queries across multiple targets.

### Assigning a group

Set the `group` field in any target file:

```yaml
group: prod-fleet
```

A target can belong to only one group. Use `tags` for multi-dimensional classification.

### Listing targets by group

```bash
dbxcli target list group=prod-fleet
```

### Running a command against all targets in a group

```bash
dbxcli db tablespace list group=prod-fleet
```

### Fleet-wide health check

```bash
dbxcli target test group=prod-fleet
```

Output includes per-target status, latency, and any connection errors.

### Recommended group naming

| Group name | Scope |
|------------|-------|
| `prod-fleet` | All production Oracle targets |
| `uat-fleet` | All UAT targets |
| `pg-fleet` | All PostgreSQL targets |
| `web-fleet` | All application servers |
| `monitoring-fleet` | Monitoring and observability hosts |

---

## Environment Variable Overrides

Target fields can be overridden at runtime via environment variables without modifying the YAML file. This is useful for CI pipelines and ephemeral environments.

### Per-target field syntax

```
DBX_TARGET_<ENTITY_NAME_UPPER>_<FIELD_UPPER>=value
```

Hyphens in entity names are converted to underscores for the variable name.

### Examples

```bash
# Override host for target prod-orcl
export DBX_TARGET_PROD_ORCL_HOST=db-failover.example.com

# Override port for target prod-pg
export DBX_TARGET_PROD_PG_PORT=5433

# Override vault path
export DBX_TARGET_PROD_ORCL_VAULT_PATH=secret/data/oracle/prod-orcl-dr
```

### Global defaults

These variables apply to all targets unless overridden at the target level:

```bash
DBX_DEFAULT_TIMEOUT=60          # global timeout in seconds
DBX_DEFAULT_FORMAT=json         # default output format
DBX_INSTANT_CLIENT_PATH=/opt/oracle/instantclient_21_12
DBX_VAULT_ADDR=https://vault.example.com
DBX_VAULT_ROLE_ID=<role-id>
DBX_VAULT_SECRET_ID=<secret-id>
```

### Active target shortcut

```bash
DBX_TARGET=prod-orcl dbxcli db tablespace list
```

Equivalent to `dbxcli db tablespace list -t prod-orcl`.

---

## SEE ALSO

- [Vault Integration](vault.md) — Storing and rotating credentials in HashiCorp Vault
- [Oracle License Declaration](oracle-license.md) — Declaring licensed editions and options per target
- [dbxcli target add](../cli/dbxcli_target_add.md) — Add a target from the CLI
- [dbxcli target list](../cli/dbxcli_target_list.md) — List registered targets
- [dbxcli target test](../cli/dbxcli_target_test.md) — Test target connectivity
