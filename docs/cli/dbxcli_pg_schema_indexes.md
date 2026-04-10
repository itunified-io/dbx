## dbxcli pg schema indexes

List indexes

### Synopsis

List all indexes for a table in the specified schema.

```
dbxcli pg schema indexes [flags]
```

### Examples

```
  dbxcli pg schema indexes schema=public table=users --target prod-pg
```

### Options

```
  -h, --help   help for indexes
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg schema](dbxcli_pg_schema.md)	 - Schema browser

