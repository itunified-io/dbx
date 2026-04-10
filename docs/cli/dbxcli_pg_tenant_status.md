## dbxcli pg tenant status

Show tenant health status

### Synopsis

Show health status for a specific tenant database.

```
dbxcli pg tenant status [flags]
```

### Examples

```
  dbxcli pg tenant status tenant_id=acme-corp --target prod-pg
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg tenant](dbxcli_pg_tenant.md)	 - Multi-tenant management

