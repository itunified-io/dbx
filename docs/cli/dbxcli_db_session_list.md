## dbxcli db session list

List active user sessions

### Synopsis

List all active user sessions from V$SESSION excluding background processes.

```
dbxcli db session list [flags]
```

### Examples

```
  dbxcli db session list --target prod-db
  dbxcli db session list --target prod-db --format json
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

* [dbxcli db session](dbxcli_db_session.md)	 - Oracle session operations

