## dbxcli pg cnpg clusters

List CNPG clusters

### Synopsis

List all CloudNativePG clusters in the specified namespace.

```
dbxcli pg cnpg clusters [flags]
```

### Examples

```
  dbxcli pg cnpg clusters namespace=default --target prod-pg
```

### Options

```
  -h, --help   help for clusters
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg cnpg](dbxcli_pg_cnpg.md)	 - CloudNativePG

