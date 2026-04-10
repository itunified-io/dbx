## dbxcli db schema list

List schemas with object counts

### Synopsis

List all schemas with counts of tables, indexes, views, packages, and other objects.

```
dbxcli db schema list [flags]
```

### Examples

```
  dbxcli db schema list --target prod-db
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

* [dbxcli db schema](dbxcli_db_schema.md)	 - Oracle schema browser

