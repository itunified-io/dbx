// Package security provides PostgreSQL security audit tools.
package security

import (
	"context"
	"fmt"

	pginternal "github.com/itunified-io/dbx/pkg/pg/internal"
)

// SSLConnection represents a connection's SSL status.
type SSLConnection struct {
	PID      int32   `json:"pid"`
	SSL      bool    `json:"ssl"`
	Version  *string `json:"version"`
	Cipher   *string `json:"cipher"`
	Bits     *int32  `json:"bits"`
	ClientDN string  `json:"client_dn"`
}

// HBARule represents a pg_hba.conf rule with risk assessment.
type HBARule struct {
	LineNumber int32  `json:"line_number"`
	Type       string `json:"type"`
	Database   string `json:"database"`
	UserName   string `json:"user_name"`
	Address    string `json:"address"`
	AuthMethod string `json:"auth_method"`
	RiskLevel  string `json:"risk_level"`
}

// PasswordPolicyResult represents password encryption audit results.
type PasswordPolicyResult struct {
	PasswordEncryption string `json:"password_encryption"`
	AuthDelay          string `json:"auth_delay_status"`
	WeakPasswords      int    `json:"weak_password_count"`
}

// PrivilegeEscalation represents a privilege escalation risk.
type PrivilegeEscalation struct {
	RoleName      string   `json:"role_name"`
	IsSuperuser   bool     `json:"is_superuser"`
	CanCreateDB   bool     `json:"can_create_db"`
	CanCreateRole bool     `json:"can_create_role"`
	Risks         []string `json:"risks"`
}

const sqlSSLAudit = `
SELECT s.pid, s.ssl, s.version, s.cipher, s.bits,
       COALESCE(s.client_dn, '') AS client_dn
FROM pg_stat_ssl s
JOIN pg_stat_activity a ON a.pid = s.pid
WHERE a.backend_type = 'client backend'`

const sqlHBARules = `
SELECT line_number, type,
       array_to_string(database, ',') AS database,
       array_to_string(user_name, ',') AS user_name,
       COALESCE(address, 'local') AS address,
       auth_method
FROM pg_hba_file_rules
WHERE error IS NULL
ORDER BY line_number`

// SSLAudit returns SSL status for all client connections.
func SSLAudit(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]SSLConnection, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlSSLAudit)
	if err != nil {
		return nil, fmt.Errorf("pg ssl audit: %w", err)
	}
	defer rows.Close()
	var results []SSLConnection
	for rows.Next() {
		var s SSLConnection
		if err := rows.Scan(&s.PID, &s.SSL, &s.Version, &s.Cipher, &s.Bits, &s.ClientDN); err != nil {
			return nil, fmt.Errorf("pg ssl audit scan: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// HBARules returns pg_hba.conf rules with risk classification.
func HBARules(ctx context.Context, q pginternal.Querier, _ map[string]any) ([]HBARule, error) {
	rows, err := pginternal.QueryRows(ctx, q, sqlHBARules)
	if err != nil {
		return nil, fmt.Errorf("pg hba rules: %w", err)
	}
	defer rows.Close()
	var results []HBARule
	for rows.Next() {
		var r HBARule
		if err := rows.Scan(&r.LineNumber, &r.Type, &r.Database, &r.UserName,
			&r.Address, &r.AuthMethod); err != nil {
			return nil, fmt.Errorf("pg hba rules scan: %w", err)
		}
		r.RiskLevel = classifyHBARisk(r)
		results = append(results, r)
	}
	return results, rows.Err()
}

func classifyHBARisk(r HBARule) string {
	switch {
	case r.AuthMethod == "trust":
		return "HIGH"
	case r.AuthMethod == "md5":
		return "MEDIUM"
	case r.AuthMethod == "password":
		return "HIGH"
	default:
		return "LOW"
	}
}
