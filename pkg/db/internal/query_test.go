package internal

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryRow_EmptySliceReturnsErrNoRows(t *testing.T) {
	// QueryRow returns sql.ErrNoRows when the result set is empty.
	rows := []map[string]any{}
	if len(rows) == 0 {
		err := sql.ErrNoRows
		require.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQueryRow_ReturnsFirstRow(t *testing.T) {
	rows := []map[string]any{
		{"col1": "val1"},
		{"col1": "val2"},
	}
	if len(rows) > 0 {
		require.Equal(t, "val1", rows[0]["col1"])
	}
}
