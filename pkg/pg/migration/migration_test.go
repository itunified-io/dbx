package migration_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/migration"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaDiff(t *testing.T) {
	mockA, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockA.Close()

	mockB, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockB.Close()

	mockA.ExpectQuery("SELECT table_name FROM information_schema.tables").
		WithArgs("public").
		WillReturnRows(pgxmock.NewRows([]string{"table_name"}).
			AddRow("users").AddRow("orders").AddRow("products"))

	mockB.ExpectQuery("SELECT table_name FROM information_schema.tables").
		WithArgs("public").
		WillReturnRows(pgxmock.NewRows([]string{"table_name"}).
			AddRow("users").AddRow("orders"))

	report, err := migration.SchemaDiff(context.Background(), mockA, mockB, map[string]any{
		"schema_a": "public",
		"schema_b": "public",
	})
	require.NoError(t, err)
	assert.Len(t, report.MissingTables, 1)
	assert.Equal(t, "products", report.MissingTables[0])
}

func TestDataMigrationRequiresConfirm(t *testing.T) {
	_, err := migration.DataMigration(context.Background(), nil, nil, map[string]any{
		"confirm": false,
	})
	assert.ErrorContains(t, err, "confirm gate")
}
