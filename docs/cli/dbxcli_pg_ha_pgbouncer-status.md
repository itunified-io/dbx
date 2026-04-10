## dbxcli pg ha pgbouncer-status

Show PgBouncer pool status

### Synopsis

Show PgBouncer connection pool status and statistics.

```
dbxcli pg ha pgbouncer-status [flags]
```

### Examples

```
  dbxcli pg ha pgbouncer-status --target prod-pg
```

### Options

```
  -h, --help   help for pgbouncer-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg ha](dbxcli_pg_ha.md)	 - High availability

