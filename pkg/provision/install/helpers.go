// Package install ships Oracle install primitives — runInstaller,
// root.sh, asmca, netca, oracleasm/afd disk labeling — invoked by
// /lab-up Phase D skills via dbxcli provision install <action>.
//
// All functions in this package require Enterprise license tier
// (license.RequireTier checked at the cobra layer, not here).
package install

import (
	"context"
	"fmt"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
)

// probeFile checks whether a file exists on the target host and returns
// its content. Used by detection probes (e.g., cat /etc/oraInst.loc).
func probeFile(ctx context.Context, exec host.Executor, path string) (exists bool, content string, err error) {
	cmd := fmt.Sprintf("test -f %s && cat %s", shellEscape(path), shellEscape(path))
	res, err := exec.Run(ctx, cmd)
	if err != nil {
		return false, "", err
	}
	if res.ExitCode != 0 {
		return false, "", nil
	}
	return true, res.Stdout, nil
}

// probeDirNonEmpty returns true when the directory exists and contains
// at least one entry. Used to detect partial installs (inventory dir
// present but missing files).
func probeDirNonEmpty(ctx context.Context, exec host.Executor, path string) (bool, error) {
	cmd := fmt.Sprintf("test -d %s && ls -A %s | head -1", shellEscape(path), shellEscape(path))
	res, err := exec.Run(ctx, cmd)
	if err != nil {
		return false, err
	}
	if res.ExitCode != 0 {
		return false, nil
	}
	return strings.TrimSpace(res.Stdout) != "", nil
}

// tailLog returns the last n lines of s. Used to attach installer
// stdout/stderr to InstallResult without ballooning audit records.
// A trailing newline is stripped before splitting so that "line\nline\n"
// is treated as 2 lines, not 3.
func tailLog(s string, n int) string {
	s = strings.TrimRight(s, "\n")
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n") + "\n"
	}
	return strings.Join(lines[len(lines)-n:], "\n") + "\n"
}

// shellEscape wraps a path in single quotes and escapes embedded single
// quotes — sufficient for paths that MAY contain spaces but never quotes.
func shellEscape(s string) string {
	if !strings.ContainsAny(s, " \t'\"$\\;|&<>(){}[]*?!#") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
