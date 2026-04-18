# Quick Start

Get up and running with dbx in 5 minutes.

## Install

### Homebrew (macOS/Linux)

```bash
brew tap itunified-io/dbx
brew install dbx
```

### Go Install

```bash
go install github.com/itunified-io/dbx/cmd/dbxcli@latest
go install github.com/itunified-io/dbx/cmd/dbxctl@latest
```

### Docker

```bash
docker pull ghcr.io/itunified-io/dbx:latest
docker run --rm ghcr.io/itunified-io/dbx:latest version
```

### Linux Packages

```bash
# DEB (Debian/Ubuntu)
wget https://github.com/itunified-io/dbx/releases/latest/download/dbx_linux_amd64.deb
sudo dpkg -i dbx_linux_amd64.deb

# RPM (RHEL/Fedora)
sudo rpm -i https://github.com/itunified-io/dbx/releases/latest/download/dbx_linux_amd64.rpm
```

## Verify

```bash
dbxcli version
dbxctl version
```

## Connect to a Database

### Oracle

```bash
dbxcli oracle connect --host 10.0.0.10 --port 1521 --sid ORCL --user system
dbxcli oracle sessions list
```

### PostgreSQL

```bash
dbxcli pg connect --host localhost --port 5432 --dbname mydb --user postgres
dbxcli pg sessions list
```

## Next Steps

- [CLI Setup Guide](cli-setup.md) -- targets, Vault integration, SSH keys
- [MCP Setup Guide](mcp-setup.md) -- AI-assisted DBA with IDE integration
- [Monitoring Setup](monitoring-setup.md) -- Docker/K8s deployment with Grafana
