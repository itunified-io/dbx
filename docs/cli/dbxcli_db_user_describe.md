## dbxcli db user describe

Describe user with roles and privileges

### Synopsis

Show detailed user information including granted roles, system/object privileges, and quotas.

```
dbxcli db user describe [flags]
```

### Examples

```
  dbxcli db user describe username=SCOTT --target prod-db
```

### Options

```
  -h, --help   help for describe
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db user](dbxcli_db_user.md)	 - Oracle user operations

