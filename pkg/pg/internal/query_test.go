package pginternal_test

import (
	"context"
	"testing"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryRows(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	rows := pgxmock.NewRows([]string{"id", "name"}).
		AddRow(1, "alpha").
		AddRow(2, "beta")
	mock.ExpectQuery("SELECT id, name FROM test_tbl WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	result, err := pginternal.QueryRows(context.Background(), mock,
		"SELECT id, name FROM test_tbl WHERE id = $1", 1)
	require.NoError(t, err)
	defer result.Close()

	var count int
	for result.Next() {
		var id int
		var name string
		require.NoError(t, result.Scan(&id, &name))
		count++
	}
	assert.Equal(t, 2, count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryRow(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM test_tbl").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(42))

	row := pginternal.QueryRow(context.Background(), mock,
		"SELECT count(*) FROM test_tbl")
	var count int
	require.NoError(t, row.Scan(&count))
	assert.Equal(t, 42, count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExec(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectExec("INSERT INTO test_tbl").
		WithArgs(3, "gamma").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	affected, err := pginternal.Exec(context.Background(), mock,
		"INSERT INTO test_tbl (id, name) VALUES ($1, $2)", 3, "gamma")
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)
	require.NoError(t, mock.ExpectationsWereMet())
}
