## dbxcli pg dr sync-status

Show replication sync status

### Synopsis

Show the replication synchronization status for a DR configuration.

```
dbxcli pg dr sync-status [flags]
```

### Examples

```
  dbxcli pg dr sync-status name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for sync-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

