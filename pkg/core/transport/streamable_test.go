package transport_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/itunified-io/dbx/pkg/core/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamableHandler_PostMessage(t *testing.T) {
	handler := transport.NewStreamableHandler()

	body := `{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2025-03-26","capabilities":{}}}`
	req := httptest.NewRequest("POST", "/mcp/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	respBody, err := io.ReadAll(rec.Body)
	require.NoError(t, err)
	assert.Contains(t, string(respBody), "protocolVersion")
}

func TestStreamableHandler_GetStream(t *testing.T) {
	handler := transport.NewStreamableHandler()

	// Use a real HTTP server so that context cancellation and flushing work properly.
	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, "GET", srv.URL+"/mcp/", nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	cancel() // stop the SSE stream
}
