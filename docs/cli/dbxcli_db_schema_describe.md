## dbxcli db schema describe

Describe a specific object

### Synopsis

Show DDL and metadata for a specific database object (table, index, view, package, etc.).

```
dbxcli db schema describe [flags]
```

### Examples

```
  dbxcli db schema describe owner=HR object_name=EMPLOYEES --target prod-db
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

* [dbxcli db schema](dbxcli_db_schema.md)	 - Oracle schema browser

