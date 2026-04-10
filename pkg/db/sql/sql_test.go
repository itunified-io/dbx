package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadOnlyGuard_AllowsSelect(t *testing.T) {
	cases := []string{
		"SELECT 1 FROM dual",
		"  SELECT * FROM v$session",
		"select count(*) from dba_users",
		"WITH cte AS (SELECT 1) SELECT * FROM cte",
		"EXPLAIN PLAN FOR SELECT 1",
	}
	for _, stmt := range cases {
		assert.NoError(t, ReadOnlyGuard(stmt), "should allow: %s", stmt)
	}
}

func TestReadOnlyGuard_BlocksDML(t *testing.T) {
	cases := []struct {
		stmt   string
		prefix string
	}{
		{"INSERT INTO t VALUES (1)", "INSERT"},
		{"UPDATE t SET x = 1", "UPDATE"},
		{"DELETE FROM t", "DELETE"},
		{"DROP TABLE t", "DROP"},
		{"ALTER TABLE t ADD col NUMBER", "ALTER"},
		{"CREATE TABLE t (id NUMBER)", "CREATE"},
		{"TRUNCATE TABLE t", "TRUNCATE"},
		{"MERGE INTO t USING s ON (1=1) WHEN MATCHED THEN UPDATE SET x=1", "MERGE"},
		{"GRANT SELECT ON t TO u", "GRANT"},
		{"REVOKE SELECT ON t FROM u", "REVOKE"},
		{"PURGE RECYCLEBIN", "PURGE"},
		{"BEGIN NULL; END;", "BEGIN"},
		{"DECLARE x NUMBER; BEGIN NULL; END;", "DECLARE"},
		{"EXEC dbms_stats.gather_schema_stats('HR')", "EXEC"},
		{"CALL dbms_stats.gather_schema_stats('HR')", "CALL"},
	}
	for _, tc := range cases {
		err := ReadOnlyGuard(tc.stmt)
		require.Error(t, err, "should block: %s", tc.stmt)
		assert.Contains(t, err.Error(), tc.prefix)
	}
}

func TestReadOnlyGuard_BlocksSelectForUpdate(t *testing.T) {
	err := ReadOnlyGuard("SELECT * FROM t FOR UPDATE")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "FOR UPDATE")
}

func TestReadOnlyGuard_BlocksUnknownStatements(t *testing.T) {
	err := ReadOnlyGuard("ANALYZE TABLE t COMPUTE STATISTICS")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only SELECT/WITH/EXPLAIN")
}

func TestReadOnlyGuard_WhitespaceHandling(t *testing.T) {
	assert.NoError(t, ReadOnlyGuard("  \t SELECT 1 FROM dual"))
	assert.Error(t, ReadOnlyGuard("  \t INSERT INTO t VALUES (1)"))
}
