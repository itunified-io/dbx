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
// the node the primitive runs against.
type Step struct {
	Order     int    `json:"order" yaml:"order"`
	Primitive string `json:"primitive" yaml:"primitive"` // asm-label|grid|root-sh|asmca|dbhome|netca|dbca|pdb
	Scope     Scope  `json:"scope" yaml:"scope"`
	Target    string `json:"target" yaml:"target"` // node name the primitive runs on
	Detail    string `json:"detail" yaml:"detail"`
}

// Plan derives the ordered provisioning sequence from a manifest.
//
// RAC ordering (Grid present):
//
//	asm-label (cluster) → grid (cluster) → root-sh:grid (per-node)
//	→ asmca (cluster) → for each db_home: dbhome (cluster) → root-sh:dbhome (per-node)
//	→ netca (cluster) → for each database: dbca (cluster) → for each pdb: pdb (cluster)
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
	add := func(prim string, scope Scope, target, detail string) {
		ord++
		steps = append(steps, Step{Order: ord, Primitive: prim, Scope: scope, Target: target, Detail: detail})
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
		add("asm-label", ScopeCluster, first, dgDetail)

		// 2) Grid Infrastructure runInstaller (cluster-wide).
		add("grid", ScopeCluster, first, fmt.Sprintf("GI %s home=%s", m.Spec.Grid.Version, m.Spec.Grid.GridHome))

		// 3) root.sh for grid on EVERY node, first node first.
		for _, n := range nodes {
			add("root-sh", ScopePerNode, n, fmt.Sprintf("grid root.sh on %s (home=%s)", n, m.Spec.Grid.GridHome))
		}

		// 4) Create ASM diskgroups (cluster-wide).
		add("asmca", ScopeCluster, first, "create ASM diskgroups")
	}

	// 5) For each DB home: runInstaller (cluster) + root.sh per node.
	for _, h := range m.Spec.DBHomes {
		add("dbhome", ScopeCluster, first, fmt.Sprintf("DB home %s %s home=%s", h.Name, h.Version, h.OracleHome))
		for _, n := range nodes {
			add("root-sh", ScopePerNode, n, fmt.Sprintf("dbhome %s root.sh on %s", h.Name, n))
		}
	}

	// 6) Listener (cluster-wide).
	add("netca", ScopeCluster, first, "create listener")

	// 7) For each database: create CDB, then each PDB.
	for _, db := range m.Spec.Databases {
		h, ok := m.dbHome(db.DBHomeRef)
		if !ok {
			return nil, fmt.Errorf("database %s: db_home_ref %q not resolvable", db.CDBName, db.DBHomeRef)
		}
		add("dbca", ScopeCluster, first, fmt.Sprintf("create CDB %s (unique=%s, home=%s)", db.CDBName, db.DBUniqueName, h.OracleHome))
		for _, p := range db.PDBs {
			add("pdb", ScopeCluster, first, fmt.Sprintf("create PDB %s in %s", p.Name, db.CDBName))
		}
	}

	return steps, nil
}
