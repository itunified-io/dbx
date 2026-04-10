// Package mcp provides the MCP JSON-RPC adapter skeleton for dbx.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/itunified-io/dbx/pkg/pipeline"
)

// ToolDefinition describes an MCP tool for the tools/list response.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// Adapter is the MCP server adapter.
type Adapter struct {
	pipeline *pipeline.Pipeline
	tools    []ToolDefinition
}

// NewAdapter creates an MCP adapter.
func NewAdapter(p *pipeline.Pipeline) *Adapter {
	return &Adapter{
		pipeline: p,
		tools:    defaultTools(),
	}
}

// RegisterTool adds a tool definition to the adapter.
func (a *Adapter) RegisterTool(t ToolDefinition) {
	a.tools = append(a.tools, t)
}

// Tools returns the registered tool definitions.
func (a *Adapter) Tools() []ToolDefinition {
	return a.tools
}

// ServeStdio runs the MCP server over stdin/stdout using JSON-RPC.
func (a *Adapter) ServeStdio() error {
	return a.serve(os.Stdin, os.Stdout)
}

func (a *Adapter) serve(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	encoder := json.NewEncoder(out)

	for scanner.Scan() {
		line := scanner.Bytes()
		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			encoder.Encode(jsonRPCError(nil, -32700, "parse error"))
			continue
		}

		switch req.Method {
		case "initialize":
			encoder.Encode(jsonRPCResult(req.ID, map[string]any{
				"protocolVersion": "2024-11-05",
				"serverInfo": map[string]string{
					"name":    "dbx",
					"version": "dev",
				},
				"capabilities": map[string]any{
					"tools": map[string]bool{"listChanged": false},
				},
			}))
		case "tools/list":
			encoder.Encode(jsonRPCResult(req.ID, map[string]any{
				"tools": a.tools,
			}))
		case "tools/call":
			encoder.Encode(jsonRPCResult(req.ID, map[string]any{
				"content": []map[string]string{
					{"type": "text", "text": "tool execution not yet implemented"},
				},
			}))
		default:
			encoder.Encode(jsonRPCError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method)))
		}
	}
	return scanner.Err()
}

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func jsonRPCResult(id any, result any) jsonRPCResponse {
	return jsonRPCResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func jsonRPCError(id any, code int, message string) jsonRPCResponse {
	return jsonRPCResponse{JSONRPC: "2.0", ID: id, Error: map[string]any{"code": code, "message": message}}
}

func defaultTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "dbx_target_list",
			Description: "List all registered targets",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "dbx_target_test",
			Description: "Test connectivity to a target",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"entity_name": map[string]string{"type": "string", "description": "Target name"},
				},
				"required": []string{"entity_name"},
			},
		},
		{
			Name:        "dbx_license_status",
			Description: "Show license status",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
}
