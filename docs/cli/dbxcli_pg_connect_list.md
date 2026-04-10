## dbxcli pg connect list

List all connection profiles

### Synopsis

List all registered PostgreSQL connection profiles.

```
dbxcli pg connect list [flags]
```

### Examples

```
  dbxcli pg connect list --target prod-pg
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg connect](dbxcli_pg_connect.md)	 - PostgreSQL connection management

