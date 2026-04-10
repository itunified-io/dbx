## dbxcli pg crud delete

Delete rows

### Synopsis

Delete rows matching the WHERE clause from the specified table.

```
dbxcli pg crud delete [flags]
```

### Examples

```
  dbxcli pg crud delete schema=public table=users where="id=1" --target prod-pg
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg crud](dbxcli_pg_crud.md)	 - Data manipulation

