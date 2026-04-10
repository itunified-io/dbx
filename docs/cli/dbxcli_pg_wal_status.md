## dbxcli pg wal status

Show WAL generation stats

### Synopsis

Show WAL generation statistics and current LSN.

```
dbxcli pg wal status [flags]
```

### Examples

```
  dbxcli pg wal status --target prod-pg
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg wal](dbxcli_pg_wal.md)	 - WAL management

