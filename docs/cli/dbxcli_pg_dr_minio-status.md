## dbxcli pg dr minio-status

Show MinIO WAL archive status

### Synopsis

Show the MinIO WAL archive status and bucket health.

```
dbxcli pg dr minio-status [flags]
```

### Examples

```
  dbxcli pg dr minio-status --target prod-pg
```

### Options

```
  -h, --help   help for minio-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

