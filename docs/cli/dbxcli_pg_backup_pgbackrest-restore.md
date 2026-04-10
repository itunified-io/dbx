## dbxcli pg backup pgbackrest-restore

Restore from pgBackRest

### Synopsis

Restore from a pgBackRest backup with optional point-in-time target.

```
dbxcli pg backup pgbackrest-restore [flags]
```

### Examples

```
  dbxcli pg backup pgbackrest-restore stanza=main target_time="2026-01-01 12:00:00" --target prod-pg
```

### Options

```
  -h, --help   help for pgbackrest-restore
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg backup](dbxcli_pg_backup.md)	 - Backup operations

