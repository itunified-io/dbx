## dbxcli db sql exec

Execute a read-only SELECT statement

### Synopsis

Execute a SELECT query and return results. DML/DDL statements are rejected.

```
dbxcli db sql exec [flags]
```

### Examples

```
  dbxcli db sql exec query="SELECT sysdate FROM dual" --target prod-db
  dbxcli db sql exec query="SELECT username, account_status FROM dba_users" --target prod-db --format json
```

### Options

```
  -h, --help   help for exec
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli db sql](dbxcli_db_sql.md)	 - Read-only SQL execution

