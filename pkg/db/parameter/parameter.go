// Package parameter provides read-only Oracle init parameter operations.
package parameter

import (
	"context"
	"database/sql"

	dbinternal "github.com/itunified-io/dbx/pkg/db/internal"
)

// ListSQL returns all visible parameters.
const ListSQL = `SELECT name, value, display_value, isdefault, ismodified,
       issys_modifiable, isinstance_modifiable, description
FROM v$parameter ORDER BY name`

// DescribeSQL returns detail for a specific parameter.
const DescribeSQL = `SELECT name, value, display_value, isdefault, ismodified,
       issys_modifiable, isinstance_modifiable, type, description,
       update_comment
FROM v$parameter WHERE name = :1`

// ModifiedSQL returns parameters that have been modified from default.
const ModifiedSQL = `SELECT name, value, display_value, ismodified,
       issys_modifiable, description
FROM v$parameter WHERE isdefault = 'FALSE'
ORDER BY name`

// HiddenSQL returns hidden (underscore) parameters.
const HiddenSQL = `SELECT a.ksppinm AS name, b.ksppstvl AS value,
       a.ksppdesc AS description
FROM x$ksppi a JOIN x$ksppcv b ON a.indx = b.indx
WHERE a.ksppinm LIKE '\_%' ESCAPE '\'
ORDER BY a.ksppinm`

// List returns all visible parameters.
func List(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ListSQL)
}

// Describe returns detail for a specific parameter.
func Describe(ctx context.Context, db *sql.DB, name string) (map[string]any, error) {
	return dbinternal.QueryRow(ctx, db, DescribeSQL, name)
}

// Modified returns parameters that have been changed from their default.
func Modified(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ModifiedSQL)
}

// Hidden returns hidden underscore parameters (requires SYS access to x$ tables).
func Hidden(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, HiddenSQL)
}
