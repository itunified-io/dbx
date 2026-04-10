## dbxcli db redo switch-history

Show log switch frequency

### Synopsis

Show redo log switch frequency per hour — useful for sizing redo logs.

```
dbxcli db redo switch-history [flags]
```

### Examples

```
  dbxcli db redo switch-history --target prod-db
```

### Options

```
  -h, --help   help for switch-history
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db redo](dbxcli_db_redo.md)	 - Oracle redo log operations

