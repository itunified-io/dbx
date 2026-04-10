## dbxcli db undo list

List undo tablespace usage

### Synopsis

Show undo tablespace utilization including active, unexpired, and expired extents.

```
dbxcli db undo list [flags]
```

### Examples

```
  dbxcli db undo list --target prod-db
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

* [dbxcli db undo](dbxcli_db_undo.md)	 - Oracle undo/rollback operations

