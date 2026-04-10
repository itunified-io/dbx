## dbxcli pg policy list

List available policies

### Synopsis

List all available policies that can be checked against databases.

```
dbxcli pg policy list [flags]
```

### Examples

```
  dbxcli pg policy list --target prod-pg
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg policy](dbxcli_pg_policy.md)	 - Policy engine

