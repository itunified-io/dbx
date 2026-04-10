## dbxcli pg migrate pg-upgrade-run

Run pg_upgrade

### Synopsis

Run pg_upgrade from the old version to the new version.

```
dbxcli pg migrate pg-upgrade-run [flags]
```

### Examples

```
  dbxcli pg migrate pg-upgrade-run old_version=15 new_version=16 --target prod-pg
```

### Options

```
  -h, --help   help for pg-upgrade-run
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg migrate](dbxcli_pg_migrate.md)	 - Migration operations

