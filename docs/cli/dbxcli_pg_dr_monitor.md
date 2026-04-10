## dbxcli pg dr monitor

Start continuous DR monitoring

### Synopsis

Start continuous disaster recovery monitoring for a configuration.

```
dbxcli pg dr monitor [flags]
```

### Examples

```
  dbxcli pg dr monitor name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for monitor
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

