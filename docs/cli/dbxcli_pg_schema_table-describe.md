## dbxcli pg schema table-describe

Describe a table

### Synopsis

Show column details, constraints, and metadata for a specific table.

```
dbxcli pg schema table-describe [flags]
```

### Examples

```
  dbxcli pg schema table-describe schema=public table=users --target prod-pg
```

### Options

```
  -h, --help   help for table-describe
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg schema](dbxcli_pg_schema.md)	 - Schema browser

