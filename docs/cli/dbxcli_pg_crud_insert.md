## dbxcli pg crud insert

Insert a row

### Synopsis

Insert a new row into the specified table.

```
dbxcli pg crud insert [flags]
```

### Examples

```
  dbxcli pg crud insert schema=public table=users data='{"name":"alice"}' --target prod-pg
```

### Options

```
  -h, --help   help for insert
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg crud](dbxcli_pg_crud.md)	 - Data manipulation

