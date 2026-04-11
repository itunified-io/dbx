package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// StreamableHandler implements the MCP Streamable HTTP Transport.
// POST /mcp/ — receive JSON-RPC messages
// GET /mcp/ — SSE event stream
type StreamableHandler struct {
	mu       sync.Mutex
	sessions map[string]*mcpSession
}

type mcpSession struct {
	events chan []byte
}

// NewStreamableHandler creates a handler for MCP streamable HTTP transport.
func NewStreamableHandler() *StreamableHandler {
	return &StreamableHandler{
		sessions: make(map[string]*mcpSession),
	}
}

func (h *StreamableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodGet:
		h.handleGetStream(w, r)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *StreamableHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"read failed"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var msg struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		ID      any    `json:"id"`
	}
	if err := json.Unmarshal(body, &msg); err != nil {
		http.Error(w, `{"error":"invalid json-rpc"}`, http.StatusBadRequest)
		return
	}

	switch msg.Method {
	case "initialize":
		resp := map[string]any{
			"jsonrpc": "2.0",
			"id":      msg.ID,
			"result": map[string]any{
				"protocolVersion": "2025-03-26",
				"capabilities": map[string]any{
					"tools": map[string]bool{"listChanged": false},
				},
				"serverInfo": map[string]string{
					"name":    "dbxctl",
					"version": "2026.4.10.1",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	default:
		// Forward to command router (placeholder for actual tool dispatch)
		resp := map[string]any{
			"jsonrpc": "2.0",
			"id":      msg.ID,
			"result":  map[string]string{"status": "ok", "method": msg.Method},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func (h *StreamableHandler) handleGetStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Send initial endpoint event per MCP spec
	fmt.Fprintf(w, "event: endpoint\ndata: /mcp/\n\n")
	flusher.Flush()

	// Hold connection open until client disconnects
	<-r.Context().Done()
}
