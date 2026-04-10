## dbxcli pg dba long-queries

Show long-running queries

### Synopsis

Show queries that have been running longer than the configured threshold.

```
dbxcli pg dba long-queries [flags]
```

### Examples

```
  dbxcli pg dba long-queries --target prod-pg
```

### Options

```
  -h, --help   help for long-queries
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba](dbxcli_pg_dba.md)	 - DBA operations

