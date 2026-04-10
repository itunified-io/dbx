## dbxcli serve

Start REST API server

### Synopsis

Start the dbx REST API server. Routes are auto-generated from the Cobra command tree.
The API exposes all CLI operations as HTTP endpoints with JWT authentication.

```
dbxcli serve [flags]
```

### Examples

```
  dbxcli serve
  dbxcli serve --port 9090 --auth-mode basic
```

### Options

```
      --auth-mode string   auth mode: jwt, basic, none (default "jwt")
  -h, --help               help for serve
      --port int           listen port (default 8080)
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
```

### SEE ALSO

* [dbxcli](dbxcli.md)	 - dbx — multi-database management platform

