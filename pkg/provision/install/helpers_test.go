package install

import (
	"context"
	"strings"
	"testing"

	"github.com/itunified-io/dbx/pkg/host/hosttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProbeFile_Exists(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f /etc/oraInst.loc").Returns(0, "", "")

	exists, err := probeFile(context.Background(), mock, "/etc/oraInst.loc")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestProbeFile_Absent(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f /missing").Returns(1, "", "")

	exists, err := probeFile(context.Background(), mock, "/missing")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestProbeFileContents_Exists(t *testing.T) {
	mock := hosttest.NewMockExecutor()
	mock.OnCommand("test -f /etc/oraInst.loc && cat /etc/oraInst.loc").
		Returns(0, "inventory_loc=/u01/app/oraInventory\ninst_group=oinstall\n", "")

	exists, content, err := probeFileContents(context.Background(), mock, "/etc/oraInst.loc")
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Contains(t, content, "inventory_loc=/u01/app/oraInventory")
}

func TestTailLog_LimitsLines(t *testing.T) {
	long := strings.Repeat("line\n", 200)
	tail := tailLog(long, 100)
	lines := strings.Count(tail, "\n")
	assert.Equal(t, 100, lines)
}

func TestTailLog_Empty(t *testing.T) {
	assert.Equal(t, "", tailLog("", 100))
}

func TestShellEscape_RejectsNewlineByQuoting(t *testing.T) {
	// shellEscape treats \n / \r as escape-triggers (defense in depth).
	out := shellEscape("foo\nbar")
	assert.True(t, strings.HasPrefix(out, "'") && strings.HasSuffix(out, "'"),
		"newline should force quoting: got %q", out)
}
