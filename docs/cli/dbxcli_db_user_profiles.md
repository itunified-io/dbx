## dbxcli db user profiles

List database profiles

### Synopsis

List all database profiles with their resource limits and password parameters.

```
dbxcli db user profiles [flags]
```

### Examples

```
  dbxcli db user profiles --target prod-db
```

### Options

```
  -h, --help   help for profiles
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db user](dbxcli_db_user.md)	 - Oracle user operations

