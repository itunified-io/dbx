## dbxcli pg observe metrics

Export Prometheus-compatible metrics

### Synopsis

Export PostgreSQL metrics in Prometheus-compatible format.

```
dbxcli pg observe metrics [flags]
```

### Examples

```
  dbxcli pg observe metrics --target prod-pg
```

### Options

```
  -h, --help   help for metrics
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg observe](dbxcli_pg_observe.md)	 - Observability

