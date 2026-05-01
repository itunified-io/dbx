// Package install ships Oracle install primitives — runInstaller,
// root.sh, asmca, netca, oracleasm/afd disk labeling — invoked by
// /lab-up Phase D skills via dbxcli provision install <action>.
//
// All functions in this package require Enterprise license tier
// (license.RequireTier checked at the cobra layer, not here).
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

// Validate returns an error if required fields are missing.
func (s InstallSpec) Validate() error {
	if strings.TrimSpace(s.Target) == "" {
		return fmt.Errorf("install: target is required")
	}
	if strings.TrimSpace(s.OracleHome) == "" {
		return fmt.Errorf("install: oracle_home is required")
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
