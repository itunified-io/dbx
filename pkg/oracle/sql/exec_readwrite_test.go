package sql

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/itunified-io/dbx/pkg/host"
	"github.com/itunified-io/dbx/pkg/host/hosttest"
)

func goodOpts() ExecOptions {
	return ExecOptions{
		OracleSID:    "ORCLPRI",
		OracleHome:   "/u01/app/oracle/product/19c/dbhome_1",
		LogTailLines: 0,
	}
}

func TestExecReadWrite_HappyPath(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	mock.OnCommandPattern(`(?s)sudo -u oracle bash -lc.*sqlplus.*FORCE LOGGING`).
		Returns(0, "Database altered.\n", "")
	res, err := execReadWriteWithExec(context.Background(), mock, "ALTER DATABASE FORCE LOGGING;", goodOpts())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", res.ExitCode)
	}
	if !strings.Contains(res.Stdout, "Database altered") {
		t.Errorf("Stdout missing expected text: %q", res.Stdout)
	}
	if len(res.Statements) != 1 {
		t.Errorf("Statements len = %d, want 1: %v", len(res.Statements), res.Statements)
	}
}

func TestExecReadWrite_ExitCodeNonZero(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	mock.OnCommandPattern(`(?s)sqlplus`).Returns(1, "", "ORA-01031: insufficient privileges\n")
	res, err := execReadWriteWithExec(context.Background(), mock, "ALTER DATABASE FLASHBACK ON;", goodOpts())
	if err == nil {
		t.Fatalf("expected error for non-zero exit, got nil")
	}
	if res == nil || res.ExitCode != 1 {
		t.Errorf("res.ExitCode = %v, want 1", res)
	}
	if !strings.Contains(res.Stderr, "ORA-01031") {
		t.Errorf("Stderr missing ORA: %q", res.Stderr)
	}
	if !strings.Contains(err.Error(), "exit 1") {
		t.Errorf("err = %v, want 'exit 1'", err)
	}
}

func TestExecReadWrite_CtxCancelled(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	mock.OnCommandPattern(`(?s)sqlplus`).Returns(0, "", "") // fallback if called
	// The MockExecutor doesn't itself respect ctx; we wrap to inject the cancel error.
	wrapped := &cancellingExec{inner: mock, simulateErr: errors.New("ssh: signal: terminated")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res, err := execReadWriteWithExec(ctx, wrapped, "ALTER SYSTEM SWITCH LOGFILE;", goodOpts())
	if err == nil {
		t.Fatalf("expected ErrCancelled, got nil")
	}
	if !errors.Is(err, ErrCancelled) {
		t.Errorf("err = %v, want errors.Is(ErrCancelled)", err)
	}
	if res == nil {
		t.Errorf("expected partial result, got nil")
	}
}

func TestExecReadWrite_EmptySQL(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	_, err := execReadWriteWithExec(context.Background(), mock, "   \n  ", goodOpts())
	if err == nil || !strings.Contains(err.Error(), "sql is empty") {
		t.Fatalf("want 'sql is empty', got %v", err)
	}
}

func TestExecReadWrite_MultiStatement(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	mock.OnCommandPattern(`(?s)sqlplus`).Returns(0, "Database altered.\nDatabase altered.\nDatabase altered.\n", "")
	multi := `ALTER DATABASE FORCE LOGGING;
ALTER DATABASE FLASHBACK ON;
ALTER DATABASE ADD STANDBY LOGFILE THREAD 1 SIZE 200M;
`
	res, err := execReadWriteWithExec(context.Background(), mock, multi, goodOpts())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Statements) != 3 {
		t.Errorf("Statements len = %d, want 3: %#v", len(res.Statements), res.Statements)
	}
}

func TestExecReadWrite_HeredocInjectionGuard(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	// SQL containing the reserved terminator token.
	_, err := execReadWriteWithExec(context.Background(), mock,
		"SELECT 1 FROM dual;\nDBX_RW_END_OF_INPUT\nrm -rf /\n", goodOpts())
	if err == nil || !strings.Contains(err.Error(), "reserved heredoc terminator") {
		t.Fatalf("want heredoc-terminator rejection, got %v", err)
	}
	// SQL containing a bare 'EOF' line.
	_, err2 := execReadWriteWithExec(context.Background(), mock,
		"SELECT 1 FROM dual;\nEOF\nrm -rf /\n", goodOpts())
	if err2 == nil || !strings.Contains(err2.Error(), "bare 'EOF'") {
		t.Fatalf("want bare-EOF rejection, got %v", err2)
	}
}

func TestExecReadWrite_RequiredOpts(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	cases := []struct {
		name string
		opts ExecOptions
		want string
	}{
		{"missing sid", ExecOptions{OracleHome: "/u01"}, "oracle_sid is required"},
		{"missing home", ExecOptions{OracleSID: "ORCL"}, "oracle_home is required"},
		{"sid newline", ExecOptions{OracleSID: "ORCL\n", OracleHome: "/u01"}, "disallowed character"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := execReadWriteWithExec(context.Background(), mock, "SELECT 1 FROM dual;", tc.opts)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %v, want %q", err, tc.want)
			}
		})
	}
}

func TestExecReadWrite_LogTail(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	mock.OnCommandPattern(`(?s)sqlplus`).Returns(0,
		"line1\nline2\nline3\nline4\nline5\n", "")
	opts := goodOpts()
	opts.LogTailLines = 2
	res, err := execReadWriteWithExec(context.Background(), mock, "SELECT 1 FROM dual;", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.LogTail, "line4") || !strings.Contains(res.LogTail, "line5") {
		t.Errorf("LogTail missing expected lines: %q", res.LogTail)
	}
	if strings.Contains(res.LogTail, "line1") {
		t.Errorf("LogTail should not contain line1: %q", res.LogTail)
	}
}

// cancellingExec wraps a mock and forces a transport-style error on Run,
// simulating ssh-killed-by-context behaviour.
type cancellingExec struct {
	inner       *hosttest.MockExecutor
	simulateErr error
}

func (c *cancellingExec) Run(ctx context.Context, _ string) (*host.RunResult, error) {
	// Block briefly to ensure the test ctx is observably cancelled.
	select {
	case <-time.After(1 * time.Millisecond):
	case <-ctx.Done():
	}
	return nil, c.simulateErr
}
