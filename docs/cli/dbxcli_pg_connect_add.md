## dbxcli pg connect add

Register a new connection profile

### Synopsis

Register a new PostgreSQL connection profile with host, port, database, and user.

```
dbxcli pg connect add [flags]
```

### Examples

```
  dbxcli pg connect add name=prod host=db01.example.com port=5432 database=app user=admin --target prod-pg
```

### Options

```
  -h, --help   help for add
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg connect](dbxcli_pg_connect.md)	 - PostgreSQL connection management

