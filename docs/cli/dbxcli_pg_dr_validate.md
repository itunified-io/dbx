## dbxcli pg dr validate

Validate DR setup

### Synopsis

Validate that the DR setup is correctly configured and operational.

```
dbxcli pg dr validate [flags]
```

### Examples

```
  dbxcli pg dr validate name=prod-dr --target prod-pg
```

### Options

```
  -h, --help   help for validate
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

