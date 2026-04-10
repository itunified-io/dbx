## dbxcli pg wal size

Show WAL directory size

### Synopsis

Show the current WAL directory size on disk.

```
dbxcli pg wal size [flags]
```

### Examples

```
  dbxcli pg wal size --target prod-pg
```

### Options

```
  -h, --help   help for size
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg wal](dbxcli_pg_wal.md)	 - WAL management

