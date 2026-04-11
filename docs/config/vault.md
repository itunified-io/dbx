# Vault Integration Reference

dbx integrates with HashiCorp Vault to retrieve database credentials, SSH private keys, and service tokens at runtime. No credentials need to be stored in target YAML files or environment variables on the filesystem.

Vault is the recommended credential source for all production targets. Plaintext `password` fields in target files are supported for development and testing only.

---

## Vault Path Layout

All dbx secrets are stored under the KV v2 secrets engine. The default mount is `secret/`. Adjust the mount path if your Vault uses a non-default mount point.

### Oracle database credentials

```
secret/data/oracle/<target-name>
```

Required keys:

| Key | Description |
|-----|-------------|
| `username` | Database username |
| `password` | Database password |

Optional keys:

| Key | Description |
|-----|-------------|
| `sysdba_password` | Password for SYSDBA connections when different from the standard user |
| `wallet_password` | Oracle Wallet password for `tcps` connections |
| `asmsnmp_password` | ASM monitoring password |

Example write:

```bash
vault kv put secret/oracle/prod-orcl \
  username="dba_user" \
  password="<redacted>"
```

### PostgreSQL credentials

```
secret/data/postgres/<target-name>
```

Required keys:

| Key | Description |
|-----|-------------|
| `username` | PostgreSQL role name |
| `password` | PostgreSQL password |

Optional keys:

| Key | Description |
|-----|-------------|
| `ssl_cert` | PEM-encoded client certificate |
| `ssl_key` | PEM-encoded client private key |
| `ssl_rootcert` | PEM-encoded CA certificate |
| `replication_password` | Password for replication slot operations |

Example write:

```bash
vault kv put secret/postgres/prod-pg \
  username="dba_user" \
  password="<redacted>"
```

### SSH credentials

```
secret/data/ssh/<target-name>
```

Keys:

| Key | Description |
|-----|-------------|
| `oracle_ssh_key` | PEM-encoded Ed25519 or RSA private key for the `oracle` OS user |
| `root_ssh_key` | PEM-encoded private key for root access |
| `<consumer>_ssh_key` | Private key for a named consumer (e.g., `backup_ssh_key`, `gg_ssh_key`) |

Example write:

```bash
vault kv put secret/ssh/db01 \
  oracle_ssh_key="$(cat ~/.ssh/oracle_ed25519)"
```

### Service tokens and API credentials

```
secret/data/monitoring/<target-name>
secret/data/goldengate/<target-name>
secret/data/oem/<target-name>
```

Structure is free-form for service-specific credentials. Reference the key names in the target YAML `vault` block.

---

## AppRole Setup

AppRole is the recommended Vault auth method for automated (non-interactive) dbx operation. It issues short-lived tokens scoped to the minimum required policies.

### 1. Enable AppRole auth

```bash
vault auth enable approle
```

### 2. Write a policy

The policy grants read access to all dbx secret paths and the ability to renew tokens.

```hcl
# dbx-policy.hcl
path "secret/data/oracle/*" {
  capabilities = ["read"]
}

path "secret/data/postgres/*" {
  capabilities = ["read"]
}

path "secret/data/ssh/*" {
  capabilities = ["read"]
}

path "secret/data/monitoring/*" {
  capabilities = ["read"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}
```

Write the policy:

```bash
vault policy write dbx-policy dbx-policy.hcl
```

### 3. Create the AppRole

```bash
vault write auth/approle/role/dbx \
  token_policies="dbx-policy" \
  token_ttl=1h \
  token_max_ttl=4h \
  secret_id_ttl=0 \
  secret_id_num_uses=0
```

| Parameter | Value | Notes |
|-----------|-------|-------|
| `token_ttl` | `1h` | Token lifetime. dbx renews automatically before expiry. |
| `token_max_ttl` | `4h` | Maximum token lifetime before forced re-auth. |
| `secret_id_ttl` | `0` | No expiry on the Secret ID. Set to a positive duration in high-security environments. |
| `secret_id_num_uses` | `0` | Unlimited uses. Set to `1` for one-time bootstrap flows. |

### 4. Retrieve Role ID

```bash
vault read auth/approle/role/dbx/role-id
```

Store the `role_id` value.

### 5. Generate Secret ID

```bash
vault write -f auth/approle/role/dbx/secret-id
```

Store the `secret_id` value. Treat it as a password — it is the credential that authenticates dbx to Vault.

### 6. Configure dbx

Set the Vault address and AppRole credentials. The recommended approach is the global environment:

```bash
export VAULT_ADDR=https://vault.example.com
export DBX_VAULT_ROLE_ID=<role-id>
export DBX_VAULT_SECRET_ID=<secret-id>
```

Or write to `~/.dbx/config.yaml`:

```yaml
vault:
  addr: https://vault.example.com
  auth_method: approle
  role_id: "<role-id>"
  secret_id_env: DBX_VAULT_SECRET_ID   # read secret-id from this env var at runtime
  namespace: ""                         # Vault Enterprise namespace, if applicable
  tls_skip_verify: false
  timeout_seconds: 10
```

`secret_id` must not be written to the config file. Use `secret_id_env` to reference an environment variable, or use `secret_id_file` to reference a file path that contains only the secret ID.

---

## Per-Target Credential Configuration

Reference a Vault path from any target YAML file using the `vault` block:

```yaml
entity_name: prod-orcl
entity_type: oracle_database
host: db01.example.com
port: 1521
service: ORCL

vault:
  path: secret/data/oracle/prod-orcl
  user_key: username          # default: username
  password_key: password      # default: password
  mount: secret               # default: secret
```

If `user_key` and `password_key` are omitted, dbx uses `username` and `password` as defaults.

To use a different Vault namespace per target (Vault Enterprise):

```yaml
vault:
  path: secret/data/oracle/prod-orcl
  namespace: prod/oracle
```

---

## SSH Key Storage and Retrieval

SSH private keys are stored as multi-line string values in Vault KV. dbx writes the key to a temporary file with mode `0600` before establishing the SSH connection, then removes it after the session ends.

### Write a key

```bash
vault kv put secret/ssh/db01 \
  oracle_ssh_key="$(cat /path/to/oracle_ed25519)"
```

The key must be PEM-encoded. Ed25519 keys are recommended for new deployments. RSA 4096-bit keys are supported for compatibility with legacy systems.

### Reference in target YAML

```yaml
entity_name: db01-ssh
entity_type: oracle_host
host: db01.example.com
user: oracle

vault:
  path: secret/data/ssh/db01
  ssh_key: oracle_ssh_key
```

The `ssh_key` field names the KV key within the path that contains the PEM private key.

---

## Dynamic Secrets and Lease Management

When Vault is configured with the PostgreSQL database secrets engine, dbx can request short-lived credentials instead of static usernames and passwords.

### Configure the database secrets engine

```bash
vault secrets enable database

vault write database/config/prod-pg \
  plugin_name=postgresql-database-plugin \
  allowed_roles="dbx-role" \
  connection_url="postgresql://{{username}}:{{password}}@pg01.example.com:5432/postgres" \
  username="vault_admin" \
  password="<redacted>"

vault write database/roles/dbx-role \
  db_name=prod-pg \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}' INHERIT; GRANT dba_group TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="4h"
```

### Reference dynamic credentials in a target

```yaml
entity_name: prod-pg
entity_type: pg_database
host: pg01.example.com
port: 5432
database: appdb

vault:
  path: database/creds/dbx-role
  user_key: username
  password_key: password
  dynamic: true               # instructs dbx to treat the path as a dynamic creds path
```

When `dynamic: true`, dbx reads a fresh credential pair on each connection and honors the Vault lease TTL. Credentials are renewed automatically if the session exceeds the TTL. On renewal failure, the session is terminated and a new credential is requested.

### Lease behavior

| Event | dbx behavior |
|-------|-------------|
| Lease at 75% of TTL | Attempt renewal |
| Renewal succeeds | Lease extended; session continues |
| Renewal fails | Force reconnect with a new credential |
| Lease expires with no renewal | Session terminated; error returned to caller |

---

## Credential Rotation

### Manual rotation

Rotate a static secret and update the Vault path in one step:

```bash
vault kv put secret/oracle/prod-orcl \
  username="dba_user" \
  password="<new-password>"
```

dbx reads credentials on each new connection. Running connections are not interrupted; the new password takes effect on the next connect.

### Automated rotation with Vault

Vault can rotate static database credentials on a schedule using the database secrets engine static roles:

```bash
vault write database/static-roles/prod-orcl-dba \
  db_name=prod-orcl-oracle \
  username="dba_user" \
  rotation_period=86400
```

dbx fetches the current credential from `database/static-creds/prod-orcl-dba` on each connection. No manual intervention is required after rotation.

Reference in target YAML:

```yaml
vault:
  path: database/static-creds/prod-orcl-dba
  user_key: username
  password_key: password
  dynamic: true
```

---

## Offline Fallback

When Vault is unreachable, dbx can fall back to locally cached credentials for a configurable grace period.

### Configuration

```yaml
# ~/.dbx/config.yaml
vault:
  offline_fallback:
    enabled: true
    grace_period_minutes: 60    # default: 30. maximum: 480
    cache_path: ~/.dbx/.credential-cache
    cache_encryption_key_env: DBX_CACHE_KEY  # AES-256 key as hex string
```

The credential cache is AES-256-GCM encrypted at rest. The encryption key must be provided via environment variable; it is never stored on disk by dbx.

### Fallback behavior

| Condition | dbx behavior |
|-----------|-------------|
| Vault reachable | Always fetch live credential; do not use cache |
| Vault unreachable, within grace period | Use cached credential; log warning |
| Vault unreachable, grace period expired | Refuse to connect; return error |
| Vault unreachable, credential never cached | Refuse to connect; return error |

### Priming the cache

```bash
dbxcli target test entity_name=prod-orcl
```

A successful connection primes the credential cache for all targets that were accessed. Run this from a cron job or CI step to keep the cache fresh.

---

## Vault HA Considerations

### Active/Standby clusters

Point `VAULT_ADDR` at a load balancer VIP or DNS name that resolves to the active Vault node. Vault Raft HA and Consul HA both support this pattern.

```bash
export VAULT_ADDR=https://vault.example.com   # resolves to active node
```

### Multiple Vault addresses with failover

```yaml
# ~/.dbx/config.yaml
vault:
  addr: https://vault-primary.example.com
  fallback_addrs:
    - https://vault-secondary.example.com
  failover_timeout_seconds: 5
```

dbx attempts the primary address first. On timeout or connection error, it retries each fallback address in order.

### Namespace isolation (Vault Enterprise)

Different target classes can use separate Vault namespaces:

```yaml
# In target YAML
vault:
  path: secret/data/oracle/prod-orcl
  namespace: prod/oracle
```

Namespace is appended to the token request header for all operations against this target.

### TLS verification

Vault TLS is verified by default. To use a private CA:

```bash
export VAULT_CACERT=/etc/ssl/certs/vault-ca.crt
```

Or in `~/.dbx/config.yaml`:

```yaml
vault:
  tls_ca_cert: /etc/ssl/certs/vault-ca.crt
  tls_skip_verify: false       # never set true in production
```

---

## SEE ALSO

- [Target YAML Reference](targets.md) — Full target schema including `vault` block
- [Oracle License Declaration](oracle-license.md) — Per-target license declarations
- [dbxcli pg vault status](../cli/dbxcli_pg_vault_status.md) — Check Vault connectivity and credential lease status
- [dbxcli pg vault rotate](../cli/dbxcli_pg_vault_rotate.md) — Trigger credential rotation via Vault
