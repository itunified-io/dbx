## dbxcli pg vault status

Show Vault lease status

### Synopsis

Show the current Vault lease status for database credentials.

```
dbxcli pg vault status [flags]
```

### Examples

```
  dbxcli pg vault status --target prod-pg
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg vault](dbxcli_pg_vault.md)	 - Vault credential management

