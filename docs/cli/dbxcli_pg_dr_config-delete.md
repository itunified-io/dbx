## dbxcli pg dr config-delete

Delete DR configuration

### Synopsis

Delete a disaster recovery configuration by name.

```
dbxcli pg dr config-delete [flags]
```

### Examples

```
  dbxcli pg dr config-delete name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for config-delete
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

