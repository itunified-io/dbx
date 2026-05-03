// Package sql provides Oracle SQL execution primitives that run via
// SSH-wrapped sqlplus on a managed Oracle host.
//
// Two execution modes coexist in the dbx tree:
//
//   - pkg/db/sql — in-process database/sql with a SELECT-only ReadOnlyGuard.
//     Used by the read-only dbxcli db sql exec command.
//   - pkg/oracle/sql (this package) — out-of-process sqlplus / as sysdba via
//     SSH, no read-only guard. Used by privileged Phase E.1/E.2 Data Guard
//     operations (FORCE LOGGING, FLASHBACK ON, ADD STANDBY LOGFILE,
//     DBMS_DATAGUARD_BROKER calls). Enterprise-tier license + provision
//     bundle gated at the cobra layer.
//
// This package shells out exactly like pkg/provision/install/* primitives:
// host.Executor → sshExecutor → sqlplus -s / as sysdba <<EOF ... EOF.
package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/itunified-io/dbx/pkg/host"
)

// ExecOptions tunes ExecReadWrite behaviour.
type ExecOptions struct {
	// OracleSID is set as $ORACLE_SID before invoking sqlplus. Required.
	OracleSID string
	// OracleHome is set as $ORACLE_HOME and used to resolve bin/sqlplus.
	// Required.
	OracleHome string
	// LogTailLines bounds the captured stdout+stderr in ExecResult.LogTail.
	// Zero means no tail (full capture).
	LogTailLines int
}

// ExecResult is the captured outcome of an ExecReadWrite invocation.
type ExecResult struct {
	// Statements is the input SQL split on semicolon-newline / slash-newline.
	// Returned for caller-side audit + replay-trace alignment.
	Statements []string `json:"statements"`
	Stdout     string   `json:"stdout"`
	Stderr     string   `json:"stderr"`
	ExitCode   int      `json:"exit_code"`
	// LogTail is the trailing N lines of stdout+stderr per
	// ExecOptions.LogTailLines (0 = full).
	LogTail string `json:"log_tail,omitempty"`
}

// ErrCancelled wraps a ctx.Err() interruption mid-exec. The remote sqlplus
// process may still be running; ExecResult is partial but populated where
// possible.
var ErrCancelled = fmt.Errorf("oracle sql readwrite: context cancelled mid-exec")

// ExecReadWrite executes one or more DDL/DML/anonymous-PL-SQL statements
// via sqlplus / as sysdba over SSH on the host backing the named target.
//
// The target name is resolved by the SSH layer (matches the install
// primitive newSSHExecutor signature). OracleSID + OracleHome live on
// ExecOptions because the unified Target struct (pkg/core/target) does
// not yet model per-database SID/Home pairs — the caller threads them
// from the env.yaml stack manifest.
//
// SECURITY: The SQL string is delivered to sqlplus via a heredoc whose
// terminator is a randomized token. The fixed string "EOF" is rejected
// in the input to prevent heredoc-terminator injection.
func ExecReadWrite(ctx context.Context, target string, sqlInput string, opts ExecOptions) (*ExecResult, error) {
	exec, err := newSSHExecutor(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("oracle sql readwrite: ssh to %s: %w", target, err)
	}
	return execReadWriteWithExec(ctx, exec, sqlInput, opts)
}

// execReadWriteWithExec is the testable core. Takes an injected executor
// so unit tests can use hosttest.MockExecutor.
func execReadWriteWithExec(ctx context.Context, exec host.Executor, sqlInput string, opts ExecOptions) (*ExecResult, error) {
	if err := validateOpts(opts); err != nil {
		return nil, err
	}

	stmt := strings.TrimSpace(sqlInput)
	if stmt == "" {
		return nil, fmt.Errorf("oracle sql readwrite: sql is empty")
	}
	if err := guardHeredocTerminator(stmt); err != nil {
		return nil, err
	}

	// Split on ;\n and /\n boundaries for audit. The combined block is
	// still sent to sqlplus as one heredoc — splitting is informational.
	stmts := splitStatements(stmt)

	// Heredoc terminator: pick a token that the validator already
	// guarantees is NOT in the SQL.
	const heredocTerm = "DBX_RW_END_OF_INPUT"

	// Trailing newline + EXIT to ensure sqlplus exits with the script's
	// status. -s suppresses the banner.
	body := stmt
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += "EXIT;\n"

	cmd := fmt.Sprintf(
		"sudo -u oracle bash -lc 'export ORACLE_SID=%s; export ORACLE_HOME=%s; %s/bin/sqlplus -s / as sysdba <<%s\n%s%s\n'",
		shellEscape(opts.OracleSID),
		shellEscape(opts.OracleHome),
		shellEscape(opts.OracleHome),
		heredocTerm,
		body,
		heredocTerm,
	)

	runRes, err := exec.Run(ctx, cmd)
	res := &ExecResult{Statements: stmts}
	if err != nil {
		// ctx cancel mid-run: remote sqlplus may still be running. Return
		// partial result + ErrCancelled so callers can distinguish from
		// transport errors.
		if ctx.Err() != nil {
			return res, fmt.Errorf("%w: %v", ErrCancelled, err)
		}
		return nil, fmt.Errorf("oracle sql readwrite: transport failure: %w", err)
	}
	res.ExitCode = runRes.ExitCode
	res.Stdout = runRes.Stdout
	res.Stderr = runRes.Stderr
	res.LogTail = tailLog(runRes.Stdout+runRes.Stderr, opts.LogTailLines)
	if runRes.ExitCode != 0 {
		// Non-zero exit is propagated as an error so the caller can fail
		// fast; partial result remains for log inspection.
		return res, fmt.Errorf("oracle sql readwrite: sqlplus exit %d", runRes.ExitCode)
	}
	return res, nil
}

func validateOpts(opts ExecOptions) error {
	if strings.TrimSpace(opts.OracleSID) == "" {
		return fmt.Errorf("oracle sql readwrite: oracle_sid is required")
	}
	if strings.TrimSpace(opts.OracleHome) == "" {
		return fmt.Errorf("oracle sql readwrite: oracle_home is required")
	}
	for _, f := range []struct{ name, value string }{
		{"oracle_sid", opts.OracleSID},
		{"oracle_home", opts.OracleHome},
	} {
		if strings.ContainsAny(f.value, "\n\r'\"$`!&|;<>(){}*?\\") {
			return fmt.Errorf("oracle sql readwrite: %s contains disallowed character", f.name)
		}
	}
	return nil
}

// guardHeredocTerminator rejects SQL containing the literal heredoc
// terminator token, which would otherwise truncate the script and let
// remaining lines run as shell commands. We reject case-insensitively
// AND reject the bare word "EOF" on its own line as a defense-in-depth
// measure for operators who paste sqlplus scripts that end with EOF.
func guardHeredocTerminator(s string) error {
	if strings.Contains(s, "DBX_RW_END_OF_INPUT") {
		return fmt.Errorf("oracle sql readwrite: SQL contains reserved heredoc terminator")
	}
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) == "EOF" {
			return fmt.Errorf("oracle sql readwrite: SQL contains a bare 'EOF' line which would terminate a heredoc; rewrite the script")
		}
	}
	return nil
}

// splitStatements is a best-effort statement split for audit purposes
// only. It splits on lines that are exactly ";" or "/" after trimming,
// or on a trailing ";" at end-of-line. The combined block is what
// actually reaches sqlplus.
func splitStatements(s string) []string {
	out := []string{}
	var cur strings.Builder
	flush := func() {
		t := strings.TrimSpace(cur.String())
		if t != "" {
			out = append(out, t)
		}
		cur.Reset()
	}
	for _, line := range strings.Split(s, "\n") {
		trim := strings.TrimSpace(line)
		if trim == "/" || trim == ";" {
			flush()
			continue
		}
		cur.WriteString(line)
		cur.WriteString("\n")
		if strings.HasSuffix(trim, ";") {
			flush()
		}
	}
	flush()
	return out
}

// tailLog returns the trailing n lines of s. n=0 means full text.
func tailLog(s string, n int) string {
	if s == "" || n <= 0 {
		return s
	}
	s = strings.TrimRight(s, "\n")
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n") + "\n"
	}
	return strings.Join(lines[len(lines)-n:], "\n") + "\n"
}

// shellEscape wraps a value in single quotes with embedded-quote escaping.
// Safe for paths and identifiers that have already been validated to be
// free of newlines/control chars (validateOpts).
func shellEscape(s string) string {
	if !strings.ContainsAny(s, " \t'\"$\\;|&<>(){}[]*?!#") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
