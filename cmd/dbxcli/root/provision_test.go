package root_test

import (
	"bytes"
	"testing"

	"github.com/itunified-io/dbx/cmd/dbxcli/root"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvisionCmd_HelpShowsProvision(t *testing.T) {
	cmd := root.NewProvisionCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})
	// Help always returns nil (cobra calls os.Exit(0) on --help by default;
	// we override that via SilenceUsage + DisableFlagParsing is not set, so
	// we use RunE absence — cobra will call Help() and return nil).
	err := cmd.Help()
	require.NoError(t, err)
	helpText := out.String()
	assert.Contains(t, helpText, "provision")
	// dbca + pdb subcommands added in Tasks 7 + 8; this test will
	// gain assertions for them then.
}

func TestProvisionCmd_RegisteredOnRootCmd(t *testing.T) {
	rootCmd := root.New("test")
	names := make([]string, 0)
	for _, c := range rootCmd.Commands() {
		names = append(names, c.Name())
	}
	assert.Contains(t, names, "provision", "provision subcommand must be registered on root")
}
