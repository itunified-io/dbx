package policy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/itunified-io/dbx/pkg/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPolicy_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
metadata:
  name: Test CIS Linux
  version: "1.0.0"
  framework: cis
  scope: host
  distros: [ubuntu, rhel]
rules:
  - id: "1.1.1"
    title: "Test rule"
    severity: medium
    check:
      type: kernel_module
      module: cramfs
      expected: disabled
    remediation:
      command: "echo test"
      confirm_required: true
`
	err := os.WriteFile(filepath.Join(dir, "test.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	p, err := policy.LoadFile(filepath.Join(dir, "test.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "Test CIS Linux", p.Metadata.Name)
	assert.Equal(t, "cis", p.Metadata.Framework)
	assert.Equal(t, "host", p.Metadata.Scope)
	assert.Len(t, p.Rules, 1)
	assert.NotEmpty(t, p.SHA256)
}

func TestLoadPolicy_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte("{{invalid}}"), 0644)
	require.NoError(t, err)
	_, err = policy.LoadFile(filepath.Join(dir, "bad.yaml"))
	assert.Error(t, err)
}

func TestLoadDirectory_MultiplePolicies(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.yaml", "b.yaml"} {
		content := `
metadata:
  name: ` + name + `
  version: "1.0.0"
  framework: cis
  scope: host
rules: []
`
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
		require.NoError(t, err)
	}
	policies, err := policy.LoadDirectory(dir)
	require.NoError(t, err)
	assert.Len(t, policies, 2)
}

func TestLoadPolicy_SHA256_Deterministic(t *testing.T) {
	dir := t.TempDir()
	content := `
metadata:
  name: SHA Test
  version: "1.0.0"
  framework: cis
  scope: host
rules: []
`
	path := filepath.Join(dir, "sha.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	p1, _ := policy.LoadFile(path)
	p2, _ := policy.LoadFile(path)
	assert.Equal(t, p1.SHA256, p2.SHA256)
	assert.NotEmpty(t, p1.SHA256)
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	reg := policy.NewRegistry()
	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{
			Name:      "CIS Linux L1",
			Framework: "cis",
			Scope:     "host",
		},
	}
	reg.Register(p)
	got := reg.Get("host", "cis")
	assert.Len(t, got, 1)
	assert.Equal(t, "CIS Linux L1", got[0].Metadata.Name)

	// Get by scope only
	all := reg.Get("host", "")
	assert.Len(t, all, 1)

	// Get nonexistent
	none := reg.Get("oracle_database", "cis")
	assert.Len(t, none, 0)
}

func TestRegistry_All(t *testing.T) {
	reg := policy.NewRegistry()
	reg.Register(&policy.Policy{Metadata: policy.PolicyMetadata{Scope: "host", Framework: "cis"}})
	reg.Register(&policy.Policy{Metadata: policy.PolicyMetadata{Scope: "host", Framework: "stig"}})
	reg.Register(&policy.Policy{Metadata: policy.PolicyMetadata{Scope: "oracle_database", Framework: "cis"}})
	assert.Len(t, reg.All(), 3)
}

func TestRegistry_Reload(t *testing.T) {
	dir := t.TempDir()
	content := `
metadata:
  name: Reload Test
  version: "1.0.0"
  framework: custom
  scope: host
rules: []
`
	err := os.WriteFile(filepath.Join(dir, "r.yaml"), []byte(content), 0644)
	require.NoError(t, err)

	reg := policy.NewRegistry()
	err = reg.Reload(dir)
	require.NoError(t, err)
	assert.Len(t, reg.All(), 1)

	// Reload clears and reloads
	err = reg.Reload(dir)
	require.NoError(t, err)
	assert.Len(t, reg.All(), 1) // still 1, not 2
}
