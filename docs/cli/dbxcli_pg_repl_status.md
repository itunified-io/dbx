## dbxcli pg repl status

Show replication status

### Synopsis

Show the current replication status from pg_stat_replication.

```
dbxcli pg repl status [flags]
```

### Examples

```
  dbxcli pg repl status --target prod-pg
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg repl](dbxcli_pg_repl.md)	 - Replication management

