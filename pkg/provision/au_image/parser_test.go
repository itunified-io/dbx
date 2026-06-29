package au_image

import (
	"os"
	"path/filepath"
	"testing"
)

// fixtureNames lists every oneoff_*.xml file under testdata/.
var fixtureNames = []string{
	"oneoff_29517242.xml",
	"oneoff_29585399.xml",
	"oneoff_34672698.xml",
	"oneoff_37847857.xml",
	"oneoff_37860476.xml",
	"oneoff_37960098.xml",
	"oneoff_38194420.xml",
}

func loadFixture(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("open fixture %s: %v", name, err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func TestParseOneoffInventory_Fixtures(t *testing.T) {
	cases := map[string]struct {
		patchID       string
		descContains  string
		minOpatch     string
		baseBugsAtLeast int
	}{
		"oneoff_29517242.xml": {
			patchID:         "29517242",
			descContains:    "Database Release Update : 19.3.0.0.190416",
			minOpatch:       "12.2.0.1.14",
			baseBugsAtLeast: 1,
		},
		"oneoff_29585399.xml": {
			patchID:         "29585399",
			descContains:    "OCW RELEASE UPDATE 19.3.0.0.0",
			minOpatch:       "",
			baseBugsAtLeast: 0,
		},
		"oneoff_34672698.xml": {
			patchID:         "34672698",
			descContains:    "ORA-00800",
			minOpatch:       "12.2.0.1.45",
			baseBugsAtLeast: 1,
		},
		"oneoff_37847857.xml": {
			patchID:         "37847857",
			descContains:    "OJVM RELEASE UPDATE: 19.28",
			minOpatch:       "",
			baseBugsAtLeast: 0,
		},
		"oneoff_37860476.xml": {
			patchID:         "37860476",
			descContains:    "JDK BUNDLE PATCH",
			minOpatch:       "",
			baseBugsAtLeast: 0,
		},
		"oneoff_37960098.xml": {
			patchID:         "37960098",
			descContains:    "Database Release Update : 19.28.0.0.250715",
			minOpatch:       "12.2.0.1.46",
			baseBugsAtLeast: 100, // huge fixture
		},
		"oneoff_38194420.xml": {
			patchID:         "38194420",
			descContains:    "MERGE ON DATABASE RU 19.28",
			minOpatch:       "",
			baseBugsAtLeast: 0,
		},
	}

	for name, exp := range cases {
		t.Run(name, func(t *testing.T) {
			f := loadFixture(t, name)
			got, err := ParseOneoffInventory(f)
			if err != nil {
				t.Fatalf("parse %s: %v", name, err)
			}
			if got.PatchID != exp.patchID {
				t.Errorf("patch_id: got %q want %q", got.PatchID, exp.patchID)
			}
			if exp.descContains != "" && !contains(got.Description, exp.descContains) {
				t.Errorf("description: got %q, want substring %q", got.Description, exp.descContains)
			}
			if exp.minOpatch != "" && got.MinOpatchVersion != exp.minOpatch {
				t.Errorf("min_opatch_version: got %q want %q", got.MinOpatchVersion, exp.minOpatch)
			}
			if exp.baseBugsAtLeast > 0 && len(got.BaseBugs) < exp.baseBugsAtLeast {
				t.Errorf("base_bugs len: got %d, want >= %d", len(got.BaseBugs), exp.baseBugsAtLeast)
			}
		})
	}
}

func TestClassifyPatches_RealLabFixture(t *testing.T) {
	var all []OneoffPatch
	for _, name := range fixtureNames {
		f := loadFixture(t, name)
		op, err := ParseOneoffInventory(f)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		all = append(all, *op)
	}

	ru, oneoffs, err := ClassifyPatches(all)
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if ru == nil {
		t.Fatalf("classify: ru is nil")
	}
	if ru.Label != "19.28" {
		t.Errorf("ru.label: got %q want %q", ru.Label, "19.28")
	}
	if ru.PatchID != "37960098" {
		t.Errorf("ru.patch_id: got %q want %q", ru.PatchID, "37960098")
	}

	// Real (non-bundled) oneoffs must be exactly:
	// 34672698, 37847857, 37860476, 38194420.
	var real []string
	for _, o := range oneoffs {
		if !o.IsOpatchBundled {
			real = append(real, o.PatchID)
		}
	}
	want := map[string]bool{
		"34672698": true,
		"37847857": true,
		"37860476": true,
		"38194420": true,
	}
	if len(real) != len(want) {
		t.Errorf("real oneoffs len: got %d (%v), want %d (%v)", len(real), real, len(want), keys(want))
	}
	for _, id := range real {
		if !want[id] {
			t.Errorf("unexpected real oneoff: %s", id)
		}
	}

	// Bundled set must include 29517242 (19.3 RU shadow) + 29585399 (OCW).
	bundled := map[string]bool{}
	for _, o := range oneoffs {
		if o.IsOpatchBundled {
			bundled[o.PatchID] = true
		}
	}
	for _, id := range []string{"29517242", "29585399"} {
		if !bundled[id] {
			t.Errorf("expected %s to be flagged IsOpatchBundled", id)
		}
	}
}

func TestComputeAuID(t *testing.T) {
	ru := &RUPatch{Label: "19.28", PatchID: "37960098"}
	cases := []struct {
		name    string
		oneoffs []OneoffPatch
		want    string
	}{
		{
			name:    "no oneoffs",
			oneoffs: nil,
			want:    "au19.28v1",
		},
		{
			name:    "only bundled (still no real)",
			oneoffs: []OneoffPatch{{PatchID: "29517242", IsOpatchBundled: true}},
			want:    "au19.28v1",
		},
		{
			name: "four real oneoffs (lab fixture)",
			oneoffs: []OneoffPatch{
				{PatchID: "34672698"},
				{PatchID: "37847857"},
				{PatchID: "37860476"},
				{PatchID: "38194420"},
				{PatchID: "29517242", IsOpatchBundled: true},
			},
			want: "au19.28+86fbedv1",
		},
		{
			name: "order independence",
			oneoffs: []OneoffPatch{
				{PatchID: "38194420"},
				{PatchID: "34672698"},
				{PatchID: "37860476"},
				{PatchID: "37847857"},
			},
			want: "au19.28+86fbedv1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeAuID(ru, tc.oneoffs, 1)
			if got != tc.want {
				t.Errorf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestMapPatchLevelToRU(t *testing.T) {
	cases := map[string]string{
		"0": "19.3 (base)",
		"1": "19.26",
		"2": "19.27",
		"3": "19.28",
	}
	for level, want := range cases {
		got, err := MapPatchLevelToRU(level)
		if err != nil {
			t.Errorf("MapPatchLevelToRU(%q) err: %v", level, err)
			continue
		}
		if got != want {
			t.Errorf("MapPatchLevelToRU(%q): got %q want %q", level, got, want)
		}
	}
	if _, err := MapPatchLevelToRU(""); err == nil {
		t.Errorf("MapPatchLevelToRU(\"\"): expected err for empty input")
	}
	if _, err := MapPatchLevelToRU("99"); err == nil {
		t.Errorf("MapPatchLevelToRU(\"99\"): expected err for unknown level")
	}
}

func TestParseCompsXML(t *testing.T) {
	f := loadFixture(t, "comps_head.xml")
	level, err := ParseCompsXML(f)
	if err != nil {
		t.Fatalf("ParseCompsXML: %v", err)
	}
	// The truncated head fixture does not include a PATCH_LEVEL
	// attribute (only present on patched homes); empty is the
	// documented contract.
	if level != "" {
		t.Logf("comps_head.xml PATCH_LEVEL = %q (note: head fixture is base-stamped)", level)
	}
}

// helpers

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || stringContains(haystack, needle))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
