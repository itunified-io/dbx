package oracle

import "fmt"

// Scope describes how a step fans out across the cluster.
type Scope string

const (
	// ScopeCluster runs once for the whole cluster (on the first node);
	// the installer/tool itself propagates to the other RAC nodes.
	ScopeCluster Scope = "cluster"
	// ScopePerNode runs once on every node (e.g. root.sh).
	ScopePerNode Scope = "per-node"
)

// Step is one ordered primitive invocation in the provisioning sequence.
// It maps 1:1 onto a `dbxcli provision install <primitive>` call; Target is
// the node the primitive runs against. The ref fields (HomeRef/DiskGroup/
// CDB/PDB/Grid) let BuildSpecs resolve the concrete install spec without
// re-deriving the plan order.
type Step struct {
	Order     int    `json:"order" yaml:"order"`
	Primitive string `json:"primitive" yaml:"primitive"` // asm-label|grid|root-sh|asmca|dbhome|netca|dbca|pdb
	Scope     Scope  `json:"scope" yaml:"scope"`
	Target    string `json:"target" yaml:"target"` // node name the primitive runs on
	Detail    string `json:"detail" yaml:"detail"`

	// Grid is true for steps that operate against the Grid home
	// (grid, grid root.sh, asm-label, and the listener when GI is present).
	Grid bool `json:"grid,omitempty" yaml:"grid,omitempty"`
	// HomeRef is the db_home name for db-home-scoped steps (dbhome, its
	// root.sh, dbca, pdb, and netca when no grid).
	HomeRef string `json:"home_ref,omitempty" yaml:"home_ref,omitempty"`
	// DiskGroup is the ASM diskgroup name for asmca steps.
	DiskGroup string `json:"disk_group,omitempty" yaml:"disk_group,omitempty"`
	// CDB is the database's DB_UNIQUE_NAME for dbca/pdb steps.
	CDB string `json:"cdb,omitempty" yaml:"cdb,omitempty"`
	// PDB is the pluggable database name for pdb steps.
	PDB string `json:"pdb,omitempty" yaml:"pdb,omitempty"`
}

// Plan derives the ordered provisioning sequence from a manifest.
//
// RAC ordering (Grid present):
//
//	asm-label (cluster) → grid (cluster) → root-sh:grid (per-node)
//	→ asmca per diskgroup (cluster) → for each db_home: dbhome (cluster)
//	→ root-sh:dbhome (per-node) → netca (cluster)
//	→ for each database: dbca (cluster) → for each pdb: pdb (cluster)
//
// Single-instance (no Grid): the grid/asm-label/asmca/grid-root.sh steps are
// omitted; dbhome runs without per-node root.sh duplication.
func Plan(m *Manifest) ([]Step, error) {
	if m == nil {
		return nil, fmt.Errorf("nil manifest")
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	nodes := m.Spec.NodesRef
	first := nodes[0]
	var steps []Step
	ord := 0
	add := func(s Step) {
		ord++
		s.Order = ord
		steps = append(steps, s)
	}

	hasGrid := m.Spec.Grid != nil

	if hasGrid {
		// 1) Label raw ASM disks (cluster-wide; discovered on all nodes).
		dgDetail := "diskgroups:"
		if m.Spec.ASM != nil {
			for _, dg := range m.Spec.ASM.Diskgroups {
				dgDetail += fmt.Sprintf(" %s(%s)", dg.Name, dg.DisksTag)
			}
		}
		add(Step{Primitive: "asm-label", Scope: ScopeCluster, Target: first, Grid: true, Detail: dgDetail})

		// 2) Grid Infrastructure runInstaller (cluster-wide).
		add(Step{Primitive: "grid", Scope: ScopeCluster, Target: first, Grid: true,
			Detail: fmt.Sprintf("GI %s home=%s", m.Spec.Grid.Version, m.Spec.Grid.GridHome)})

		// 3) root.sh for grid on EVERY node, first node first.
		for _, n := range nodes {
			add(Step{Primitive: "root-sh", Scope: ScopePerNode, Target: n, Grid: true,
				Detail: fmt.Sprintf("grid root.sh on %s (home=%s)", n, m.Spec.Grid.GridHome)})
		}

		// 4) Create ASM diskgroups (cluster-wide) — one asmca call per diskgroup.
		if m.Spec.ASM != nil {
			for _, dg := range m.Spec.ASM.Diskgroups {
				add(Step{Primitive: "asmca", Scope: ScopeCluster, Target: first, Grid: true, DiskGroup: dg.Name,
					Detail: fmt.Sprintf("create ASM diskgroup %s (%s, %s)", dg.Name, dg.Redundancy, dg.DisksTag)})
			}
		}
	}

	// 5) For each DB home: runInstaller (cluster) + root.sh per node.
	for _, h := range m.Spec.DBHomes {
		add(Step{Primitive: "dbhome", Scope: ScopeCluster, Target: first, HomeRef: h.Name,
			Detail: fmt.Sprintf("DB home %s %s home=%s", h.Name, h.Version, h.OracleHome)})
		for _, n := range nodes {
			add(Step{Primitive: "root-sh", Scope: ScopePerNode, Target: n, HomeRef: h.Name,
				Detail: fmt.Sprintf("dbhome %s root.sh on %s", h.Name, n)})
		}
	}

	// 6) Listener (cluster-wide) — from the grid home when GI is present.
	add(Step{Primitive: "netca", Scope: ScopeCluster, Target: first, Grid: hasGrid, Detail: "create listener"})

	// 7) For each database: create CDB, then each PDB.
	for _, db := range m.Spec.Databases {
		h, ok := m.dbHome(db.DBHomeRef)
		if !ok {
			return nil, fmt.Errorf("database %s: db_home_ref %q not resolvable", db.CDBName, db.DBHomeRef)
		}
		add(Step{Primitive: "dbca", Scope: ScopeCluster, Target: first, HomeRef: h.Name, CDB: db.DBUniqueName,
			Detail: fmt.Sprintf("create CDB %s (unique=%s, home=%s)", db.CDBName, db.DBUniqueName, h.OracleHome)})
		for _, p := range db.PDBs {
			add(Step{Primitive: "pdb", Scope: ScopeCluster, Target: first, HomeRef: h.Name, CDB: db.DBUniqueName, PDB: p.Name,
				Detail: fmt.Sprintf("create PDB %s in %s", p.Name, db.CDBName)})
		}
	}

	return steps, nil
}
