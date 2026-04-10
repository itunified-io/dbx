## dbxcli pg tenant migrate

Run tenant schema migration

### Synopsis

Run schema migrations on a specific tenant database.

```
dbxcli pg tenant migrate [flags]
```

### Examples

```
  dbxcli pg tenant migrate tenant_id=acme-corp --target prod-pg
```

### Options

```
  -h, --help   help for migrate
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg tenant](dbxcli_pg_tenant.md)	 - Multi-tenant management

