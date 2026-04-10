## dbxcli pg wal archive-status

Check WAL archiving status

### Synopsis

Check the WAL archiving status and any failures.

```
dbxcli pg wal archive-status [flags]
```

### Examples

```
  dbxcli pg wal archive-status --target prod-pg
```

### Options

```
  -h, --help   help for archive-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg wal](dbxcli_pg_wal.md)	 - WAL management

