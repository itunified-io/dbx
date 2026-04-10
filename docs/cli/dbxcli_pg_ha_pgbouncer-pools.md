## dbxcli pg ha pgbouncer-pools

List PgBouncer pools

### Synopsis

List all PgBouncer connection pools with their configuration.

```
dbxcli pg ha pgbouncer-pools [flags]
```

### Examples

```
  dbxcli pg ha pgbouncer-pools --target prod-pg
```

### Options

```
  -h, --help   help for pgbouncer-pools
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg ha](dbxcli_pg_ha.md)	 - High availability

