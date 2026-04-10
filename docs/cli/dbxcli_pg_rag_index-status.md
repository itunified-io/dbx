## dbxcli pg rag index-status

Show vector index stats

### Synopsis

Show vector index statistics for a collection.

```
dbxcli pg rag index-status [flags]
```

### Examples

```
  dbxcli pg rag index-status collection=embeddings --target prod-pg
```

### Options

```
  -h, --help   help for index-status
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg rag](dbxcli_pg_rag.md)	 - RAG/pgvector

