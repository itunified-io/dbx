## dbxcli pg connect status

Show current connection status

### Synopsis

Show the current active PostgreSQL connection status and details.

```
dbxcli pg connect status [flags]
```

### Examples

```
  dbxcli pg connect status --target prod-pg
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg connect](dbxcli_pg_connect.md)	 - PostgreSQL connection management

