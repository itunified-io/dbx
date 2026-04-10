## dbxcli db session describe

Describe a session by SID

### Synopsis

Show detailed information for a specific session including SQL, wait event, and program.

```
dbxcli db session describe [flags]
```

### Examples

```
  dbxcli db session describe sid=142 --target prod-db
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

* [dbxcli db session](dbxcli_db_session.md)	 - Oracle session operations

