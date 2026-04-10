## dbxcli db tablespace list

List tablespaces with usage metrics

### Synopsis

List all tablespaces with total size, used space, free space, and percentage utilization.

```
dbxcli db tablespace list [flags]
```

### Examples

```
  dbxcli db tablespace list --target prod-db
  dbxcli db tablespace list --target prod-db --format json
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db tablespace](dbxcli_db_tablespace.md)	 - Oracle tablespace operations

