## dbxcli pg cnpg switchover

Trigger CNPG switchover

### Synopsis

Trigger a switchover on a CloudNativePG cluster.

```
dbxcli pg cnpg switchover [flags]
```

### Examples

```
  dbxcli pg cnpg switchover name=my-cluster namespace=default --target prod-pg
```

### Options

```
  -h, --help   help for switchover
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg cnpg](dbxcli_pg_cnpg.md)	 - CloudNativePG

