## dbxcli pg dr config-set

Set DR configuration

### Synopsis

Create or update a disaster recovery configuration.

```
dbxcli pg dr config-set [flags]
```

### Examples

```
  dbxcli pg dr config-set name=prod-dr primary=dc1-pg standby=dc2-pg --target prod-pg
```

### Options

```
  -h, --help   help for config-set
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

