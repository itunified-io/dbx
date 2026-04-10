## dbxcli pg ha patroni-restart

Restart Patroni member

### Synopsis

Restart a specific Patroni cluster member.

```
dbxcli pg ha patroni-restart [flags]
```

### Examples

```
  dbxcli pg ha patroni-restart member=node1 --target prod-pg
```

### Options

```
  -h, --help   help for patroni-restart
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg ha](dbxcli_pg_ha.md)	 - High availability

