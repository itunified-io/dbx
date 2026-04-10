## dbxcli pg ha pgbouncer-resume

Resume PgBouncer connections

### Synopsis

Resume all PgBouncer connection pools.

```
dbxcli pg ha pgbouncer-resume [flags]
```

### Examples

```
  dbxcli pg ha pgbouncer-resume --target prod-pg
```

### Options

```
  -h, --help   help for pgbouncer-resume
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg ha](dbxcli_pg_ha.md)	 - High availability

