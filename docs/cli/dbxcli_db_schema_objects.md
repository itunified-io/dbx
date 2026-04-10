## dbxcli db schema objects

List objects in a schema

### Synopsis

List all objects in a given schema filtered by object type.

```
dbxcli db schema objects [flags]
```

### Examples

```
  dbxcli db schema objects owner=HR --target prod-db
  dbxcli db schema objects owner=HR object_type=TABLE --target prod-db
```

### Options

```
  -h, --help   help for objects
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db schema](dbxcli_db_schema.md)	 - Oracle schema browser

