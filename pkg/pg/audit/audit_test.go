package audit_test

import (
	"context"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/pg/audit"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionAudit(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT pid, usename").
		WillReturnRows(pgxmock.NewRows([]string{
			"pid", "usename", "datname", "client_addr", "backend_start", "state",
		}).
			AddRow(int32(100), "appuser", "mydb", "10.0.0.5", &now, "active"))

	results, err := audit.ConnectionAudit(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "appuser", results[0].Username)
}

func TestPermissionChanges(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT grantee").
		WillReturnRows(pgxmock.NewRows([]string{
			"grantee", "object_name", "object_type", "privilege_type", "is_grantable",
		}).
			AddRow("appuser", "public.users", "TABLE", "SELECT", "NO"))

	results, err := audit.PermissionChanges(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "SELECT", results[0].PrivilegeType)
}
