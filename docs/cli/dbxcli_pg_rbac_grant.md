## dbxcli pg rbac grant

Grant role or privilege

### Synopsis

Grant a role or privilege to a target role.

```
dbxcli pg rbac grant [flags]
```

### Examples

```
  dbxcli pg rbac grant role=readonly target=app_user --target prod-pg
```

### Options

```
  -h, --help   help for grant
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg rbac](dbxcli_pg_rbac.md)	 - Role-based access control

