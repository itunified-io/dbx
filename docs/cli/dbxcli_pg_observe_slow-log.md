## dbxcli pg observe slow-log

Analyze slow query log

### Synopsis

Analyze the PostgreSQL slow query log for patterns.

```
dbxcli pg observe slow-log [flags]
```

### Examples

```
  dbxcli pg observe slow-log --target prod-pg
```

### Options

```
  -h, --help   help for slow-log
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg observe](dbxcli_pg_observe.md)	 - Observability

