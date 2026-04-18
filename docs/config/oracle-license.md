# Oracle License Declaration Reference

dbx enforces Oracle licensing boundaries at the tool level. Before a licensed tool can execute, dbx checks whether the required edition and options are declared on the target. This prevents accidental use of features that require licenses you do not hold, and provides an audit trail for compliance reporting.

This document covers the declaration model, supported editions and options, enforcement modes, and compliance reporting commands.

---

## Why License Declaration Matters

Oracle Database licensing is metric-based and feature-gated. Using features such as SQL Tuning Advisor, Active Session History, or Partition pruning on an unlicensed database is a license violation regardless of whether the feature is technically accessible. Oracle auditors examine AWR data, alert logs, and feature usage views (`V$OPTION`, `DBA_FEATURE_USAGE_STATISTICS`) during audits.

dbx implements tool-level gating to:

1. Prevent accidental use of licensed features by operational staff
2. Record feature access attempts in the audit log for compliance review
3. Surface licensing gaps during fleet-wide audits before Oracle engagement

Declaration is not a substitute for a contractual Oracle license. It is a governance control that reflects the licenses you hold.

---

## Supported Editions

| Edition key | Oracle edition |
|-------------|----------------|
| `enterprise` | Oracle Database Enterprise Edition |
| `standard2` | Oracle Database Standard Edition 2 |

Enterprise Edition is required for all options listed below. Standard Edition 2 does not support any separately licensed options.

Declare the edition on each target:

```bash
dbxcli target set prod-orcl oracle_edition=enterprise
```

Or in the target YAML file:

```yaml
oracle_edition: enterprise
```

---

## Supported Options

The following Oracle options are modeled in dbx. Declare only the options covered by your license agreement.

| Option key | Oracle product name | Affected tools |
|------------|---------------------|----------------|
| `diagnostics_pack` | Oracle Diagnostics Pack | AWR reports, ASH analysis, ADDM, metric history, `db advisor segment` |
| `tuning_pack` | Oracle Tuning Pack | SQL Tuning Advisor, SQL Access Advisor, `db advisor sql-tuning`, automatic plan evolution |
| `partitioning` | Oracle Partitioning | Partition enumeration, partition-level statistics, `db schema` partition metadata |
| `advanced_security` | Oracle Advanced Security | TDE wallet status, network encryption config, data redaction |
| `olap` | Oracle OLAP | OLAP cube metadata, analytic workspace queries |
| `spatial` | Oracle Spatial and Graph | SDO metadata queries, geometry type introspection |
| `advanced_compression` | Oracle Advanced Compression | Compression advisor, segment compression status |
| `real_application_testing` | Oracle Real Application Testing | Database Replay capture/replay, SQL Performance Analyzer |
| `label_security` | Oracle Label Security | OLS policy metadata, label component queries |

Declare options as a list:

```bash
dbxcli target set prod-orcl oracle_options="diagnostics_pack,tuning_pack,partitioning"
```

Or in the target YAML file:

```yaml
oracle_options:
  - diagnostics_pack
  - tuning_pack
  - partitioning
```

---

## OEM Management Packs

Oracle Enterprise Manager Management Packs are licensed separately from database options. The following packs are relevant to dbx tools that surface OEM-sourced metrics.

| Pack | Declaration key | Notes |
|------|-----------------|-------|
| Diagnostics Pack | `oem_diagnostics_pack` | Required for OEM AWR and ASH access |
| Tuning Pack | `oem_tuning_pack` | Required for OEM SQL Tuning access |
| Lifecycle Management Pack | `oem_lifecycle_pack` | Required for OEM patch automation tools |
| Cloud Management Pack | `oem_cloud_pack` | Required for cloud target management from OEM |

OEM pack declarations apply to the OEM management server target, not to individual database targets:

```yaml
entity_name: prod-oem
entity_type: oracle_host
description: OEM Management Server

oracle_oem_packs:
  - oem_diagnostics_pack
  - oem_tuning_pack
```

---

## Enforcement Modes

Three enforcement modes are available. Set globally in `~/.dbx/config.yaml` or override per target.

### `strict`

Blocks tool execution when the required license is not declared. Returns a non-zero exit code and an explanatory error message. This is the recommended mode for production environments.

```
ERROR: tool 'db advisor sql-tuning' requires oracle_options=[tuning_pack].
Target 'prod-orcl' has not declared this option.
To declare it: dbxcli target set prod-orcl oracle_options="tuning_pack"
To check what is declared: dbxcli target get prod-orcl
```

### `warn`

Allows tool execution but prints a warning to stderr and records the event in the license audit log. Useful during a transition period or in non-production environments.

```
WARNING: tool 'db advisor sql-tuning' requires oracle_options=[tuning_pack].
Target 'prod-orcl' has not declared this option. Proceeding anyway (enforcement=warn).
```

### `audit-only`

Executes the tool with no visible warning. Records the event silently in the audit log. Use for retroactive compliance analysis where you want to discover unlicensed usage without disrupting operations.

### Global configuration

```yaml
# ~/.dbx/config.yaml
oracle_license:
  enforcement_mode: strict    # strict | warn | audit-only. default: strict
```

### Per-target override

```yaml
# ~/.dbx/targets/dev-orcl.yaml
oracle_license_enforcement: warn
```

Per-target enforcement is applied after global enforcement. You can set globally to `strict` and relax individual non-production targets to `warn`.

---

## Per-Target Declaration

### Using the CLI

Set edition:

```bash
dbxcli target set prod-orcl oracle_edition=enterprise
```

Set options (comma-separated list, replaces existing options):

```bash
dbxcli target set prod-orcl oracle_options="diagnostics_pack,tuning_pack,partitioning"
```

Add a single option without replacing existing ones:

```bash
dbxcli target set prod-orcl oracle_options_add=advanced_compression
```

Remove a single option:

```bash
dbxcli target set prod-orcl oracle_options_remove=advanced_compression
```

View current declaration for a target:

```bash
dbxcli target get prod-orcl --show-license
```

Output:

```
Target:          prod-orcl
Edition:         enterprise
Options:
  diagnostics_pack       declared
  tuning_pack            declared
  partitioning           declared
Enforcement:     strict (global)
```

### In target YAML

```yaml
entity_name: prod-orcl
entity_type: oracle_database
host: db01.example.com
port: 1521
service: ORCL

oracle_edition: enterprise
oracle_options:
  - diagnostics_pack
  - tuning_pack
  - partitioning
oracle_license_enforcement: strict  # optional: override global setting
```

---

## Fleet-Wide Audit

`dbxcli license oracle-audit` queries all registered Oracle targets, reads `DBA_FEATURE_USAGE_STATISTICS`, and compares detected feature usage against the declared options on each target. Targets that show feature usage without a corresponding declaration are flagged.

```bash
dbxcli license oracle-audit
```

```bash
# Scope to a group
dbxcli license oracle-audit group=prod-fleet

# Scope to a single target
dbxcli license oracle-audit entity_name=prod-orcl

# Output as JSON for ingestion into a SIEM or reporting tool
dbxcli license oracle-audit --format json

# Output as CSV
dbxcli license oracle-audit --format csv > oracle-license-audit.csv
```

### Sample audit output

```
Oracle License Audit — 2026-04-10

Target           Edition        Detected Options          Undeclared
prod-orcl        enterprise     diagnostics_pack          —
                                tuning_pack               —
                                partitioning              —
prod-orcl-2      enterprise     diagnostics_pack          tuning_pack (USED, not declared)
dev-orcl         enterprise     diagnostics_pack          —
uat-orcl         enterprise     —                         —

Summary:
  Targets audited:     4
  Clean targets:       3
  Targets with gaps:   1 (prod-orcl-2)
  Undeclared options:  1
```

The `USED, not declared` flag means Oracle recorded actual feature usage in `DBA_FEATURE_USAGE_STATISTICS`. This is a license violation risk requiring immediate attention.

---

## Compliance Reporting

### Generate a compliance report

```bash
dbxcli license oracle-report
```

Produces a structured report covering:

- All declared editions and options across the fleet
- Feature usage gaps (used but not declared)
- Targets with no edition declared (unmanaged risk)
- Enforcement mode per target
- Historical trend (first seen, last seen) from the dbx audit log

```bash
# Export to PDF-compatible Markdown
dbxcli license oracle-report --format markdown > oracle-compliance-$(date +%Y-%m-%d).md

# Export as JSON for automated processing
dbxcli license oracle-report --format json
```

### Audit log

Every tool execution that touches a licensed feature is recorded:

```
~/.dbx/audit/oracle-license.log
```

Log format (JSON Lines):

```json
{"ts":"2026-04-10T14:32:01Z","target":"prod-orcl","tool":"db advisor sql-tuning","required_option":"tuning_pack","declared":true,"enforcement":"strict","action":"allowed","user":"dba@workstation.example.com"}
{"ts":"2026-04-10T14:35:12Z","target":"prod-orcl-2","tool":"db advisor sql-tuning","required_option":"tuning_pack","declared":false,"enforcement":"warn","action":"allowed_with_warning","user":"ops@workstation.example.com"}
```

Log rotation follows the global dbx log rotation policy (`~/.dbx/config.yaml` → `logging.max_age_days`).

### Checking license status for the active target

```bash
dbxcli license status
```

Output:

```
dbx License
  Plan:        Enterprise
  Licensed to: Example Corp
  Valid until: 2027-04-10
  Phone-home:  last checked 2026-04-10T12:00:00Z (OK)

Oracle Declarations — active target: prod-orcl
  Edition:     enterprise
  Options:     diagnostics_pack, tuning_pack, partitioning
  Enforcement: strict
```

### Activate a dbx license

```bash
dbxcli license activate license_key=XXXX-XXXX-XXXX-XXXX
```

See [dbxcli license activate](../cli/dbxcli_license_activate.md) for full details including offline activation.

---

## SEE ALSO

- [Target YAML Reference](targets.md) — Full schema for `oracle_edition`, `oracle_options`, `oracle_license_enforcement`
- [dbxcli license status](../cli/dbxcli_license_status.md) — Show dbx and Oracle license status
- [dbxcli license activate](../cli/dbxcli_license_activate.md) — Activate a dbx license key
- [dbxcli db advisor](../cli/dbxcli_db_advisor.md) — Licensed advisor tools
