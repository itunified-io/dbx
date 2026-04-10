## dbxcli pg dba terminate

Terminate a backend

### Synopsis

Terminate a backend process by PID (pg_terminate_backend).

```
dbxcli pg dba terminate [flags]
```

### Examples

```
  dbxcli pg dba terminate pid=12345 --target prod-pg
```

### Options

```
  -h, --help   help for terminate
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba](dbxcli_pg_dba.md)	 - DBA operations

