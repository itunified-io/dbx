## dbxcli pg observe stat-bgwriter

Show background writer stats

### Synopsis

Show background writer statistics from pg_stat_bgwriter.

```
dbxcli pg observe stat-bgwriter [flags]
```

### Examples

```
  dbxcli pg observe stat-bgwriter --target prod-pg
```

### Options

```
  -h, --help   help for stat-bgwriter
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg observe](dbxcli_pg_observe.md)	 - Observability

