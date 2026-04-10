## dbxcli db redo list

List redo log groups

### Synopsis

List all redo log groups with status, size, member count, and sequence number.

```
dbxcli db redo list [flags]
```

### Examples

```
  dbxcli db redo list --target prod-db
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

* [dbxcli db redo](dbxcli_db_redo.md)	 - Oracle redo log operations

