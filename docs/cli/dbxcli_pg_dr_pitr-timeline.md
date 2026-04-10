## dbxcli pg dr pitr-timeline

Show PITR timeline

### Synopsis

Show the PITR timeline and available recovery points.

```
dbxcli pg dr pitr-timeline [flags]
```

### Examples

```
  dbxcli pg dr pitr-timeline name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for pitr-timeline
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

