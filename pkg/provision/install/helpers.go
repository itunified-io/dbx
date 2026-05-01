package install

import (
	"context"
	"fmt"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
)

// probeFile checks whether a file exists on the target host. It does
// NOT return file content — InstallResult.LogTail is opaque to
// secret-redaction, so propagating arbitrary file content into result
// fields risks leaking credentials, host keys, or other sensitive
// data into audit records. Callers that genuinely need the bytes (and
// have separately confirmed the path is non-sensitive) must use
// probeFileContents.
func probeFile(ctx context.Context, exec host.Executor, path string) (exists bool, err error) {
	cmd := fmt.Sprintf("test -f %s", shellEscape(path))
	res, err := exec.Run(ctx, cmd)
	if err != nil {
		return false, err
	}
	return res.ExitCode == 0, nil
}

// probeFileContents returns existence + content. Package-internal
// only — see probeFile godoc for the redaction rationale. Use this
// only for paths whose contents are known-safe (well-known Oracle
// metadata files like /etc/oraInst.loc), and never propagate the
// returned content into InstallResult or any audit-bound field.
func probeFileContents(ctx context.Context, exec host.Executor, path string) (exists bool, content string, err error) {
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
// is treated as 2 lines, not 3. Empty input returns "" (no synthetic
// trailing newline).
func tailLog(s string, n int) string {
	if s == "" {
		return ""
	}
	s = strings.TrimRight(s, "\n")
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n") + "\n"
	}
	return strings.Join(lines[len(lines)-n:], "\n") + "\n"
}

// shellEscape wraps a path in single quotes and escapes embedded single
// quotes — sufficient for paths that MAY contain spaces but never quotes.
// Newline + carriage return are treated as escape-triggers as a defense-
// in-depth measure (InstallSpec.Validate also rejects them earlier).
func shellEscape(s string) string {
	if !strings.ContainsAny(s, " \t\n\r'\"$\\;|&<>(){}[]*?!#") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
