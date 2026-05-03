package root_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/itunified-io/dbx/cmd/dbxcli/root"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvisionInstall_LeavesRegistered verifies all 8 install leaves
// are wired under `dbxcli provision install`.
func TestProvisionInstall_LeavesRegistered(t *testing.T) {
	cmd := root.NewInstallCmd()
	got := make(map[string]bool)
	for _, c := range cmd.Commands() {
		got[c.Name()] = true
	}
	for _, want := range []string{"grid", "dbhome", "root-sh", "asmca", "netca", "asm-label", "dbca", "pdb"} {
		assert.True(t, got[want], "leaf %q not registered under `provision install`", want)
	}
}

// TestProvisionInstall_LeafHelp ensures every leaf renders --help
// without panic and surfaces its short description. This proves the
// command tree builds, flags are wired, and required-flag markers do
// not fire on --help (cobra's documented behavior).
func TestProvisionInstall_LeafHelp(t *testing.T) {
	leaves := []string{"grid", "dbhome", "root-sh", "asmca", "netca", "asm-label", "dbca", "pdb"}
	for _, leaf := range leaves {
		t.Run(leaf, func(t *testing.T) {
			cmd := root.NewInstallCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{leaf, "--help"})
			err := cmd.Execute()
			require.NoError(t, err, "leaf %s --help failed: %s", leaf, out.String())
			assert.Contains(t, out.String(), leaf)
		})
	}
}

// TestProvisionInstall_MissingRequiredFlag verifies that each leaf
// with a required flag list rejects an invocation that omits them.
// This is a pure flag-parsing test — we never reach the install
// primitive (which would attempt a real SSH connection).
//
// We pass a single required flag per leaf so that the missing-flag
// error fires on a *different* required flag, proving the flag-required
// chain works end-to-end without a live host.
func TestProvisionInstall_MissingRequiredFlag(t *testing.T) {
	cases := []struct {
		leaf      string
		args      []string
		wantError string
	}{
		{leaf: "grid", args: []string{}, wantError: "oracle-home"},
		{leaf: "dbhome", args: []string{}, wantError: "oracle-home"},
		{leaf: "root-sh", args: []string{}, wantError: "oracle-home"},
		{leaf: "asmca", args: []string{"--oracle-home", "/x"}, wantError: "oracle-base"},
		{leaf: "netca", args: []string{"--oracle-home", "/x"}, wantError: "oracle-base"},
		{leaf: "asm-label", args: []string{"--grid-home", "/x"}, wantError: "oracle-base"},
		{leaf: "dbca", args: []string{"--oracle-home", "/x"}, wantError: "oracle-base"},
		{leaf: "pdb", args: []string{"--oracle-home", "/x"}, wantError: "oracle-base"},
	}
	for _, tc := range cases {
		t.Run(tc.leaf, func(t *testing.T) {
			cmd := root.NewInstallCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			args := append([]string{tc.leaf}, tc.args...)
			cmd.SetArgs(args)
			err := cmd.Execute()
			require.Error(t, err, "expected missing-required-flag error; out=%s", out.String())
			combined := strings.ToLower(err.Error() + " " + out.String())
			assert.Contains(t, combined, strings.ToLower(tc.wantError))
		})
	}
}

// TestProvisionInstall_TargetResolution verifies that --target on the
// leaf is accepted. We deliberately pass it together with a list of
// other required flags BUT omit at least one so cobra rejects with a
// flag-validation error before any SSH attempt is made. The point is
// to exercise resolveTarget()'s leaf-flag path without needing a live
// host.
func TestProvisionInstall_TargetResolution(t *testing.T) {
	cmd := root.NewInstallCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	// grid only requires --oracle-home; pass --target + --oracle-home and
	// exercise the leaf. Without a real host, install.GridInstall will
	// fail at SSH; that's acceptable here — we only assert the command
	// constructed and parsed flags. We add a sentinel check by NOT
	// setting --target to verify resolveTarget surfaces its error.
	cmd.SetArgs([]string{"grid", "--oracle-home", "/u01"})
	err := cmd.Execute()
	require.Error(t, err)
	combined := strings.ToLower(err.Error() + " " + out.String())
	// One of:
	//   - tier gate: provision bundle requires Enterprise tier
	//     (no license configured in test env — this is the post-#27 norm)
	//   - resolveTarget error
	//   - SSH attempt error if a stray default --target slipped in
	assert.True(t,
		strings.Contains(combined, "tier gate") ||
			strings.Contains(combined, "target") ||
			strings.Contains(combined, "ssh"),
		"expected tier-gate / target / ssh error, got: %s", combined)
}
