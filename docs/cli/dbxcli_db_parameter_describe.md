## dbxcli db parameter describe

Describe a specific parameter

### Synopsis

Show full details for a parameter including description, default, range, and whether it requires restart.

```
dbxcli db parameter describe [flags]
```

### Examples

```
  dbxcli db parameter describe name=sga_target --target prod-db
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db parameter](dbxcli_db_parameter.md)	 - Oracle init parameter operations

