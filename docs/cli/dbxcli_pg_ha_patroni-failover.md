## dbxcli pg ha patroni-failover

Initiate Patroni failover

### Synopsis

Initiate a Patroni failover to the specified target node.

```
dbxcli pg ha patroni-failover [flags]
```

### Examples

```
  dbxcli pg ha patroni-failover target_node=replica1 --target prod-pg
```

### Options

```
  -h, --help   help for patroni-failover
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg ha](dbxcli_pg_ha.md)	 - High availability

