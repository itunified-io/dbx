## dbxcli pg dba connections

Show active connections

### Synopsis

Show all active connections from pg_stat_activity.

```
dbxcli pg dba connections [flags]
```

### Examples

```
  dbxcli pg dba connections --target prod-pg
```

### Options

```
  -h, --help   help for connections
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba](dbxcli_pg_dba.md)	 - DBA operations

