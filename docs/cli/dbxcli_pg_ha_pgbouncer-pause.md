## dbxcli pg ha pgbouncer-pause

Pause PgBouncer connections

### Synopsis

Pause all PgBouncer connection pools.

```
dbxcli pg ha pgbouncer-pause [flags]
```

### Examples

```
  dbxcli pg ha pgbouncer-pause --target prod-pg
```

### Options

```
  -h, --help   help for pgbouncer-pause
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg ha](dbxcli_pg_ha.md)	 - High availability

