// Package rbac provides PostgreSQL role-based access control tools.
package rbac

import (
	"context"
	"fmt"
	"time"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// Role represents a PostgreSQL role.
type Role struct {
	RoleName       string     `json:"role_name"`
	IsSuperuser    bool       `json:"is_superuser"`
	CanCreateRole  bool       `json:"can_create_role"`
	CanCreateDB    bool       `json:"can_create_db"`
	CanLogin       bool       `json:"can_login"`
	IsReplication  bool       `json:"is_replication"`
	ConnLimit      int32      `json:"conn_limit"`
	ValidUntil     *time.Time `json:"valid_until"`
}

// Grant represents a table privilege grant.
type Grant struct {
	Grantee       string `json:"grantee"`
	TableSchema   string `json:"table_schema"`
	TableName     string `json:"table_name"`
	PrivilegeType string `json:"privilege_type"`
	IsGrantable   string `json:"is_grantable"`
}

// RLSPolicy represents a row-level security policy.
type RLSPolicy struct {
	SchemaName string `json:"schema"`
	TableName  string `json:"table"`
	PolicyName string `json:"policy_name"`
	Permissive string `json:"permissive"`
	Roles      string `json:"roles"`
	Command    string `json:"cmd"`
	Qual       string `json:"qual"`
	WithCheck  string `json:"with_check"`
}

// RLSAuditResult represents a table's RLS status.
type RLSAuditResult struct {
	SchemaName string `json:"schema"`
	TableName  string `json:"table"`
	RLSEnabled bool   `json:"rls_enabled"`
	PolicyCount int   `json:"policy_count"`
}

const sqlRoleList = `
SELECT rolname, rolsuper, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication,
       rolconnlimit, rolvaliduntil
FROM pg_roles WHERE rolname NOT LIKE 'pg_%' ORDER BY rolname`

const sqlGrantMatrix = `
SELECT grantee, table_schema, table_name, privilege_type, is_grantable
FROM information_schema.table_privileges
WHERE table_schema = $1 ORDER BY grantee, table_name, privilege_type`

const sqlRLSPolicies = `
SELECT schemaname, tablename, policyname, permissive, roles::text, cmd,
       COALESCE(qual, '') AS qual, COALESCE(with_check, '') AS with_check
FROM pg_policies WHERE schemaname = $1 ORDER BY tablename, policyname`

const sqlRLSAudit = `
SELECT n.nspname, c.relname, c.relrowsecurity,
       (SELECT count(*) FROM pg_policies p WHERE p.schemaname = n.nspname AND p.tablename = c.relname) AS policy_count
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'r' AND n.nspname = $1
ORDER BY c.relname`

// RoleList returns all non-system roles.
func RoleList(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]Role, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlRoleList)
	if err != nil {
		return nil, fmt.Errorf("pg role list: %w", err)
	}
	defer rows.Close()
	var results []Role
	for rows.Next() {
		var r Role
		if err := rows.Scan(&r.RoleName, &r.IsSuperuser, &r.CanCreateRole, &r.CanCreateDB,
			&r.CanLogin, &r.IsReplication, &r.ConnLimit, &r.ValidUntil); err != nil {
			return nil, fmt.Errorf("pg role list scan: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GrantMatrix returns all table privileges in a schema.
func GrantMatrix(ctx context.Context, q pginternal.Querier, params map[string]any) ([]Grant, error) {
	schema, _ := params["schema"].(string)
	if schema == "" {
		schema = "public"
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlGrantMatrix, schema)
	if err != nil {
		return nil, fmt.Errorf("pg grant matrix: %w", err)
	}
	defer rows.Close()
	var results []Grant
	for rows.Next() {
		var g Grant
		if err := rows.Scan(&g.Grantee, &g.TableSchema, &g.TableName,
			&g.PrivilegeType, &g.IsGrantable); err != nil {
			return nil, fmt.Errorf("pg grant matrix scan: %w", err)
		}
		results = append(results, g)
	}
	return results, rows.Err()
}

// RLSPolicies returns row-level security policies in a schema.
func RLSPolicies(ctx context.Context, q pginternal.Querier, params map[string]any) ([]RLSPolicy, error) {
	schema, _ := params["schema"].(string)
	if schema == "" {
		schema = "public"
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlRLSPolicies, schema)
	if err != nil {
		return nil, fmt.Errorf("pg rls policies: %w", err)
	}
	defer rows.Close()
	var results []RLSPolicy
	for rows.Next() {
		var p RLSPolicy
		if err := rows.Scan(&p.SchemaName, &p.TableName, &p.PolicyName,
			&p.Permissive, &p.Roles, &p.Command, &p.Qual, &p.WithCheck); err != nil {
			return nil, fmt.Errorf("pg rls policies scan: %w", err)
		}
		results = append(results, p)
	}
	return results, rows.Err()
}

// RLSAudit returns RLS status for all tables in a schema.
func RLSAudit(ctx context.Context, q pginternal.Querier, params map[string]any) ([]RLSAuditResult, error) {
	schema, _ := params["schema"].(string)
	if schema == "" {
		schema = "public"
	}
	rows, err := pginternal.QueryRows(ctx, q, sqlRLSAudit, schema)
	if err != nil {
		return nil, fmt.Errorf("pg rls audit: %w", err)
	}
	defer rows.Close()
	var results []RLSAuditResult
	for rows.Next() {
		var r RLSAuditResult
		if err := rows.Scan(&r.SchemaName, &r.TableName, &r.RLSEnabled, &r.PolicyCount); err != nil {
			return nil, fmt.Errorf("pg rls audit scan: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
