package vault_pg_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/itunified-io/dbx/pkg/pg/vault_pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockVaultClient struct {
	secrets map[string]map[string]any
}

func (m *mockVaultClient) ReadSecret(_ context.Context, path string) (map[string]any, error) {
	if s, ok := m.secrets[path]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("secret not found: %s", path)
}

func (m *mockVaultClient) WriteSecret(_ context.Context, path string, data map[string]any) error {
	m.secrets[path] = data
	return nil
}

func TestVaultConnect(t *testing.T) {
	client := &mockVaultClient{
		secrets: map[string]map[string]any{
			"secret/data/pg/mydb": {
				"username": "appuser",
				"password": "secret123",
			},
		},
	}

	result, err := vault_pg.VaultConnect(context.Background(), client, "secret/data/pg/mydb")
	require.NoError(t, err)
	assert.True(t, result.HasUsername)
	assert.True(t, result.HasPassword)
	assert.Equal(t, "secret/data/pg/mydb", result.VaultPath)
}

func TestVaultConnectNotFound(t *testing.T) {
	client := &mockVaultClient{secrets: map[string]map[string]any{}}
	_, err := vault_pg.VaultConnect(context.Background(), client, "secret/data/pg/missing")
	assert.ErrorContains(t, err, "secret not found")
}

func TestVaultRotateRequiresConfirm(t *testing.T) {
	client := &mockVaultClient{
		secrets: map[string]map[string]any{
			"secret/data/pg/mydb": {"username": "appuser", "password": "old"},
		},
	}
	err := vault_pg.VaultRotate(context.Background(), client, "secret/data/pg/mydb", false)
	assert.ErrorContains(t, err, "confirm gate")
}
