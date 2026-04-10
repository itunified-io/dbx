package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itunified-io/dbx/pkg/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(os.Getenv("HOME"), ".dbx"), cfg.DataDir)
	assert.Equal(t, "strict", cfg.OracleGateMode)
	assert.Equal(t, "file", cfg.AuditSink)
	assert.Equal(t, 8080, cfg.RESTPort)
}

func TestLoadFromYAML(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	os.WriteFile(cfgFile, []byte(`
data_dir: /opt/dbx
oracle_gate_mode: warn
audit_sink: stdout
rest_port: 9090
vault:
  address: https://vault.example.com:8200
  auth_method: approle
  role_id_env: VAULT_ROLE_ID
  secret_id_env: VAULT_SECRET_ID
`), 0644)

	cfg, err := config.Load(cfgFile)
	require.NoError(t, err)
	assert.Equal(t, "/opt/dbx", cfg.DataDir)
	assert.Equal(t, "warn", cfg.OracleGateMode)
	assert.Equal(t, "stdout", cfg.AuditSink)
	assert.Equal(t, 9090, cfg.RESTPort)
	assert.Equal(t, "https://vault.example.com:8200", cfg.Vault.Address)
	assert.Equal(t, "approle", cfg.Vault.AuthMethod)
}

func TestLoadEnvOverride(t *testing.T) {
	t.Setenv("DBX_DATA_DIR", "/tmp/dbx-test")
	t.Setenv("DBX_ORACLE_GATE_MODE", "audit-only")

	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/dbx-test", cfg.DataDir)
	assert.Equal(t, "audit-only", cfg.OracleGateMode)
}
