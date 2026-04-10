## dbxcli pg schema sequences

List sequences

### Synopsis

List all sequences in the specified schema.

```
dbxcli pg schema sequences [flags]
```

### Examples

```
  dbxcli pg schema sequences schema=public --target prod-pg
```

### Options

```
  -h, --help   help for sequences
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg schema](dbxcli_pg_schema.md)	 - Schema browser

