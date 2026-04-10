## dbxcli pg perf wait-events

Show wait event statistics

### Synopsis

Show wait event statistics from pg_stat_activity.

```
dbxcli pg perf wait-events [flags]
```

### Examples

```
  dbxcli pg perf wait-events --target prod-pg
```

### Options

```
  -h, --help   help for wait-events
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg perf](dbxcli_pg_perf.md)	 - Performance analysis

