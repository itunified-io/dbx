package install

import (
	"fmt"
	"strings"
)

// InstallSpec is the common input for all install primitives. Tool-
// specific extensions (GridSpec, DBHomeSpec, etc.) embed it.
type InstallSpec struct {
	// Target is the dbx target name (resolves to host + SSH config via
	// dbx/pkg/core/target).
	Target string `json:"target"`
	// OracleHome is the destination $ORACLE_HOME (or $GRID_HOME) path.
	OracleHome string `json:"oracle_home"`
	// OracleBase is /u01/app/oracle (or /u01/app/grid).
	OracleBase string `json:"oracle_base"`
	// SoftwareStaging is the local-on-host path where unzipped Oracle
	// software lives (e.g. /smb/software/oracle/19c/grid_home).
	SoftwareStaging string `json:"software_staging"`
	// ResponseFilePath is the absolute path on the target host where
	// the rendered response file (.rsp) was placed by the caller (skill).
	ResponseFilePath string `json:"response_file_path"`
}

// Validate returns an error if required fields are missing or contain
// disallowed characters. \n and \r are rejected on every field as a
// command-injection guard: shellEscape quoting alone does not stop
// embedded newlines from being interpreted by some shell contexts.
func (s InstallSpec) Validate() error {
	if strings.TrimSpace(s.Target) == "" {
		return fmt.Errorf("install: target is required")
	}
	if strings.TrimSpace(s.OracleHome) == "" {
		return fmt.Errorf("install: oracle_home is required")
	}
	for _, f := range []struct{ name, value string }{
		{"target", s.Target},
		{"oracle_home", s.OracleHome},
		{"oracle_base", s.OracleBase},
		{"software_staging", s.SoftwareStaging},
		{"response_file_path", s.ResponseFilePath},
	} {
		if strings.ContainsAny(f.value, "\n\r") {
			return fmt.Errorf("install: field contains control character: %s", f.name)
		}
	}
	return nil
}

// InstallResult is the common output for all install primitives.
type InstallResult struct {
	// Detected is the pre-flight detection result.
	Detected DetectionState `json:"detected"`
	// Skipped is true when Detected != Absent and operator did not
	// pass --reset (no work performed).
	Skipped bool `json:"skipped"`
	// LogTail captures the last 100 lines of installer/script stdout+stderr
	// for debugging. Empty when Skipped=true.
	LogTail string `json:"log_tail,omitempty"`
	// ExitCode is the wrapped command's exit code (0 = success, non-zero
	// is preserved in Error before propagation up the call stack).
	ExitCode int `json:"exit_code"`
}

// DetectionState is the pre-flight idempotency probe outcome.
type DetectionState int

const (
	DetectionStateAbsent    DetectionState = iota // No prior install evidence on target
	DetectionStatePartial                         // Some evidence; install was started but not finished
	DetectionStateInstalled                       // Full prior install detected; safe to skip

	// DetectionStateUnset is a test-only sentinel for table-driven
	// cases that don't care to assert a specific Detected state.
	// Real probes never return Unset; it lives outside the iota
	// sequence so the zero value of DetectionState remains Absent.
	DetectionStateUnset DetectionState = -1
)

// String implements fmt.Stringer.
func (s DetectionState) String() string {
	switch s {
	case DetectionStateAbsent:
		return "absent"
	case DetectionStatePartial:
		return "partial"
	case DetectionStateInstalled:
		return "installed"
	default:
		return "unknown"
	}
}

// AsmcaSpec configures an initial ASM diskgroup creation. Used ONLY for
// the first DG (DATA, RECO) — subsequent diskgroup operations go through
// mcp-oracle-ee-asm tools which assume ASM is already up.
type AsmcaSpec struct {
	InstallSpec
	DGName     string   `json:"dg_name"`    // e.g. "DATA"
	Redundancy string   `json:"redundancy"` // "EXTERNAL" | "NORMAL" | "HIGH"
	AUSizeMB   int      `json:"au_size_mb"` // e.g. 4
	Disks      []string `json:"disks"`      // ["/dev/sdb", "/dev/sdc"] OR ["AFD:DATA1", ...]
}

// Validate extends InstallSpec.Validate with DG-specific checks.
// OracleBase is required (sentinels live under <oracle_base>/cfgtoollogs/dbx/).
func (s AsmcaSpec) Validate() error {
	if err := s.InstallSpec.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(s.OracleBase) == "" {
		return fmt.Errorf("install: oracle_base is required for asmca (sentinel path)")
	}
	if strings.TrimSpace(s.DGName) == "" {
		return fmt.Errorf("install: dg_name is required")
	}
	switch s.Redundancy {
	case "EXTERNAL", "NORMAL", "HIGH":
	default:
		return fmt.Errorf("install: redundancy must be EXTERNAL/NORMAL/HIGH, got %q", s.Redundancy)
	}
	if s.AUSizeMB <= 0 {
		return fmt.Errorf("install: au_size_mb must be > 0")
	}
	if len(s.Disks) == 0 {
		return fmt.Errorf("install: disks list is required")
	}
	// Reject control chars, comma (the join separator), and shell
	// metacharacters that would survive shellEscape on the joined
	// argument. The whole join is shell-escaped at the call site, but
	// individual entries are still defended-in-depth here.
	const disallowed = "\n\r, \t$`!&|;'\"\\<>*?(){}[]"
	for _, d := range s.Disks {
		if strings.ContainsAny(d, disallowed) {
			return fmt.Errorf("install: disk entry contains disallowed character: %q", d)
		}
	}
	return nil
}

// AsmDiskLabelImpl identifies which raw-disk labeling backend is used.
// One of "asmlib" (Oracle ASMlib via /usr/sbin/oracleasm) or "afd"
// (Oracle ASM Filter Driver via <grid_home>/bin/asmcmd afd_label).
const (
	AsmDiskLabelImplAsmlib = "asmlib"
	AsmDiskLabelImplAFD    = "afd"
)

// AsmLabelEntry pairs a raw block device with the ASM label to assign.
type AsmLabelEntry struct {
	Name   string `json:"name"`   // e.g. "DATA1"
	Device string `json:"device"` // e.g. "/dev/sdb"
}

// AsmDiskLabelSpec configures raw-disk → ASM-discoverable labeling.
// Phase D.1 prerequisite to AsmcaSilent. Per-label sentinels live at
// <oracle_base>/cfgtoollogs/dbx/asm-label.<NAME>.{partial,installed}.
type AsmDiskLabelSpec struct {
	Target         string          `json:"target"`
	GridHome       string          `json:"grid_home"`      // $GRID_HOME (provides bin/asmcmd for AFD)
	OracleBase     string          `json:"oracle_base"`    // sentinel root
	Implementation string          `json:"implementation"` // "asmlib" | "afd"
	Labels         []AsmLabelEntry `json:"labels"`
}

// AsmDiskLabelResult lists per-label results.
type AsmDiskLabelResult struct {
	Implementation string                   `json:"implementation"`
	Labels         []AsmDiskLabelLabelResult `json:"labels,omitempty"`
}

// AsmDiskLabelLabelResult is the per-label outcome.
type AsmDiskLabelLabelResult struct {
	Name     string         `json:"name"`
	Device   string         `json:"device"`
	Detected DetectionState `json:"detected"`
	Skipped  bool           `json:"skipped"`
	ExitCode int            `json:"exit_code"`
	LogTail  string         `json:"log_tail,omitempty"`
}

// Validate returns an error if required fields are missing or contain
// disallowed characters. Mirrors AsmcaSpec/NetcaSpec defense-in-depth:
// every shell-interpolated string is checked for control chars + shell
// metacharacters in addition to shellEscape at the call site.
func (s AsmDiskLabelSpec) Validate() error {
	if strings.TrimSpace(s.Target) == "" {
		return fmt.Errorf("install: target is required")
	}
	if strings.TrimSpace(s.GridHome) == "" {
		return fmt.Errorf("install: grid_home is required")
	}
	if strings.TrimSpace(s.OracleBase) == "" {
		return fmt.Errorf("install: oracle_base is required for asm-label (sentinel path)")
	}
	switch s.Implementation {
	case AsmDiskLabelImplAsmlib, AsmDiskLabelImplAFD:
	default:
		return fmt.Errorf("install: implementation must be %q or %q, got %q",
			AsmDiskLabelImplAsmlib, AsmDiskLabelImplAFD, s.Implementation)
	}
	if len(s.Labels) == 0 {
		return fmt.Errorf("install: labels list is required")
	}
	// Reject control chars on the path-like fields.
	for _, f := range []struct{ name, value string }{
		{"target", s.Target},
		{"grid_home", s.GridHome},
		{"oracle_base", s.OracleBase},
	} {
		if strings.ContainsAny(f.value, "\n\r") {
			return fmt.Errorf("install: field contains control character: %s", f.name)
		}
	}
	// Per-label defense-in-depth: reject control chars + shell metachars
	// on both Name and Device. shellEscape at the call site handles
	// quoting, but these fields end up in sentinel filenames + lsblk-
	// adjacent commands where embedded metachars must never reach.
	const disallowed = "\n\r \t$`!&|;'\"\\<>*?(){}[]"
	seen := map[string]bool{}
	for i, l := range s.Labels {
		if strings.TrimSpace(l.Name) == "" {
			return fmt.Errorf("install: labels[%d].name is required", i)
		}
		if strings.TrimSpace(l.Device) == "" {
			return fmt.Errorf("install: labels[%d].device is required", i)
		}
		if strings.ContainsAny(l.Name, disallowed) {
			return fmt.Errorf("install: labels[%d].name contains disallowed character: %q", i, l.Name)
		}
		if strings.ContainsAny(l.Device, disallowed+",") {
			return fmt.Errorf("install: labels[%d].device contains disallowed character: %q", i, l.Device)
		}
		if seen[l.Name] {
			return fmt.Errorf("install: labels[%d].name %q is duplicated", i, l.Name)
		}
		seen[l.Name] = true
	}
	return nil
}

// DbcaCreateDbSpec configures silent CDB creation via `dbca -silent
// -createDatabase -responseFile <path>`. Phase D.4 of /lab-up.
//
// OracleBase is required because the two-phase sentinel pair lives under
// <OracleBase>/cfgtoollogs/dbx/dbca.<DB_UNIQUE_NAME>.{partial,installed}.
//
// Passwords MUST be passed via files on the target host (SysPasswordFile,
// SystemPasswordFile) — never on the command line where they would leak
// into ps(1) output and audit records. The caller (skill) is responsible
// for placing those files (mode 0600, oracle-owned) on the host before
// invoking this primitive.
type DbcaCreateDbSpec struct {
	InstallSpec
	// DbUniqueName is the DB_UNIQUE_NAME for the database (used in the
	// sentinel path AND as the live-probe key for `srvctl status database`).
	DbUniqueName string `json:"db_unique_name"`
	// SysPasswordFile is an optional absolute path to a file on the
	// target host containing the SYS password. dbca reads it via
	// -sysPassword via -responseFile (caller may also embed the password
	// in the response file). Both forms are supported; this field is a
	// hint for the runbook only.
	SysPasswordFile string `json:"sys_password_file,omitempty"`
	// SystemPasswordFile is the analogous path for the SYSTEM password.
	SystemPasswordFile string `json:"system_password_file,omitempty"`
}

// Validate extends InstallSpec.Validate with dbca-specific checks.
func (s DbcaCreateDbSpec) Validate() error {
	if err := s.InstallSpec.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(s.OracleBase) == "" {
		return fmt.Errorf("install: oracle_base is required for dbca (sentinel path)")
	}
	if strings.TrimSpace(s.ResponseFilePath) == "" {
		return fmt.Errorf("install: response_file_path is required (caller must render + scp the dbca .rsp first)")
	}
	if strings.TrimSpace(s.DbUniqueName) == "" {
		return fmt.Errorf("install: db_unique_name is required")
	}
	// Reject control chars + shell metachars on db_unique_name. Mirrors
	// AsmcaSpec/NetcaSpec defense-in-depth: db_unique_name is interpolated
	// into both the srvctl probe AND the sentinel filename, neither of
	// which tolerate metacharacters.
	const disallowed = "\n\r \t$`!&|;'\"\\<>*?(){}[]"
	if strings.ContainsAny(s.DbUniqueName, disallowed) {
		return fmt.Errorf("install: db_unique_name contains disallowed character: %q", s.DbUniqueName)
	}
	// Password-file paths are optional; reject control chars only.
	for _, f := range []struct{ name, value string }{
		{"sys_password_file", s.SysPasswordFile},
		{"system_password_file", s.SystemPasswordFile},
	} {
		if f.value != "" && strings.ContainsAny(f.value, "\n\r") {
			return fmt.Errorf("install: field contains control character: %s", f.name)
		}
	}
	return nil
}

// PdbCreateSpec configures silent PDB creation via `dbca -silent
// -createPluggableDatabase`. Phase D.5 of /lab-up — runs after the
// parent CDB exists (Phase D.4 dbca -createDatabase) and before any
// PDB-scoped tooling.
//
// Sentinels are keyed on the parent CDB's DB_UNIQUE_NAME PLUS the PDB
// name so a same-named PDB in two different CDBs on the same host
// does not collide:
//
//	<OracleBase>/cfgtoollogs/dbx/pdb.<CDB>.<PDB>.{partial,installed}
//
// Passwords MUST be passed via files on the target host
// (AdminPasswordFile) — never on the command line where they would
// leak into ps(1) output and audit records. The caller (skill) is
// responsible for placing those files (mode 0600, oracle-owned) on
// the host before invoking this primitive.
//
// ResponseFilePath is OPTIONAL: when set, dbca is invoked with
// -responseFile <path>; when empty, dbca is invoked with direct CLI
// args (-pdbName, -createAsContainerDatabase). The CLI-arg form is
// the common case; the response-file form supports advanced templates
// (custom datafile placement, SEED FILE_NAME_CONVERT, etc.).
type PdbCreateSpec struct {
	InstallSpec
	// CdbName is the parent CDB's DB_UNIQUE_NAME (used in the sentinel
	// path AND as the sqlplus connection key for the live probe).
	CdbName string `json:"cdb_name"`
	// PdbName is the new PDB's name (used in the sentinel path AND as
	// the v$pdbs probe key + dbca -pdbName argument).
	PdbName string `json:"pdb_name"`
	// AdminPasswordFile is the absolute path on the target host to a
	// file containing the new PDB's admin user password (mode 0600,
	// oracle-owned). REQUIRED.
	AdminPasswordFile string `json:"admin_password_file"`
	// DatafileDest is an optional `+DATA` ASM diskgroup or absolute
	// filesystem path. When empty, dbca uses the CDB's default
	// (DB_CREATE_FILE_DEST or the response-file value).
	DatafileDest string `json:"datafile_dest,omitempty"`
}

// Validate extends InstallSpec.Validate with PDB-specific checks.
// ResponseFilePath is OPTIONAL for PdbCreateSpec (overrides the
// embedded InstallSpec validation which treats it as required-when-set).
func (s PdbCreateSpec) Validate() error {
	// We can't call s.InstallSpec.Validate() blindly because some leaves
	// require ResponseFilePath and some don't; instead repeat the
	// generic checks here and treat ResponseFilePath as optional.
	if strings.TrimSpace(s.Target) == "" {
		return fmt.Errorf("install: target is required")
	}
	if strings.TrimSpace(s.OracleHome) == "" {
		return fmt.Errorf("install: oracle_home is required")
	}
	for _, f := range []struct{ name, value string }{
		{"target", s.Target},
		{"oracle_home", s.OracleHome},
		{"oracle_base", s.OracleBase},
		{"software_staging", s.SoftwareStaging},
		{"response_file_path", s.ResponseFilePath},
	} {
		if strings.ContainsAny(f.value, "\n\r") {
			return fmt.Errorf("install: field contains control character: %s", f.name)
		}
	}
	if strings.TrimSpace(s.OracleBase) == "" {
		return fmt.Errorf("install: oracle_base is required for pdb create (sentinel path)")
	}
	if strings.TrimSpace(s.CdbName) == "" {
		return fmt.Errorf("install: cdb_name is required")
	}
	if strings.TrimSpace(s.PdbName) == "" {
		return fmt.Errorf("install: pdb_name is required")
	}
	if strings.TrimSpace(s.AdminPasswordFile) == "" {
		return fmt.Errorf("install: admin_password_file is required (passwords must NEVER be on the command line)")
	}
	// Reject control chars + shell metachars on cdb_name and pdb_name.
	// Both are interpolated into the sqlplus probe AND the sentinel
	// filename, neither of which tolerate metacharacters.
	const disallowed = "\n\r \t$`!&|;'\"\\<>*?(){}[]"
	if strings.ContainsAny(s.CdbName, disallowed) {
		return fmt.Errorf("install: cdb_name contains disallowed character: %q", s.CdbName)
	}
	if strings.ContainsAny(s.PdbName, disallowed) {
		return fmt.Errorf("install: pdb_name contains disallowed character: %q", s.PdbName)
	}
	// Reject control chars on the optional path-like fields.
	for _, f := range []struct{ name, value string }{
		{"admin_password_file", s.AdminPasswordFile},
		{"datafile_dest", s.DatafileDest},
	} {
		if f.value != "" && strings.ContainsAny(f.value, "\n\r") {
			return fmt.Errorf("install: field contains control character: %s", f.name)
		}
	}
	return nil
}

// NetcaSpec configures listener creation via netca silent. Used during
// Phase D.2 (post-Grid, pre-DBCA) to ensure a LISTENER exists for client
// connections AND during Phase E.2 to add static services on a standby
// for RMAN DUPLICATE.
//
// OracleBase is required because the two-phase sentinel pair lives under
// <OracleBase>/cfgtoollogs/dbx/netca.<LISTENER>.{partial,installed}.
type NetcaSpec struct {
	InstallSpec
	ListenerName string `json:"listener_name"` // e.g. "LISTENER"
	Port         int    `json:"port"`          // e.g. 1521
}

// Validate extends InstallSpec.Validate with listener-specific checks.
func (s NetcaSpec) Validate() error {
	if err := s.InstallSpec.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(s.OracleBase) == "" {
		return fmt.Errorf("install: oracle_base is required for netca (sentinel path)")
	}
	if strings.TrimSpace(s.ListenerName) == "" {
		return fmt.Errorf("install: listener_name is required")
	}
	// Reject control chars + shell metachars on the listener name. Mirrors
	// AsmcaSpec.Validate's defense-in-depth check on disk entries; the
	// listener name is interpolated into both the lsnrctl status command
	// and the sentinel filename, neither of which tolerate metacharacters.
	const disallowed = "\n\r \t$`!&|;'\"\\<>*?(){}[]"
	if strings.ContainsAny(s.ListenerName, disallowed) {
		return fmt.Errorf("install: listener_name contains disallowed character: %q", s.ListenerName)
	}
	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("install: port must be 1-65535, got %d", s.Port)
	}
	return nil
}
