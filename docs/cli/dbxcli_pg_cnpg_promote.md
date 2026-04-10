## dbxcli pg cnpg promote

Promote CNPG replica

### Synopsis

Promote a CloudNativePG replica to primary.

```
dbxcli pg cnpg promote [flags]
```

### Examples

```
  dbxcli pg cnpg promote name=my-cluster namespace=default --target prod-pg
```

### Options

```
  -h, --help   help for promote
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg cnpg](dbxcli_pg_cnpg.md)	 - CloudNativePG

