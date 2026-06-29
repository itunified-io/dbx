package root_test

import (
	"bytes"
	"testing"

	"github.com/itunified-io/dbx/cmd/dbxcli/root"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgCmdSubcommands(t *testing.T) {
	cmd := root.NewPgCmd()
	names := make([]string, 0, len(cmd.Commands()))
	for _, c := range cmd.Commands() {
		names = append(names, c.Name())
	}
	expected := []string{
		"connect", "query", "schema", "crud", "dba", "dba-adv", "perf", "health",
		"security", "audit", "comply", "rbac", "repl", "ha", "backup", "migrate",
		"observe", "tenant", "wal", "cnpg", "dr", "rag", "vault", "policy",
	}
	// 24 groups in OSS (capacity is EE-only, registered by dbx-ee)
	assert.Len(t, cmd.Commands(), 24)
	for _, exp := range expected {
		assert.Contains(t, names, exp, "missing subcommand: %s", exp)
	}
}

// runPg builds a fresh pgCmd, sets args, silences usage, and returns the error
// from Execute. This lets each test exercise the full cobra dispatch path.
func runPg(t *testing.T, args ...string) error {
	t.Helper()
	cmd := root.NewPgCmd()
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	// Propagate silence to all children so failure output is clean.
	var silence func(*cobra.Command)
	silence = func(c *cobra.Command) {
		c.SilenceUsage = true
		c.SilenceErrors = true
		for _, ch := range c.Commands() {
			silence(ch)
		}
	}
	silence(cmd)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// ADR-0047: pg crud delete must reject a call with no confirm_table.
func TestPgCrudDeleteMissingRestatementBlocks(t *testing.T) {
	err := runPg(t, "crud", "delete", "schema=public", "table=users", "where=id=1")
	require.Error(t, err)
	assert.ErrorContains(t, err, "identifier confirmation required")
}

func TestPgCrudDeleteWrongRestatementBlocks(t *testing.T) {
	err := runPg(t, "crud", "delete", "schema=public", "table=users", "where=id=1", "confirm_table=orders")
	require.Error(t, err)
	assert.ErrorContains(t, err, "identifier confirmation mismatch")
}

func TestPgCrudDeleteCorrectRestatementProceeds(t *testing.T) {
	err := runPg(t, "crud", "delete", "schema=public", "table=users", "where=id=1", "confirm_table=users")
	assert.NoError(t, err)
}

// ADR-0047: pg rag collection-drop must reject a call with no confirm_name.
func TestPgRagCollectionDropMissingRestatementBlocks(t *testing.T) {
	err := runPg(t, "rag", "collection-drop", "name=embeddings")
	require.Error(t, err)
	assert.ErrorContains(t, err, "identifier confirmation required")
}

func TestPgRagCollectionDropWrongRestatementBlocks(t *testing.T) {
	err := runPg(t, "rag", "collection-drop", "name=embeddings", "confirm_name=wrong-collection")
	require.Error(t, err)
	assert.ErrorContains(t, err, "identifier confirmation mismatch")
}

func TestPgRagCollectionDropCorrectRestatementProceeds(t *testing.T) {
	err := runPg(t, "rag", "collection-drop", "name=embeddings", "confirm_name=embeddings")
	assert.NoError(t, err)
}
