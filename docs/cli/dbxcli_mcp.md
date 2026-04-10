## dbxcli mcp

Start MCP server

### Synopsis

Start the dbx MCP (Model Context Protocol) server for AI integration.
Supports stdio transport (default) and SSE transport for remote connections.

```
dbxcli mcp [flags]
```

### Examples

```
  dbxcli mcp                  # stdio transport (for Claude Code, IDE extensions)
  dbxcli mcp --port 3001      # SSE transport (for remote MCP clients)
```

### Options

```
  -h, --help       help for mcp
      --port int   SSE transport port (0 = stdio)
```

### Options inherited from parent commands

```
      --format string   output format: table, json, yaml (default "table")
```

### SEE ALSO

* [dbxcli](dbxcli.md)	 - dbx — multi-database management platform

