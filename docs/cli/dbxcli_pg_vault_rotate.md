## dbxcli pg vault rotate

Rotate database credentials

### Synopsis

Rotate database credentials for the specified Vault role.

```
dbxcli pg vault rotate [flags]
```

### Examples

```
  dbxcli pg vault rotate role=app-readonly --target prod-pg
```

### Options

```
  -h, --help   help for rotate
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg vault](dbxcli_pg_vault.md)	 - Vault credential management

