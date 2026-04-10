## dbxcli pg rag collection-drop

Drop vector collection

### Synopsis

Drop a vector collection and all its data.

```
dbxcli pg rag collection-drop [flags]
```

### Examples

```
  dbxcli pg rag collection-drop name=embeddings --target prod-pg
```

### Options

```
  -h, --help   help for collection-drop
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg rag](dbxcli_pg_rag.md)	 - RAG/pgvector

