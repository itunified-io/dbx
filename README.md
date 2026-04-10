# dbx

Multi-database lifecycle management framework — CLI, REST API, MCP adapters, and Web UI for Oracle, PostgreSQL, Host/OS, and Engineered Systems.

## Architecture

```
Layer 0: dbx Go framework (dbxcli + dbxctl + REST API + Go library)
Layer 1: MCP adapters (TypeScript npm wrappers calling dbxcli)
Layer 2: AI skills (Claude Code skills, IDE integrations)
```

Layer 0 works standalone — MCP and AI are optional.

## Engines

| Engine | Tools | Status |
|--------|-------|--------|
| Oracle Database | 350+ | Active |
| PostgreSQL | 124 | Active |
| Host / OS | 80+ | Active |
| Engineered Systems | 120+ | Active |
| SQL Server | — | Roadmap |
| MySQL / MongoDB | — | Roadmap |

## Quick Start

```bash
# Build
make build

# Run CLI
./bin/dbxcli version
./bin/dbxcli target list

# Run service controller
./bin/dbxctl serve
```

## Interfaces

| Interface | Binary | Description |
|-----------|--------|-------------|
| CLI | `dbxcli` | Interactive terminal, scripts, cron |
| Service | `dbxctl` | Long-running daemon, health, scheduling |
| REST API | `dbxctl serve` | HTTP/JSON for integrations |
| MCP | Layer 1 adapters | AI model integration |
| Web UI | `dbxctl ui` | Browser-based operations |

## License

AGPL-3.0 — see [LICENSE](LICENSE).
