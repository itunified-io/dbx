## dbxcli pg audit pgaudit-status

Show pgAudit configuration

### Synopsis

Show the current pgAudit extension configuration and logging settings.

```
dbxcli pg audit pgaudit-status [flags]
```

### Examples

```
  dbxcli pg audit pgaudit-status --target prod-pg
```

### Options

```
  -h, --help   help for pgaudit-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg audit](dbxcli_pg_audit.md)	 - Audit logging

