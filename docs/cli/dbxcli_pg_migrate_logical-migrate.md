## dbxcli pg migrate logical-migrate

Logical migration via pub/sub

### Synopsis

Perform a logical migration using publication/subscription between source and target.

```
dbxcli pg migrate logical-migrate [flags]
```

### Examples

```
  dbxcli pg migrate logical-migrate source=old-db target=new-db --target prod-pg
```

### Options

```
  -h, --help   help for logical-migrate
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg migrate](dbxcli_pg_migrate.md)	 - Migration operations

