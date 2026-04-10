## dbxcli pg rbac effective-perms

Show effective permissions

### Synopsis

Show effective permissions for a role including inherited grants.

```
dbxcli pg rbac effective-perms [flags]
```

### Examples

```
  dbxcli pg rbac effective-perms role=app_user --target prod-pg
```

### Options

```
  -h, --help   help for effective-perms
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg rbac](dbxcli_pg_rbac.md)	 - Role-based access control

