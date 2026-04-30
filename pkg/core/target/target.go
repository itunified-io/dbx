// Package target defines the unified target model for all database engines.
package target

import "gopkg.in/yaml.v3"

// EntityType identifies the kind of managed system.
type EntityType string

const (
	// Oracle entities — aligned with Cloud Control / OEM target_type
	// registry (ADR-0097 in itunified-io/infrastructure). String values
	// match `mgmt$target.target_type` so manifest entities round-trip
	// 1:1 with Cloud Control discovery + `emcli add_target`.
	TypeOracleDatabase   EntityType = "oracle_database"    // single-instance DB
	TypeRACDatabase      EntityType = "rac_database"       // clustered DB
	TypeOracleInstance   EntityType = "oracle_instance"    // per-node DB instance
	TypeOraclePDB        EntityType = "oracle_pdb"         // pluggable database
	TypeOracleListener   EntityType = "oracle_listener"    // TNS listener
	TypeOracleASM        EntityType = "oracle_asm"         // ASM instance
	TypeOracleHome       EntityType = "oracle_home"        // discrete Oracle home install
	TypeOracleDGTopology EntityType = "oracle_dg_topology" // Data Guard broker config
	TypeCluster          EntityType = "cluster"            // GI/CRS cluster
	// Oracle Engineered Systems
	TypeExadata EntityType = "exadata"
	TypeODA     EntityType = "oda"
	TypeZDLRA   EntityType = "zdlra"
	// Legacy alias retained for back-compat with pre-ADR-0097 dbx targets;
	// new manifests should use TypeHost (matches OEM canonical naming).
	TypeOracleHost EntityType = "oracle_host"
	// PostgreSQL
	TypePGDatabase EntityType = "pg_database"
	TypePGCluster  EntityType = "pg_cluster"
	// Generic
	TypeHost EntityType = "host"
)

// Target is the unified target model for all database engines.
type Target struct {
	Name        string     `yaml:"name"`
	Type        EntityType `yaml:"type"`
	Description string     `yaml:"description"`

	// Oracle-specific
	OracleLicense *OracleLicense `yaml:"oracle_license,omitempty"`

	// Connection endpoints
	Primary *Endpoint  `yaml:"primary,omitempty"`
	Standby *Endpoint  `yaml:"standby,omitempty"`
	Replica *Endpoint  `yaml:"replica,omitempty"`
	ASM     *Endpoint  `yaml:"asm,omitempty"`
	SSH     *SSHConfig `yaml:"ssh,omitempty"`

	// PostgreSQL-specific
	CNPG      *CNPGConfig `yaml:"cnpg,omitempty"`
	PgBouncer *Endpoint   `yaml:"pgbouncer,omitempty"`
	DR        *DRConfig   `yaml:"dr,omitempty"`

	// Shared
	Monitoring *MonitoringConfig `yaml:"monitoring,omitempty"`
	GoldenGate *Endpoint         `yaml:"goldengate,omitempty"`
	OEM        *Endpoint         `yaml:"oem,omitempty"`
}

// OracleLicense declares the Oracle edition, options, and OEM packs.
type OracleLicense struct {
	Edition  string   `yaml:"edition"`
	Options  []string `yaml:"options"`
	OEMPacks []string `yaml:"oem_packs"`
}

// Endpoint represents a connection endpoint.
type Endpoint struct {
	Host       string `yaml:"host,omitempty"`
	Port       int    `yaml:"port,omitempty"`
	Service    string `yaml:"service,omitempty"`
	Database   string `yaml:"database,omitempty"`
	SSLMode    string `yaml:"sslmode,omitempty"`
	Role       string `yaml:"role,omitempty"`
	Credential string `yaml:"credential,omitempty"`
	VaultPath  string `yaml:"vault_path,omitempty"`
	RestURL    string `yaml:"rest_url,omitempty"`
}

// SSHConfig holds SSH connection details.
type SSHConfig struct {
	Host      string `yaml:"host"`
	User      string `yaml:"user"`
	KeyPath   string `yaml:"key_path,omitempty"`
	VaultPath string `yaml:"vault_path,omitempty"`
}

// CNPGConfig holds CloudNativePG cluster metadata.
type CNPGConfig struct {
	ClusterName string `yaml:"cluster_name"`
	Namespace   string `yaml:"namespace"`
	K8sContext  string `yaml:"k8s_context"`
}

// DRConfig holds cross-cluster disaster recovery settings.
type DRConfig struct {
	RemoteCluster string `yaml:"remote_cluster"`
	RemoteContext  string `yaml:"remote_context"`
	WALArchive    string `yaml:"wal_archive"`
}

// MonitoringConfig holds monitoring agent settings.
type MonitoringConfig struct {
	AgentPort int `yaml:"agent_port"`
}

// Parse deserializes a YAML byte slice into a Target.
func Parse(data []byte) (*Target, error) {
	var t Target
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// IsOracle returns true if this target is an Oracle entity type.
func (t *Target) IsOracle() bool {
	switch t.Type {
	case TypeOracleDatabase, TypeRACDatabase, TypeOracleInstance, TypeOraclePDB,
		TypeOracleListener, TypeOracleASM, TypeOracleHome, TypeOracleDGTopology,
		TypeCluster, TypeOracleHost, TypeExadata, TypeODA, TypeZDLRA:
		return true
	}
	return false
}

// IsPostgres returns true if this target is a PostgreSQL entity type.
func (t *Target) IsPostgres() bool {
	return t.Type == TypePGDatabase || t.Type == TypePGCluster
}

// IsHost returns true if this target is a generic host.
func (t *Target) IsHost() bool {
	return t.Type == TypeHost
}
