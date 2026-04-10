## dbxcli pg dba reload

Reload configuration

### Synopsis

Reload PostgreSQL server configuration without restart (pg_reload_conf).

```
dbxcli pg dba reload [flags]
```

### Examples

```
  dbxcli pg dba reload --target prod-pg
```

### Options

```
  -h, --help   help for reload
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba](dbxcli_pg_dba.md)	 - DBA operations

