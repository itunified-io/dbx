// Package oracle orchestrates the Oracle install primitives
// (pkg/provision/install) from a single infrastructure OracleDatabase
// (DbSys) manifest, deriving the ordered per-node provisioning sequence.
//
// The install primitives (grid, dbhome, root-sh, asmca, asm-label, netca,
// dbca, pdb) each provision one step against one target. This package reads
// the high-level `kind: OracleDatabase` manifest emitted by the
// infrastructure repo (stacks/<stack>/databases/<dbsys>.yaml) and turns it
// into the ordered Plan that drives those primitives across the cluster's
// nodes — the piece the infra yaml referred to as the (previously unbuilt)
// `dbxcli provision oracle` consumer.
package oracle

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest is the subset of the infrastructure `kind: OracleDatabase`
// (DbSys) manifest needed to plan provisioning. Unknown fields are ignored,
// so it tolerates the full infra manifest without tracking every attribute.
type Manifest struct {
	Version  string   `yaml:"version"`
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`
	Spec     Spec     `yaml:"spec"`
}

// Metadata carries the DbSys identity.
type Metadata struct {
	Name string `yaml:"name"`
}

// Spec is the engine/topology/role + the component lists to provision.
type Spec struct {
	Engine    string     `yaml:"engine"`
	Edition   string     `yaml:"edition"`
	Topology  string     `yaml:"topology"` // single-instance | rac | rac-extended
	Role      string     `yaml:"role"`
	NodesRef  []string   `yaml:"nodes_ref"`
	Grid      *Grid      `yaml:"grid"`
	DBHomes   []DBHome   `yaml:"db_homes"`
	ASM       *ASM       `yaml:"asm"`
	Databases []Database `yaml:"databases"`
}

// Grid is the Grid Infrastructure home (present for rac; nil for a
// single-instance non-ASM deployment).
type Grid struct {
	Version  string `yaml:"version"`
	GridBase string `yaml:"grid_base"`
	GridHome string `yaml:"grid_home"`
}

// DBHome is one Oracle Database home. Multiple may exist (e.g. 19c + 23ai).
type DBHome struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	OracleBase string `yaml:"oracle_base"`
	OracleHome string `yaml:"oracle_home"`
}

// ASM is cluster-wide ASM storage.
type ASM struct {
	Diskgroups []Diskgroup `yaml:"diskgroups"`
}

// Diskgroup is one ASM diskgroup definition.
type Diskgroup struct {
	Name       string `yaml:"name"`
	Redundancy string `yaml:"redundancy"`
	DisksTag   string `yaml:"disks_tag"`
	AUSize     string `yaml:"au_size"`
}

// Database is one container database bound to a db_home, with its PDBs.
type Database struct {
	CDBName      string `yaml:"cdb_name"`
	DBUniqueName string `yaml:"db_unique_name"`
	DBHomeRef    string `yaml:"db_home_ref"`
	PDBs         []PDB  `yaml:"pdbs"`
}

// PDB is one pluggable database.
type PDB struct {
	Name string `yaml:"name"`
}

// IsRAC reports whether this DbSys is a (multi-node) RAC topology.
func (m *Manifest) IsRAC() bool {
	return strings.HasPrefix(strings.ToLower(m.Spec.Topology), "rac")
}

// LoadManifest reads and parses an OracleDatabase manifest from path.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", path, err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

// Validate checks the manifest is a usable OracleDatabase spec.
func (m *Manifest) Validate() error {
	if m.Kind != "OracleDatabase" {
		return fmt.Errorf("manifest kind %q is not OracleDatabase", m.Kind)
	}
	if strings.TrimSpace(m.Metadata.Name) == "" {
		return fmt.Errorf("manifest metadata.name is required")
	}
	if e := strings.ToLower(m.Spec.Engine); e != "" && e != "oracle" {
		return fmt.Errorf("engine %q is not oracle", m.Spec.Engine)
	}
	if len(m.Spec.NodesRef) == 0 {
		return fmt.Errorf("spec.nodes_ref is required (at least one node)")
	}
	if len(m.Spec.DBHomes) == 0 {
		return fmt.Errorf("spec.db_homes is required (at least one db home)")
	}
	if len(m.Spec.Databases) == 0 {
		return fmt.Errorf("spec.databases is required (at least one database)")
	}
	if m.IsRAC() && m.Spec.Grid == nil {
		return fmt.Errorf("topology %q requires spec.grid", m.Spec.Topology)
	}
	// Every database must bind a db_home that exists.
	homes := map[string]bool{}
	for _, h := range m.Spec.DBHomes {
		homes[h.Name] = true
	}
	for _, db := range m.Spec.Databases {
		if db.CDBName == "" {
			return fmt.Errorf("database is missing cdb_name")
		}
		if db.DBHomeRef != "" && !homes[db.DBHomeRef] {
			return fmt.Errorf("database %s: db_home_ref %q not found in db_homes", db.CDBName, db.DBHomeRef)
		}
	}
	return nil
}

// dbHome returns the DBHome named ref, or the first home when ref is empty.
func (m *Manifest) dbHome(ref string) (DBHome, bool) {
	if ref == "" && len(m.Spec.DBHomes) > 0 {
		return m.Spec.DBHomes[0], true
	}
	for _, h := range m.Spec.DBHomes {
		if h.Name == ref {
			return h, true
		}
	}
	return DBHome{}, false
}
