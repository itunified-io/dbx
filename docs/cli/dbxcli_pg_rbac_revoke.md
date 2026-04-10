## dbxcli pg rbac revoke

Revoke role or privilege

### Synopsis

Revoke a role or privilege from a target role.

```
dbxcli pg rbac revoke [flags]
```

### Examples

```
  dbxcli pg rbac revoke role=readonly target=app_user --target prod-pg
```

### Options

```
  -h, --help   help for revoke
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg rbac](dbxcli_pg_rbac.md)	 - Role-based access control

