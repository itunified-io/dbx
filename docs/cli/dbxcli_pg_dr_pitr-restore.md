## dbxcli pg dr pitr-restore

Point-in-time recovery

### Synopsis

Perform a point-in-time recovery to a specific timestamp.

```
dbxcli pg dr pitr-restore [flags]
```

### Examples

```
  dbxcli pg dr pitr-restore name=prod-dr target_time="2026-01-01 12:00:00" --target prod-pg
```

### Options

```
  -h, --help   help for pitr-restore
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

