## dbxcli pg dba blocker-tree

Show blocking lock tree

### Synopsis

Show the blocking lock tree — which sessions block which.

```
dbxcli pg dba blocker-tree [flags]
```

### Examples

```
  dbxcli pg dba blocker-tree --target prod-pg
```

### Options

```
  -h, --help   help for blocker-tree
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba](dbxcli_pg_dba.md)	 - DBA operations

