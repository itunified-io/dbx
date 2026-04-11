// Package oracle provides SQL-based policy check executors for Oracle Database.
package oracle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/itunified-io/dbx/pkg/policy"
)

// DBQuerier abstracts database query execution.
type DBQuerier interface {
	QueryRows(ctx context.Context, query string, args ...any) ([]map[string]any, error)
}

// SQLExecutor evaluates SQL-based policy checks.
type SQLExecutor struct{ db DBQuerier }

func NewSQLExecutor(db DBQuerier) *SQLExecutor {
	return &SQLExecutor{db: db}
}

func (e *SQLExecutor) Execute(ctx context.Context, check policy.RuleCheck) (policy.CheckResult, error) {
	rows, err := e.db.QueryRows(ctx, check.Query)
	if err != nil {
		return policy.CheckResult{Status: "error", Message: err.Error(), EvaluatedAt: time.Now()}, nil
	}

	// expected_empty: check that query returns zero rows
	if check.ExpectedEmpty {
		if len(rows) == 0 {
			return policy.CheckResult{Status: "pass", Actual: "0 rows", EvaluatedAt: time.Now()}, nil
		}
		return policy.CheckResult{
			Status: "fail", Actual: fmt.Sprintf("%d rows returned", len(rows)),
			Expected: "0 rows", EvaluatedAt: time.Now(),
		}, nil
	}

	// expected_not_contain: check that no row value contains the string
	if check.ExpectedNotContain != "" {
		for _, row := range rows {
			for _, v := range row {
				s := fmt.Sprintf("%v", v)
				if strings.Contains(strings.ToLower(s), strings.ToLower(check.ExpectedNotContain)) {
					return policy.CheckResult{
						Status: "fail", Actual: s,
						Expected:    "must not contain: " + check.ExpectedNotContain,
						EvaluatedAt: time.Now(),
					}, nil
				}
			}
		}
		return policy.CheckResult{Status: "pass", EvaluatedAt: time.Now()}, nil
	}

	// expected: exact match (string or []string)
	if check.Expected != nil && len(rows) > 0 {
		var actual string
		for _, v := range rows[0] {
			actual = fmt.Sprintf("%v", v)
			break
		}

		switch exp := check.Expected.(type) {
		case string:
			if actual != exp {
				return policy.CheckResult{Status: "fail", Actual: actual, Expected: exp, EvaluatedAt: time.Now()}, nil
			}
		case []any:
			found := false
			var expStrs []string
			for _, e := range exp {
				s := fmt.Sprintf("%v", e)
				expStrs = append(expStrs, s)
				if actual == s {
					found = true
				}
			}
			if !found {
				return policy.CheckResult{
					Status: "fail", Actual: actual,
					Expected:    strings.Join(expStrs, " | "),
					EvaluatedAt: time.Now(),
				}, nil
			}
		}
	}

	return policy.CheckResult{Status: "pass", Actual: "matched", EvaluatedAt: time.Now()}, nil
}

// RegisterOracleExecutors registers SQL-based executors for Oracle policies.
func RegisterOracleExecutors(eng *policy.Engine, db DBQuerier) {
	eng.RegisterExecutor("sql", NewSQLExecutor(db))
}
