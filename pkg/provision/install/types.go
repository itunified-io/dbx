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
	DetectionStatePartial                          // Some evidence; install was started but not finished
	DetectionStateInstalled                        // Full prior install detected; safe to skip
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
	for _, d := range s.Disks {
		if strings.ContainsAny(d, "\n\r,") {
			return fmt.Errorf("install: disk entry contains control character or comma: %q", d)
		}
	}
	return nil
}
