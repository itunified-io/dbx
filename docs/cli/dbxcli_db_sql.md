## dbxcli db sql

Read-only SQL execution

### Synopsis

Execute read-only SQL statements against an Oracle database.
Only SELECT statements are permitted — DML/DDL is blocked by the ReadOnlyGuard.

### Options

```
  -h, --help   help for sql
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db](dbxcli_db.md)	 - Oracle database read-only operations
* [dbxcli db sql exec](dbxcli_db_sql_exec.md)	 - Execute a read-only SELECT statement
* [dbxcli db sql explain](dbxcli_db_sql_explain.md)	 - Generate execution plan for a SELECT

