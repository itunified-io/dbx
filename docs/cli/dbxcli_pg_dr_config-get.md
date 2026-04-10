## dbxcli pg dr config-get

Get DR configuration

### Synopsis

Get a specific disaster recovery configuration by name.

```
dbxcli pg dr config-get [flags]
```

### Examples

```
  dbxcli pg dr config-get name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for config-get
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

