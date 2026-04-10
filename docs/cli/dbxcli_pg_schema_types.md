## dbxcli pg schema types

List custom types

### Synopsis

List all custom types in the specified schema.

```
dbxcli pg schema types [flags]
```

### Examples

```
  dbxcli pg schema types schema=public --target prod-pg
```

### Options

```
  -h, --help   help for types
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg schema](dbxcli_pg_schema.md)	 - Schema browser

