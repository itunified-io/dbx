package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func racManifest() *Manifest {
	return &Manifest{
		Version: "1", Kind: "OracleDatabase",
		Metadata: Metadata{Name: "clext6"},
		Spec: Spec{
			Engine: "oracle", Edition: "enterprise", Topology: "rac",
			NodesRef: []string{"ext6adm1", "ext6adm2"},
			Grid:     &Grid{Version: "19.26", GridBase: "/u01/app/grid", GridHome: "/u01/app/19.0.0/grid"},
			DBHomes:  []DBHome{{Name: "dbhome_19c_1", Version: "19.26", OracleBase: "/u01/app/oracle", OracleHome: "/u01/app/oracle/product/19.0.0/dbhome_1"}},
			ASM:      &ASM{Diskgroups: []Diskgroup{{Name: "CRS", DisksTag: "asm-crs"}, {Name: "DATA", DisksTag: "asm-data"}, {Name: "RECO", DisksTag: "asm-reco"}}},
			Databases: []Database{{
				CDBName: "CLEXT6", DBUniqueName: "clext6pri", DBHomeRef: "dbhome_19c_1",
				PDBs: []PDB{{Name: "pdb1"}, {Name: "pdb2"}},
			}},
		},
	}
}

func primitives(steps []Step) []string {
	out := make([]string, len(steps))
	for i, s := range steps {
		out[i] = s.Primitive
	}
	return out
}

func TestPlan_RAC_OrderedSequence(t *testing.T) {
	steps, err := Plan(racManifest())
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	want := []string{
		"asm-label", "grid", "root-sh", "root-sh", "asmca", "asmca", "asmca",
		"dbhome", "root-sh", "root-sh",
		"netca", "dbca", "pdb", "pdb",
	}
	got := primitives(steps)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("plan sequence mismatch:\n got=%v\nwant=%v", got, want)
	}
	// Order field must be 1..N contiguous.
	for i, s := range steps {
		if s.Order != i+1 {
			t.Errorf("step %d has Order=%d, want %d", i, s.Order, i+1)
		}
	}
	// grid root.sh must run on each node, first node first.
	if steps[2].Target != "ext6adm1" || steps[3].Target != "ext6adm2" {
		t.Errorf("grid root-sh targets: got %s,%s want ext6adm1,ext6adm2", steps[2].Target, steps[3].Target)
	}
	// cluster-scoped steps target the first node.
	if steps[1].Scope != ScopeCluster || steps[1].Target != "ext6adm1" {
		t.Errorf("grid step should be cluster-scoped on ext6adm1, got scope=%s target=%s", steps[1].Scope, steps[1].Target)
	}
}

func TestPlan_SingleInstance_OmitsGridSteps(t *testing.T) {
	m := racManifest()
	m.Spec.Topology = "single-instance"
	m.Spec.Grid = nil
	m.Spec.ASM = nil
	m.Spec.NodesRef = []string{"solo1"}
	m.Spec.Databases[0].PDBs = []PDB{{Name: "pdb1"}}
	steps, err := Plan(m)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	want := []string{"dbhome", "root-sh", "netca", "dbca", "pdb"}
	if got := primitives(steps); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("single-instance plan mismatch:\n got=%v\nwant=%v", got, want)
	}
}

func TestPlan_PdbFanOut(t *testing.T) {
	m := racManifest()
	m.Spec.Databases[0].PDBs = []PDB{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	steps, _ := Plan(m)
	n := 0
	for _, s := range steps {
		if s.Primitive == "pdb" {
			n++
		}
	}
	if n != 3 {
		t.Errorf("expected 3 pdb steps, got %d", n)
	}
}

func TestValidate_Errors(t *testing.T) {
	cases := map[string]func(*Manifest){
		"kind":     func(m *Manifest) { m.Kind = "Env" },
		"name":     func(m *Manifest) { m.Metadata.Name = "" },
		"nodes":    func(m *Manifest) { m.Spec.NodesRef = nil },
		"dbhomes":  func(m *Manifest) { m.Spec.DBHomes = nil },
		"rac-grid": func(m *Manifest) { m.Spec.Grid = nil },
		"badref":   func(m *Manifest) { m.Spec.Databases[0].DBHomeRef = "nope" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m := racManifest()
			mutate(m)
			if err := m.Validate(); err == nil {
				t.Errorf("%s: expected validation error, got nil", name)
			}
		})
	}
}

func TestLoadManifest_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	yaml := `version: "1"
kind: OracleDatabase
metadata:
  name: clext6
spec:
  engine: oracle
  edition: enterprise
  topology: rac
  nodes_ref: ["ext6adm1", "ext6adm2"]
  grid:
    version: "19.26"
    grid_base: /u01/app/grid
    grid_home: /u01/app/19.0.0/grid
  db_homes:
    - name: dbhome_19c_1
      version: "19.26"
      oracle_base: /u01/app/oracle
      oracle_home: /u01/app/oracle/product/19.0.0/dbhome_1
  asm:
    diskgroups:
      - { name: DATA, redundancy: external, disks_tag: asm-data, au_size: 4M }
  databases:
    - cdb_name: CLEXT6
      db_unique_name: clext6pri
      db_home_ref: dbhome_19c_1
      pdbs:
        - { name: pdb1 }
        - { name: pdb2 }
`
	p := filepath.Join(dir, "clext6-orcl.yaml")
	if err := os.WriteFile(p, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := LoadManifest(p)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if m.Metadata.Name != "clext6" || !m.IsRAC() || len(m.Spec.NodesRef) != 2 {
		t.Errorf("unexpected parse: name=%s rac=%v nodes=%d", m.Metadata.Name, m.IsRAC(), len(m.Spec.NodesRef))
	}
	steps, err := Plan(m)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(steps) != 12 {
		t.Errorf("expected 12 steps for the clext6 manifest, got %d", len(steps))
	}
}
