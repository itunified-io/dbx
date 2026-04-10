package rbac_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/rbac"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT rolname, rolsuper").
		WillReturnRows(pgxmock.NewRows([]string{
			"rolname", "rolsuper", "rolcreaterole", "rolcreatedb",
			"rolcanlogin", "rolreplication", "rolconnlimit", "rolvaliduntil",
		}).
			AddRow("postgres", true, true, true, true, true, int32(-1), nil).
			AddRow("appuser", false, false, false, true, false, int32(10), nil))

	roles, err := rbac.RoleList(context.Background(), mock, nil)
	require.NoError(t, err)
	assert.Len(t, roles, 2)
	assert.True(t, roles[0].IsSuperuser)
	assert.False(t, roles[1].IsSuperuser)
}

func TestGrantMatrix(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT grantee, table_schema").
		WithArgs("public").
		WillReturnRows(pgxmock.NewRows([]string{
			"grantee", "table_schema", "table_name", "privilege_type", "is_grantable",
		}).
			AddRow("appuser", "public", "users", "SELECT", "NO").
			AddRow("appuser", "public", "users", "INSERT", "NO"))

	grants, err := rbac.GrantMatrix(context.Background(), mock, map[string]any{"schema": "public"})
	require.NoError(t, err)
	assert.Len(t, grants, 2)
}

func TestRLSPolicies(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	mock.ExpectQuery("SELECT schemaname, tablename").
		WithArgs("public").
		WillReturnRows(pgxmock.NewRows([]string{
			"schemaname", "tablename", "policyname", "permissive",
			"roles", "cmd", "qual", "with_check",
		}).
			AddRow("public", "orders", "tenant_isolation", "PERMISSIVE",
				"{appuser}", "ALL", "(tenant_id = current_setting('app.tenant_id')::int)", ""))

	policies, err := rbac.RLSPolicies(context.Background(), mock, map[string]any{"schema": "public"})
	require.NoError(t, err)
	assert.Len(t, policies, 1)
	assert.Equal(t, "tenant_isolation", policies[0].PolicyName)
}
