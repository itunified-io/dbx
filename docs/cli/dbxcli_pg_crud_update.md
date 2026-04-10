## dbxcli pg crud update

Update rows

### Synopsis

Update rows matching the WHERE clause in the specified table.

```
dbxcli pg crud update [flags]
```

### Examples

```
  dbxcli pg crud update schema=public table=users set='{"active":true}' where="id=1" --target prod-pg
```

### Options

```
  -h, --help   help for update
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg crud](dbxcli_pg_crud.md)	 - Data manipulation

