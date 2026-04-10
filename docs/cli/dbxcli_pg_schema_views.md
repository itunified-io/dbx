## dbxcli pg schema views

List views

### Synopsis

List all views in the specified schema.

```
dbxcli pg schema views [flags]
```

### Examples

```
  dbxcli pg schema views schema=public --target prod-pg
```

### Options

```
  -h, --help   help for views
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg schema](dbxcli_pg_schema.md)	 - Schema browser

