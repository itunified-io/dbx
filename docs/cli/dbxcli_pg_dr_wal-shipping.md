## dbxcli pg dr wal-shipping

Configure WAL shipping

### Synopsis

Configure WAL shipping for a DR configuration.

```
dbxcli pg dr wal-shipping [flags]
```

### Examples

```
  dbxcli pg dr wal-shipping name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for wal-shipping
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

