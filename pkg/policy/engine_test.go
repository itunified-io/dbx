package policy_test

import (
	"context"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExecutor returns a predetermined result for any check.
type mockExecutor struct {
	status string
	actual string
}

func (m *mockExecutor) Execute(_ context.Context, _ policy.RuleCheck) (policy.CheckResult, error) {
	return policy.CheckResult{Status: m.status, Actual: m.actual, EvaluatedAt: time.Now()}, nil
}

func TestEngine_ScanAllPass(t *testing.T) {
	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{
			Name: "test", Framework: "cis", Scope: "host",
		},
		Rules: []policy.Rule{
			{ID: "1.1", Title: "Test A", Severity: "medium", Check: policy.RuleCheck{Type: "sysctl_value", Key: "fs.suid_dumpable", Expected: "0"}},
			{ID: "1.2", Title: "Test B", Severity: "high", Check: policy.RuleCheck{Type: "sysctl_value", Key: "kernel.randomize_va_space", Expected: "2"}},
		},
		SHA256: "abc123",
	}

	eng := policy.NewEngine(policy.EngineOpts{Concurrency: 2})
	eng.RegisterExecutor("sysctl_value", &mockExecutor{status: "pass", actual: "0"})

	result, err := eng.Scan(context.Background(), "db-prod", "host", p)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Summary.Total)
	assert.Equal(t, 2, result.Summary.Passed)
	assert.Equal(t, 0, result.Summary.Failed)
}

func TestEngine_ScanWithFailures(t *testing.T) {
	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{Name: "test", Framework: "cis", Scope: "host"},
		Rules: []policy.Rule{
			{ID: "1.1", Title: "Failing", Severity: "critical", Check: policy.RuleCheck{Type: "file_content"}},
		},
		SHA256: "def456",
	}

	eng := policy.NewEngine(policy.EngineOpts{Concurrency: 1})
	eng.RegisterExecutor("file_content", &mockExecutor{status: "fail", actual: "yes"})

	result, err := eng.Scan(context.Background(), "db-prod", "host", p)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Summary.Failed)
}

func TestEngine_UnknownCheckType_ReturnsError(t *testing.T) {
	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{Name: "test", Framework: "cis", Scope: "host"},
		Rules: []policy.Rule{
			{ID: "1.1", Title: "Unknown", Severity: "medium", Check: policy.RuleCheck{Type: "nonexistent"}},
		},
		SHA256: "ghi789",
	}

	eng := policy.NewEngine(policy.EngineOpts{Concurrency: 1})
	result, err := eng.Scan(context.Background(), "db-prod", "host", p)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Summary.Errors)
	assert.Equal(t, "error", result.Results[0].Status)
}

func TestEngine_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{Name: "test", Framework: "cis", Scope: "host"},
		Rules: []policy.Rule{
			{ID: "1.1", Title: "Cancelled", Severity: "medium", Check: policy.RuleCheck{Type: "sysctl_value"}},
		},
		SHA256: "jkl012",
	}

	eng := policy.NewEngine(policy.EngineOpts{Concurrency: 1})
	eng.RegisterExecutor("sysctl_value", &mockExecutor{status: "pass"})

	_, err := eng.Scan(ctx, "db-prod", "host", p)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestEngine_DefaultConcurrency(t *testing.T) {
	eng := policy.NewEngine(policy.EngineOpts{})
	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{Name: "test", Framework: "cis", Scope: "host"},
		Rules: []policy.Rule{
			{ID: "1.1", Title: "A", Severity: "low", Check: policy.RuleCheck{Type: "cmd"}},
			{ID: "1.2", Title: "B", Severity: "low", Check: policy.RuleCheck{Type: "cmd"}},
			{ID: "1.3", Title: "C", Severity: "low", Check: policy.RuleCheck{Type: "cmd"}},
			{ID: "1.4", Title: "D", Severity: "low", Check: policy.RuleCheck{Type: "cmd"}},
			{ID: "1.5", Title: "E", Severity: "low", Check: policy.RuleCheck{Type: "cmd"}},
		},
		SHA256: "mno345",
	}
	eng.RegisterExecutor("cmd", &mockExecutor{status: "pass"})
	result, err := eng.Scan(context.Background(), "srv1", "host", p)
	require.NoError(t, err)
	assert.Equal(t, 5, result.Summary.Total)
	assert.Equal(t, 5, result.Summary.Passed)
}

func TestComplianceScore(t *testing.T) {
	sr := &policy.ScanResult{
		Summary: policy.ScanSummary{Total: 10, Passed: 8, Failed: 1, Errors: 0, Skipped: 1},
	}
	score := policy.ComplianceScore(sr)
	// 8 passed out of 9 applicable (10 - 1 skipped) = 88.88...
	assert.InDelta(t, 88.89, score, 0.1)

	// All skipped = 100%
	sr2 := &policy.ScanResult{
		Summary: policy.ScanSummary{Total: 5, Skipped: 5},
	}
	assert.Equal(t, 100.0, policy.ComplianceScore(sr2))
}
