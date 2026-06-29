package oracle

import (
	"context"
	"testing"
)

// findStep returns the first StepSpec matching primitive (and optional pdb name).
func findSpec(specs []StepSpec, primitive string) (StepSpec, bool) {
	for _, s := range specs {
		if s.Step.Primitive == primitive {
			return s, true
		}
	}
	return StepSpec{}, false
}

func TestBuildSpecs_GridAndHomePaths(t *testing.T) {
	m := racManifest()
	m.Spec.Grid.SoftwareStaging.Source = "/smb/software"
	m.Spec.DBHomes[0].SoftwareStaging.Source = "/smb/software"
	specs, err := BuildSpecs(m, ApplyOptions{})
	if err != nil {
		t.Fatalf("BuildSpecs: %v", err)
	}
	grid, ok := findSpec(specs, "grid")
	if !ok || grid.Install == nil {
		t.Fatal("no grid Install spec")
	}
	if grid.Install.OracleHome != "/u01/app/19.0.0/grid" || grid.Install.OracleBase != "/u01/app/grid" {
		t.Errorf("grid home/base: %s / %s", grid.Install.OracleHome, grid.Install.OracleBase)
	}
	if grid.Install.SoftwareStaging != "/smb/software" {
		t.Errorf("grid staging: %s", grid.Install.SoftwareStaging)
	}
	dbh, ok := findSpec(specs, "dbhome")
	if !ok || dbh.Install == nil {
		t.Fatal("no dbhome Install spec")
	}
	if dbh.Install.OracleHome != "/u01/app/oracle/product/19.0.0/dbhome_1" {
		t.Errorf("dbhome oracle_home: %s", dbh.Install.OracleHome)
	}
}

func TestBuildSpecs_AsmcaPerDiskgroup(t *testing.T) {
	m := racManifest()
	for i := range m.Spec.ASM.Diskgroups {
		m.Spec.ASM.Diskgroups[i].Redundancy = "external"
		m.Spec.ASM.Diskgroups[i].AUSize = "4M"
	}
	opts := ApplyOptions{DisksByTag: map[string][]string{"asm-data": {"/dev/sdb", "/dev/sdc"}}}
	specs, err := BuildSpecs(m, opts)
	if err != nil {
		t.Fatalf("BuildSpecs: %v", err)
	}
	var asmca []StepSpec
	for _, s := range specs {
		if s.Step.Primitive == "asmca" {
			asmca = append(asmca, s)
		}
	}
	if len(asmca) != 3 {
		t.Fatalf("expected 3 asmca specs, got %d", len(asmca))
	}
	for _, s := range asmca {
		if s.Asmca == nil {
			t.Fatal("asmca spec nil")
		}
		if s.Asmca.Redundancy != "EXTERNAL" {
			t.Errorf("redundancy not upper-cased: %s", s.Asmca.Redundancy)
		}
		if s.Asmca.AUSizeMB != 4 {
			t.Errorf("au_size_mb: got %d want 4", s.Asmca.AUSizeMB)
		}
		if s.Asmca.DGName == "DATA" && len(s.Asmca.Disks) != 2 {
			t.Errorf("DATA disks not wired from tag: %v", s.Asmca.Disks)
		}
	}
}

func TestBuildSpecs_DbcaAndPdb(t *testing.T) {
	m := racManifest()
	specs, err := BuildSpecs(m, ApplyOptions{PdbAdminPasswordFile: "/tmp/pw"})
	if err != nil {
		t.Fatalf("BuildSpecs: %v", err)
	}
	dbca, ok := findSpec(specs, "dbca")
	if !ok || dbca.Dbca == nil || dbca.Dbca.DbUniqueName != "clext6pri" {
		t.Errorf("dbca spec: %+v", dbca.Dbca)
	}
	var pdbs []string
	for _, s := range specs {
		if s.Step.Primitive == "pdb" {
			if s.Pdb == nil || s.Pdb.CdbName != "clext6pri" || s.Pdb.AdminPasswordFile != "/tmp/pw" {
				t.Errorf("pdb spec wrong: %+v", s.Pdb)
			}
			pdbs = append(pdbs, s.Pdb.PdbName)
		}
	}
	if len(pdbs) != 2 || pdbs[0] != "pdb1" || pdbs[1] != "pdb2" {
		t.Errorf("pdb names: %v", pdbs)
	}
}

func TestApply_DryRun_NoExecution(t *testing.T) {
	m := racManifest()
	res, err := Apply(context.Background(), m, ApplyOptions{}, false, false)
	if err != nil {
		t.Fatalf("Apply dry-run: %v", err)
	}
	if len(res) != 14 {
		t.Errorf("expected 14 dry-run results, got %d", len(res))
	}
	for _, r := range res {
		if r.Executed {
			t.Errorf("step %d marked executed in dry-run", r.Step.Order)
		}
	}
}

func TestAuSizeMB(t *testing.T) {
	for in, want := range map[string]int{"4M": 4, "8": 8, "16MB": 16, "": 0, "1M": 1} {
		if got := auSizeMB(in); got != want {
			t.Errorf("auSizeMB(%q)=%d want %d", in, got, want)
		}
	}
}
