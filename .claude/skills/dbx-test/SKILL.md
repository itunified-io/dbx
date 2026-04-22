---
name: dbx-test
description: Full live integration test suite. Provisions a 3-target test env (Oracle 19c + PG 16 + plain OL9 host) on Proxmox via proxctl, applies OS-level config via linuxctl, runs dbxcli smoke tests against each target, posts results to Slack C0ALHK18VC5, tears down. Use when running /dbx-test or when asked to verify dbx against real targets (dbx#15).
---

# /dbx-test

Full end-to-end integration test suite for dbx. Provisions real test VMs, runs dbxcli smoke tests, tears down.

## When to use

- CI run: "run dbx live tests"
- Pre-release verification before tagging dbx
- Post-bugfix regression check
- Any user says `/dbx-test`

## Prerequisites (enforced by skill)

- proxctl v2026.04.11.6+ installed and `~/.proxctl/config.yaml` has a working Proxmox context
- linuxctl v2026.04.11.6+ installed
- dbxcli + dbx-ee installed and on PATH
- Vault AppRole credentials loaded (`source ~/.secrets/.env`)
- Slack webhook in env (`SLACK_WH_DBX_TEST`) — posts to C0ALHK18VC5
- Test env YAML: `docs/testing/dbx-test-env/env.yaml`

## Steps

1. **Pre-flight** — verify prerequisites; fail fast if anything missing.
   - `proxctl version` / `linuxctl version` meet minimum CalVer
   - `proxctl config validate docs/testing/dbx-test-env/env.yaml`
   - `[ -n "$SLACK_WH_DBX_TEST" ]`
   - Vault token / AppRole reachable
2. **Up phase**:
   - `proxctl workflow up docs/testing/dbx-test-env/env.yaml --yes` (3 VMs concurrent via MultiNodeWorkflow)
   - Wait for SSH reachable on all 3 nodes (proxctl already blocks until cloud-init done)
3. **OS layout phase**:
   - For each node: `linuxctl apply apply docs/testing/dbx-test-env/env.yaml --host <node> --yes`
   - Verify zero drift: `linuxctl diff docs/testing/dbx-test-env/env.yaml --host <node>` (expect empty)
   - (Optional) `linuxctl cluster setup-ssh docs/testing/dbx-test-env/env.yaml` for SSH mesh
4. **dbx target registration**:
   - For each VM, auto-generate `~/.dbx/targets/<name>.yaml` with connection info (pulls from env.yaml / hypervisor.yaml)
5. **Smoke tests** (run + collect pass/fail):
   - Oracle: `dbxcli policy scan -t dbx-test-oracle-19c --profile cis-oracle-19c` — expect CIS findings
   - Oracle: `dbxcli db session list -t dbx-test-oracle-19c`
   - Oracle: `dbxcli db tablespace list -t dbx-test-oracle-19c`
   - PG: `dbxcli pg schema list -t dbx-test-pg-16`
   - PG: `dbxcli pg connection status -t dbx-test-pg-16`
   - Host: `dbxcli host info -t dbx-test-plain-ol9` (20 tools — iterate full host subcommand list)
   - RAG: `dbxcli rag index-status`
   - Cloud: `dbxcli cloud estimate docs/testing/dbx-test-env/env.yaml --dry-run`
6. **Slack summary**:
   - Build a formatted summary: targets up, tests run, pass/fail counts, run duration, snapshot names (if any)
   - Post to C0ALHK18VC5 via `$SLACK_WH_DBX_TEST`
7. **Cleanup** (always runs, even on failure):
   - `proxctl workflow down docs/testing/dbx-test-env/env.yaml --yes` (double-confirm gate auto-acknowledged via --yes)
   - Remove `~/.dbx/targets/dbx-test-*.yaml`
   - If apply failed, a snapshot was auto-created (proxctl default); mention snapshot ID in Slack

## Failure modes

- VM provision fails → stop, post failure, teardown anything partial
- Smoke test fails → continue remaining tests, aggregate failures, teardown
- Slack unreachable → log locally, don't fail the overall run
- Teardown fails → post warning; manual cleanup runbook linked

## Exit codes

- 0: all pass
- 1: any smoke test failed
- 2: provision/apply/teardown failed
- 3: prerequisites missing

## Outputs

- `/tmp/dbx-test-<timestamp>/run.log` — full run log
- `/tmp/dbx-test-<timestamp>/results.json` — pass/fail per test
- `/tmp/dbx-test-<timestamp>/slack-posted.json` — Slack webhook response
- Slack message ID posted to C0ALHK18VC5

## References

- Plan: infrastructure/docs/plans/polished-purring-piglet.md
- Plan (P6): infrastructure#389
- Env YAML: docs/testing/dbx-test-env/env.yaml
- Closes: itunified-io/dbx#15
