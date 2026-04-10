## dbxcli pg ha vip-status

Show VIP failover status

### Synopsis

Show the virtual IP failover status.

```
dbxcli pg ha vip-status [flags]
```

### Examples

```
  dbxcli pg ha vip-status --target prod-pg
```

### Options

```
  -h, --help   help for vip-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg ha](dbxcli_pg_ha.md)	 - High availability

