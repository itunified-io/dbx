## dbxcli pg backup pgbackrest-status

Show pgBackRest backup status

### Synopsis

Show the current pgBackRest backup status and history.

```
dbxcli pg backup pgbackrest-status [flags]
```

### Examples

```
  dbxcli pg backup pgbackrest-status --target prod-pg
```

### Options

```
  -h, --help   help for pgbackrest-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg backup](dbxcli_pg_backup.md)	 - Backup operations

