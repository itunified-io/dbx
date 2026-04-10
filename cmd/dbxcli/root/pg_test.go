package root_test

import (
	"testing"

	"github.com/itunified-io/dbx/cmd/dbxcli/root"
	"github.com/stretchr/testify/assert"
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
