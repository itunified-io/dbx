package wal_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/wal"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWALStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT pg_current_wal_lsn").
		WillReturnRows(pgxmock.NewRows([]string{
			"current_lsn", "current_segment", "total_wal_bytes",
		}).AddRow("0/5000000", "000000010000000000000005", int64(83886080)))

	result, err := wal.WALStatus(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Equal(t, "0/5000000", result.CurrentLSN)
	assert.Equal(t, int64(83886080), result.TotalWALBytes)
}

func TestArchiveStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	lastWAL := "000000010000000000000004"
	mock.ExpectQuery("SELECT archived_count").
		WillReturnRows(pgxmock.NewRows([]string{
			"archived_count", "failed_count", "last_archived_wal",
			"last_archived_time", "last_failed_wal", "last_failed_time",
		}).AddRow(int64(1500), int64(2), &lastWAL, nil, nil, nil))

	result, err := wal.ArchiveStatus(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1500), result.ArchivedCount)
	assert.Equal(t, int64(2), result.FailedCount)
}
