## dbxcli pg dba cancel-query

Cancel a running query

### Synopsis

Cancel a running query by backend PID (pg_cancel_backend).

```
dbxcli pg dba cancel-query [flags]
```

### Examples

```
  dbxcli pg dba cancel-query pid=12345 --target prod-pg
```

### Options

```
  -h, --help   help for cancel-query
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba](dbxcli_pg_dba.md)	 - DBA operations

