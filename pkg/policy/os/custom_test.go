package os_test

import (
	"context"
	"testing"

	"github.com/itunified-io/dbx/pkg/policy"
	pos "github.com/itunified-io/dbx/pkg/policy/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterCustomExecutors(t *testing.T) {
	eng := policy.NewEngine(policy.EngineOpts{Concurrency: 1})
	ssh := &mockSSH{outputs: map[string]string{
		"sysctl -n net.ipv4.ip_forward": "0",
	}}
	pos.RegisterCustomExecutors(eng, ssh)

	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{Name: "custom-test", Framework: "custom", Scope: "host"},
		Rules: []policy.Rule{
			{ID: "CUSTOM-01", Title: "IP forwarding", Severity: "medium", Check: policy.RuleCheck{
				Type: "sysctl_value", Key: "net.ipv4.ip_forward", Expected: "0",
			}},
		},
		SHA256: "custom",
	}
	result, err := eng.Scan(context.Background(), "srv1", "host", p)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Summary.Passed)
}
