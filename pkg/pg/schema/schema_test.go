package schema_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/schema"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT t.table_name").
		WithArgs("public").
		WillReturnRows(pgxmock.NewRows([]string{
			"table_name", "table_type", "row_estimate", "total_size", "index_size",
		}).
			AddRow("users", "BASE TABLE", int64(1500), "128 kB", "32 kB").
			AddRow("orders", "BASE TABLE", int64(50000), "8192 kB", "2048 kB"))

	tables, err := schema.TableList(context.Background(), mock, map[string]any{"schema": "public"})
	require.NoError(t, err)
	assert.Len(t, tables, 2)
	assert.Equal(t, "users", tables[0].TableName)
	assert.Equal(t, int64(1500), tables[0].RowEstimate)
}

func TestTableDescribe(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	defaultVal := "nextval('users_id_seq'::regclass)"
	maxLen255 := int32(255)
	mock.ExpectQuery("SELECT c.column_name").
		WithArgs("public", "users").
		WillReturnRows(pgxmock.NewRows([]string{
			"column_name", "data_type", "is_nullable", "column_default", "character_maximum_length",
		}).
			AddRow("id", "integer", "NO", &defaultVal, nil).
			AddRow("name", "character varying", "YES", nil, &maxLen255).
			AddRow("email", "character varying", "NO", nil, &maxLen255))

	cols, err := schema.TableDescribe(context.Background(), mock, map[string]any{"schema": "public", "table": "users"})
	require.NoError(t, err)
	assert.Len(t, cols, 3)
	assert.Equal(t, "id", cols[0].ColumnName)
	assert.Equal(t, "NO", cols[0].IsNullable)
}

func TestIndexList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT i.indexname").
		WithArgs("public", "users").
		WillReturnRows(pgxmock.NewRows([]string{
			"indexname", "indexdef", "idx_size", "idx_scan", "idx_tup_read",
		}).
			AddRow("users_pkey", "CREATE UNIQUE INDEX users_pkey ON public.users USING btree (id)", "16 kB", int64(9500), int64(9500)))

	indexes, err := schema.IndexList(context.Background(), mock, map[string]any{"schema": "public", "table": "users"})
	require.NoError(t, err)
	assert.Len(t, indexes, 1)
	assert.Equal(t, "users_pkey", indexes[0].IndexName)
}

func TestSchemaList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT n.nspname").
		WillReturnRows(pgxmock.NewRows([]string{"schema_name", "owner", "table_count"}).
			AddRow("public", "postgres", int64(12)).
			AddRow("app", "appuser", int64(8)))

	schemas, err := schema.SchemaList(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, schemas, 2)
}
