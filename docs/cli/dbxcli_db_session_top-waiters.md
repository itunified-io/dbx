## dbxcli db session top-waiters

Show top sessions by wait time

### Synopsis

Show sessions with the highest cumulative wait time, useful for identifying performance bottlenecks.

```
dbxcli db session top-waiters [flags]
```

### Examples

```
  dbxcli db session top-waiters --target prod-db
```

### Options

```
  -h, --help   help for top-waiters
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db session](dbxcli_db_session.md)	 - Oracle session operations

