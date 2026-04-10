## dbxcli target add

Register a new target

### Synopsis

Register a new target (database, host, or service) in the target registry.

```
dbxcli target add [flags]
```

### Examples

```
  dbxcli target add entity_name=prod-db entity_type=oracle_db host=db01.example.com port=1521 service=ORCL
  dbxcli target add entity_name=web01 entity_type=oracle_host host=web01.example.com
```

### Options

```
  -h, --help   help for add
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
```

### SEE ALSO

* [dbxcli target](dbxcli_target.md)	 - Manage system targets

