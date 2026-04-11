# Air-Gap / Offline Installation Guide

**Audience:** DBA / Infrastructure Engineer in regulated or isolated environments  
**Estimated setup time:** ~45 minutes  
**Applies to:** All dbx components and MCP adapters

---

## Table of Contents

1. [Overview](#overview)
2. [Air-Gap Bundle Creation (Connected Environment)](#air-gap-bundle-creation-connected-environment)
3. [Air-Gap Deployment (Isolated Network)](#air-gap-deployment-isolated-network)
4. [License Offline Activation](#license-offline-activation)
5. [Vault Initialization Without Internet](#vault-initialization-without-internet)
6. [Update Procedure for Air-Gap Environments](#update-procedure-for-air-gap-environments)

---

## Overview

Air-gap installation is for environments where the target infrastructure has no internet access. The process consists of two phases:

1. **Bundle creation** — performed on a connected machine (jump host, build workstation, or CI pipeline)
2. **Deployment** — performed on or from the isolated network using only the transferred bundle

The bundle contains:
- `dbxcli` binary (Linux amd64 and/or arm64)
- Docker images (saved as `.tar.gz`)
- npm packages (packed as `.tgz`)
- Pre-activated license file
- Checksums for integrity verification

No outbound connections are made during or after deployment in air-gap mode. Vault is initialized in server mode without any network dependencies.

---

## Air-Gap Bundle Creation (Connected Environment)

Perform these steps on a machine with internet access. The output is a single directory that is transferred to the isolated environment.

### Step 1: Create the bundle directory

```bash
mkdir -p ~/dbx-airgap-bundle/{binaries,images,npm,license,certs,docs}
cd ~/dbx-airgap-bundle
```

### Step 2: Download the dbxcli binary

```bash
# Linux amd64
curl -sSL https://releases.example.com/dbx/cli/latest/dbxcli-linux-amd64 \
  -o binaries/dbxcli-linux-amd64
chmod +x binaries/dbxcli-linux-amd64

# Linux arm64 (if required for target hosts)
curl -sSL https://releases.example.com/dbx/cli/latest/dbxcli-linux-arm64 \
  -o binaries/dbxcli-linux-arm64
chmod +x binaries/dbxcli-linux-arm64
```

### Step 3: Pull and save Docker images

Pull all images required for the monitoring stack, then save them as tarballs:

```bash
# Pull images
docker pull ghcr.io/itunified.io/dbmon-central:latest
docker pull victoriametrics/victoria-metrics:latest
docker pull grafana/grafana:latest
docker pull prom/alertmanager:latest
docker pull hashicorp/vault:latest

# Save images to tarballs
docker save ghcr.io/itunified.io/dbmon-central:latest \
  | gzip > images/dbmon-central.tar.gz

docker save victoriametrics/victoria-metrics:latest \
  | gzip > images/victoriametrics.tar.gz

docker save grafana/grafana:latest \
  | gzip > images/grafana.tar.gz

docker save prom/alertmanager:latest \
  | gzip > images/alertmanager.tar.gz

docker save hashicorp/vault:latest \
  | gzip > images/vault.tar.gz
```

### Step 4: Pack npm packages

Pack each required MCP adapter as a self-contained npm tarball:

```bash
# Example: pack the Oracle engine packages
npm pack @itunified.io/mcp-oracle --pack-destination npm/
npm pack @itunified.io/mcp-oracle-full --pack-destination npm/

# Example: pack the PostgreSQL packages
npm pack @itunified.io/mcp-postgres --pack-destination npm/
npm pack @itunified.io/mcp-postgres-enterprise --pack-destination npm/

# Example: pack the host and monitoring packages
npm pack @itunified.io/mcp-host --pack-destination npm/
npm pack @itunified.io/mcp-host-enterprise --pack-destination npm/
npm pack @itunified.io/mcp-dbmonitor --pack-destination npm/
npm pack @itunified.io/mcp-dbmonitor-ee --pack-destination npm/
```

Each pack command produces a `.tgz` file in `npm/`. The tarballs contain the package and all production dependencies, so no npm registry access is needed during installation.

### Step 5: Pre-activate the license

Pre-activation binds the license to the target environment's hardware fingerprint. You will need the hardware fingerprint from the isolated machine before building the bundle.

Obtain the fingerprint from the target machine (can be done without internet):

```bash
# On the target host (isolated environment)
dbxcli license fingerprint
# Output: hw-fingerprint: a3b7c912e4f08d1b (example)
```

Activate the license against the fingerprint using the dbx license server (from the connected machine):

```bash
dbxcli license activate \
  license_key=<your-license-key> \
  fingerprint=a3b7c912e4f08d1b \
  output=license/dbx-license.json \
  mode=offline
```

The output file `license/dbx-license.json` is a signed offline license token valid for the declared fingerprint and expiry period.

### Step 6: Generate checksums

Generate SHA-256 checksums for all bundle files to allow integrity verification on the isolated side:

```bash
cd ~/dbx-airgap-bundle
find . -type f ! -name checksums.sha256 \
  -exec sha256sum {} \; > checksums.sha256

echo "Bundle checksums written to checksums.sha256"
```

### Step 7: Transfer the bundle

Transfer the bundle to the isolated environment using your approved media transfer procedure (removable media, secure file transfer, etc.):

```
~/dbx-airgap-bundle/
  binaries/
    dbxcli-linux-amd64
    dbxcli-linux-arm64
  images/
    dbmon-central.tar.gz
    victoriametrics.tar.gz
    grafana.tar.gz
    alertmanager.tar.gz
    vault.tar.gz
  npm/
    itunified.io-mcp-oracle-<version>.tgz
    itunified.io-mcp-postgres-<version>.tgz
    ...
  license/
    dbx-license.json
  certs/
    (optional: CA certificates for internal PKI)
  checksums.sha256
```

---

## Air-Gap Deployment (Isolated Network)

Perform these steps on or from within the isolated network. No outbound connections are made.

### Step 1: Verify bundle integrity

Before proceeding, verify all bundle files match their checksums:

```bash
cd /path/to/dbx-airgap-bundle
sha256sum --check checksums.sha256
```

All lines must show `OK`. If any file fails, discard the bundle and request a fresh transfer.

### Step 2: Install the dbxcli binary

```bash
install -o root -g root -m 755 \
  binaries/dbxcli-linux-amd64 \
  /usr/local/bin/dbxcli

# Verify
dbxcli version
```

### Step 3: Load Docker images

```bash
# Load all images from the bundle
for f in images/*.tar.gz; do
  echo "Loading $f..."
  gunzip -c "$f" | docker load
done

# Confirm images are available
docker images | grep -E 'dbmon-central|victoriametrics|grafana|alertmanager|vault'
```

### Step 4: Install npm packages locally

npm packages from the bundle are installed using the local path — no registry access is required:

```bash
# Install globally from the local tarballs
npm install -g npm/itunified.io-mcp-oracle-*.tgz
npm install -g npm/itunified.io-mcp-postgres-*.tgz
npm install -g npm/itunified.io-mcp-host-*.tgz
npm install -g npm/itunified.io-mcp-dbmonitor-*.tgz

# Verify installation
npx @itunified.io/mcp-oracle --version
```

### Step 5: Install the offline license

```bash
mkdir -p ~/.dbx/license
cp license/dbx-license.json ~/.dbx/license/dbx-license.json
chmod 600 ~/.dbx/license/dbx-license.json

# Register the license with dbxcli
dbxcli license install path=~/.dbx/license/dbx-license.json
```

### Step 6: Deploy the monitoring stack

Use the Docker Compose configuration from the air-gap bundle. All images reference the locally-loaded tags — no image pulls will occur:

```bash
mkdir -p ~/dbx-monitoring
cp -r /path/to/dbx-airgap-bundle/compose/* ~/dbx-monitoring/

cd ~/dbx-monitoring
docker compose up -d
```

For a minimal air-gap deployment (no internet-hosted images needed), pin all image references in `docker-compose.yml` to the exact digest loaded during Step 3:

```yaml
services:
  dbmon-central:
    image: ghcr.io/itunified.io/dbmon-central@sha256:<digest>
```

---

## License Offline Activation

If the license was not pre-activated before transfer, use the manual activation procedure.

### Manual activation flow

**Step 1:** Generate the hardware fingerprint on the isolated host:

```bash
dbxcli license fingerprint
```

Record the fingerprint value (e.g., `hw-fingerprint: a3b7c912e4f08d1b`).

**Step 2:** On a connected machine (outside the isolated network), generate an offline activation token:

```bash
dbxcli license generate-token \
  license_key=<your-license-key> \
  fingerprint=a3b7c912e4f08d1b \
  expiry_days=365
```

This outputs a base64-encoded activation token. The token is self-contained and does not require further network calls.

**Step 3:** Transfer the token to the isolated host via approved media.

**Step 4:** Apply the token on the isolated host:

```bash
dbxcli license apply-token \
  token=<base64-token> \
  output=~/.dbx/license/dbx-license.json
```

**Step 5:** Verify the license is active:

```bash
dbxcli license status
```

Expected output:

```
License Status  : ACTIVE (offline)
Edition         : Full Platform
Fingerprint     : a3b7c912e4f08d1b (match: OK)
Expiry          : 2027-04-11
Features        : oracle, postgres, host, monitoring
```

---

## Vault Initialization Without Internet

HashiCorp Vault in air-gap mode runs entirely from the locally-loaded Docker image. No outbound connections to HashiCorp license servers or telemetry endpoints are made.

### Start Vault in server mode

Create a minimal Vault configuration:

```bash
mkdir -p ~/vault/config ~/vault/data ~/vault/logs

cat > ~/vault/config/vault.hcl << 'EOF'
ui            = false
disable_mlock = true

storage "file" {
  path = "/vault/data"
}

listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_disable   = true
}
EOF
```

Start Vault using the locally-loaded image:

```bash
docker run -d \
  --name vault \
  --cap-add IPC_LOCK \
  -p 8200:8200 \
  -v ~/vault/config:/vault/config:ro \
  -v ~/vault/data:/vault/data \
  hashicorp/vault:latest \
  server -config=/vault/config/vault.hcl
```

### Initialize and unseal Vault

```bash
export VAULT_ADDR=http://127.0.0.1:8200

vault operator init -key-shares=5 -key-threshold=3 \
  -format=json > ~/vault-init.json

chmod 600 ~/vault-init.json
```

Store `~/vault-init.json` securely. It contains the unseal keys and root token. Follow your organization's key ceremony procedure for distributing unseal keys.

Unseal with 3 of 5 keys:

```bash
vault operator unseal $(cat ~/vault-init.json | python3 -c "import sys,json; print(json.load(sys.stdin)['unseal_keys_b64'][0])")
vault operator unseal $(cat ~/vault-init.json | python3 -c "import sys,json; print(json.load(sys.stdin)['unseal_keys_b64'][1])")
vault operator unseal $(cat ~/vault-init.json | python3 -c "import sys,json; print(json.load(sys.stdin)['unseal_keys_b64'][2])")
```

Enable the KV v2 secrets engine and configure dbx:

```bash
ROOT_TOKEN=$(cat ~/vault-init.json | python3 -c "import sys,json; print(json.load(sys.stdin)['root_token'])")
export VAULT_TOKEN=$ROOT_TOKEN

vault secrets enable -path=secret kv-v2

# Seed the dbx license key into Vault
vault kv put secret/dbx/license \
  oracle_full=<your-license-key>
```

Configure dbxcli to use the local Vault instance:

```bash
dbxcli vault configure \
  vault_addr=http://127.0.0.1:8200 \
  vault_mount=secret \
  vault_auth_method=token \
  token_env=VAULT_TOKEN
```

---

## Update Procedure for Air-Gap Environments

Updating an air-gap deployment follows the same two-phase process as the initial installation.

### Step 1: Build an update bundle (connected environment)

Pull only the components that changed in the new version:

```bash
mkdir -p ~/dbx-update-<version>/{images,npm,license}

# Pull updated images (only those with new versions)
docker pull ghcr.io/itunified.io/dbmon-central:<new-version>
docker save ghcr.io/itunified.io/dbmon-central:<new-version> \
  | gzip > ~/dbx-update-<version>/images/dbmon-central-<new-version>.tar.gz

# Pack updated npm packages
npm pack @itunified.io/mcp-oracle@<new-version> \
  --pack-destination ~/dbx-update-<version>/npm/

# Download updated dbxcli binary
curl -sSL https://releases.example.com/dbx/cli/<new-version>/dbxcli-linux-amd64 \
  -o ~/dbx-update-<version>/binaries/dbxcli-linux-amd64

# Generate checksums
cd ~/dbx-update-<version>
find . -type f -exec sha256sum {} \; > checksums.sha256
```

If the license expiry is within 60 days, generate a renewed offline license token and include it in the update bundle.

### Step 2: Apply the update (isolated environment)

Verify checksums, then apply each component:

```bash
# Verify
sha256sum --check checksums.sha256

# Load updated Docker images
for f in images/*.tar.gz; do
  gunzip -c "$f" | docker load
done

# Update npm packages
npm install -g npm/itunified.io-mcp-oracle-<new-version>.tgz

# Update dbxcli binary
install -o root -g root -m 755 \
  binaries/dbxcli-linux-amd64 \
  /usr/local/bin/dbxcli

# Update docker-compose.yml image tags if pinned to digests
# Restart affected services
docker compose pull   # uses only local images
docker compose up -d
```

### Step 3: Verify the update

```bash
dbxcli version
dbxcli monitor stack status
dbxcli license status
```

Confirm that all components report the expected new version and that the monitoring stack is healthy before closing the change window.
