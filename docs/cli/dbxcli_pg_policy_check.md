## dbxcli pg policy check

Run policy check against database

### Synopsis

Run a specific policy check against the target database.

```
dbxcli pg policy check [flags]
```

### Examples

```
  dbxcli pg policy check policy_name=no-superuser-apps --target prod-pg
```

### Options

```
  -h, --help   help for check
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg policy](dbxcli_pg_policy.md)	 - Policy engine

