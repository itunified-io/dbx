## dbxcli db advisor sql-tuning

List SQL tuning advisor tasks

### Synopsis

List SQL tuning advisor tasks and their recommendations (DBA_ADVISOR_TASKS).

```
dbxcli db advisor sql-tuning [flags]
```

### Examples

```
  dbxcli db advisor sql-tuning --target prod-db
```

### Options

```
  -h, --help   help for sql-tuning
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db advisor](dbxcli_db_advisor.md)	 - Oracle advisor operations

