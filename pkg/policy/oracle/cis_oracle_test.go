package oracle_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/policy"
	poracl "github.com/itunified-io/dbx/pkg/policy/oracle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDB struct {
	results []map[string]any
}

func (m *mockDB) QueryRows(_ context.Context, _ string, _ ...any) ([]map[string]any, error) {
	return m.results, nil
}

func TestSQLExecutor_ExactMatch(t *testing.T) {
	db := &mockDB{results: []map[string]any{{"value": "DB,EXTENDED"}}}
	exec := poracl.NewSQLExecutor(db)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type:     "sql",
		Query:    "SELECT value FROM v$parameter WHERE name = 'audit_trail'",
		Expected: []any{"DB", "DB,EXTENDED", "XML", "XML,EXTENDED", "OS"},
	})
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestSQLExecutor_ExactMatch_Fail(t *testing.T) {
	db := &mockDB{results: []map[string]any{{"value": "NONE"}}}
	exec := poracl.NewSQLExecutor(db)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type:     "sql",
		Query:    "SELECT value FROM v$parameter WHERE name = 'audit_trail'",
		Expected: []any{"DB", "DB,EXTENDED"},
	})
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
}

func TestSQLExecutor_StringMatch(t *testing.T) {
	db := &mockDB{results: []map[string]any{{"value": "on"}}}
	exec := poracl.NewSQLExecutor(db)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "sql", Query: "SHOW ssl", Expected: "on",
	})
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestSQLExecutor_NotContain(t *testing.T) {
	db := &mockDB{results: []map[string]any{{"value": "LISTENER_PROD"}}}
	exec := poracl.NewSQLExecutor(db)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type:               "sql",
		Query:              "SELECT value FROM v$parameter WHERE name = 'local_listener'",
		ExpectedNotContain: "extproc",
	})
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestSQLExecutor_NotContain_Fail(t *testing.T) {
	db := &mockDB{results: []map[string]any{{"value": "EXTPROC_CONNECTION_DATA"}}}
	exec := poracl.NewSQLExecutor(db)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type:               "sql",
		Query:              "SELECT value FROM v$parameter WHERE name = 'local_listener'",
		ExpectedNotContain: "extproc",
	})
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
}

func TestSQLExecutor_ExpectedEmpty_Fail(t *testing.T) {
	db := &mockDB{results: []map[string]any{{"line_number": 5, "auth_method": "trust"}}}
	exec := poracl.NewSQLExecutor(db)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type:          "sql",
		Query:         "SELECT line_number, auth_method FROM pg_hba_file_rules WHERE auth_method = 'trust'",
		ExpectedEmpty: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
}

func TestSQLExecutor_ExpectedEmpty_Pass(t *testing.T) {
	db := &mockDB{results: []map[string]any{}}
	exec := poracl.NewSQLExecutor(db)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type:          "sql",
		Query:         "SELECT * FROM sensitive_view WHERE condition = 'bad'",
		ExpectedEmpty: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestRegisterOracleExecutors(t *testing.T) {
	eng := policy.NewEngine(policy.EngineOpts{Concurrency: 1})
	db := &mockDB{results: []map[string]any{{"value": "DB,EXTENDED"}}}
	poracl.RegisterOracleExecutors(eng, db)

	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{Name: "CIS Oracle", Framework: "cis", Scope: "oracle_database"},
		Rules: []policy.Rule{
			{ID: "2.2.1", Title: "Audit trail", Severity: "high", Check: policy.RuleCheck{
				Type: "sql", Query: "SELECT value FROM v$parameter WHERE name = 'audit_trail'", Expected: "DB,EXTENDED",
			}},
		},
		SHA256: "test",
	}
	result, err := eng.Scan(context.Background(), "prod-db", "oracle_database", p)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Summary.Passed)
}
