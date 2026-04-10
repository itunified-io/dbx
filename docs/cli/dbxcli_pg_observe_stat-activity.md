## dbxcli pg observe stat-activity

Show pg_stat_activity snapshot

### Synopsis

Show a snapshot of pg_stat_activity with all backend details.

```
dbxcli pg observe stat-activity [flags]
```

### Examples

```
  dbxcli pg observe stat-activity --target prod-pg
```

### Options

```
  -h, --help   help for stat-activity
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg observe](dbxcli_pg_observe.md)	 - Observability

