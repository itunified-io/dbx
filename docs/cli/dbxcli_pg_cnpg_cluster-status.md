## dbxcli pg cnpg cluster-status

Show CNPG cluster status

### Synopsis

Show detailed status for a specific CloudNativePG cluster.

```
dbxcli pg cnpg cluster-status [flags]
```

### Examples

```
  dbxcli pg cnpg cluster-status name=my-cluster namespace=default --target prod-pg
```

### Options

```
  -h, --help   help for cluster-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg cnpg](dbxcli_pg_cnpg.md)	 - CloudNativePG

