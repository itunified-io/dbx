## dbxcli pg query explain

Generate execution plan

### Synopsis

Generate an EXPLAIN ANALYZE execution plan for a query.

```
dbxcli pg query explain [flags]
```

### Examples

```
  dbxcli pg query explain query="SELECT * FROM orders WHERE status='pending'" --target prod-pg
```

### Options

```
  -h, --help   help for explain
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg query](dbxcli_pg_query.md)	 - SQL query execution

