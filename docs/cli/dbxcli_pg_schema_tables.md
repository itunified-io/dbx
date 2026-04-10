## dbxcli pg schema tables

List tables

### Synopsis

List all tables in the specified schema.

```
dbxcli pg schema tables [flags]
```

### Examples

```
  dbxcli pg schema tables schema=public --target prod-pg
```

### Options

```
  -h, --help   help for tables
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg schema](dbxcli_pg_schema.md)	 - Schema browser

