// Package config provides YAML + env var configuration loading with sensible defaults.
package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// AppConfig holds all dbx configuration.
type AppConfig struct {
	DataDir        string      `yaml:"data_dir"`
	OracleGateMode string      `yaml:"oracle_gate_mode"`
	AuditSink      string      `yaml:"audit_sink"`
	RESTPort       int         `yaml:"rest_port"`
	Vault          VaultConfig `yaml:"vault"`
}

// VaultConfig holds Vault connection settings.
type VaultConfig struct {
	Address     string `yaml:"address"`
	AuthMethod  string `yaml:"auth_method"`
	RoleIDEnv   string `yaml:"role_id_env"`
	SecretIDEnv string `yaml:"secret_id_env"`
}

// Load reads config from file (if provided), then overlays env vars.
func Load(path string) (*AppConfig, error) {
	cfg := defaults()

	if path != "" {
		if err := loadFile(cfg, path); err != nil {
			return nil, err
		}
	}

	applyEnv(cfg)
	return cfg, nil
}

func defaults() *AppConfig {
	home, _ := os.UserHomeDir()
	return &AppConfig{
		DataDir:        filepath.Join(home, ".dbx"),
		OracleGateMode: "strict",
		AuditSink:      "file",
		RESTPort:       8080,
		Vault: VaultConfig{
			AuthMethod:  "approle",
			RoleIDEnv:   "VAULT_ROLE_ID",
			SecretIDEnv: "VAULT_SECRET_ID",
		},
	}
}

func loadFile(cfg *AppConfig, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

func applyEnv(cfg *AppConfig) {
	if v := os.Getenv("DBX_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}
	if v := os.Getenv("DBX_ORACLE_GATE_MODE"); v != "" {
		cfg.OracleGateMode = v
	}
	if v := os.Getenv("DBX_AUDIT_SINK"); v != "" {
		cfg.AuditSink = v
	}
	if v := os.Getenv("DBX_REST_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.RESTPort = port
		}
	}
	if v := os.Getenv("DBX_VAULT_ADDRESS"); v != "" {
		cfg.Vault.Address = v
	}
}
