# dbx Quick Start

Get from zero to your first database query in under 5 minutes.

dbx is a multi-database management platform built around `dbxcli`, a Go-based CLI that speaks to Oracle, PostgreSQL, and other database engines through a unified command interface. MCP adapters expose the same capabilities to AI assistants such as Claude.

---

## 1. Install

### macOS (Homebrew)

```bash
brew tap itunified-io/dbx
brew install dbx
```

### Linux (install script)

```bash
curl -fsSL https://get.itunified.de/dbx | sh
```

The script detects your architecture, downloads the appropriate binary, and places it in `/usr/local/bin`.

### Docker (no local install)

```bash
docker run --rm -it ghcr.io/itunified-io/dbx:latest dbxcli version
```

Use the Docker image to evaluate dbx or run it in CI pipelines without installing the binary on the host.

### npm — MCP Adapters Only

The npm packages expose MCP tool interfaces for use with AI assistants. They do not include the `dbxcli` binary.

```bash
# Oracle MCP adapter (28 tools)
npm install -g @itunified.io/mcp-oracle

# PostgreSQL MCP adapter (27 tools)
npm install -g @itunified.io/mcp-postgres
```

---

## 2. Verify Installation

Confirm the binary is on your PATH and the version is correct:

```bash
dbxcli version
```

Expected output:

```
dbxcli v2026.04.11.1
Build: go1.22.3 darwin/arm64
Commit: a1b2c3d
```

List available command groups:

```bash
dbxcli --help
```

```
Usage:
  dbxcli [command]

Available Commands:
  target      Manage database connection targets
  db          Oracle database operations
  pg          PostgreSQL database operations
  host        Host and OS-level diagnostics
  config      Configuration and credential management
  version     Print version information

Flags:
  -h, --help       Help for dbxcli
  -o, --output     Output format: table, json, yaml (default: table)
  -t, --target     Target entity name (overrides default)

Use "dbxcli [command] --help" for more information about a command.
```

---

## 3. Connect to Oracle

### Add a target

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

`--ask-password` prompts for the password interactively and stores it in the local credential store. The password is never written to shell history.

### Test the connection

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

---

## 4. Connect to PostgreSQL

### Add a target

```bash
dbxcli target add \
  entity_name=prod-pg \
  entity_type=pg_database \
  host=pg01.example.com \
  port=5432 \
  database=appdb \
  user=dba_user \
  --ask-password
```

### Test the connection

```bash
dbxcli target test entity_name=prod-pg
```

Expected output:

```
Target:  prod-pg
Type:    pg_database
Host:    pg01.example.com:5432/appdb
Status:  connected
Version: PostgreSQL 16.2 on aarch64-unknown-linux-gnu
Latency: 2ms
```

---

## 5. First Commands

### Oracle

List active sessions on the database:

```bash
dbxcli db session list
```

List all tablespaces with size and free space:

```bash
dbxcli db tablespace list
```

Display OS and host information for the database server:

```bash
dbxcli host info
```

### PostgreSQL

Show connection pool status and active connections:

```bash
dbxcli pg connection status
```

List all tables in the current schema with row estimates and sizes:

```bash
dbxcli pg table list
```

Report database sizes across all databases in the cluster:

```bash
dbxcli pg database size
```

All commands accept `-t <entity_name>` to target a specific connection when you have multiple targets configured:

```bash
dbxcli db session list -t prod-orcl
dbxcli pg table list -t prod-pg
```

---

## 6. MCP Setup for AI Assistants

The MCP adapters allow Claude and other AI assistants to execute dbx operations through natural language. The MCP server connects to your configured targets using the same credential store as `dbxcli`.

### Claude Code

Register the Oracle MCP server:

```bash
claude mcp add oracle -- npx -y @itunified.io/mcp-oracle
```

Register the PostgreSQL MCP server:

```bash
claude mcp add postgres -- npx -y @itunified.io/mcp-postgres
```

Verify the servers are registered:

```bash
claude mcp list
```

### Claude Desktop

Add the following entries to your `claude_desktop_config.json` (typically located at `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "mcp-oracle": {
      "command": "npx",
      "args": ["-y", "@itunified.io/mcp-oracle"],
      "env": {
        "DBX_TARGET": "prod-orcl"
      }
    },
    "mcp-postgres": {
      "command": "npx",
      "args": ["-y", "@itunified.io/mcp-postgres"],
      "env": {
        "DBX_TARGET": "prod-pg"
      }
    }
  }
}
```

Restart Claude Desktop after saving the file. The MCP tools will be available in your next conversation.

---

## 7. What's Next

### Installation Guides

- [Oracle Database Setup](install/oracle-setup.md) — Oracle client libraries, TNS configuration, wallet-based authentication, and multi-tenant (CDB/PDB) support
- [PostgreSQL Setup](install/postgres-setup.md) — SSL certificates, pg_hba configuration, and connection pooling with PgBouncer
- [Host Agent Setup](install/host-setup.md) — SSH-based host diagnostics and OS-level metrics collection

### Configuration

- [Vault Integration](config/vault.md) — Store and rotate database credentials in HashiCorp Vault instead of the local credential store

### Administration

- [Administration Guide](admin/administration-guide.md) — Target management, role-based access, audit logging, and fleet-scale operations
