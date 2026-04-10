## dbxcli db parameter hidden

List hidden underscore parameters

### Synopsis

List hidden (underscore-prefixed) parameters. Use with caution — these are unsupported by Oracle.

```
dbxcli db parameter hidden [flags]
```

### Examples

```
  dbxcli db parameter hidden --target prod-db
```

### Options

```
  -h, --help   help for hidden
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db parameter](dbxcli_db_parameter.md)	 - Oracle init parameter operations

