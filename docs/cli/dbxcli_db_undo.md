## dbxcli db undo

Oracle undo/rollback operations

### Synopsis

Query Oracle undo tablespace usage and rollback segment details (V$UNDOSTAT, DBA_UNDO_EXTENTS). Read-only.

### Options

```
  -h, --help   help for undo
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db](dbxcli_db.md)	 - Oracle database read-only operations
* [dbxcli db undo list](dbxcli_db_undo_list.md)	 - List undo tablespace usage
* [dbxcli db undo segments](dbxcli_db_undo_segments.md)	 - Show undo segment details

