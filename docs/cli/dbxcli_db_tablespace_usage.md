## dbxcli db tablespace usage

Show aggregated tablespace usage summary

### Synopsis

Show aggregated storage utilization across all tablespaces with alert thresholds.

```
dbxcli db tablespace usage [flags]
```

### Examples

```
  dbxcli db tablespace usage --target prod-db
```

### Options

```
  -h, --help   help for usage
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db tablespace](dbxcli_db_tablespace.md)	 - Oracle tablespace operations

