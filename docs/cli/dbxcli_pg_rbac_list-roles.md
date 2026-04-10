## dbxcli pg rbac list-roles

List all roles with members

### Synopsis

List all roles with their member relationships.

```
dbxcli pg rbac list-roles [flags]
```

### Examples

```
  dbxcli pg rbac list-roles --target prod-pg
```

### Options

```
  -h, --help   help for list-roles
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg rbac](dbxcli_pg_rbac.md)	 - Role-based access control

