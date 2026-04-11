# Monitoring Stack Installation Guide

**Audience:** DBA / DevOps  
**Estimated setup time:** ~30 minutes  
**Applies to:** dbx Monitoring Engine (mcp-dbmonitor OSS + mcp-dbmonitor-ee Licensed)

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Docker Compose Monitoring Stack](#docker-compose-monitoring-stack)
3. [Agent Deployment to Target Hosts](#agent-deployment-to-target-hosts)
4. [Kubernetes / Helm Deployment](#kubernetes--helm-deployment)
5. [Grafana Dashboard Import](#grafana-dashboard-import)
6. [VictoriaMetrics Integration](#victoriametrics-integration)
7. [Alert Configuration](#alert-configuration)
8. [Tools Unlocked by Tier](#tools-unlocked-by-tier)

---

## Architecture Overview

The dbx monitoring stack consists of four components that work together to collect, store, and visualize database and host metrics:

```
Target Hosts / Databases
        |
        | (SSH / DB connection, no agent required for dbxcli-native targets)
        v
  dbmon-agent                     <- Lightweight collection agent (optional for non-dbx hosts)
        |
        | HTTP POST (gzip+json)
        v
  dbmon-central                   <- Aggregation and routing layer
        |
        |-- remote_write -------> VictoriaMetrics   <- Long-term time-series storage
        |
        |-- alert evaluation ---> Alertmanager      <- Threshold evaluation + routing
        |
        v
      Grafana                     <- Dashboards, explore, alert UI
```

**Component roles:**

| Component | Description | Required |
|---|---|---|
| `dbmon-central` | Receives metrics from agents and dbx targets; provides query API; evaluates alert rules | Yes |
| `dbmon-agent` | Lightweight systemd-based agent for hosts not managed via dbxcli SSH | Optional |
| VictoriaMetrics | High-performance time-series database; compatible with Prometheus remote_write | Recommended |
| Grafana | Visualization and alerting UI; pre-built dashboards included | Recommended |
| Alertmanager | Alert routing to Slack, PagerDuty, email | Optional |

---

## Docker Compose Monitoring Stack

The monitoring stack is distributed as a Docker Compose bundle. Run it on any host with Docker 24+ and Docker Compose v2.

### Download and configure

```bash
mkdir -p ~/dbx-monitoring && cd ~/dbx-monitoring

# Download the compose bundle
curl -sSL https://releases.example.com/dbx/monitoring/latest/docker-compose.yml -o docker-compose.yml
curl -sSL https://releases.example.com/dbx/monitoring/latest/dbmon-central.yaml -o config/dbmon-central.yaml
```

### Configure `config/dbmon-central.yaml`

```yaml
central:
  listen_addr: "0.0.0.0:9100"
  retention_days: 30
  auth:
    enabled: true
    token_vault_path: secret/dbx/monitoring/central-token

remote_write:
  enabled: true
  url: http://victoriametrics:8428/api/v1/write
  flush_interval: 15s

alert_rules_dir: /etc/dbmon/alerts/

logging:
  level: info
  format: json
```

### `docker-compose.yml`

```yaml
services:
  dbmon-central:
    image: ghcr.io/itunified.io/dbmon-central:latest
    container_name: dbmon-central
    ports:
      - "9100:9100"
    volumes:
      - ./config/dbmon-central.yaml:/etc/dbmon/central.yaml:ro
      - ./config/alerts:/etc/dbmon/alerts:ro
      - dbmon-data:/var/lib/dbmon
    restart: unless-stopped

  victoriametrics:
    image: victoriametrics/victoria-metrics:latest
    container_name: victoriametrics
    ports:
      - "8428:8428"
    volumes:
      - vm-data:/storage
    command:
      - "--storageDataPath=/storage"
      - "--retentionPeriod=12"
      - "--httpListenAddr=:8428"
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./config/grafana/provisioning:/etc/grafana/provisioning:ro
    environment:
      - GF_SECURITY_ADMIN_PASSWORD__FILE=/run/secrets/grafana_admin_password
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:latest
    container_name: alertmanager
    ports:
      - "9093:9093"
    volumes:
      - ./config/alertmanager.yaml:/etc/alertmanager/alertmanager.yml:ro
    restart: unless-stopped

volumes:
  dbmon-data:
  vm-data:
  grafana-data:
```

### Start the stack

```bash
docker compose up -d

# Verify all containers are running
docker compose ps
```

Expected output:

```
NAME              STATUS     PORTS
dbmon-central     running    0.0.0.0:9100->9100/tcp
victoriametrics   running    0.0.0.0:8428->8428/tcp
grafana           running    0.0.0.0:3000->3000/tcp
alertmanager      running    0.0.0.0:9093->9093/tcp
```

---

## Agent Deployment to Target Hosts

For hosts managed outside dbxcli (no SSH key access), deploy the `dbmon-agent` as a systemd service.

### Download the agent binary

```bash
# On the target host
curl -sSL https://releases.example.com/dbx/agent/latest/dbmon-agent-linux-amd64 \
  -o /usr/local/bin/dbmon-agent
chmod +x /usr/local/bin/dbmon-agent
```

### Create the agent configuration

```bash
mkdir -p /etc/dbmon
cat > /etc/dbmon/agent.yaml << 'EOF'
agent:
  agent_id: prod-host-01
  central_url: https://monitoring.example.com:9100
  auth_token_env: DBMON_AUTH_TOKEN
  collect_interval: 30s
  flush_interval: 60s

targets:
  - name: self
    type: host
    collect_interval: 30s
  - name: local-postgres
    type: postgres
    host: localhost
    port: 5432
    database: appdb
    credential: vault
    vault_path: secret/dbx/monitoring/local-postgres
    collect_interval: 60s

logging:
  level: info
  output: journald
EOF
```

**Key agent configuration fields:**

| Field | Description |
|---|---|
| `agent_id` | Unique identifier for this agent in dbmon-central |
| `central_url` | URL of the dbmon-central aggregation endpoint |
| `auth_token_env` | Environment variable holding the agent authentication token |
| `collect_interval` | Default metric collection interval (per-target override supported) |
| `flush_interval` | How often collected metrics are pushed to central |

### Create the systemd service

```bash
cat > /etc/systemd/system/dbmon-agent.service << 'EOF'
[Unit]
Description=dbx Monitoring Agent
After=network.target

[Service]
Type=simple
User=dbmon
ExecStart=/usr/local/bin/dbmon-agent --config /etc/dbmon/agent.yaml
EnvironmentFile=/etc/dbmon/agent.env
Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

Create the environment file with the auth token:

```bash
# /etc/dbmon/agent.env — do not commit this file
echo "DBMON_AUTH_TOKEN=<token-from-vault>" > /etc/dbmon/agent.env
chmod 600 /etc/dbmon/agent.env
```

### Enable and start the service

```bash
useradd --system --no-create-home --shell /sbin/nologin dbmon
systemctl daemon-reload
systemctl enable dbmon-agent
systemctl start dbmon-agent
systemctl status dbmon-agent
```

### Verify agent registration in dbmon-central

```bash
dbxcli monitor agent list
```

Expected output:

```
Agent ID        Status    Last Seen         Targets
prod-host-01    ACTIVE    2 seconds ago     2 (host: self, postgres: local-postgres)
```

---

## Kubernetes / Helm Deployment

For Kubernetes environments, use the dbx Helm chart to deploy the monitoring stack.

### Add the Helm repository

```bash
helm repo add dbx https://charts.example.com/dbx
helm repo update
```

### Create a values override file

```yaml
# dbx-monitoring-values.yaml

central:
  replicaCount: 1
  ingress:
    enabled: true
    host: monitoring.example.com
    tls:
      enabled: true
      secretName: monitoring-tls
  persistence:
    storageClass: standard
    size: 20Gi

victoriametrics:
  enabled: true
  retention: "12"
  persistence:
    storageClass: standard
    size: 100Gi

grafana:
  enabled: true
  adminPasswordSecret: grafana-admin-secret
  ingress:
    enabled: true
    host: grafana.example.com

alertmanager:
  enabled: true
  configSecret: alertmanager-config

vault:
  enabled: true
  address: https://vault.example.com:8200
  authMethod: kubernetes
  roleName: dbx-monitoring
```

### Deploy

```bash
kubectl create namespace dbx-monitoring

helm install dbx-monitoring dbx/dbx-monitoring \
  --namespace dbx-monitoring \
  --values dbx-monitoring-values.yaml
```

### Verify deployment

```bash
kubectl get pods -n dbx-monitoring
```

Expected output:

```
NAME                               READY   STATUS    RESTARTS
dbmon-central-7d9f8c4b5-xk2lp     1/1     Running   0
victoriametrics-0                  1/1     Running   0
grafana-6b8c7f9d4-m3np9            1/1     Running   0
alertmanager-5f7b9d8c6-p8xnt       1/1     Running   0
```

---

## Grafana Dashboard Import

dbx ships built-in Grafana dashboards for each supported database and host type.

### Built-in dashboards

| Dashboard | ID | Description |
|---|---|---|
| Oracle Overview | `dbx-oracle-overview` | Sessions, tablespaces, wait events, redo, ASM |
| Oracle Performance | `dbx-oracle-perf` | AWR summary, SQL top, CPU, I/O, buffer cache |
| Oracle Data Guard | `dbx-oracle-dg` | Log apply lag, transport status, redo rate |
| Oracle RAC | `dbx-oracle-rac` | Per-instance stats, interconnect, GCS/GES |
| PostgreSQL Overview | `dbx-pg-overview` | Connections, query stats, bloat, checkpoints |
| PostgreSQL CNPG | `dbx-pg-cnpg` | Cluster topology, replication lag, backup status |
| Host Overview | `dbx-host-overview` | CPU, memory, disk I/O, network, processes |
| Alert Summary | `dbx-alerts` | Active alerts, alert history, resolution rate |

### Import via dbxcli

```bash
dbxcli monitor dashboard import all --grafana-url=http://grafana.example.com:3000
```

Import a specific dashboard:

```bash
dbxcli monitor dashboard import dbx-oracle-overview \
  --grafana-url=http://grafana.example.com:3000
```

### Manual import via Grafana UI

1. In Grafana, go to **Dashboards → Import**
2. Enter the dashboard ID (e.g., `dbx-oracle-overview`) or upload the JSON from `/usr/share/dbx/dashboards/`
3. Select the VictoriaMetrics data source
4. Click **Import**

---

## VictoriaMetrics Integration

dbmon-central writes metrics to VictoriaMetrics via the Prometheus `remote_write` protocol.

Configure the remote_write endpoint in `dbmon-central.yaml`:

```yaml
remote_write:
  enabled: true
  url: http://victoriametrics:8428/api/v1/write
  flush_interval: 15s
  queue_size: 10000
  max_retries: 3
  retry_interval: 5s
  labels:
    environment: production
    cluster: prod-db-cluster
```

To use an external VictoriaMetrics cluster:

```yaml
remote_write:
  enabled: true
  url: https://vm.example.com:8480/insert/0/prometheus/api/v1/write
  tls:
    ca_file: /etc/dbmon/certs/ca.crt
    cert_file: /etc/dbmon/certs/client.crt
    key_file: /etc/dbmon/certs/client.key
```

To query metrics directly from VictoriaMetrics using MetricsQL:

```bash
curl 'http://victoriametrics:8428/api/v1/query?query=dbx_oracle_sessions_active{entity="prod-oracle-01"}'
```

---

## Alert Configuration

Alerts are defined as YAML files in the alert rules directory (`/etc/dbmon/alerts/`). Each file can contain multiple rules.

### Alert rule format

```yaml
# /etc/dbmon/alerts/oracle-sessions.yaml

groups:
  - name: oracle.sessions
    rules:
      - alert: OracleActiveSessionsHigh
        expr: dbx_oracle_sessions_active > 400
        for: 5m
        severity: warning
        labels:
          engine: oracle
          scope: sessions
        annotations:
          summary: "High active session count on {{ $labels.entity }}"
          description: "Active sessions: {{ $value }}"

      - alert: OracleActiveSessionsCritical
        expr: dbx_oracle_sessions_active > 600
        for: 2m
        severity: critical
        labels:
          engine: oracle
          scope: sessions
        annotations:
          summary: "Critical active session count on {{ $labels.entity }}"
          description: "Active sessions: {{ $value }}. Investigate immediately."
```

### Slack routing configuration (`config/alertmanager.yaml`)

```yaml
global:
  resolve_timeout: 5m

route:
  receiver: default
  group_by: [alertname, entity, severity]
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  routes:
    - match:
        severity: critical
      receiver: critical-alerts
      continue: true
    - match:
        severity: warning
      receiver: warning-alerts

receivers:
  - name: default
    slack_configs:
      - api_url: "<slack-webhook-url>"
        channel: "#infra-monitoring"
        title: "{{ .GroupLabels.alertname }}"
        text: "{{ range .Alerts }}{{ .Annotations.description }}{{ end }}"

  - name: warning-alerts
    slack_configs:
      - api_url: "<slack-webhook-url>"
        channel: "#infra-alerts"

  - name: critical-alerts
    slack_configs:
      - api_url: "<slack-webhook-url>"
        channel: "#infra-alerts"
```

Reload alert rules without restarting:

```bash
dbxcli monitor alerts reload
```

---

## Tools Unlocked by Tier

### Free Tier — mcp-dbmonitor, 16 tools, 4 skills

| Category | Tools |
|---|---|
| Agent | `monitor_agent_list`, `monitor_agent_status`, `monitor_agent_restart` |
| Metrics | `monitor_query`, `monitor_metric_list`, `monitor_metric_latest` |
| Dashboards | `monitor_dashboard_list`, `monitor_dashboard_import`, `monitor_dashboard_export` |
| Alerts | `monitor_alert_list`, `monitor_alert_silence`, `monitor_alert_history` |
| Health | `monitor_stack_status`, `monitor_target_health`, `monitor_connectivity_test` |

**Free tier skills:**

| Skill | Description |
|---|---|
| `/monitor-health` | Check status of all monitoring components and agents |
| `/monitor-alerts` | List and summarize active alerts across all targets |
| `/monitor-dashboards` | Import or refresh built-in Grafana dashboards |
| `/monitor-test` | End-to-end connectivity test: agent → central → VictoriaMetrics |

### Licensed Tier — mcp-dbmonitor-ee, 35 tools, 6 skills

Additional tools beyond the free tier:

| Category | Count | Tools |
|---|---|---|
| Anomaly Detection | 5 | ML-based query anomaly, session spike detection, wait event outlier, slow replication alert, disk growth anomaly |
| Capacity Forecasting | 4 | Tablespace growth forecast, disk capacity projection, memory trend, connection pool saturation |
| Compliance Reporting | 6 | SOC 2 monitoring posture, CIS metric coverage, audit completeness report, retention policy check, data access trend, GDPR metric audit |
| SLA / Availability | 4 | Uptime SLA report, MTTR calculation, incident frequency, availability heatmap |
| Cross-Target Correlation | 5 | Correlated event search, host-database event overlay, replication lag + host I/O correlation, alert grouping by impact, root cause suggestion |
| Advanced Alerting | 5 | Alert dependency chains, suppression rules, escalation policies, alert deduplication, on-call schedule integration |
| Reporting | 6 | Weekly DBA digest, monthly availability report, capacity trend PDF export, executive summary, alert trend analysis, change impact report |

**Licensed tier skills:**

| Skill | Description |
|---|---|
| `/monitor-forecast` | 30/60/90 day capacity forecast for all registered targets |
| `/monitor-anomalies` | Detect and summarize metric anomalies over the past 24 hours |
| `/monitor-sla` | Generate SLA availability report for a specified time range |
| `/monitor-compliance` | Run monitoring compliance check against SOC 2 / CIS criteria |
| `/monitor-weekly` | Generate and post weekly DBA monitoring digest to Slack |
| `/monitor-incident` | Correlate alerts into an incident timeline for root cause analysis |
