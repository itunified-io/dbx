// Package hosttest provides test helpers for the host.Executor interface.
package hosttest

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/itunified-io/dbx/pkg/host"
)

// mockEntry is a single registered command expectation.
type mockEntry struct {
	exact   string         // exact command match (if non-empty)
	pattern *regexp.Regexp // regexp match (if non-nil)
	result  *host.RunResult
}

// MockExecutor implements host.Executor for unit tests. Register
// expectations with OnCommand or OnCommandPattern before calling
// gridInstallWithExec (or any function under test).
//
// Unregistered commands return an error; this surfaces unexpected calls
// immediately rather than silently returning zero values.
//
// MockExecutor is safe for concurrent use from multiple goroutines.
type MockExecutor struct {
	mu      sync.Mutex
	entries []mockEntry
	calls   []string // recorded in invocation order
}

// NewMockExecutor returns an empty MockExecutor.
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{}
}

// stub is a builder used by OnCommand / OnCommandPattern to attach a Returns call.
type stub struct {
	m   *MockExecutor
	idx int
}

// Returns registers the exit code, stdout, and stderr for the stub.
func (s *stub) Returns(exitCode int, stdout, stderr string) {
	s.m.mu.Lock()
	defer s.m.mu.Unlock()
	s.m.entries[s.idx].result = &host.RunResult{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}
}

// OnCommand registers an exact-match expectation. The first matching entry wins.
func (m *MockExecutor) OnCommand(exact string) *stub {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, mockEntry{exact: exact})
	return &stub{m: m, idx: len(m.entries) - 1}
}

// OnCommandPattern registers a regexp-match expectation.
func (m *MockExecutor) OnCommandPattern(pattern string) *stub {
	re := regexp.MustCompile(pattern)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, mockEntry{pattern: re})
	return &stub{m: m, idx: len(m.entries) - 1}
}

// Run satisfies host.Executor. It records the command and returns the result
// of the first matching entry, or an error if no entry matches.
func (m *MockExecutor) Run(_ context.Context, command string) (*host.RunResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, command)
	for _, e := range m.entries {
		if e.exact != "" && e.exact == command {
			return e.result, nil
		}
		if e.pattern != nil && e.pattern.MatchString(command) {
			return e.result, nil
		}
	}
	return nil, fmt.Errorf("hosttest: unexpected command: %q", command)
}

// Calls returns a copy of the recorded command history (in invocation order).
// Safe to call concurrently with Run().
func (m *MockExecutor) Calls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.calls))
	copy(out, m.calls)
	return out
}

// AssertCalled fails the test if cmd was NOT among recorded calls.
// Substring match (uses strings.Contains) so callers can assert on a
// fragment of a longer composed command.
func (m *MockExecutor) AssertCalled(t testing.TB, cmd string) {
	t.Helper()
	for _, c := range m.Calls() {
		if strings.Contains(c, cmd) {
			return
		}
	}
	t.Fatalf("expected command containing %q, got %d calls: %v", cmd, len(m.Calls()), m.Calls())
}

// AssertCallCount fails the test if the number of calls containing cmd
// (substring match) is not exactly n.
func (m *MockExecutor) AssertCallCount(t testing.TB, cmd string, n int) {
	t.Helper()
	count := 0
	for _, c := range m.Calls() {
		if strings.Contains(c, cmd) {
			count++
		}
	}
	if count != n {
		t.Fatalf("expected %d calls containing %q, got %d: %v", n, cmd, count, m.Calls())
	}
}
