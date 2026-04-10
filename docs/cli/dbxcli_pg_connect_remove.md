## dbxcli pg connect remove

Remove a connection profile

### Synopsis

Remove an existing PostgreSQL connection profile by name.

```
dbxcli pg connect remove [flags]
```

### Examples

```
  dbxcli pg connect remove name=prod --target prod-pg
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg connect](dbxcli_pg_connect.md)	 - PostgreSQL connection management

