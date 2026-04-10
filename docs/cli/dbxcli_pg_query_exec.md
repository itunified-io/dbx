## dbxcli pg query exec

Execute a read-only SELECT

### Synopsis

Execute a read-only SELECT query and return the results.

```
dbxcli pg query exec [flags]
```

### Examples

```
  dbxcli pg query exec query="SELECT * FROM users LIMIT 10" --target prod-pg
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg query](dbxcli_pg_query.md)	 - SQL query execution

