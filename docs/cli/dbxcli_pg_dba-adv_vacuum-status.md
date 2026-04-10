## dbxcli pg dba-adv vacuum-status

Show vacuum status

### Synopsis

Show autovacuum and manual vacuum status for all tables.

```
dbxcli pg dba-adv vacuum-status [flags]
```

### Examples

```
  dbxcli pg dba-adv vacuum-status --target prod-pg
```

### Options

```
  -h, --help   help for vacuum-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba-adv](dbxcli_pg_dba-adv.md)	 - Advanced DBA operations

