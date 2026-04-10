## dbxcli pg wal

WAL management

### Synopsis

PostgreSQL WAL (Write-Ahead Log) management — status, archiving, size, retention, and replay lag.

### Options

```
  -h, --help   help for wal
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg](dbxcli_pg.md)	 - PostgreSQL database operations
* [dbxcli pg wal archive-status](dbxcli_pg_wal_archive-status.md)	 - Check WAL archiving status
* [dbxcli pg wal replay-lag](dbxcli_pg_wal_replay-lag.md)	 - Show WAL replay lag on standbys
* [dbxcli pg wal retention](dbxcli_pg_wal_retention.md)	 - Show WAL retention policy
* [dbxcli pg wal size](dbxcli_pg_wal_size.md)	 - Show WAL directory size
* [dbxcli pg wal status](dbxcli_pg_wal_status.md)	 - Show WAL generation stats

