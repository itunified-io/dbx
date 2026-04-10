## dbxcli pg comply privilege-audit

Audit excessive privileges

### Synopsis

Audit roles for excessive privileges (superuser, CREATEROLE, etc.).

```
dbxcli pg comply privilege-audit [flags]
```

### Examples

```
  dbxcli pg comply privilege-audit --target prod-pg
```

### Options

```
  -h, --help   help for privilege-audit
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg comply](dbxcli_pg_comply.md)	 - Compliance checks

