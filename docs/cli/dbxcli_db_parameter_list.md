## dbxcli db parameter list

List all visible parameters

### Synopsis

List all non-hidden initialization parameters with current values and whether they are modified.

```
dbxcli db parameter list [flags]
```

### Examples

```
  dbxcli db parameter list --target prod-db
  dbxcli db parameter list --target prod-db --format json
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

* [dbxcli db parameter](dbxcli_db_parameter.md)	 - Oracle init parameter operations

