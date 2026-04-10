## dbxcli pg security roles

List roles and privileges

### Synopsis

List all PostgreSQL roles with their privileges and membership.

```
dbxcli pg security roles [flags]
```

### Examples

```
  dbxcli pg security roles --target prod-pg
```

### Options

```
  -h, --help   help for roles
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg security](dbxcli_pg_security.md)	 - Security audit

