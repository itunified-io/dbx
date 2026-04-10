## dbxcli pg dr test-failover

Dry-run failover test

### Synopsis

Run a dry-run failover test without actually switching over.

```
dbxcli pg dr test-failover [flags]
```

### Examples

```
  dbxcli pg dr test-failover name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for test-failover
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

