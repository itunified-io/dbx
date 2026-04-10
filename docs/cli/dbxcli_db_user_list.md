## dbxcli db user list

List database users

### Synopsis

List all database users with account status, default tablespace, and profile.

```
dbxcli db user list [flags]
```

### Examples

```
  dbxcli db user list --target prod-db
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

* [dbxcli db user](dbxcli_db_user.md)	 - Oracle user operations

