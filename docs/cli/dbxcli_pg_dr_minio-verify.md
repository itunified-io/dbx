## dbxcli pg dr minio-verify

Verify MinIO archive integrity

### Synopsis

Verify the integrity of the MinIO WAL archive.

```
dbxcli pg dr minio-verify [flags]
```

### Examples

```
  dbxcli pg dr minio-verify --target prod-pg
```

### Options

```
  -h, --help   help for minio-verify
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg dr](dbxcli_pg_dr.md)	 - Disaster recovery

