package hosttest

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeT is a minimal testing.TB stub that captures failure without calling
// runtime.Goexit (which panics when used outside of a real tRunner goroutine).
type fakeT struct {
	testing.TB
	failed bool
}

func (f *fakeT) Helper()                           {}
func (f *fakeT) Fatalf(_ string, _ ...interface{}) { f.failed = true }

func TestMockExecutor_RecordsCalls(t *testing.T) {
	m := NewMockExecutor()
	m.OnCommand("foo").Returns(0, "ok", "")
	m.OnCommand("bar").Returns(0, "ok", "")

	_, _ = m.Run(context.Background(), "foo")
	_, _ = m.Run(context.Background(), "bar")

	assert.Equal(t, []string{"foo", "bar"}, m.Calls())
	m.AssertCalled(t, "foo")
	m.AssertCallCount(t, "foo", 1)
	m.AssertCallCount(t, "bar", 1)
	m.AssertCallCount(t, "missing", 0)
}

func TestMockExecutor_AssertCalled_FailsOnAbsent(t *testing.T) {
	m := NewMockExecutor()
	m.OnCommand("foo").Returns(0, "", "")
	_, _ = m.Run(context.Background(), "foo")

	ft := &fakeT{}
	m.AssertCalled(ft, "missing")
	assert.True(t, ft.failed, "AssertCalled should fail when cmd absent")
}

func TestMockExecutor_ParallelSafe(t *testing.T) {
	m := NewMockExecutor()
	m.OnCommandPattern(`.*`).Returns(0, "ok", "")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = m.Run(context.Background(), "concurrent")
		}()
	}
	wg.Wait()
	require.Len(t, m.Calls(), 50)
}

func TestMockExecutor_OnCommandPattern_Matches(t *testing.T) {
	m := NewMockExecutor()
	m.OnCommandPattern(`runInstaller .*-silent`).Returns(0, "Installation Successful.", "")

	res, err := m.Run(context.Background(), "/u01/app/19c/grid/runInstaller -silent -responseFile /tmp/grid.rsp")
	require.NoError(t, err)
	assert.Equal(t, 0, res.ExitCode)
	assert.Contains(t, res.Stdout, "Successful")
}

func TestMockExecutor_UnmatchedCommand_ReturnsError(t *testing.T) {
	m := NewMockExecutor()
	m.OnCommand("expected").Returns(0, "", "")

	_, err := m.Run(context.Background(), "unexpected")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected")
}
