package performance_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/performance"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlowQueries(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT queryid, query").
		WithArgs(1000.0, 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"queryid", "query", "calls", "total_exec_time_ms",
			"mean_exec_time_ms", "stddev_exec_time_ms",
			"min_exec_time_ms", "max_exec_time_ms",
			"rows", "shared_blks_hit", "shared_blks_read",
		}).
			AddRow(int64(12345), "SELECT * FROM orders WHERE ...", int64(500),
				float64(125000.0), float64(2500.0), float64(500.0),
				float64(100.0), float64(15000.0),
				int64(50000), int64(900000), int64(100000)))

	results, err := performance.SlowQueries(context.Background(), mock, map[string]any{
		"min_duration_ms": 1000.0,
		"limit":           20,
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, int64(12345), results[0].QueryID)
	assert.Greater(t, results[0].MeanTime, 1000.0)
}

func TestIndexAdvisor(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT schemaname, relname").
		WillReturnRows(pgxmock.NewRows([]string{
			"schemaname", "relname", "seq_scan", "seq_tup_read",
			"idx_scan", "table_size", "seq_scan_pct",
		}).
			AddRow("public", "orders", int64(15000), int64(2000000),
				int64(500), "256 MB", float64(96.8)))

	results, err := performance.IndexAdvisor(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Contains(t, results[0].Suggestion, "sequential scans")
}
