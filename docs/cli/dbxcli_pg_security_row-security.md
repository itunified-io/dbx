## dbxcli pg security row-security

Show row-level security policies

### Synopsis

Show all row-level security (RLS) policies defined on tables.

```
dbxcli pg security row-security [flags]
```

### Examples

```
  dbxcli pg security row-security --target prod-pg
```

### Options

```
  -h, --help   help for row-security
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg security](dbxcli_pg_security.md)	 - Security audit

