## dbxcli pg rag ingest

Ingest documents

### Synopsis

Ingest documents into a vector collection from the specified source.

```
dbxcli pg rag ingest [flags]
```

### Examples

```
  dbxcli pg rag ingest collection=embeddings source=/path/to/docs --target prod-pg
```

### Options

```
  -h, --help   help for ingest
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg rag](dbxcli_pg_rag.md)	 - RAG/pgvector

