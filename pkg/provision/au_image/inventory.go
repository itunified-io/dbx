package au_image

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/itunified-io/dbx/pkg/host"
	"github.com/itunified-io/dbx/pkg/otel"
)

// InventorySpec is the input for /oracle-au-image-inventory.
type InventorySpec struct {
	// Target is the dbx target (a builder VM with /smb mounted +
	// unzip + Oracle Java if a clone+lsinventory pass is desired).
	Target string

	// ArtifactPath is the absolute path on the target to the legacy
	// au-image zip (e.g. /smb/software/oracle/19/db_home/legacy/foo.zip).
	ArtifactPath string

	// HomeType is "grid_home" | "db_home". Caller-derived from path
	// or operator-supplied; the legacy zip itself does not encode it.
	HomeType string

	// StagePath is the extraction dir on the target. Default
	// `/u01/au-stage/<basename(artifact)>`.
	StagePath string

	// KeepStage preserves the stage dir for post-mortem.
	KeepStage bool

	// ZipListingOnly skips the runInstaller -clone + opatch
	// lsinventory pass and parses XML directly. Faster, no JRE
	// required, and is the only path on hosts without an Oracle
	// stack pre-installed.
	ZipListingOnly bool
}

// Validate returns an error if required fields are missing or contain
// disallowed characters.
func (s InventorySpec) Validate() error {
	if strings.TrimSpace(s.Target) == "" {
		return fmt.Errorf("au_image: target is required")
	}
	if strings.TrimSpace(s.ArtifactPath) == "" {
		return fmt.Errorf("au_image: artifact_path is required")
	}
	switch s.HomeType {
	case "grid_home", "db_home", "":
		// "" is acceptable: caller may want auto-derivation later.
	default:
		return fmt.Errorf("au_image: home_type must be \"grid_home\" or \"db_home\", got %q", s.HomeType)
	}
	for _, f := range []struct{ name, value string }{
		{"target", s.Target},
		{"artifact_path", s.ArtifactPath},
		{"stage_path", s.StagePath},
	} {
		if strings.ContainsAny(f.value, "\n\r") {
			return fmt.Errorf("au_image: field contains control character: %s", f.name)
		}
	}
	return nil
}

// Inventory is the public entrypoint. Resolves an SSH executor for
// spec.Target and dispatches to inventoryWithExec.
func Inventory(ctx context.Context, spec InventorySpec) (*InventoryResult, error) {
	exe, err := newSSHExecutor(ctx, spec.Target)
	if err != nil {
		return nil, fmt.Errorf("ssh to %s: %w", spec.Target, err)
	}
	return inventoryWithExec(ctx, exe, spec)
}

// inventoryWithExec is the testable core. Steps:
//
//  1. Validate spec; default StagePath if empty.
//  2. mkdir stage; unzip artifact into stage.
//  3. Walk inventory/oneoffs/<id>/etc/config/inventory.xml — parse each.
//  4. Read inventory/ContentsXML/comps.xml — extract PATCH_LEVEL.
//  5. ClassifyPatches → ComputeAuID → assemble result.
//  6. (Optional) rm -rf stage.
//
// Phase-2 scope: XML-direct only. The runInstaller -clone +
// opatch lsinventory branch is reserved for a follow-up because it
// requires a fully-licensed Oracle home on the builder host.
func inventoryWithExec(ctx context.Context, exe host.Executor, spec InventorySpec) (res *InventoryResult, retErr error) {
	sb := otel.NewSpan("provision.au_image.inventory", "dbxcli").
		WithAttrs(
			otel.StringAttr(otel.AttrDbxHost, spec.Target),
			otel.StringAttr(otel.AttrDbxEntityType, "oracle_au_image"),
			otel.StringAttr(otel.AttrDbxEntityName, filepath.Base(spec.ArtifactPath)),
		)
	defer func() { emitSpan(ctx, sb, retErr) }()

	if err := spec.Validate(); err != nil {
		return nil, err
	}

	stage := spec.StagePath
	if stage == "" {
		base := filepath.Base(spec.ArtifactPath)
		base = strings.TrimSuffix(base, filepath.Ext(base))
		stage = "/u01/au-stage/" + base
	}

	// 1. mkdir + unzip.
	if _, err := runOK(ctx, exe, fmt.Sprintf("mkdir -p %s", shellEscape(stage))); err != nil {
		return nil, fmt.Errorf("mkdir stage: %w", err)
	}
	if _, err := runOK(ctx, exe, fmt.Sprintf("cd %s && unzip -oq %s", shellEscape(stage), shellEscape(spec.ArtifactPath))); err != nil {
		return nil, fmt.Errorf("unzip artifact: %w", err)
	}

	// 2. List oneoff patch IDs from inventory/oneoffs/.
	listCmd := fmt.Sprintf("ls -1 %s 2>/dev/null", shellEscape(stage+"/inventory/oneoffs"))
	listRes, err := exe.Run(ctx, listCmd)
	if err != nil {
		return nil, fmt.Errorf("list oneoffs: %w", err)
	}
	patchIDs := splitLines(listRes.Stdout)

	var allPatches []OneoffPatch
	for _, pid := range patchIDs {
		invPath := stage + "/inventory/oneoffs/" + pid + "/etc/config/inventory.xml"
		got, content, err := probeFileContents(ctx, exe, invPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", invPath, err)
		}
		if !got {
			continue
		}
		op, err := ParseOneoffInventory(strings.NewReader(content))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", invPath, err)
		}
		allPatches = append(allPatches, *op)
	}

	// 3. comps.xml (best-effort PATCH_LEVEL hint, may be empty).
	compsPath := stage + "/inventory/ContentsXML/comps.xml"
	if _, content, err := probeFileContents(ctx, exe, compsPath); err == nil && content != "" {
		if level, perr := ParseCompsXML(strings.NewReader(content)); perr == nil && level != "" {
			// PATCH_LEVEL is informational only when we also have
			// a DB-RU oneoff; the latter wins. Future enhancement:
			// reconcile + warn if they disagree.
			_ = level
		}
	}

	// 4. Classify + assemble.
	ru, oneoffs, err := ClassifyPatches(allPatches)
	if err != nil {
		return nil, fmt.Errorf("classify patches: %w", err)
	}

	// Split overlay-style patches (OJVM, JDK) into their dedicated
	// fields while keeping them in Oneoffs for AuID stability.
	var ojvm *OJVMRU
	var jdk *JDKOverlay
	for _, o := range oneoffs {
		if o.IsOpatchBundled {
			continue
		}
		switch {
		case strings.HasPrefix(o.Description, "OJVM RELEASE UPDATE"):
			ojvm = &OJVMRU{Label: ru.Label, PatchID: o.PatchID}
		case strings.HasPrefix(o.Description, "JDK BUNDLE PATCH"):
			jdk = &JDKOverlay{PatchID: o.PatchID}
		}
	}

	out := &InventoryResult{
		HomeType:         spec.HomeType,
		Release:          "19c",
		Arch:             "x86_64",
		Topology:         "unknown",
		Base:             BaseRelease{Release: "19.3.0.0.0"},
		RU:               *ru,
		Oneoffs:          oneoffs,
		OJVM:             ojvm,
		JDK:              jdk,
		ParsedVia:        "xml_inventory",
		SourceLegacyPath: spec.ArtifactPath,
	}
	out.AuID = ComputeAuID(ru, oneoffs, 1)

	// 5. Cleanup.
	if !spec.KeepStage {
		_, _ = exe.Run(ctx, fmt.Sprintf("rm -rf %s", shellEscape(stage)))
	}
	return out, nil
}

// runOK executes a remote command and treats any non-zero exit as an
// error — convenience for steps where we don't care about stdout but
// need failure to surface.
func runOK(ctx context.Context, exe host.Executor, cmd string) (*host.RunResult, error) {
	r, err := exe.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if r.ExitCode != 0 {
		return r, fmt.Errorf("exit %d: %s", r.ExitCode, strings.TrimSpace(r.Stderr))
	}
	return r, nil
}

// splitLines returns non-empty trimmed lines from s.
func splitLines(s string) []string {
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			out = append(out, ln)
		}
	}
	return out
}

// shellEscape mirrors pkg/provision/install/helpers.go shellEscape.
// Duplicated here to avoid a cross-package dependency on a sibling
// internal helper.
func shellEscape(s string) string {
	if !strings.ContainsAny(s, " \t\n\r'\"$\\;|&<>(){}[]*?!#") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// probeFileContents returns existence + content for paths whose
// content we know is metadata (no secrets). Mirrors the helper in
// pkg/provision/install/.
func probeFileContents(ctx context.Context, exe host.Executor, path string) (bool, string, error) {
	cmd := fmt.Sprintf("test -f %s && cat %s", shellEscape(path), shellEscape(path))
	res, err := exe.Run(ctx, cmd)
	if err != nil {
		return false, "", err
	}
	if res.ExitCode != 0 {
		return false, "", nil
	}
	return true, res.Stdout, nil
}

// emitSpan exports a single span via the package-level OTEL exporter.
// Mirrors pkg/provision/install/otel.go.
func emitSpan(ctx context.Context, sb *otel.SpanBuilder, retErr error) {
	if sb == nil {
		return
	}
	var span otel.Span
	if retErr != nil {
		span = sb.EndError(retErr)
	} else {
		span = sb.EndOK()
	}
	_ = otel.GlobalExporter().Export(ctx, []otel.Span{span})
}

// sshExecutor is the local-ssh implementation of host.Executor used
// when callers don't supply an executor.
type sshExecutor struct {
	target  string
	host    string
	user    string
	keyPath string
}

func newSSHExecutor(_ context.Context, name string) (host.Executor, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("target must not be empty")
	}
	e := &sshExecutor{target: name}
	if t, err := target.Load(name); err == nil && t.SSH != nil {
		e.host = t.SSH.Host
		e.user = t.SSH.User
		e.keyPath = t.SSH.KeyPath
	}
	return e, nil
}

func (e *sshExecutor) Run(ctx context.Context, command string) (*host.RunResult, error) {
	args := []string{
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
	}
	if e.keyPath != "" {
		args = append(args, "-i", e.keyPath)
	}
	dest := e.target
	if e.host != "" {
		if e.user != "" {
			dest = e.user + "@" + e.host
		} else {
			dest = e.host
		}
	}
	args = append(args, dest, command)
	cmd := exec.CommandContext(ctx, "ssh", args...) //nolint:gosec
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
			err = nil
		} else {
			return nil, err
		}
	}
	return &host.RunResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}
