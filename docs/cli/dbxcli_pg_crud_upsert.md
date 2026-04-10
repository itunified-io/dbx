## dbxcli pg crud upsert

Upsert a row

### Synopsis

Insert a row or update it on conflict using the specified conflict key.

```
dbxcli pg crud upsert [flags]
```

### Examples

```
  dbxcli pg crud upsert schema=public table=users data='{"id":1,"name":"alice"}' conflict=id --target prod-pg
```

### Options

```
  -h, --help   help for upsert
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg crud](dbxcli_pg_crud.md)	 - Data manipulation

