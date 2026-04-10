package rag_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/rag"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVectorStoreStatus(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT extversion FROM pg_extension WHERE extname = 'vector'").
		WillReturnRows(pgxmock.NewRows([]string{"extversion"}).AddRow("0.7.0"))

	mock.ExpectQuery("SELECT count").
		WillReturnRows(pgxmock.NewRows([]string{"doc_count", "total_chunks"}).
			AddRow(int64(150), int64(4500)))

	status, err := rag.VectorStoreStatus(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Equal(t, "0.7.0", status.PgvectorVersion)
	assert.Equal(t, int64(150), status.DocumentCount)
}

func TestExceptionMatch(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT category, pattern, action, description").
		WithArgs("deadlock detected").
		WillReturnRows(pgxmock.NewRows([]string{
			"category", "pattern", "action", "description",
		}).
			AddRow("lock", "deadlock detected", "retry_transaction",
				"Deadlock detected. Retry the transaction with exponential backoff."))

	result, err := rag.ExceptionMatch(context.Background(), mock, map[string]any{
		"error_text": "deadlock detected",
	})
	require.NoError(t, err)
	assert.Equal(t, "retry_transaction", result.Action)
}
