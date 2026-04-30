package target

import "testing"

func TestOTELAttrs_OracleDatabase(t *testing.T) {
	tgt := &Target{
		Name: "ext3-orcl",
		Type: TypeOracleDatabase,
		Primary: &Endpoint{
			Host:     "ext3adm1.itunified.io",
			Port:     1521,
			Database: "ORCLPRI",
		},
	}
	got := tgt.OTELAttrs()
	want := map[string]string{
		"dbx.entity_type":    "oracle_database",
		"dbx.entity_name":    "ext3-orcl",
		"dbx.db_unique_name": "ORCLPRI",
		"dbx.host":           "ext3adm1.itunified.io",
	}
	gotMap := map[string]string{}
	for _, a := range got {
		gotMap[a.Key] = a.Value
	}
	for k, v := range want {
		if gotMap[k] != v {
			t.Errorf("attr %s: got %q want %q", k, gotMap[k], v)
		}
	}
}

func TestOTELAttrs_HostFallsBackToSSH(t *testing.T) {
	tgt := &Target{
		Name: "myhost",
		Type: TypeHost,
		SSH:  &SSHConfig{Host: "myhost.example.com", User: "root"},
	}
	got := tgt.OTELAttrs()
	gotMap := map[string]string{}
	for _, a := range got {
		gotMap[a.Key] = a.Value
	}
	if gotMap["dbx.host"] != "myhost.example.com" {
		t.Errorf("expected SSH host fallback, got %q", gotMap["dbx.host"])
	}
}

func TestOTELAttrs_AlwaysIncludesTypeAndName(t *testing.T) {
	tgt := &Target{Name: "x", Type: TypeOraclePDB}
	got := tgt.OTELAttrs()
	if len(got) != 2 {
		t.Fatalf("minimal target should have 2 attrs (type+name), got %d: %+v", len(got), got)
	}
	gotMap := map[string]string{}
	for _, a := range got {
		gotMap[a.Key] = a.Value
	}
	if gotMap["dbx.entity_type"] != "oracle_pdb" || gotMap["dbx.entity_name"] != "x" {
		t.Errorf("attrs: %+v", gotMap)
	}
}

func TestOTELAttrs_DGTopology(t *testing.T) {
	tgt := &Target{
		Name: "ORCL_DG",
		Type: TypeOracleDGTopology,
	}
	got := tgt.OTELAttrs()
	gotMap := map[string]string{}
	for _, a := range got {
		gotMap[a.Key] = a.Value
	}
	if gotMap["dbx.entity_type"] != "oracle_dg_topology" {
		t.Errorf("DG topology type: %q", gotMap["dbx.entity_type"])
	}
	if gotMap["dbx.entity_name"] != "ORCL_DG" {
		t.Errorf("DG name: %q", gotMap["dbx.entity_name"])
	}
}
