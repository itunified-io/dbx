# Troubleshooting Guide

This guide covers the most common operational issues encountered with dbx, grouped by subsystem. Each issue includes the symptom, root cause, and resolution steps.

For issues not covered here, run the built-in diagnostics first:

```bash
dbxcli target test entity_name=<name>          # connectivity check
dbxcli license status                          # license and Vault check
dbxcli --log-level debug <command>             # verbose output
```

---

## Oracle

### ORA-12541: TNS: no listener

**Symptom:**
```
ORA-12541: TNS:no listener
Error: connection failed for target prod-orcl: dial tcp db01.example.com:1521: connect: connection refused
```

**Cause:** The Oracle Net listener is not running on the target host, or it is bound to a different port or interface.

**Resolution:**

1. Verify the listener is running on the host:
   ```bash
   dbxcli linux security service-status entity_name=db01-ssh service_name=oracle-listener
   # or directly via SSH:
   ssh oracle@db01.example.com "lsnrctl status"
   ```

2. If the listener is down, start it:
   ```bash
   ssh oracle@db01.example.com "lsnrctl start"
   ```

3. Confirm the port and interface from `listener.ora` on the host:
   ```bash
   ssh oracle@db01.example.com "cat \$ORACLE_HOME/network/admin/listener.ora"
   ```

4. Check that the firewall permits traffic from the dbx client to port 1521:
   ```bash
   dbxcli linux network dns-check entity_name=db01-ssh
   ```

5. Update the target if the port has changed:
   ```bash
   dbxcli target set prod-orcl port=1522
   ```

---

### ORA-12514: TNS: listener does not currently know of service requested in connect descriptor

**Symptom:**
```
ORA-12514: TNS:listener does not currently know of service requested in connect descriptor
```

**Cause:** The service name in the target does not match any service registered with the listener. This commonly occurs after a failover, PDB rename, or misconfigured target.

**Resolution:**

1. List services currently registered with the listener:
   ```bash
   ssh oracle@db01.example.com "lsnrctl services"
   ```

2. Find the correct service name:
   ```bash
   ssh oracle@db01.example.com "sqlplus -S / as sysdba <<< 'SELECT name, open_mode FROM v\$pdbs;'"
   ```

3. Update the target with the correct service name:
   ```bash
   dbxcli target set prod-orcl service=PRODDB_NEW
   ```

4. If using a CDB, confirm whether you are connecting to a PDB or the CDB root. The service name for a PDB is typically the PDB name; for the CDB root it is the `DB_UNIQUE_NAME` or the configured `SERVICE_NAMES` parameter.

---

### ORA-01017: invalid username/password; logon denied

**Symptom:**
```
ORA-01017: invalid username/password; logon denied
Error: authentication failed for target prod-orcl
```

**Cause:** The username or password stored in Vault (or the target file) does not match the database account, or the account is locked.

**Resolution:**

1. Verify the account status in the database:
   ```bash
   # Connect as SYSDBA from the host
   ssh oracle@db01.example.com "sqlplus -S / as sysdba <<< \"SELECT username, account_status FROM dba_users WHERE username = UPPER('dba_user');\""
   ```

2. If the account is locked (`LOCKED` or `EXPIRED & LOCKED`), unlock it:
   ```bash
   ssh oracle@db01.example.com "sqlplus -S / as sysdba <<< \"ALTER USER dba_user ACCOUNT UNLOCK;\""
   ```

3. Verify the credential in Vault matches the current database password:
   ```bash
   vault kv get secret/oracle/prod-orcl
   ```

4. Reset the Vault credential if it is out of sync:
   ```bash
   vault kv put secret/oracle/prod-orcl username="dba_user" password="<new-password>"
   # Then update the database to match:
   ssh oracle@db01.example.com "sqlplus -S / as sysdba <<< \"ALTER USER dba_user IDENTIFIED BY \\\"<new-password>\\\";\""
   ```

5. Check that the password does not contain characters that require escaping in the Oracle Net connect string (e.g., `@`, `/`, `"`). If so, use Oracle Wallet or Vault instead of plaintext passwords.

---

### TNS resolution failure (tnsnames.ora, LDAP, EZConnect)

**Symptom:**
```
ORA-12154: TNS: could not resolve the connect identifier specified
```
or
```
Error: could not locate Oracle client libraries — no tnsnames.ora or TNS_ADMIN found
```

**Cause:** The TNS alias specified in the target does not resolve, or `TNS_ADMIN` is not set to the directory containing `tnsnames.ora`.

**Resolution:**

1. If the target uses `tns_alias`, verify `TNS_ADMIN` is set correctly:
   ```bash
   echo $TNS_ADMIN
   ls $TNS_ADMIN/tnsnames.ora
   ```

2. Set `TNS_ADMIN` globally or per target:
   ```bash
   export TNS_ADMIN=/etc/oracle/network/admin
   # Or in ~/.dbx/config.yaml:
   # oracle:
   #   tns_admin: /etc/oracle/network/admin
   ```

3. Verify the alias exists in `tnsnames.ora`:
   ```bash
   grep -i "PRODDB" $TNS_ADMIN/tnsnames.ora
   ```

4. Switch to EZConnect format to bypass TNS resolution entirely:
   ```bash
   dbxcli target set prod-orcl tns_alias="" host=db01.example.com port=1521 service=PRODDB
   ```

---

### Oracle Instant Client setup issues

**Symptom:**
```
Error: Oracle client libraries not found. Set oracle_instant_client_path or DBX_INSTANT_CLIENT_PATH.
```
or
```
dyld: Library not loaded: /usr/lib/libocci.dylib
```

**Cause:** The Oracle Instant Client is not installed, is the wrong version, or the shared library path is not configured.

**Resolution:**

1. Download the appropriate Instant Client package for your platform from Oracle's website (Basic package is sufficient for dbx; Basic Lite is not supported).

2. Extract to a stable directory:
   ```bash
   mkdir -p /opt/oracle
   unzip instantclient-basic-macos.arm64-21.12.0.0.0dbru.zip -d /opt/oracle
   ```

3. On macOS, remove the quarantine attribute:
   ```bash
   xattr -d com.apple.quarantine /opt/oracle/instantclient_21_12/*.dylib 2>/dev/null || true
   ```

4. Configure dbx to use the Instant Client:
   ```bash
   export DBX_INSTANT_CLIENT_PATH=/opt/oracle/instantclient_21_12
   ```
   Or in `~/.dbx/config.yaml`:
   ```yaml
   oracle:
     instant_client_path: /opt/oracle/instantclient_21_12
   ```

5. On Linux, update the shared library cache:
   ```bash
   echo /opt/oracle/instantclient_21_12 > /etc/ld.so.conf.d/oracle-instantclient.conf
   ldconfig
   ```

6. Verify the client is detected:
   ```bash
   dbxcli target test entity_name=prod-orcl
   ```

---

### NLS_LANG encoding issues

**Symptom:** Query results contain garbled characters, or `ORA-12715: invalid character set specified` appears during connection.

**Cause:** The `NLS_LANG` environment variable does not match the database character set, causing incorrect transcoding.

**Resolution:**

1. Determine the database character set:
   ```bash
   dbxcli db sql exec entity_name=prod-orcl sql="SELECT value FROM nls_database_parameters WHERE parameter = 'NLS_CHARACTERSET'"
   ```

2. Set `NLS_LANG` to match (format: `LANGUAGE_TERRITORY.CHARACTERSET`):
   ```bash
   export NLS_LANG="AMERICAN_AMERICA.AL32UTF8"
   ```

3. Set per target to avoid global conflicts:
   ```yaml
   # In target YAML
   nls_lang: "AMERICAN_AMERICA.AL32UTF8"
   ```

4. For ASCII-only environments with a non-Unicode database, use the database's actual character set (e.g., `WE8MSWIN1252`) to suppress conversion overhead:
   ```bash
   export NLS_LANG="AMERICAN_AMERICA.WE8MSWIN1252"
   ```

---

## PostgreSQL

### Connection refused

**Symptom:**
```
Error: connection refused for target prod-pg: dial tcp pg01.example.com:5432: connect: connection refused
```

**Cause:** The PostgreSQL server is not running, is bound to a different interface, or the port is blocked.

**Resolution:**

1. Verify the server is running:
   ```bash
   dbxcli pg dba uptime entity_name=prod-pg
   # or via SSH on the host:
   ssh admin@pg01.example.com "systemctl status postgresql"
   ```

2. Check what address PostgreSQL is listening on (`postgresql.conf`):
   ```bash
   ssh admin@pg01.example.com "grep listen_addresses /etc/postgresql/16/main/postgresql.conf"
   ```
   A value of `localhost` means connections from remote hosts will fail. Change to `'*'` or the specific IP and reload:
   ```bash
   ssh admin@pg01.example.com "psql -c \"ALTER SYSTEM SET listen_addresses = '*';\" && systemctl reload postgresql"
   ```

3. Check that the firewall permits port 5432:
   ```bash
   dbxcli linux security firewall-list entity_name=pg01-ssh
   ```

---

### SSL certificate errors

**Symptom:**
```
Error: SSL connection failed: x509: certificate signed by unknown authority
```
or
```
Error: SSL connection failed: x509: certificate has expired or is not yet valid
```

**Cause:** The client does not trust the server's certificate, or the certificate has expired.

**Resolution:**

1. Check certificate expiry on the server:
   ```bash
   ssh admin@pg01.example.com "openssl x509 -noout -dates -in /etc/ssl/certs/pg-server.crt"
   ```

2. If expired, renew the certificate and restart PostgreSQL:
   ```bash
   # Renew using your CA or ACME provider, then:
   ssh admin@pg01.example.com "systemctl restart postgresql"
   ```

3. If using a private CA, add the CA certificate to the target:
   ```yaml
   sslrootcert: /etc/ssl/certs/pg-ca.crt
   sslmode: verify-full
   ```

4. For development environments where certificate validation is not required, downgrade `sslmode`:
   ```yaml
   sslmode: require    # encrypts but does not verify the certificate
   ```
   Do not use `sslmode: disable` in any environment that handles real data.

---

### pg_hba.conf denials

**Symptom:**
```
FATAL: no pg_hba.conf entry for host "10.0.1.50", user "dba_user", database "appdb", no encryption
```

**Cause:** The client IP address, user, or database combination does not match any rule in `pg_hba.conf`, or the authentication method is incompatible with the connection parameters.

**Resolution:**

1. Review the current rules on the server:
   ```bash
   dbxcli pg security pg-hba entity_name=prod-pg
   ```

2. Add a rule for the client:
   ```bash
   ssh admin@pg01.example.com "echo 'host appdb dba_user 10.0.1.0/24 scram-sha-256' >> /etc/postgresql/16/main/pg_hba.conf"
   ssh admin@pg01.example.com "psql -c 'SELECT pg_reload_conf();'"
   ```

3. Verify connectivity after reload:
   ```bash
   dbxcli target test entity_name=prod-pg
   ```

4. If the error says "no encryption" and you are connecting with `sslmode: require`, ensure the `hostssl` rule is present instead of (or in addition to) the `host` rule:
   ```
   hostssl appdb dba_user 10.0.1.0/24 scram-sha-256
   ```

---

### Password authentication failures

**Symptom:**
```
FATAL: password authentication failed for user "dba_user"
```

**Cause:** The password in the target or Vault does not match the current PostgreSQL role password.

**Resolution:**

1. Verify the credential stored in Vault:
   ```bash
   vault kv get secret/postgres/prod-pg
   ```

2. Test connectivity with an explicit password override:
   ```bash
   DBX_TARGET_PROD_PG_PASSWORD="<test-password>" dbxcli target test entity_name=prod-pg
   ```

3. Reset the role password if necessary:
   ```bash
   ssh admin@pg01.example.com "psql -c \"ALTER ROLE dba_user PASSWORD '<new-password>';\""
   ```

4. Update the Vault secret to match:
   ```bash
   vault kv put secret/postgres/prod-pg username="dba_user" password="<new-password>"
   ```

5. If password hashing in `pg_hba.conf` is set to `md5` but the server is configured for `scram-sha-256`, re-create the role password to force re-hashing:
   ```bash
   ssh admin@pg01.example.com "psql -c \"ALTER ROLE dba_user PASSWORD '<new-password>';\""
   ```

---

### CNPG kubectl access issues

**Symptom:**
```
Error: kubectl not found or not configured for CNPG operations
```
or
```
Error: cluster.postgresql.cnpg.io "prod-pg" not found in namespace default
```

**Cause:** The `kubectl` binary is not on the PATH, the kubeconfig context is not set, or the CNPG cluster name or namespace does not match the target.

**Resolution:**

1. Verify `kubectl` is available:
   ```bash
   kubectl version --client
   ```

2. Check the current kubeconfig context:
   ```bash
   kubectl config current-context
   ```

3. Confirm the CNPG cluster exists in the expected namespace:
   ```bash
   kubectl get cluster -n <namespace>
   ```

4. Update the target YAML with the correct cluster reference:
   ```yaml
   # In pg_cluster target YAML
   cnpg:
     cluster_name: prod-pg-cluster
     namespace: databases
     kubeconfig_context: prod-k8s
   ```

5. If running dbx from a host that does not have direct cluster access, ensure the kubeconfig merges the correct cluster context:
   ```bash
   export KUBECONFIG=~/.kube/config:~/.kube/prod-cluster.yaml
   kubectl config use-context prod-k8s
   ```

---

## SSH

### Key permissions (600/700)

**Symptom:**
```
Error: SSH key /home/user/.ssh/id_ed25519 has incorrect permissions (0644). Must be 0600 or stricter.
```

**Cause:** OpenSSH and dbx refuse to use private key files with world-readable or group-readable permissions.

**Resolution:**

```bash
chmod 600 ~/.ssh/id_ed25519
chmod 700 ~/.ssh
```

For keys retrieved from Vault and written to a temporary file, dbx sets `0600` automatically. If this error appears for a Vault-sourced key, it indicates that an external process modified the temporary file before dbx used it — report this as a security incident.

---

### Host key verification failure

**Symptom:**
```
Error: SSH host key verification failed for db01.example.com:
  REMOTE HOST IDENTIFICATION HAS CHANGED
  Offending key in /Users/user/.ssh/known_hosts:14
```

**Cause:** The server's host key has changed since the last connection. This can indicate a legitimate host rebuild or an active MITM attack.

**Resolution:**

1. If the server was legitimately rebuilt or its key was rotated, remove the stale entry:
   ```bash
   ssh-keygen -R db01.example.com
   ```

2. Re-connect to accept the new key:
   ```bash
   ssh oracle@db01.example.com "exit"
   ```

3. If you cannot physically verify the new key fingerprint with the server team, do not proceed. Treat unexplained key changes as a potential security event.

4. To automate known_hosts management for trusted internal hosts, use `ssh-keyscan` from a trusted network:
   ```bash
   ssh-keyscan -H db01.example.com >> ~/.ssh/known_hosts
   ```

Do not set `strict_host_key_checking: false` in production target files.

---

### SSH agent forwarding

**Symptom:** Tools that chain SSH connections (e.g., connecting to a bastion, then to a database host) fail to authenticate on the second hop.

**Resolution:**

1. Start `ssh-agent` and add your key:
   ```bash
   eval "$(ssh-agent -s)"
   ssh-add ~/.ssh/id_ed25519
   ```

2. Verify the agent has the key:
   ```bash
   ssh-add -l
   ```

3. Enable agent forwarding in the target:
   ```yaml
   # oracle_host target YAML
   ssh_forward_agent: true
   ```

4. Ensure the bastion server allows agent forwarding (`AllowAgentForwarding yes` in `/etc/ssh/sshd_config`).

---

### Jump host configuration

**Symptom:**
```
Error: jump host 'bastion.example.com' connection failed: dial tcp bastion.example.com:22: i/o timeout
```

**Cause:** The jump host is unreachable from the dbx client, or the jump host address is incorrect.

**Resolution:**

1. Test direct SSH to the jump host:
   ```bash
   ssh user@bastion.example.com "exit"
   ```

2. Update the jump host in the target:
   ```yaml
   ssh_jump_host: user@bastion.example.com:22
   ```

3. If the jump host itself requires a different key, set the key path in the target:
   ```yaml
   ssh_jump_key_path: ~/.ssh/bastion_ed25519
   ```

---

### Connection timeout

**Symptom:**
```
Error: SSH connection to db01.example.com:22 timed out after 30s
```

**Cause:** The host is unreachable (network issue, host down), or `timeout_seconds` is too short for a high-latency connection.

**Resolution:**

1. Increase `timeout_seconds` on the target:
   ```bash
   dbxcli target set db01-ssh timeout_seconds=60
   ```

2. Verify network reachability:
   ```bash
   dbxcli linux network dns-check entity_name=db01-ssh
   ```

3. Check whether SSH is listening on the expected port:
   ```bash
   # From another host on the same network:
   nc -zv db01.example.com 22
   ```

---

## License

### Activation failure

**Symptom:**
```
Error: license activation failed: invalid license key or key already in use on another machine
```

**Cause:** The license key is typed incorrectly, has been activated on the maximum number of machines, or has expired.

**Resolution:**

1. Re-enter the license key carefully:
   ```bash
   dbxcli license activate license_key=XXXX-XXXX-XXXX-XXXX
   ```

2. Check the current activation count in your account at `https://portal.example.com/licenses`.

3. Deactivate an existing machine if you have reached the seat limit, then retry activation.

4. For volume licenses (fleet activation), use the fleet key format:
   ```bash
   dbxcli license activate license_key=FLEET-XXXX-XXXX-XXXX
   ```

---

### Phone-home failures

**Symptom:**
```
WARNING: license phone-home check failed (last success: 2026-04-08T10:00:00Z). Grace period: 48h remaining.
```

**Cause:** The dbx process cannot reach the license validation endpoint over HTTPS. This may be due to network egress restrictions, a proxy, or a DNS issue.

**Resolution:**

1. Check outbound HTTPS connectivity from the dbx client:
   ```bash
   curl -v https://license.example.com/v1/ping
   ```

2. If a proxy is required, set the standard proxy environment variables:
   ```bash
   export HTTPS_PROXY=http://proxy.example.com:8080
   export NO_PROXY=10.0.0.0/8,*.example.com
   ```

3. Verify DNS resolution for the license endpoint:
   ```bash
   dig +short license.example.com
   ```

4. For air-gapped environments, request an offline license token from support. See the offline licensing section below.

---

### Grace period behavior

When phone-home fails, dbx enters a grace period. All licensed tools remain available during the grace period.

| Grace period status | Tools available | Action required |
|---------------------|-----------------|-----------------|
| Within grace period | Yes | Restore connectivity or obtain offline token |
| Grace period expired | No | License enforcement blocks execution |

Check the current grace period status:

```bash
dbxcli license status
```

The default grace period is 72 hours. This is not configurable for Standard plans. Enterprise plans can negotiate extended grace periods.

---

### Offline licensing

For environments with no outbound internet access:

1. Run the offline activation command to generate a machine fingerprint:
   ```bash
   dbxcli license offline-fingerprint
   ```
   Output: a `DBX-MACHINE-ID: <base64>` string.

2. Provide the machine ID to your account manager or the support portal to obtain an offline token.

3. Apply the offline token:
   ```bash
   dbxcli license offline-activate token=<base64-token>
   ```

4. The offline token encodes the license terms and a validity window. Renew it before expiry using the same fingerprint/token exchange process.

---

### Bundle mismatch

**Symptom:**
```
Error: tool 'pg cnpg cluster-status' is not available in your current plan (Free).
Upgrade to Standard or Enterprise to access CNPG tools.
```

**Cause:** The tool belongs to a plan tier above the activated license.

**Resolution:**

1. Check your current plan and what bundles are included:
   ```bash
   dbxcli license status
   ```

2. Review the tool-to-plan mapping in the [CLI Reference](cli/dbxcli.md).

3. To upgrade, contact your account manager or visit the portal.

---

## Monitoring

### Agent not connecting to central

**Symptom:** The monitoring dashboard shows a target as `offline` or `no data` even though the database is running and accessible via `dbxcli target test`.

**Cause:** The host monitoring agent cannot reach the central collection endpoint, or the agent is not running.

**Resolution:**

1. Check agent status on the host:
   ```bash
   dbxcli linux security service-status entity_name=<host-target> service_name=dbx-agent
   ```

2. Restart the agent if it is stopped:
   ```bash
   ssh admin@monitored-host.example.com "systemctl restart dbx-agent"
   ```

3. Verify the agent's configured central endpoint (typically in `/etc/dbx-agent/config.yaml` on the host):
   ```bash
   ssh admin@monitored-host.example.com "cat /etc/dbx-agent/config.yaml"
   ```

4. Test connectivity from the host to the central endpoint:
   ```bash
   ssh admin@monitored-host.example.com "curl -sv https://dbx-central.example.com/api/v1/health"
   ```

---

### Metric gaps

**Symptom:** Time-series graphs show gaps in data, but the agent is running.

**Cause:** The agent collection interval is too long for the graph resolution, the agent process was briefly stopped, or the target database was unavailable during the gap.

**Resolution:**

1. Check the agent collection log for errors during the gap time window:
   ```bash
   ssh admin@monitored-host.example.com "journalctl -u dbx-agent --since '1 hour ago'"
   ```

2. Confirm the collection interval matches the expected resolution:
   ```bash
   ssh admin@monitored-host.example.com "grep collect_interval /etc/dbx-agent/config.yaml"
   ```

3. For Oracle, ensure the AWR snapshot interval is compatible with the dbx collection schedule. A 60-minute AWR interval combined with a 5-minute dbx interval will result in gaps in AWR-sourced metrics.

---

### VictoriaMetrics write failures

**Symptom:**
```
ERROR [agent] failed to push metrics to VictoriaMetrics: connection refused at http://victoria:8428/api/v1/import/prometheus
```

**Cause:** VictoriaMetrics is down, the endpoint URL is wrong, or the network path between the agent and VictoriaMetrics is blocked.

**Resolution:**

1. Check VictoriaMetrics health:
   ```bash
   curl -s http://victoria.example.com:8428/health
   ```

2. Update the VictoriaMetrics endpoint in the agent configuration:
   ```bash
   ssh admin@monitored-host.example.com "sed -i 's|victoria_url:.*|victoria_url: http://victoria.example.com:8428|' /etc/dbx-agent/config.yaml"
   ssh admin@monitored-host.example.com "systemctl restart dbx-agent"
   ```

3. Verify the network path allows outbound traffic on port 8428 from the monitored host.

---

## MCP

### Claude Desktop not finding server

**Symptom:** Claude Desktop shows "MCP server not found" or the dbx tools do not appear in the tool list.

**Cause:** The MCP server entry is missing from `claude_desktop_config.json`, the `npx` command path is wrong, or the server process fails to start.

**Resolution:**

1. Locate and open the Claude Desktop configuration file:
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`

2. Verify the server entry is present and correctly formatted:
   ```json
   {
     "mcpServers": {
       "mcp-oracle": {
         "command": "npx",
         "args": ["-y", "@itunified.io/mcp-oracle"],
         "env": {
           "DBX_TARGET": "prod-orcl",
           "VAULT_ADDR": "https://vault.example.com",
           "DBX_VAULT_ROLE_ID": "<role-id>",
           "DBX_VAULT_SECRET_ID": "<secret-id>"
         }
       }
     }
   }
   ```

3. Test the server startup manually from a terminal:
   ```bash
   npx -y @itunified.io/mcp-oracle
   ```
   If the command fails, the error message will indicate whether it is a missing dependency, a missing environment variable, or a credential failure.

4. Restart Claude Desktop after modifying the configuration.

---

### stdio transport issues

**Symptom:** The MCP server starts but Claude immediately disconnects, or tool calls return `transport error`.

**Cause:** The server process is printing non-JSON output to stdout (which corrupts the stdio protocol), or it is exiting immediately due to a startup error.

**Resolution:**

1. Check that all log output from the MCP server goes to stderr, not stdout. Stdout is reserved for the MCP protocol.

2. Run the server in standalone mode and check stderr:
   ```bash
   npx -y @itunified.io/mcp-oracle 2>&1 | head -50
   ```

3. If the server exits immediately, a required environment variable is likely missing. Common required variables:
   ```bash
   VAULT_ADDR
   DBX_VAULT_ROLE_ID
   DBX_VAULT_SECRET_ID
   DBX_TARGET
   ```

4. Set all required variables in the `env` block of the Claude Desktop config rather than relying on shell environment inheritance. Claude Desktop does not inherit the shell environment on macOS.

---

### Environment variables not loaded

**Symptom:** The MCP server cannot find Vault credentials or the active target, even though the variables are set in the shell.

**Cause:** Claude Desktop (and other GUI launchers on macOS) do not inherit environment variables from the login shell. Variables set in `.zshrc` or `.bashrc` are not available to GUI-launched processes.

**Resolution:**

1. Move all required variables to the `env` block in `claude_desktop_config.json` (see MCP setup above).

2. Alternatively, use a wrapper script that sources the variables:
   ```bash
   #!/usr/bin/env bash
   # ~/.local/bin/mcp-oracle-wrapper.sh
   source ~/.secrets/.env
   exec npx -y @itunified.io/mcp-oracle "$@"
   ```
   Reference the wrapper in the config:
   ```json
   {
     "command": "/Users/<user>/.local/bin/mcp-oracle-wrapper.sh",
     "args": []
   }
   ```

3. Verify the variables are present inside the running MCP process by adding a temporary diagnostic tool call:
   ```
   # In Claude: ask the MCP server to report its environment
   # (only works if the server exposes a diagnostics tool)
   ```

---

## SEE ALSO

- [Target YAML Reference](config/targets.md) — Full target schema and field descriptions
- [Vault Integration](config/vault.md) — Vault setup, AppRole, and offline fallback
- [dbxcli target test](cli/dbxcli_target_test.md) — Connectivity testing
- [dbxcli license status](cli/dbxcli_license_status.md) — License and phone-home status
