package dba_test

import (
	"context"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/pg/dba"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now()
	wetStr := "Client"
	weStr := "ClientRead"
	mock.ExpectQuery("SELECT pid, usename").
		WillReturnRows(pgxmock.NewRows([]string{
			"pid", "usename", "application_name", "client_addr", "state",
			"wait_event_type", "wait_event", "query", "query_start",
			"state_change", "backend_type",
		}).
			AddRow(int32(123), "appuser", "myapp", "10.0.0.5", "active",
				&wetStr, &weStr, "SELECT 1", &now, &now, "client backend"))

	sessions, err := dba.SessionList(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, int32(123), sessions[0].PID)
	assert.Equal(t, "appuser", sessions[0].Username)
	assert.Equal(t, "active", sessions[0].State)
}

func TestSessionKillRequiresConfirm(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	_, err = dba.SessionKill(context.Background(), mock, map[string]any{
		"pid":     123,
		"confirm": false,
	})
	assert.ErrorContains(t, err, "confirm gate")
}

func TestBloatCheck(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT schemaname, relname").
		WillReturnRows(pgxmock.NewRows([]string{
			"schemaname", "relname", "n_live_tup", "n_dead_tup",
			"dead_ratio", "total_bytes", "bloat_bytes",
		}).
			AddRow("public", "orders", int64(50000), int64(15000),
				float64(23.1), int64(10485760), int64(2420000)))

	results, err := dba.BloatCheck(context.Background(), mock, map[string]any{"threshold": 20.0})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Greater(t, results[0].DeadRatio, 20.0)
}

func TestLockTree(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT blocked_locks.pid AS blocked_pid").
		WillReturnRows(pgxmock.NewRows([]string{
			"blocked_pid", "blocked_user", "blocking_pid", "blocking_user",
			"blocked_query", "blocking_query",
		}).
			AddRow(int32(456), "appuser", int32(123), "admin",
				"UPDATE orders SET ...", "ALTER TABLE orders ..."))

	locks, err := dba.LockTree(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, locks, 1)
	assert.Equal(t, int32(456), locks[0].BlockedPID)
	assert.Equal(t, int32(123), locks[0].BlockingPID)
}
