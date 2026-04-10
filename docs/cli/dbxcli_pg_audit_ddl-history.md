## dbxcli pg audit ddl-history

Show DDL change history

### Synopsis

Show the history of DDL changes (CREATE, ALTER, DROP) from event triggers or logs.

```
dbxcli pg audit ddl-history [flags]
```

### Examples

```
  dbxcli pg audit ddl-history --target prod-pg
```

### Options

```
  -h, --help   help for ddl-history
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg audit](dbxcli_pg_audit.md)	 - Audit logging

