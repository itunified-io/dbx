package compliance_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/compliance"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSLCompliance(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT count").
		WillReturnRows(pgxmock.NewRows([]string{
			"total", "ssl_count", "non_ssl_count",
		}).AddRow(10, 10, 0))

	result, err := compliance.SSLCompliance(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.True(t, result.Compliant)
	assert.Equal(t, 0, result.NonSSLCount)
}

func TestGDPRCheck(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT table_schema, table_name").
		WillReturnRows(pgxmock.NewRows([]string{
			"table_schema", "table_name", "column_name", "data_type",
		}).
			AddRow("public", "users", "email", "character varying").
			AddRow("public", "users", "phone_number", "character varying"))

	results, err := compliance.GDPRCheck(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestCISScan(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT setting FROM pg_settings").
		WillReturnRows(pgxmock.NewRows([]string{"setting"}).AddRow("on"))

	report, err := compliance.CISScan(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, report.Passed)
	assert.Equal(t, 0, report.Failed)
}
