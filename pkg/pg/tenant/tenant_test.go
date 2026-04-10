package tenant_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/tenant"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT n.nspname AS schema").
		WillReturnRows(pgxmock.NewRows([]string{
			"schema", "owner", "table_count", "total_size",
		}).
			AddRow("tenant_a", "appuser", int64(15), "64 MB").
			AddRow("tenant_b", "appuser", int64(15), "128 MB"))

	results, err := tenant.TenantList(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "tenant_a", results[0].SchemaName)
}

func TestDriftDetect(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT 'MISSING' AS drift_type").
		WithArgs("template", "tenant_a").
		WillReturnRows(pgxmock.NewRows([]string{
			"drift_type", "table_name", "column_name",
		}).
			AddRow("MISSING", "users", "avatar_url").
			AddRow("EXTRA", "orders", "legacy_field"))

	report, err := tenant.DriftDetect(context.Background(), mock, map[string]any{
		"template": "template",
		"tenant":   "tenant_a",
	})
	require.NoError(t, err)
	assert.Len(t, report.Drifts, 2)
	assert.Equal(t, "MISSING", report.Drifts[0].DriftType)
	assert.Equal(t, "EXTRA", report.Drifts[1].DriftType)
}

func TestDriftDetectRequiresParams(t *testing.T) {
	_, err := tenant.DriftDetect(context.Background(), nil, map[string]any{})
	assert.ErrorContains(t, err, "template and tenant parameters are required")
}
