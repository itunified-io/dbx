## dbxcli pg health check

Run comprehensive health check

### Synopsis

Run a comprehensive health check covering connections, replication, locks, and storage.

```
dbxcli pg health check [flags]
```

### Examples

```
  dbxcli pg health check --target prod-pg
```

### Options

```
  -h, --help   help for check
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg health](dbxcli_pg_health.md)	 - Cluster health

