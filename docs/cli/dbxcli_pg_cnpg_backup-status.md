## dbxcli pg cnpg backup-status

Show CNPG backup status

### Synopsis

Show backup status for a CloudNativePG cluster.

```
dbxcli pg cnpg backup-status [flags]
```

### Examples

```
  dbxcli pg cnpg backup-status name=my-cluster namespace=default --target prod-pg
```

### Options

```
  -h, --help   help for backup-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg cnpg](dbxcli_pg_cnpg.md)	 - CloudNativePG

