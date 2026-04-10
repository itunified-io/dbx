## dbxcli pg perf stat-statements

Show pg_stat_statements top queries

### Synopsis

Show top queries from pg_stat_statements ordered by total time.

```
dbxcli pg perf stat-statements [flags]
```

### Examples

```
  dbxcli pg perf stat-statements --target prod-pg
```

### Options

```
  -h, --help   help for stat-statements
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg perf](dbxcli_pg_perf.md)	 - Performance analysis

