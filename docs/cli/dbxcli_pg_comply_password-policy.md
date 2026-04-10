## dbxcli pg comply password-policy

Check password policies

### Synopsis

Check password policy enforcement on all database roles.

```
dbxcli pg comply password-policy [flags]
```

### Examples

```
  dbxcli pg comply password-policy --target prod-pg
```

### Options

```
  -h, --help   help for password-policy
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg comply](dbxcli_pg_comply.md)	 - Compliance checks

