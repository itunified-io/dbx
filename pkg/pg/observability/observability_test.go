package observability_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/observability"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitEvents(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT wait_event_type").
		WillReturnRows(pgxmock.NewRows([]string{
			"wait_event_type", "wait_event", "count",
		}).
			AddRow("LWLock", "WALWriteLock", int64(5)).
			AddRow("IO", "DataFileRead", int64(12)))

	events, err := observability.WaitEvents(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, "IO", events[1].WaitEventType)
}

func TestCheckpointStats(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT checkpoints_timed").
		WillReturnRows(pgxmock.NewRows([]string{
			"checkpoints_timed", "checkpoints_req", "buffers_checkpoint",
			"buffers_clean", "buffers_backend", "maxwritten_clean",
		}).
			AddRow(int64(150), int64(3), int64(200000), int64(5000), int64(15000), int64(2)))

	stats, err := observability.CheckpointStats(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(150), stats.CheckpointsTimed)
	assert.Equal(t, int64(3), stats.CheckpointsReq)
}

func TestTableIOStats(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT schemaname, relname").
		WithArgs("public").
		WillReturnRows(pgxmock.NewRows([]string{
			"schemaname", "relname", "heap_blks_read", "heap_blks_hit",
			"cache_hit_ratio", "idx_blks_read", "idx_blks_hit",
		}).
			AddRow("public", "orders", int64(5000), int64(95000), float64(95.0), int64(1000), int64(49000)))

	results, err := observability.TableIOStats(context.Background(), mock, map[string]any{"schema": "public"})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, float64(95.0), results[0].CacheHitRatio)
}
