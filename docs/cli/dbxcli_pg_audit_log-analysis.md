## dbxcli pg audit log-analysis

Analyze PostgreSQL logs

### Synopsis

Analyze PostgreSQL server logs for errors, warnings, and patterns.

```
dbxcli pg audit log-analysis [flags]
```

### Examples

```
  dbxcli pg audit log-analysis --target prod-pg
```

### Options

```
  -h, --help   help for log-analysis
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg audit](dbxcli_pg_audit.md)	 - Audit logging

