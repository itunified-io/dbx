## dbxcli pg repl lag

Show replication lag

### Synopsis

Show replication lag for all standbys.

```
dbxcli pg repl lag [flags]
```

### Examples

```
  dbxcli pg repl lag --target prod-pg
```

### Options

```
  -h, --help   help for lag
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg repl](dbxcli_pg_repl.md)	 - Replication management

