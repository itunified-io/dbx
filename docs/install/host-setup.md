# Host/OS Monitoring Installation Guide

**Audience:** Sysadmin / DBA  
**Estimated setup time:** ~10 minutes  
**Applies to:** dbx Host Engine (OSS + Enterprise)

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Supported Distributions](#supported-distributions)
3. [Install MCP Adapter](#install-mcp-adapter)
4. [Add Host Target](#add-host-target)
5. [Verify Connection](#verify-connection)
6. [Tools Unlocked by Tier](#tools-unlocked-by-tier)
7. [Distribution-Specific Notes](#distribution-specific-notes)
8. [Host-to-Database Linking](#host-to-database-linking)

---

## Prerequisites

| Requirement | Minimum Version | Notes |
|---|---|---|
| SSH access to target host | — | Ed25519 key recommended; password auth not supported |
| SSH key pair | Ed25519 or RSA 4096 | Private key accessible to the machine running `dbxcli` |
| Node.js | 18+ | Required for MCP adapter |
| `dbxcli` binary | Latest | Required for target registration; see quick-start guide |
| Remote user sudo access | — | Read-only (`/proc`, `/sys`, system logs); Enterprise hardening tools require elevated access |

Generate a dedicated Ed25519 key for dbx if one does not exist:

```bash
ssh-keygen -t ed25519 -C "dbx-host-monitor" -f ~/.ssh/id_ed25519_dbx
```

Copy the public key to each target host:

```bash
ssh-copy-id -i ~/.ssh/id_ed25519_dbx.pub user@host.example.com
```

---

## Supported Distributions

| Distribution | Minimum Version | Package Manager | Notes |
|---|---|---|---|
| Fedora | 39+ | `dnf` | Full support including SELinux tooling |
| Ubuntu | 22.04 LTS | `apt` / `apt-get` | AppArmor tooling available in Enterprise tier |
| RHEL | 8+ | `dnf` / `yum` | Includes AlmaLinux 8+, Rocky Linux 8+ |
| SLES | 15 SP4+ | `zypper` | Requires SUSE Manager or direct Internet access for patch data |
| Oracle Linux | 8+ | `dnf` / `yum` | UEK kernel detection; Ksplice integration in Enterprise tier |

The host agent collects metrics via SSH without installing any persistent agent on the target. All reads are non-intrusive (`/proc`, `/sys`, `systemctl`, distro-specific package managers).

---

## Install MCP Adapter

### Free Tier (OSS) — 20 tools, 3 skills

```bash
npm install -g @itunified.io/mcp-host
```

### Enterprise Tier (Licensed) — 40 additional tools, 6 additional skills

Requires a valid dbx license key.

```bash
npm install -g @itunified.io/mcp-host-enterprise
```

Set the license key:

```bash
export DBX_LICENSE_KEY=<your-license-key>
```

Or reference from Vault:

```bash
dbxcli license configure \
  vault_path=secret/dbx/license \
  key_field=host_enterprise
```

Confirm installation:

```bash
npx @itunified.io/mcp-host --version
```

---

## Add Host Target

Register a host target using `dbxcli target add`. The `entity_type` must be `host`.

```bash
dbxcli target add \
  entity_name=prod-db-host-01 \
  entity_type=host \
  ssh_host=db.example.com \
  ssh_user=oracle \
  ssh_key=~/.ssh/id_ed25519_dbx \
  ssh_port=22
```

**Parameters:**

| Parameter | Required | Description |
|---|---|---|
| `entity_name` | Yes | Unique identifier for this host within dbx |
| `entity_type` | Yes | Must be `host` |
| `ssh_host` | Yes | Hostname or IP of the target server |
| `ssh_user` | Yes | OS user for SSH connection |
| `ssh_key` | Yes | Path to private key on the machine running dbxcli |
| `ssh_port` | No | SSH port (default: 22) |
| `sudo_enabled` | No | Set `true` if the ssh_user has sudo access (required for some Enterprise tools) |
| `distro_hint` | No | Override distro detection: `rhel`, `ubuntu`, `fedora`, `sles`, `ol` |

Verify target was registered:

```bash
dbxcli target list --type host
```

---

## Verify Connection

Run the following commands to confirm the SSH connection is working and basic metrics are accessible:

```bash
# Show host OS and hardware summary
dbxcli host info entity=prod-db-host-01

# Show current CPU utilization
dbxcli host cpu entity=prod-db-host-01

# Show memory usage
dbxcli host memory entity=prod-db-host-01
```

Expected output for `host info`:

```
Hostname      : prod-db-host-01.example.com
OS            : Oracle Linux 8.9 (UEK 6.1.55)
Architecture  : x86_64
CPU           : 2 x Intel(R) Xeon(R) Platinum 8375C @ 2.90GHz (64 logical cores)
Memory        : 128 GB total, 94 GB used, 34 GB free
Uptime        : 42 days, 7 hours
SSH User      : oracle
Sudo          : available
```

If connection fails:

```bash
dbxcli target ssh-test prod-db-host-01
```

---

## Tools Unlocked by Tier

### Free Tier — 20 tools, 3 skills

| Tool | Description |
|---|---|
| `host_info` | OS version, architecture, hostname, uptime |
| `host_cpu` | CPU count, utilization, load average |
| `host_memory` | Total, used, free, swap, huge pages |
| `host_disk` | Filesystem mounts, used/free space, inode usage |
| `host_io` | Block device I/O stats (reads, writes, await, util) |
| `host_network` | Interface list, TX/RX stats, errors |
| `host_process_list` | Top processes by CPU or memory |
| `host_process_detail` | Open files, threads, memory map for a given PID |
| `host_service_list` | systemd service status list |
| `host_service_status` | Status of a named service |
| `host_uptime` | Boot time, uptime, last restart reason |
| `host_kernel` | Kernel version, modules, sysctl key values |
| `host_users` | Currently logged-in users, last login |
| `host_cron_list` | Crontab entries for system and user cron |
| `host_file_stat` | File metadata (size, mtime, permissions, owner) |
| `host_tail_log` | Tail a file (e.g., `/var/log/messages`) |
| `host_syslog_search` | Search journald or syslog by keyword and time range |
| `host_open_ports` | Listening ports and associated processes |
| `host_firewall_status` | Active firewall backend (firewalld/ufw/iptables) + zone rules |
| `host_dns_config` | Resolv.conf, /etc/hosts, active DNS servers |

**Free tier skills:**

| Skill | Description |
|---|---|
| `/host-health` | Full health check: CPU, memory, disk, network, services |
| `/host-info` | Summary report: OS, hardware, uptime, kernel |
| `/host-test` | Connectivity and permission validation for this target |

### Enterprise Tier — +40 tools, +6 skills

| Category | Count | Tools |
|---|---|---|
| CIS / STIG Hardening | 8 | CIS benchmark scan, STIG profile check, compliance gap report, hardening apply, benchmark diff, audit finding export, remediation plan, baseline snapshot |
| Patch Management | 6 | Available patches list, patch apply, patch rollback, patch history, CVE cross-reference, security errata list |
| Policy Enforcement | 5 | Password policy check, SSH config audit, sudo rules audit, PAM config review, umask/suid scan |
| Security | 6 | Rootkit check (via rkhunter), AIDE integrity scan, SSSD/LDAP auth audit, failed login report, privileged user audit, setuid binary list |
| Capacity Planning | 4 | Disk growth projection, memory trend, CPU baseline, tablespace/filesystem growth forecast |
| Log Analysis | 5 | Log anomaly detection, error rate trend, login anomaly report, log archive summary, kernel OOM history |
| User Management | 3 | Local user audit, expired account check, password age report |
| Advanced | 3 | Ksplice live-patch status (Oracle Linux), AppArmor/SELinux profile status, container runtime detection |

**Enterprise skills:**

| Skill | Description |
|---|---|
| `/host-cis` | Run CIS Level 1 benchmark and return pass/fail summary |
| `/host-patch` | Apply pending security patches with pre/post health check |
| `/host-stig` | Run STIG profile check and generate gap report |
| `/host-capacity` | Capacity trend analysis with 30/60/90 day forecast |
| `/host-security-audit` | Comprehensive security posture report |
| `/host-log-audit` | Log anomaly summary across syslog, auth.log, and audit.log |

---

## Distribution-Specific Notes

### Fedora 39+

- Firewall tooling uses `firewalld` via `firewall-cmd`; nftables is the backend
- SELinux is enforcing by default; `host_selinux_status` and `host_selinux_denials` tools available in Enterprise
- DNF automatic is used for patch management (`dnf-automatic` must be installed on target for full patch history)

### Ubuntu 22.04+

- UFW is the default firewall interface; the `host_firewall_status` tool reads both UFW and nftables rules
- AppArmor is enforcing by default; Enterprise tier includes AppArmor profile audit and violation search
- `unattended-upgrades` is detected and patch history is read from `/var/log/unattended-upgrades/`

### RHEL 8+ / AlmaLinux / Rocky Linux

- `firewalld` is the default firewall; zone-based rules are fully supported
- Subscription Manager state is checked by `host_patch` to confirm patch source availability
- RHEL 9+ uses `dnf5` — dbx detects and adapts automatically
- FIPS mode detection is included in `host_info` output

### SLES 15+

- `zypper` is used for patch management; SUSE security patches (`category=security`) are filtered automatically
- `firewalld` is available but `SuSEfirewall2` may be present on older systems — both are detected
- SUSE Manager integration requires the Enterprise tier and additional configuration (`suse_manager_host`, `suse_manager_api_key`)

### Oracle Linux 8+

- UEK (Unbreakable Enterprise Kernel) version is reported in `host_info`
- Ksplice live-patch status is available in the Enterprise tier via `host_ksplice_status`
- Oracle Linux repos (`ol8_baseos_latest`, `ol8_appstream`) are detected for patch availability checks

---

## Host-to-Database Linking

Linking a host target to a database target enables correlated monitoring: OS-level metrics are presented alongside database metrics in health reports and skills.

### Link a host to an Oracle database

```bash
dbxcli target set prod-oracle-01 \
  host_target=prod-db-host-01
```

### Link a host to a PostgreSQL target

```bash
dbxcli target set prod-pg \
  host_target=prod-db-host-01
```

### Verify the link

```bash
dbxcli target show prod-oracle-01
```

Output (excerpt):

```
Entity Name  : prod-oracle-01
Type         : oracle_database
Host Target  : prod-db-host-01 (db.example.com) — LINKED
```

Once linked, correlated data is available via:

```bash
# Combined OS + DB health report
dbxcli oracle health entity=prod-oracle-01 include_host=true

# OS memory visible alongside SGA/PGA breakdown
dbxcli oracle memory entity=prod-oracle-01 correlated=true
```
