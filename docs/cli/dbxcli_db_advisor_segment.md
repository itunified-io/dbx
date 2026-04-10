## dbxcli db advisor segment

Show segment advisor recommendations

### Synopsis

Show segment advisor recommendations for table/index shrink, compression, and space reclamation.

```
dbxcli db advisor segment [flags]
```

### Examples

```
  dbxcli db advisor segment --target prod-db
```

### Options

```
  -h, --help   help for segment
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db advisor](dbxcli_db_advisor.md)	 - Oracle advisor operations

