package audit_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/core/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventCreation(t *testing.T) {
	e := audit.NewEvent("cli", "dba@acme.com", "oracle_dg_switchover", "prod-orcl")
	assert.NotEmpty(t, e.EventID)
	assert.Equal(t, "cli", e.Interface)
	assert.Equal(t, "dba@acme.com", e.User)
	assert.Equal(t, "oracle_dg_switchover", e.Tool)
	assert.Equal(t, "prod-orcl", e.Target)
	assert.WithinDuration(t, time.Now(), e.Timestamp, 2*time.Second)
}

func TestEventComplete(t *testing.T) {
	e := audit.NewEvent("rest", "api-user", "pg_table_list", "prod-pg")
	time.Sleep(5 * time.Millisecond)
	e.Complete("success", nil)
	assert.Equal(t, "success", e.Result)
	assert.True(t, e.DurationMs > 0)
}

func TestRedaction(t *testing.T) {
	e := audit.NewEvent("cli", "admin", "target_test", "prod-orcl")
	e.Params = map[string]any{
		"entity_name": "prod-orcl",
		"password":    "secret123",
		"vault_path":  "secret/data/oracle/prod",
	}
	e.Redact([]string{"password"})
	assert.Equal(t, "***REDACTED***", e.Params["password"])
	assert.Equal(t, "prod-orcl", e.Params["entity_name"])
}

func TestStdoutSink(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := audit.NewLogger(audit.WithStdoutSink(buf))

	e := audit.NewEvent("cli", "admin", "target_list", "")
	e.Complete("success", nil)
	logger.Log(e)

	var decoded audit.Event
	err := json.NewDecoder(buf).Decode(&decoded)
	require.NoError(t, err)
	assert.Equal(t, "target_list", decoded.Tool)
	assert.Equal(t, "success", decoded.Result)
}

func TestMultiSink(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	logger := audit.NewLogger(
		audit.WithStdoutSink(buf1),
		audit.WithStdoutSink(buf2),
	)

	e := audit.NewEvent("mcp", "claude", "pg_query", "prod-pg")
	e.Complete("success", nil)
	logger.Log(e)

	assert.NotEmpty(t, buf1.String())
	assert.NotEmpty(t, buf2.String())
}
