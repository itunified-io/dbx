## dbxcli pg migrate fdw-setup

Setup foreign data wrapper

### Synopsis

Setup a foreign data wrapper to access a remote server and schema.

```
dbxcli pg migrate fdw-setup [flags]
```

### Examples

```
  dbxcli pg migrate fdw-setup server=remote-db schema=public --target prod-pg
```

### Options

```
  -h, --help   help for fdw-setup
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg migrate](dbxcli_pg_migrate.md)	 - Migration operations

