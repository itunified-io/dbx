package pg_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/policy"
	ppg "github.com/itunified-io/dbx/pkg/policy/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDB struct {
	results []map[string]any
}

func (m *mockDB) QueryRows(_ context.Context, _ string, _ ...any) ([]map[string]any, error) {
	return m.results, nil
}

func TestRegisterPGExecutors(t *testing.T) {
	eng := policy.NewEngine(policy.EngineOpts{Concurrency: 1})
	db := &mockDB{results: []map[string]any{{"ssl": "on"}}}
	ppg.RegisterPGExecutors(eng, db)

	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{Name: "CIS PostgreSQL 17", Framework: "cis", Scope: "pg_database"},
		Rules: []policy.Rule{
			{ID: "6.7", Title: "SSL enabled", Severity: "high", Check: policy.RuleCheck{
				Type: "sql", Query: "SHOW ssl", Expected: "on",
			}},
		},
		SHA256: "test",
	}
	result, err := eng.Scan(context.Background(), "pg-prod", "pg_database", p)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Summary.Passed)
}

func TestPGExecutor_HBANoTrust(t *testing.T) {
	eng := policy.NewEngine(policy.EngineOpts{Concurrency: 1})
	db := &mockDB{results: []map[string]any{}}
	ppg.RegisterPGExecutors(eng, db)

	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{Name: "CIS PG", Framework: "cis", Scope: "pg_database"},
		Rules: []policy.Rule{
			{ID: "6.2", Title: "No trust auth", Severity: "critical", Check: policy.RuleCheck{
				Type: "sql", Query: "SELECT * FROM pg_hba_file_rules WHERE auth_method = 'trust'", ExpectedEmpty: true,
			}},
		},
		SHA256: "test",
	}
	result, err := eng.Scan(context.Background(), "pg-prod", "pg_database", p)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Summary.Passed)
}
