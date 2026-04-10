## dbxcli pg security pg-hba

Parse pg_hba.conf rules

### Synopsis

Parse and display the active pg_hba.conf authentication rules.

```
dbxcli pg security pg-hba [flags]
```

### Examples

```
  dbxcli pg security pg-hba --target prod-pg
```

### Options

```
  -h, --help   help for pg-hba
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg security](dbxcli_pg_security.md)	 - Security audit

