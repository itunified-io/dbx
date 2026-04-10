## dbxcli pg comply encryption-status

Check encryption at rest

### Synopsis

Check the status of data encryption at rest.

```
dbxcli pg comply encryption-status [flags]
```

### Examples

```
  dbxcli pg comply encryption-status --target prod-pg
```

### Options

```
  -h, --help   help for encryption-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg comply](dbxcli_pg_comply.md)	 - Compliance checks

