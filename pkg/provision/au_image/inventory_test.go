package au_image

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/itunified-io/dbx/pkg/host/hosttest"
)

func mustReadFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(b)
}

// TestInventoryWithExec_HappyPath exercises the full XML-direct path
// with a MockExecutor. The mock returns the lab fixtures for each
// inventory.xml read, so the result must match the proven AuID.
func TestInventoryWithExec_HappyPath(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	stage := "/u01/au-stage/legacy_au19_28"

	// 1. mkdir + unzip succeed.
	mock.OnCommandPattern(`^mkdir -p `).Returns(0, "", "")
	mock.OnCommandPattern(`^cd .* && unzip -oq `).Returns(0, "", "")

	// 2. ls of oneoffs returns the seven fixture patch ids.
	listing := strings.Join([]string{
		"29517242", "29585399", "34672698",
		"37847857", "37860476", "37960098", "38194420",
	}, "\n") + "\n"
	mock.OnCommandPattern(`^ls -1 .*/inventory/oneoffs`).Returns(0, listing, "")

	// 3. Each per-patch inventory.xml read. The probeFileContents
	//    helper emits `test -f X && cat X` — match by the stage-rooted
	//    path inside the command string.
	for _, fix := range []struct {
		patchID string
		file    string
	}{
		{"29517242", "oneoff_29517242.xml"},
		{"29585399", "oneoff_29585399.xml"},
		{"34672698", "oneoff_34672698.xml"},
		{"37847857", "oneoff_37847857.xml"},
		{"37860476", "oneoff_37860476.xml"},
		{"37960098", "oneoff_37960098.xml"},
		{"38194420", "oneoff_38194420.xml"},
	} {
		body := mustReadFixture(t, fix.file)
		// Each command string contains `oneoffs/<id>/etc/config/inventory.xml`.
		mock.OnCommandPattern(`oneoffs/` + fix.patchID + `/etc/config/inventory.xml`).Returns(0, body, "")
	}

	// 4. comps.xml read — return base head fixture (no PATCH_LEVEL).
	mock.OnCommandPattern(`ContentsXML/comps\.xml`).Returns(0, mustReadFixture(t, "comps_head.xml"), "")

	// 5. cleanup rm -rf.
	mock.OnCommandPattern(`^rm -rf `).Returns(0, "", "")

	spec := InventorySpec{
		Target:       "aubuilder1",
		ArtifactPath: "/smb/software/oracle/19/db_home/legacy/OraDB_au19.28_v1.zip",
		HomeType:     "db_home",
		StagePath:    stage,
	}

	res, err := inventoryWithExec(context.Background(), mock, spec)
	if err != nil {
		t.Fatalf("inventoryWithExec: %v", err)
	}
	if res.AuID != "au19.28+86fbedv1" {
		t.Errorf("au_id: got %q want %q", res.AuID, "au19.28+86fbedv1")
	}
	if res.RU.Label != "19.28" {
		t.Errorf("ru.label: got %q want %q", res.RU.Label, "19.28")
	}
	if res.RU.PatchID != "37960098" {
		t.Errorf("ru.patch_id: got %q want %q", res.RU.PatchID, "37960098")
	}
	if res.HomeType != "db_home" {
		t.Errorf("home_type: got %q", res.HomeType)
	}
	if res.ParsedVia != "xml_inventory" {
		t.Errorf("parsed_via: got %q", res.ParsedVia)
	}
	if res.SourceLegacyPath != spec.ArtifactPath {
		t.Errorf("source_legacy_path: got %q", res.SourceLegacyPath)
	}
	// Real (non-bundled) oneoffs len must be 4.
	var real int
	for _, o := range res.Oneoffs {
		if !o.IsOpatchBundled {
			real++
		}
	}
	if real != 4 {
		t.Errorf("real oneoffs: got %d want 4 (full set: %v)", real, res.Oneoffs)
	}
	if res.OJVM == nil || res.OJVM.PatchID != "37847857" {
		t.Errorf("ojvm: got %+v", res.OJVM)
	}
	if res.JDK == nil || res.JDK.PatchID != "37860476" {
		t.Errorf("jdk: got %+v", res.JDK)
	}
}

func TestInventorySpec_Validate(t *testing.T) {
	cases := []struct {
		name    string
		spec    InventorySpec
		wantErr bool
	}{
		{"missing target", InventorySpec{ArtifactPath: "/a.zip"}, true},
		{"missing artifact", InventorySpec{Target: "h"}, true},
		{"bad home_type", InventorySpec{Target: "h", ArtifactPath: "/a.zip", HomeType: "weird"}, true},
		{"newline in target", InventorySpec{Target: "h\n", ArtifactPath: "/a.zip"}, true},
		{"ok db_home", InventorySpec{Target: "h", ArtifactPath: "/a.zip", HomeType: "db_home"}, false},
		{"ok grid_home", InventorySpec{Target: "h", ArtifactPath: "/a.zip", HomeType: "grid_home"}, false},
		{"ok empty home_type", InventorySpec{Target: "h", ArtifactPath: "/a.zip"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("got err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
