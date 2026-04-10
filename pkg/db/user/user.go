// Package user provides read-only Oracle user/role operations.
package user

import (
	"context"
	"database/sql"

	dbinternal "github.com/itunified-io/dbx/pkg/db/internal"
)

// ListSQL returns all database users.
const ListSQL = `SELECT username, account_status, default_tablespace,
       temporary_tablespace, profile, created,
       expiry_date, lock_date
FROM dba_users ORDER BY username`

// DescribeSQL returns detailed user info including roles.
const DescribeSQL = `SELECT u.username, u.account_status, u.default_tablespace,
       u.temporary_tablespace, u.profile, u.created,
       u.expiry_date, u.lock_date,
       (SELECT LISTAGG(granted_role, ', ') WITHIN GROUP (ORDER BY granted_role)
        FROM dba_role_privs WHERE grantee = u.username) AS roles,
       (SELECT LISTAGG(privilege, ', ') WITHIN GROUP (ORDER BY privilege)
        FROM dba_sys_privs WHERE grantee = u.username) AS sys_privs
FROM dba_users u WHERE u.username = :1`

// ProfileListSQL returns all profiles.
const ProfileListSQL = `SELECT DISTINCT profile FROM dba_profiles ORDER BY profile`

// List returns all database users.
func List(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ListSQL)
}

// Describe returns detailed user info including roles and privileges.
func Describe(ctx context.Context, db *sql.DB, username string) (map[string]any, error) {
	return dbinternal.QueryRow(ctx, db, DescribeSQL, username)
}

// ProfileList returns all database profiles.
func ProfileList(ctx context.Context, db *sql.DB) ([]map[string]any, error) {
	return dbinternal.QueryRows(ctx, db, ProfileListSQL)
}
