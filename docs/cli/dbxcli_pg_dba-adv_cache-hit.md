## dbxcli pg dba-adv cache-hit

Show buffer cache hit ratio

### Synopsis

Show the buffer cache hit ratio from pg_stat_database.

```
dbxcli pg dba-adv cache-hit [flags]
```

### Examples

```
  dbxcli pg dba-adv cache-hit --target prod-pg
```

### Options

```
  -h, --help   help for cache-hit
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba-adv](dbxcli_pg_dba-adv.md)	 - Advanced DBA operations

