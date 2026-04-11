// Package policy provides the unified compliance policy engine for CIS, STIG, and custom policies.
package policy

import (
	"context"
	"time"
)

// Policy represents a loaded YAML policy file.
type Policy struct {
	Metadata PolicyMetadata `yaml:"metadata"`
	Rules    []Rule         `yaml:"rules"`
	SHA256   string         // computed at load time
	Path     string         // file path
}

// PolicyMetadata holds policy-level information.
type PolicyMetadata struct {
	Name      string   `yaml:"name"`
	Version   string   `yaml:"version"`
	Framework string   `yaml:"framework"` // cis, stig, custom
	Scope     string   `yaml:"scope"`     // host, oracle_database, pg_database
	Distros   []string `yaml:"distros,omitempty"`
	Versions  []string `yaml:"versions,omitempty"`
}

// Rule is a single policy check.
type Rule struct {
	ID          string      `yaml:"id"`
	Title       string      `yaml:"title"`
	Severity    string      `yaml:"severity"` // critical, high, medium, low, info
	Check       RuleCheck   `yaml:"check"`
	Remediation Remediation `yaml:"remediation"`
}

// RuleCheck defines how to evaluate the rule.
type RuleCheck struct {
	Type               string `yaml:"type"` // kernel_module, file_content, sql, command_output, file_permission, sysctl_value
	Module             string `yaml:"module,omitempty"`
	Path               string `yaml:"path,omitempty"`
	Pattern            string `yaml:"pattern,omitempty"`
	Expected           any    `yaml:"expected,omitempty"`
	ExpectedNotContain string `yaml:"expected_not_contain,omitempty"`
	ExpectedEmpty      bool   `yaml:"expected_empty,omitempty"`
	Query              string `yaml:"query,omitempty"`
	Command            string `yaml:"command,omitempty"`
	Key                string `yaml:"key,omitempty"`
	Permission         string `yaml:"permission,omitempty"`
	Owner              string `yaml:"owner,omitempty"`
	Group              string `yaml:"group,omitempty"`
}

// Remediation describes how to fix a failing rule.
type Remediation struct {
	Command         string `yaml:"command,omitempty"`
	Manual          string `yaml:"manual,omitempty"`
	ConfirmRequired bool   `yaml:"confirm_required"`
}

// CheckResult is the outcome of evaluating one rule.
type CheckResult struct {
	RuleID      string    `json:"rule_id"`
	Title       string    `json:"title"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"` // pass, fail, error, skip
	Actual      string    `json:"actual,omitempty"`
	Expected    string    `json:"expected,omitempty"`
	Message     string    `json:"message,omitempty"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// ScanResult is the full result of a policy scan.
type ScanResult struct {
	EntityName string        `json:"entity_name"`
	EntityType string        `json:"entity_type"`
	PolicyName string        `json:"policy_name"`
	Framework  string        `json:"framework"`
	PolicySHA  string        `json:"policy_sha256"`
	ScanTime   time.Time     `json:"scan_time"`
	Duration   time.Duration `json:"duration"`
	Results    []CheckResult `json:"results"`
	Summary    ScanSummary   `json:"summary"`
}

// ScanSummary aggregates pass/fail counts.
type ScanSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Errors  int `json:"errors"`
	Skipped int `json:"skipped"`
}

// CheckExecutor evaluates a single rule against a target.
type CheckExecutor interface {
	Execute(ctx context.Context, check RuleCheck) (CheckResult, error)
}
