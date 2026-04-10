## dbxcli db undo segments

Show undo segment details

### Synopsis

Show individual rollback segment status, size, and transaction counts.

```
dbxcli db undo segments [flags]
```

### Examples

```
  dbxcli db undo segments --target prod-db
```

### Options

```
  -h, --help   help for segments
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db undo](dbxcli_db_undo.md)	 - Oracle undo/rollback operations

