## dbxcli pg dr switchover

Execute DR switchover

### Synopsis

Execute a controlled DR switchover from primary to standby.

```
dbxcli pg dr switchover [flags]
```

### Examples

```
  dbxcli pg dr switchover name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for switchover
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

