package query_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/query"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT id, name FROM users").
		WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).
			AddRow(1, "alice").
			AddRow(2, "bob"))

	result, err := query.Execute(context.Background(), mock, map[string]any{
		"sql":      "SELECT id, name FROM users",
		"max_rows": 1000,
	})
	require.NoError(t, err)
	assert.Len(t, result.Rows, 2)
	assert.Equal(t, "alice", result.Rows[0]["name"])
}

func TestExecuteRowLimit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"id"})
	for i := 0; i < 100; i++ {
		rows.AddRow(i)
	}
	mock.ExpectQuery("SELECT id FROM big_table").WillReturnRows(rows)

	result, err := query.Execute(context.Background(), mock, map[string]any{
		"sql":      "SELECT id FROM big_table",
		"max_rows": 10,
	})
	require.NoError(t, err)
	assert.Len(t, result.Rows, 10)
	assert.True(t, result.Truncated)
}

func TestExplain(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("EXPLAIN SELECT 1").
		WillReturnRows(pgxmock.NewRows([]string{"QUERY PLAN"}).
			AddRow("Result  (cost=0.00..0.01 rows=1 width=4)"))

	plan, err := query.Explain(context.Background(), mock, map[string]any{
		"sql":     "SELECT 1",
		"analyze": false,
	})
	require.NoError(t, err)
	assert.Contains(t, plan[0], "Result")
}

func TestExplainAnalyze(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("EXPLAIN ANALYZE SELECT 1").
		WillReturnRows(pgxmock.NewRows([]string{"QUERY PLAN"}).
			AddRow("Result  (cost=0.00..0.01 rows=1 width=4) (actual time=0.001..0.001 rows=1 loops=1)").
			AddRow("Planning Time: 0.021 ms").
			AddRow("Execution Time: 0.031 ms"))

	plan, err := query.Explain(context.Background(), mock, map[string]any{
		"sql":     "SELECT 1",
		"analyze": true,
	})
	require.NoError(t, err)
	assert.Len(t, plan, 3)
	assert.Contains(t, plan[0], "actual time")
}
