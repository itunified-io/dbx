// Package au_image discovers the actual content of legacy Oracle au-image
// (gold-image) zip artifacts. The zip filename labels lie (e.g. an artifact
// named "OraDB_au19.28_v1.zip" can contain extra one-offs that the operator
// added without re-naming), so the inventory must be derived from the
// patches actually present inside the archive.
//
// Issue: itunified-io/dbx#40.
//
// Two parsing paths are supported:
//
//  1. opatch lsinventory (Phase 2, parser TBD): runs runInstaller -clone
//     and then opatch lsinventory -detail on the staged home; this is the
//     canonical "Oracle says so" view but requires a builder host with
//     working JRE + cluvfy.
//
//  2. xml_inventory (this package, Phase 1): parses
//     `<extracted>/inventory/oneoffs/<id>/etc/config/inventory.xml` per
//     patch + `<extracted>/inventory/ContentsXML/comps.xml` for the
//     PATCH_LEVEL hint. Works on any host with `unzip` — no JRE.
package au_image

import "time"

// InventoryResult is the canonical record of an au-image's actual content.
type InventoryResult struct {
	// AuID is the deterministic identifier "au<RU-label>+<sha[:6]>v1"
	// (e.g. "au19.28+86fbedv1"). When there are no one-offs, the
	// "+<sha>" suffix is omitted (e.g. "au19.28v1").
	AuID string `json:"au_id"`

	// HomeType is supplied by the caller — the legacy zip itself does
	// not encode this. Values: "grid_home" | "db_home".
	HomeType string `json:"home_type"`

	// Release is the major Oracle release ("19c", "23ai", "26ai").
	Release string `json:"release"`

	// Arch is the binary architecture; supplied by the caller.
	Arch string `json:"arch"`

	// Topology is supplied by the caller (or "unknown").
	// Values: "ha_config" | "crs_config" | "unknown".
	Topology string `json:"topology"`

	Base    BaseRelease   `json:"base"`
	Opatch  OpatchVersion `json:"opatch"`
	RU      RUPatch       `json:"ru"`
	Oneoffs []OneoffPatch `json:"oneoffs"`

	JDK  *JDKOverlay  `json:"jdk,omitempty"`
	Perl *PerlOverlay `json:"perl,omitempty"`
	OJVM *OJVMRU      `json:"ojvm,omitempty"`

	// ParsedVia is "xml_inventory" or "opatch_lsinventory".
	ParsedVia string `json:"parsed_via"`

	SourceLegacyPath   string `json:"source_legacy_path,omitempty"`
	SourceLegacySha256 string `json:"source_legacy_sha256,omitempty"`
}

// BaseRelease is the base Oracle 19.3.0.0.0 release the home was cloned
// from (or the equivalent base for 23ai/26ai).
type BaseRelease struct {
	Release string `json:"release"`        // e.g. "19.3.0.0.0"
	Zip     string `json:"zip,omitempty"`  // optional: original base zip name
	Sha256  string `json:"sha256,omitempty"`
}

// OpatchVersion is the version of OPatch shipped under $OH/OPatch.
type OpatchVersion struct {
	Version string `json:"version"` // e.g. "12.2.0.1.47"
}

// RUPatch is the Database Release Update applied (e.g. RU 19.28).
type RUPatch struct {
	Label       string `json:"label"`        // e.g. "19.28"
	PatchID     string `json:"patch_id"`     // e.g. "37960098"
	Description string `json:"description"`  // e.g. "Database Release Update : 19.28.0.0.250715 (37960098)"
}

// OneoffPatch is a per-patch inventory record (excluding the RU itself).
type OneoffPatch struct {
	PatchID          string    `json:"patch_id"`
	UniquePatchID    string    `json:"unique_patch_id,omitempty"`
	Description      string    `json:"description"`
	AppliedDate      time.Time `json:"applied_date,omitempty"`
	MinOpatchVersion string    `json:"min_opatch_version,omitempty"`
	// BaseBugs is omitted by default; populated only when the caller
	// requested verbose output. The Oracle DB RU contains thousands of
	// these and they bloat the JSON output by ~5x.
	BaseBugs []BaseBug `json:"base_bugs,omitempty"`
	// IsOpatchBundled is set on patches that are shipped as part of an
	// OPatch / cluvfy / OneOff packaging chain rather than being
	// "operator-chosen" oneoffs. They appear in the on-disk inventory
	// but are not real adds; ClassifyPatches excludes them from the
	// public Oneoffs slice (and therefore from the AuID hash input).
	IsOpatchBundled bool `json:"is_opatch_bundled,omitempty"`
}

// BaseBug is a single Oracle bug inside a patch's <base_bugs> list.
type BaseBug struct {
	Number      string `json:"number"`
	Description string `json:"description"`
}

// JDKOverlay describes the JDK version overlaid onto the base release
// (e.g. base ships 1.8.0.201, the home overlays 1.8.0.401).
type JDKOverlay struct {
	Version string `json:"version"`
	PatchID string `json:"patch_id,omitempty"`
}

// PerlOverlay describes the Perl version overlaid onto the base release.
type PerlOverlay struct {
	Version string `json:"version"`
	PatchID string `json:"patch_id,omitempty"`
}

// OJVMRU is the OJVM Release Update (db_home only — Grid does not include OJVM).
type OJVMRU struct {
	Label   string `json:"label"`
	PatchID string `json:"patch_id"`
}
