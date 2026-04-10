## dbxcli pg dr runbook

Generate DR runbook

### Synopsis

Generate a disaster recovery runbook for a configuration.

```
dbxcli pg dr runbook [flags]
```

### Examples

```
  dbxcli pg dr runbook name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for runbook
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

