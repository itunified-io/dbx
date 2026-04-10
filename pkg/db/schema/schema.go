// Package schema provides read-only Oracle schema/object browser operations.
package schema

import (
	"context"
	"database/sql"

	dbinternal "github.com/itunified-io/dbx/pkg/db/internal"
)

// ListSQL returns all schemas with object counts.
const ListSQL = `SELECT owner, COUNT(*) AS object_count,
       SUM(CASE WHEN object_type = 'TABLE' THEN 1 ELSE 0 END) AS tables,
       SUM(CASE WHEN object_type = 'INDEX' THEN 1 ELSE 0 END) AS indexes,
       SUM(CASE WHEN object_type = 'VIEW' THEN 1 ELSE 0 END) AS views,
       SUM(CASE WHEN object_type IN ('PROCEDURE','FUNCTION','PACKAGE') THEN 1 ELSE 0 END) AS code_objects
FROM dba_objects
WHERE owner NOT IN ('SYS','SYSTEM','OUTLN','DBSNMP','XDB','WMSYS','CTXSYS','MDSYS','ORDSYS','ORDDATA')
GROUP BY owner ORDER BY owner`

// ObjectListSQL returns objects for a schema, optionally filtered by type.
const ObjectListSQL = `SELECT object_name, object_type, status, created, last_ddl_time
FROM dba_objects WHERE owner = :1
ORDER BY object_type, object_name`

// ObjectListByTypeSQL returns objects for a schema filtered by object type.
const ObjectListByTypeSQL = `SELECT object_name, object_type, status, created, last_ddl_time
FROM dba_objects WHERE owner = :1 AND object_type = :2
ORDER BY object_name`

// ObjectDescribeSQL returns detailed object info.
const ObjectDescribeSQL = `SELECT object_name, object_type, owner, status, created,
       last_ddl_time, timestamp, generated, temporary, secondary
FROM dba_objects WHERE owner = :1 AND object_name = :2`

// List returns all schemas with object counts.
func List(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ListSQL)
}

// ObjectList returns objects for a schema.
func ObjectList(ctx context.Context, db *sql.DB, owner string) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ObjectListSQL, owner)
}

// ObjectListByType returns objects for a schema filtered by type.
func ObjectListByType(ctx context.Context, db *sql.DB, owner, objectType string) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ObjectListByTypeSQL, owner, objectType)
}

// ObjectDescribe returns detailed info for a specific object.
func ObjectDescribe(ctx context.Context, db *sql.DB, owner, objectName string) (map[string]any, error) {
	return dbinternal.QueryRow(ctx, db, ObjectDescribeSQL, owner, objectName)
}
