package health_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/health"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutovacuumCheck(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT schemaname, relname").
		WillReturnRows(pgxmock.NewRows([]string{
			"schemaname", "relname", "n_dead_tup", "last_autovacuum", "autovacuum_count",
			"n_live_tup", "dead_ratio",
		}).
			AddRow("public", "orders", int64(25000), nil, int64(0),
				int64(100000), float64(20.0)))

	results, err := health.AutovacuumCheck(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "orders", results[0].TableName)
	assert.Greater(t, results[0].DeadRatio, 10.0)
}

func TestConnectionLimits(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT").
		WillReturnRows(pgxmock.NewRows([]string{
			"max_connections", "current_connections", "reserved_connections",
			"available_connections", "utilization_pct",
		}).
			AddRow(int32(200), int32(45), int32(3), int32(152), float64(22.5)))

	result, err := health.ConnectionLimits(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Equal(t, int32(200), result.MaxConnections)
	assert.Equal(t, int32(45), result.CurrentConnections)
	assert.Less(t, result.UtilizationPct, 50.0)
}
