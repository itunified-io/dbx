## dbxcli pg tenant list

List tenant databases

### Synopsis

List all tenant databases with their status and metadata.

```
dbxcli pg tenant list [flags]
```

### Examples

```
  dbxcli pg tenant list --target prod-pg
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

* [dbxcli pg tenant](dbxcli_pg_tenant.md)	 - Multi-tenant management

