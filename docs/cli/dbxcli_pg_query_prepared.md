## dbxcli pg query prepared

Execute a prepared statement

### Synopsis

Execute a prepared statement by name with the given parameters.

```
dbxcli pg query prepared [flags]
```

### Examples

```
  dbxcli pg query prepared name=get_user params="[1]" --target prod-pg
```

### Options

```
  -h, --help   help for prepared
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg query](dbxcli_pg_query.md)	 - SQL query execution

