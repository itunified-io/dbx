## dbxcli pg dba locks

Show lock information

### Synopsis

Show current lock information from pg_locks.

```
dbxcli pg dba locks [flags]
```

### Examples

```
  dbxcli pg dba locks --target prod-pg
```

### Options

```
  -h, --help   help for locks
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba](dbxcli_pg_dba.md)	 - DBA operations

