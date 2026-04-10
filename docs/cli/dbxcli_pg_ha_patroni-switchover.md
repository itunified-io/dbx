## dbxcli pg ha patroni-switchover

Initiate Patroni switchover

### Synopsis

Initiate a controlled Patroni switchover to the specified target node.

```
dbxcli pg ha patroni-switchover [flags]
```

### Examples

```
  dbxcli pg ha patroni-switchover target_node=replica1 --target prod-pg
```

### Options

```
  -h, --help   help for patroni-switchover
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg ha](dbxcli_pg_ha.md)	 - High availability

