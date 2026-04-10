## dbxcli pg cnpg restart

Rolling restart CNPG cluster

### Synopsis

Perform a rolling restart of a CloudNativePG cluster.

```
dbxcli pg cnpg restart [flags]
```

### Examples

```
  dbxcli pg cnpg restart name=my-cluster namespace=default --target prod-pg
```

### Options

```
  -h, --help   help for restart
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg cnpg](dbxcli_pg_cnpg.md)	 - CloudNativePG

