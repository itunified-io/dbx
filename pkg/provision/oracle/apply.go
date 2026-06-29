package oracle

import (
	"context"
	"fmt"
	"strings"

	"github.com/itunified-io/dbx/pkg/provision/install"
)

// ApplyOptions carries the operator-supplied inputs that the OracleDatabase
// manifest deliberately does NOT contain — secrets, response files, raw disk
// devices, listener config. The manifest provides paths/homes/topology; these
// provide the run-time material a skill/operator stages (often from Vault).
type ApplyOptions struct {
	// ASMImplementation is "asmlib" or "afd" (default "asmlib").
	ASMImplementation string
	// ASMLabels maps raw devices to ASM labels for the asm-label step.
	ASMLabels []install.AsmLabelEntry
	// DisksByTag maps a diskgroup's disks_tag to its device/label list, used
	// to populate each asmca step's Disks.
	DisksByTag map[string][]string

	// Response files: absolute paths ON the target host (staged by caller).
	GridResponseFile   string
	DBHomeResponseFile string
	DbcaResponseFile   string

	// Password files: absolute paths ON the target host (mode 0600, staged).
	SysPasswordFile      string
	SystemPasswordFile   string
	PdbAdminPasswordFile string

	// Listener.
	ListenerName string // default "LISTENER"
	ListenerPort int    // default 1521

	// SoftwareRoot overrides the manifest staging.source (optional).
	SoftwareRoot string
}

func (o ApplyOptions) impl() string {
	if o.ASMImplementation != "" {
		return o.ASMImplementation
	}
	return "asmlib"
}

func (o ApplyOptions) listenerName() string {
	if o.ListenerName != "" {
		return o.ListenerName
	}
	return "LISTENER"
}

func (o ApplyOptions) listenerPort() int {
	if o.ListenerPort != 0 {
		return o.ListenerPort
	}
	return 1521
}

// StepSpec pairs a planned Step with the concrete install spec to invoke.
// Exactly one of the spec pointers is non-nil, selected by Step.Primitive.
type StepSpec struct {
	Step     Step
	Install  *install.InstallSpec      // grid, dbhome, root-sh
	Asmca    *install.AsmcaSpec        // asmca
	AsmLabel *install.AsmDiskLabelSpec // asm-label
	Netca    *install.NetcaSpec        // netca
	Dbca     *install.DbcaCreateDbSpec // dbca
	Pdb      *install.PdbCreateSpec    // pdb
}

// auSizeMB parses an au_size string like "4M" / "4" into megabytes.
func auSizeMB(s string) int {
	s = strings.TrimSpace(strings.ToUpper(s))
	s = strings.TrimSuffix(s, "MB")
	s = strings.TrimSuffix(s, "M")
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return n
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// BuildSpecs derives the concrete install spec for every planned step. Pure:
// it reads only the manifest + opts, performs no I/O, and is the unit-tested
// core of `apply`. Manifest-derivable fields (target, homes, staging,
// diskgroup attrs, names) are filled here; operator-supplied fields (secrets,
// disks, response files) come from opts and are validated per-primitive at
// execute time (see Apply), not here.
func BuildSpecs(m *Manifest, opts ApplyOptions) ([]StepSpec, error) {
	steps, err := Plan(m)
	if err != nil {
		return nil, err
	}
	var gridHome, gridBase, gridStaging string
	if m.Spec.Grid != nil {
		gridHome = m.Spec.Grid.GridHome
		gridBase = m.Spec.Grid.GridBase
		gridStaging = m.Spec.Grid.SoftwareStaging.Source
	}
	if opts.SoftwareRoot != "" {
		gridStaging = opts.SoftwareRoot
	}

	out := make([]StepSpec, 0, len(steps))
	for _, st := range steps {
		ss := StepSpec{Step: st}
		switch st.Primitive {
		case "asm-label":
			ss.AsmLabel = &install.AsmDiskLabelSpec{
				Target: st.Target, GridHome: gridHome, OracleBase: gridBase,
				Implementation: opts.impl(), Labels: opts.ASMLabels,
			}
		case "grid":
			ss.Install = &install.InstallSpec{
				Target: st.Target, OracleHome: gridHome, OracleBase: gridBase,
				SoftwareStaging: gridStaging, ResponseFilePath: opts.GridResponseFile,
			}
		case "root-sh":
			oh, ob := gridHome, gridBase
			if !st.Grid {
				h, ok := m.dbHome(st.HomeRef)
				if !ok {
					return nil, fmt.Errorf("root-sh: db_home %q not found", st.HomeRef)
				}
				oh, ob = h.OracleHome, h.OracleBase
			}
			ss.Install = &install.InstallSpec{Target: st.Target, OracleHome: oh, OracleBase: ob}
		case "asmca":
			dg, ok := diskgroup(m, st.DiskGroup)
			if !ok {
				return nil, fmt.Errorf("asmca: diskgroup %q not found", st.DiskGroup)
			}
			ss.Asmca = &install.AsmcaSpec{
				InstallSpec: install.InstallSpec{Target: st.Target, OracleHome: gridHome, OracleBase: gridBase},
				DGName:      dg.Name, Redundancy: strings.ToUpper(dg.Redundancy),
				AUSizeMB: auSizeMB(dg.AUSize), Disks: opts.DisksByTag[dg.DisksTag],
			}
		case "dbhome":
			h, ok := m.dbHome(st.HomeRef)
			if !ok {
				return nil, fmt.Errorf("dbhome: db_home %q not found", st.HomeRef)
			}
			staging := h.SoftwareStaging.Source
			if opts.SoftwareRoot != "" {
				staging = opts.SoftwareRoot
			}
			ss.Install = &install.InstallSpec{
				Target: st.Target, OracleHome: h.OracleHome, OracleBase: h.OracleBase,
				SoftwareStaging: staging, ResponseFilePath: opts.DBHomeResponseFile,
			}
		case "netca":
			oh, ob := gridHome, gridBase
			if !st.Grid {
				h, ok := m.dbHome("")
				if !ok {
					return nil, fmt.Errorf("netca: no db_home to source listener home from")
				}
				oh, ob = h.OracleHome, h.OracleBase
			}
			ss.Netca = &install.NetcaSpec{
				InstallSpec:  install.InstallSpec{Target: st.Target, OracleHome: oh, OracleBase: ob},
				ListenerName: opts.listenerName(), Port: opts.listenerPort(),
			}
		case "dbca":
			h, ok := m.dbHome(st.HomeRef)
			if !ok {
				return nil, fmt.Errorf("dbca: db_home %q not found", st.HomeRef)
			}
			ss.Dbca = &install.DbcaCreateDbSpec{
				InstallSpec: install.InstallSpec{
					Target: st.Target, OracleHome: h.OracleHome, OracleBase: h.OracleBase,
					ResponseFilePath: opts.DbcaResponseFile,
				},
				DbUniqueName: st.CDB, SysPasswordFile: opts.SysPasswordFile, SystemPasswordFile: opts.SystemPasswordFile,
			}
		case "pdb":
			h, ok := m.dbHome(st.HomeRef)
			if !ok {
				return nil, fmt.Errorf("pdb: db_home %q not found", st.HomeRef)
			}
			ss.Pdb = &install.PdbCreateSpec{
				InstallSpec:       install.InstallSpec{Target: st.Target, OracleHome: h.OracleHome, OracleBase: h.OracleBase},
				CdbName:           st.CDB,
				PdbName:           st.PDB,
				AdminPasswordFile: opts.PdbAdminPasswordFile,
			}
		default:
			return nil, fmt.Errorf("unknown primitive %q in plan", st.Primitive)
		}
		out = append(out, ss)
	}
	return out, nil
}

func diskgroup(m *Manifest, name string) (Diskgroup, bool) {
	if m.Spec.ASM == nil {
		return Diskgroup{}, false
	}
	for _, dg := range m.Spec.ASM.Diskgroups {
		if dg.Name == name {
			return dg, true
		}
	}
	return Diskgroup{}, false
}

// StepResult is the outcome of one applied step.
type StepResult struct {
	Step     Step                  `json:"step"`
	Executed bool                  `json:"executed"`
	Result   *install.InstallResult `json:"result,omitempty"`
	Err      string                `json:"error,omitempty"`
}

// Apply runs the provisioning sequence. When execute is false it is a dry-run:
// specs are built and returned without invoking any primitive (no I/O). When
// execute is true each step calls the matching install.* primitive in order,
// stopping at the first error. reset is passed through to the primitives.
//
// Execute-time validation: primitives that require operator-supplied material
// (asmca disks, pdb admin-password file) error before the call if missing.
func Apply(ctx context.Context, m *Manifest, opts ApplyOptions, execute, reset bool) ([]StepResult, error) {
	specs, err := BuildSpecs(m, opts)
	if err != nil {
		return nil, err
	}
	results := make([]StepResult, 0, len(specs))
	for _, ss := range specs {
		if !execute {
			results = append(results, StepResult{Step: ss.Step, Executed: false})
			continue
		}
		res, err := runStep(ctx, ss, reset)
		sr := StepResult{Step: ss.Step, Executed: true, Result: res}
		if err != nil {
			sr.Err = err.Error()
			results = append(results, sr)
			return results, fmt.Errorf("step %d (%s on %s): %w", ss.Step.Order, ss.Step.Primitive, ss.Step.Target, err)
		}
		results = append(results, sr)
	}
	return results, nil
}

func runStep(ctx context.Context, ss StepSpec, reset bool) (*install.InstallResult, error) {
	switch ss.Step.Primitive {
	case "asm-label":
		if len(ss.AsmLabel.Labels) == 0 {
			return nil, fmt.Errorf("asm-label requires --asm-label entries")
		}
		_, err := install.AsmDiskLabel(ctx, *ss.AsmLabel, reset)
		return nil, err
	case "grid":
		return install.GridInstall(ctx, *ss.Install, reset)
	case "root-sh":
		return install.RootSh(ctx, *ss.Install, reset)
	case "asmca":
		if len(ss.Asmca.Disks) == 0 {
			return nil, fmt.Errorf("asmca diskgroup %s requires --disks for its tag", ss.Asmca.DGName)
		}
		return install.AsmcaSilent(ctx, *ss.Asmca, reset)
	case "dbhome":
		return install.DBHomeInstall(ctx, *ss.Install, reset)
	case "netca":
		return install.NetcaSilent(ctx, *ss.Netca, reset)
	case "dbca":
		return install.DbcaCreateDb(ctx, *ss.Dbca, reset)
	case "pdb":
		if ss.Pdb.AdminPasswordFile == "" {
			return nil, fmt.Errorf("pdb %s requires --pdb-admin-password-file", ss.Pdb.PdbName)
		}
		return install.PdbCreate(ctx, *ss.Pdb, reset)
	default:
		return nil, fmt.Errorf("unknown primitive %q", ss.Step.Primitive)
	}
}
