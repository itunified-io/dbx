## dbxcli pg dba-adv unused-indexes

List unused indexes

### Synopsis

List indexes with zero or very low scan counts.

```
dbxcli pg dba-adv unused-indexes [flags]
```

### Examples

```
  dbxcli pg dba-adv unused-indexes --target prod-pg
```

### Options

```
  -h, --help   help for unused-indexes
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dba-adv](dbxcli_pg_dba-adv.md)	 - Advanced DBA operations

