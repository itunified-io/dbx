## dbxcli pg rag search

Similarity search

### Synopsis

Perform a similarity search against a vector collection.

```
dbxcli pg rag search [flags]
```

### Examples

```
  dbxcli pg rag search collection=embeddings query="machine learning" limit=10 --target prod-pg
```

### Options

```
  -h, --help   help for search
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg rag](dbxcli_pg_rag.md)	 - RAG/pgvector

