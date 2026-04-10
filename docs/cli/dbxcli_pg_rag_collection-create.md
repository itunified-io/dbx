## dbxcli pg rag collection-create

Create vector collection

### Synopsis

Create a new vector collection with the specified dimensions.

```
dbxcli pg rag collection-create [flags]
```

### Examples

```
  dbxcli pg rag collection-create name=embeddings dimensions=1536 --target prod-pg
```

### Options

```
  -h, --help   help for collection-create
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
      --target string   target name (from ~/.dbx/targets/)
```

### SEE ALSO

* [dbxcli pg rag](dbxcli_pg_rag.md)	 - RAG/pgvector

