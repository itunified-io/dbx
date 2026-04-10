## dbxcli pg wal retention

Show WAL retention policy

### Synopsis

Show the configured WAL retention policy and current usage.

```
dbxcli pg wal retention [flags]
```

### Examples

```
  dbxcli pg wal retention --target prod-pg
```

### Options

```
  -h, --help   help for retention
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg wal](dbxcli_pg_wal.md)	 - WAL management

