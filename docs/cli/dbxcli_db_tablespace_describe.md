## dbxcli db tablespace describe

Describe tablespace datafiles

### Synopsis

Show datafiles belonging to a tablespace with sizes, autoextend status, and paths.

```
dbxcli db tablespace describe [flags]
```

### Examples

```
  dbxcli db tablespace describe name=USERS --target prod-db
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

* [dbxcli db tablespace](dbxcli_db_tablespace.md)	 - Oracle tablespace operations

