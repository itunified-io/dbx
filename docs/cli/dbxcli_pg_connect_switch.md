## dbxcli pg connect switch

Switch active connection

### Synopsis

Switch the active PostgreSQL connection to a different profile.

```
dbxcli pg connect switch [flags]
```

### Examples

```
  dbxcli pg connect switch name=staging --target prod-pg
```

### Options

```
  -h, --help   help for switch
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg connect](dbxcli_pg_connect.md)	 - PostgreSQL connection management

