// Package vault_pg provides Vault-backed PostgreSQL credential management.
package vault_pg

import (
	"context"
	"fmt"
)

// VaultClient abstracts Vault operations for testing.
type VaultClient interface {
	ReadSecret(ctx context.Context, path string) (map[string]any, error)
	WriteSecret(ctx context.Context, path string, data map[string]any) error
}

// VaultStatusResult represents Vault PG credential status.
type VaultStatusResult struct {
	VaultPath    string `json:"vault_path"`
	HasUsername  bool   `json:"has_username"`
	HasPassword  bool   `json:"has_password"`
	LeaseTTL     string `json:"lease_ttl,omitempty"`
}

// VaultConnect connects to a PG database using credentials from Vault.
func VaultConnect(ctx context.Context, vaultClient VaultClient, vaultPath string) (*VaultStatusResult, error) {
	secret, err := vaultClient.ReadSecret(ctx, vaultPath)
	if err != nil {
		return nil, fmt.Errorf("vault connect: %w", err)
	}

	result := &VaultStatusResult{VaultPath: vaultPath}
	if _, ok := secret["username"]; ok {
		result.HasUsername = true
	}
	if _, ok := secret["password"]; ok {
		result.HasPassword = true
	}

	return result, nil
}

// VaultStatus checks the status of Vault PG credentials.
func VaultStatus(ctx context.Context, vaultClient VaultClient, vaultPath string) (*VaultStatusResult, error) {
	return VaultConnect(ctx, vaultClient, vaultPath)
}

// VaultRotate rotates PG credentials stored in Vault. Confirm-gated.
func VaultRotate(ctx context.Context, vaultClient VaultClient, vaultPath string, confirm bool) error {
	if !confirm {
		return fmt.Errorf("confirm gate: set confirm=true to rotate Vault PG credentials")
	}

	// Read current credentials
	_, err := vaultClient.ReadSecret(ctx, vaultPath)
	if err != nil {
		return fmt.Errorf("vault rotate read: %w", err)
	}

	// In a real implementation, this would:
	// 1. Generate new password
	// 2. ALTER ROLE ... PASSWORD '...'
	// 3. Write new credentials to Vault
	return fmt.Errorf("vault rotation requires live PG connection for ALTER ROLE")
}
