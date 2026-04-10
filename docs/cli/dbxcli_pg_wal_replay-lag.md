## dbxcli pg wal replay-lag

Show WAL replay lag on standbys

### Synopsis

Show WAL replay lag on all standby servers.

```
dbxcli pg wal replay-lag [flags]
```

### Examples

```
  dbxcli pg wal replay-lag --target prod-pg
```

### Options

```
  -h, --help   help for replay-lag
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg wal](dbxcli_pg_wal.md)	 - WAL management

