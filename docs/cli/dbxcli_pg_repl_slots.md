## dbxcli pg repl slots

List replication slots

### Synopsis

List all replication slots with their status and lag.

```
dbxcli pg repl slots [flags]
```

### Examples

```
  dbxcli pg repl slots --target prod-pg
```

### Options

```
  -h, --help   help for slots
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg repl](dbxcli_pg_repl.md)	 - Replication management

