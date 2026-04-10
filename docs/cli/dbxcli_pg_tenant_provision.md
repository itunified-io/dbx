## dbxcli pg tenant provision

Provision new tenant database

### Synopsis

Provision a new tenant database with schema and role setup.

```
dbxcli pg tenant provision [flags]
```

### Examples

```
  dbxcli pg tenant provision tenant_id=acme-corp --target prod-pg
```

### Options

```
  -h, --help   help for provision
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg tenant](dbxcli_pg_tenant.md)	 - Multi-tenant management

