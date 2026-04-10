## dbxcli pg dr

Disaster recovery

### Synopsis

PostgreSQL disaster recovery — configuration, status, switchover, failover,
WAL shipping, MinIO archive, PITR, runbooks, and compliance reporting.

### Options

```
  -h, --help   help for dr
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg](dbxcli_pg.md)	 - PostgreSQL database operations
* [dbxcli pg dr config-delete](dbxcli_pg_dr_config-delete.md)	 - Delete DR configuration
* [dbxcli pg dr config-get](dbxcli_pg_dr_config-get.md)	 - Get DR configuration
* [dbxcli pg dr config-list](dbxcli_pg_dr_config-list.md)	 - List DR configurations
* [dbxcli pg dr config-set](dbxcli_pg_dr_config-set.md)	 - Set DR configuration
* [dbxcli pg dr failover](dbxcli_pg_dr_failover.md)	 - Execute DR failover
* [dbxcli pg dr minio-status](dbxcli_pg_dr_minio-status.md)	 - Show MinIO WAL archive status
* [dbxcli pg dr minio-verify](dbxcli_pg_dr_minio-verify.md)	 - Verify MinIO archive integrity
* [dbxcli pg dr monitor](dbxcli_pg_dr_monitor.md)	 - Start continuous DR monitoring
* [dbxcli pg dr pitr-restore](dbxcli_pg_dr_pitr-restore.md)	 - Point-in-time recovery
* [dbxcli pg dr pitr-timeline](dbxcli_pg_dr_pitr-timeline.md)	 - Show PITR timeline
* [dbxcli pg dr report](dbxcli_pg_dr_report.md)	 - Generate DR compliance report
* [dbxcli pg dr runbook](dbxcli_pg_dr_runbook.md)	 - Generate DR runbook
* [dbxcli pg dr status](dbxcli_pg_dr_status.md)	 - Show DR status
* [dbxcli pg dr switchover](dbxcli_pg_dr_switchover.md)	 - Execute DR switchover
* [dbxcli pg dr sync-status](dbxcli_pg_dr_sync-status.md)	 - Show replication sync status
* [dbxcli pg dr test-failover](dbxcli_pg_dr_test-failover.md)	 - Dry-run failover test
* [dbxcli pg dr validate](dbxcli_pg_dr_validate.md)	 - Validate DR setup
* [dbxcli pg dr wal-shipping](dbxcli_pg_dr_wal-shipping.md)	 - Configure WAL shipping

