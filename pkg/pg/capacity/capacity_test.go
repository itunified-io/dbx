package capacity_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/capacity"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableSizes(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT schemaname, relname").
		WithArgs(50).
		WillReturnRows(pgxmock.NewRows([]string{
			"schemaname", "relname", "row_count",
			"total_size", "table_size", "index_size", "toast_size",
		}).
			AddRow("public", "orders", int64(50000),
				"256 MB", "180 MB", "64 MB", "12 MB"))

	results, err := capacity.TableSizes(context.Background(), mock, map[string]any{"limit": 50})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "orders", results[0].TableName)
	assert.Equal(t, int64(50000), results[0].RowCount)
}

func TestGrowthForecastNotImplemented(t *testing.T) {
	_, err := capacity.GrowthForecast(context.Background(), nil, nil)
	assert.ErrorContains(t, err, "baseline snapshot")
}
