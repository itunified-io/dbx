## dbxcli pg backup barman-status

Show Barman backup status

### Synopsis

Show the current Barman backup status and catalog.

```
dbxcli pg backup barman-status [flags]
```

### Examples

```
  dbxcli pg backup barman-status --target prod-pg
```

### Options

```
  -h, --help   help for barman-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg backup](dbxcli_pg_backup.md)	 - Backup operations

