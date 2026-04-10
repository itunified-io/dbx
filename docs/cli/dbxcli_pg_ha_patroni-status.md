## dbxcli pg ha patroni-status

Show Patroni cluster status

### Synopsis

Show the current Patroni cluster status including leader, replicas, and timeline.

```
dbxcli pg ha patroni-status [flags]
```

### Examples

```
  dbxcli pg ha patroni-status --target prod-pg
```

### Options

```
  -h, --help   help for patroni-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg ha](dbxcli_pg_ha.md)	 - High availability

