## dbxcli pg dr failover

Execute DR failover

### Synopsis

Execute an emergency DR failover to the standby.

```
dbxcli pg dr failover [flags]
```

### Examples

```
  dbxcli pg dr failover name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for failover
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

