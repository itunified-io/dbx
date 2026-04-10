## dbxcli pg vault revoke

Revoke database credentials

### Synopsis

Revoke a Vault database credential lease by ID.

```
dbxcli pg vault revoke [flags]
```

### Examples

```
  dbxcli pg vault revoke lease_id=database/creds/app-readonly/abc123 --target prod-pg
```

### Options

```
  -h, --help   help for revoke
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg vault](dbxcli_pg_vault.md)	 - Vault credential management

