## dbxcli pg backup pgbackrest-backup

Trigger pgBackRest backup

### Synopsis

Trigger a pgBackRest backup of the specified type and stanza.

```
dbxcli pg backup pgbackrest-backup [flags]
```

### Examples

```
  dbxcli pg backup pgbackrest-backup type=full stanza=main --target prod-pg
```

### Options

```
  -h, --help   help for pgbackrest-backup
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg backup](dbxcli_pg_backup.md)	 - Backup operations

