package dba_advanced_test

import (
	"context"
	"testing"

	dba_advanced "github.com/itunified-io/dbx/pkg/pg/dba_advanced"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT c.relname AS partition_name").
		WithArgs("orders").
		WillReturnRows(pgxmock.NewRows([]string{
			"partition_name", "partition_bound", "size", "row_estimate",
		}).
			AddRow("orders_2026_01", "FOR VALUES FROM ('2026-01-01') TO ('2026-02-01')", "128 MB", int64(50000)).
			AddRow("orders_2026_02", "FOR VALUES FROM ('2026-02-01') TO ('2026-03-01')", "96 MB", int64(35000)))

	results, err := dba_advanced.PartitionList(context.Background(), mock, map[string]any{"parent": "orders"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "orders_2026_01", results[0].PartitionName)
}

func TestMatviewRefreshRequiresConfirm(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	_, err = dba_advanced.MatviewRefresh(context.Background(), mock, map[string]any{
		"schema":  "public",
		"name":    "mv_summary",
		"confirm": false,
	})
	assert.ErrorContains(t, err, "confirm gate")
}

func TestMatviewList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT schemaname, matviewname").
		WithArgs("public").
		WillReturnRows(pgxmock.NewRows([]string{
			"schemaname", "matviewname", "matviewowner", "ispopulated", "size",
		}).
			AddRow("public", "mv_summary", "postgres", true, "2048 kB"))

	results, err := dba_advanced.MatviewList(context.Background(), mock, map[string]any{"schema": "public"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "mv_summary", results[0].Name)
	assert.True(t, results[0].Populated)
}
