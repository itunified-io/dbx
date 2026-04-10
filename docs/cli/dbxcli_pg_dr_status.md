## dbxcli pg dr status

Show DR status

### Synopsis

Show the current disaster recovery status for a configuration.

```
dbxcli pg dr status [flags]
```

### Examples

```
  dbxcli pg dr status name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

