## dbxcli pg security ssl-status

Show SSL/TLS connection status

### Synopsis

Show SSL/TLS connection status for all active connections.

```
dbxcli pg security ssl-status [flags]
```

### Examples

```
  dbxcli pg security ssl-status --target prod-pg
```

### Options

```
  -h, --help   help for ssl-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg security](dbxcli_pg_security.md)	 - Security audit

