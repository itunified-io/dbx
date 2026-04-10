## dbxcli db sql explain

Generate execution plan for a SELECT

### Synopsis

Generate EXPLAIN PLAN output for a SELECT statement without executing it.

```
dbxcli db sql explain [flags]
```

### Examples

```
  dbxcli db sql explain query="SELECT * FROM hr.employees WHERE department_id = 10" --target prod-db
```

### Options

```
  -h, --help   help for explain
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db sql](dbxcli_db_sql.md)	 - Read-only SQL execution

