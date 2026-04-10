## dbxcli pg tenant deprovision

Remove tenant database

### Synopsis

Remove a tenant database and all associated resources.

```
dbxcli pg tenant deprovision [flags]
```

### Examples

```
  dbxcli pg tenant deprovision tenant_id=acme-corp --target prod-pg
```

### Options

```
  -h, --help   help for deprovision
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg tenant](dbxcli_pg_tenant.md)	 - Multi-tenant management

