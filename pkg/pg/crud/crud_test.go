package crud_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/crud"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec("INSERT INTO \"public\".\"users\"").
		WithArgs("alice@example.com", "alice").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	result, err := crud.Insert(context.Background(), mock, map[string]any{
		"schema":  "public",
		"table":   "users",
		"data":    map[string]any{"name": "alice", "email": "alice@example.com"},
		"confirm": true,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestInsertNoConfirm(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	_, err = crud.Insert(context.Background(), mock, map[string]any{
		"schema":  "public",
		"table":   "users",
		"data":    map[string]any{"name": "alice"},
		"confirm": false,
	})
	assert.ErrorContains(t, err, "confirm gate")
}

// ADR-0047: a no-WHERE DELETE wipes all rows. A bare confirm boolean must not
// authorize it — the caller must restate the target table name via confirm_table.
func TestDeleteNoWhereBooleanOnlyBlocks(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()
	// No ExpectExec registered: if any SQL runs, mock.Close/expectations would flag it.

	_, err = crud.Delete(context.Background(), mock, map[string]any{
		"schema":  "public",
		"table":   "users",
		"where":   "",
		"confirm": true,
		// no confirm_table
	})
	assert.ErrorContains(t, err, "identifier confirmation required")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteNoWhereWrongTableBlocks(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	_, err = crud.Delete(context.Background(), mock, map[string]any{
		"schema":        "public",
		"table":         "users",
		"where":         "",
		"confirm":       true,
		"confirm_table": "orders", // wrong table
	})
	assert.ErrorContains(t, err, "identifier confirmation mismatch")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteNoWhereCorrectTableProceeds(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec("DELETE FROM \"public\".\"users\"").
		WillReturnResult(pgxmock.NewResult("DELETE", 9))

	result, err := crud.Delete(context.Background(), mock, map[string]any{
		"schema":        "public",
		"table":         "users",
		"where":         "",
		"confirm":       true,
		"confirm_table": "users", // restated identifier matches
	})
	require.NoError(t, err)
	assert.Equal(t, int64(9), result.RowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteWithWhere(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec("DELETE FROM \"public\".\"users\" WHERE id = \\$1").
		WithArgs(42).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	result, err := crud.Delete(context.Background(), mock, map[string]any{
		"schema":  "public",
		"table":   "users",
		"where":   "id = $1",
		"args":    []any{42},
		"confirm": true,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestUpsert(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec("INSERT INTO \"public\".\"users\"").
		WithArgs(1, "alice").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	result, err := crud.Upsert(context.Background(), mock, map[string]any{
		"schema":   "public",
		"table":    "users",
		"data":     map[string]any{"id": 1, "name": "alice"},
		"conflict": "id",
		"confirm":  true,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.RowsAffected)
}
