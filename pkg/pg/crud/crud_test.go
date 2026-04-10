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

func TestDeleteNoWhereDoubleConfirm(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	_, err = crud.Delete(context.Background(), mock, map[string]any{
		"schema":              "public",
		"table":               "users",
		"where":               "",
		"confirm":             true,
		"confirm_destructive": false,
	})
	assert.ErrorContains(t, err, "double-confirm required")
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
