## dbxcli pg migrate pg-upgrade-check

Pre-check pg_upgrade compatibility

### Synopsis

Run pg_upgrade compatibility checks without performing the upgrade.

```
dbxcli pg migrate pg-upgrade-check [flags]
```

### Examples

```
  dbxcli pg migrate pg-upgrade-check --target prod-pg
```

### Options

```
  -h, --help   help for pg-upgrade-check
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg migrate](dbxcli_pg_migrate.md)	 - Migration operations

