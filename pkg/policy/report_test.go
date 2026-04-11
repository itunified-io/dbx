package policy_test

import (
	"strings"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleScanResult() *policy.ScanResult {
	return &policy.ScanResult{
		EntityName: "db-prod",
		EntityType: "host",
		PolicyName: "CIS Linux Level 1",
		Framework:  "cis",
		PolicySHA:  "abc123",
		ScanTime:   time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC),
		Duration:   5 * time.Second,
		Results: []policy.CheckResult{
			{RuleID: "1.1.1", Title: "Disable cramfs", Severity: "medium", Status: "pass", EvaluatedAt: time.Now()},
			{RuleID: "5.2.2", Title: "SSH root login", Severity: "high", Status: "fail", Actual: "yes", Expected: "no", EvaluatedAt: time.Now()},
			{RuleID: "5.2.11", Title: "Empty passwords", Severity: "critical", Status: "fail", Actual: "yes", Expected: "no", EvaluatedAt: time.Now()},
		},
		Summary: policy.ScanSummary{Total: 3, Passed: 1, Failed: 2},
	}
}

func TestReportJSON(t *testing.T) {
	data, err := policy.ReportJSON(sampleScanResult())
	require.NoError(t, err)
	assert.Contains(t, string(data), `"entity_name"`)
	assert.Contains(t, string(data), `"db-prod"`)
	assert.Contains(t, string(data), `"fail"`)
}

func TestReportHTML(t *testing.T) {
	data, err := policy.ReportHTML(sampleScanResult())
	require.NoError(t, err)
	html := string(data)
	assert.Contains(t, html, "<html>")
	assert.Contains(t, html, "CIS Linux Level 1")
	assert.Contains(t, html, "db-prod")
	assert.Contains(t, html, "FAIL")
}

func TestReportCSV(t *testing.T) {
	data, err := policy.ReportCSV(sampleScanResult())
	require.NoError(t, err)
	csv := string(data)
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	assert.Len(t, lines, 4) // header + 3 results
	assert.Contains(t, lines[0], "rule_id,title,severity,status")
}

func TestFleetReport(t *testing.T) {
	results := []*policy.ScanResult{sampleScanResult(), sampleScanResult()}
	results[1].EntityName = "db-staging"
	data, err := policy.FleetReportCSV(results)
	require.NoError(t, err)
	csv := string(data)
	assert.Contains(t, csv, "db-prod")
	assert.Contains(t, csv, "db-staging")
}
