package replication_test

import (
	"context"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/pg/replication"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamingStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT client_addr").
		WillReturnRows(pgxmock.NewRows([]string{
			"client_addr", "state", "sent_lsn", "write_lsn", "flush_lsn",
			"replay_lsn", "replay_lag_bytes", "sync_state", "reply_time",
		}).
			AddRow("10.0.0.5", "streaming", "0/5000000", "0/5000000",
				"0/4F00000", "0/4F00000", int64(1048576), "async", &now))

	results, err := replication.StreamingStatus(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "streaming", results[0].State)
}

func TestSlotList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT slot_name").
		WillReturnRows(pgxmock.NewRows([]string{
			"slot_name", "plugin", "slot_type", "database", "active",
			"restart_lsn", "confirmed_flush_lsn", "retained_bytes",
		}).
			AddRow("my_slot", "pgoutput", "logical", "mydb", true,
				"0/4000000", "0/4500000", int64(524288)))

	results, err := replication.SlotList(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "my_slot", results[0].SlotName)
	assert.True(t, results[0].Active)
}

func TestPublicationList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT pubname").
		WillReturnRows(pgxmock.NewRows([]string{
			"pubname", "pubowner", "puballtables", "pubinsert", "pubupdate", "pubdelete",
		}).
			AddRow("my_pub", "postgres", true, true, true, true))

	results, err := replication.PublicationList(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "my_pub", results[0].PubName)
}
