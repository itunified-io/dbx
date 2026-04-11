package os_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/itunified-io/dbx/pkg/policy"
	pos "github.com/itunified-io/dbx/pkg/policy/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSSH struct {
	outputs map[string]string
}

func (m *mockSSH) Run(_ context.Context, command string) (string, error) {
	if out, ok := m.outputs[command]; ok {
		return out, nil
	}
	return "", fmt.Errorf("command not mocked: %s", command)
}

func TestKernelModuleExecutor_Disabled(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"lsmod | grep -c cramfs":  "0",
		"modprobe -n -v cramfs":   "install /bin/true",
	}}
	exec := pos.NewKernelModuleExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "kernel_module", Module: "cramfs", Expected: "disabled",
	})
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestKernelModuleExecutor_Loaded(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"lsmod | grep -c cramfs": "1",
	}}
	exec := pos.NewKernelModuleExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "kernel_module", Module: "cramfs", Expected: "disabled",
	})
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
}

func TestSysctlExecutor_Pass(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"sysctl -n fs.suid_dumpable": "0",
	}}
	exec := pos.NewSysctlExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "sysctl_value", Key: "fs.suid_dumpable", Expected: "0",
	})
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestSysctlExecutor_Fail(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"sysctl -n fs.suid_dumpable": "2",
	}}
	exec := pos.NewSysctlExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "sysctl_value", Key: "fs.suid_dumpable", Expected: "0",
	})
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
	assert.Equal(t, "2", result.Actual)
}

func TestFileContentExecutor_Match(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"grep -E '^PermitRootLogin' /etc/ssh/sshd_config": "PermitRootLogin no",
	}}
	exec := pos.NewFileContentExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "file_content", Path: "/etc/ssh/sshd_config", Pattern: "^PermitRootLogin", Expected: "no",
	})
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestFileContentExecutor_NoMatch(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"grep -E '^PermitRootLogin' /etc/ssh/sshd_config": "PermitRootLogin yes",
	}}
	exec := pos.NewFileContentExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "file_content", Path: "/etc/ssh/sshd_config", Pattern: "^PermitRootLogin", Expected: "no",
	})
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
}

func TestFilePermissionExecutor_Correct(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"stat -c '%a %U %G' /etc/ssh/sshd_config": "600 root root",
	}}
	exec := pos.NewFilePermissionExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "file_permission", Path: "/etc/ssh/sshd_config",
		Permission: "0600", Owner: "root", Group: "root",
	})
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestFilePermissionExecutor_Wrong(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"stat -c '%a %U %G' /etc/ssh/sshd_config": "644 root root",
	}}
	exec := pos.NewFilePermissionExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "file_permission", Path: "/etc/ssh/sshd_config",
		Permission: "0600", Owner: "root", Group: "root",
	})
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
}

func TestCommandOutputExecutor_EmptyExpected(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"awk -F: '($3 < 1000)' /etc/passwd": "",
	}}
	exec := pos.NewCommandOutputExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "command_output", Command: "awk -F: '($3 < 1000)' /etc/passwd", Expected: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "pass", result.Status)
}

func TestCommandOutputExecutor_NonEmpty(t *testing.T) {
	ssh := &mockSSH{outputs: map[string]string{
		"awk -F: '($3 < 1000)' /etc/passwd": "daemon",
	}}
	exec := pos.NewCommandOutputExecutor(ssh)
	result, err := exec.Execute(context.Background(), policy.RuleCheck{
		Type: "command_output", Command: "awk -F: '($3 < 1000)' /etc/passwd", Expected: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "fail", result.Status)
}

func TestRegisterOSExecutors(t *testing.T) {
	eng := policy.NewEngine(policy.EngineOpts{Concurrency: 1})
	ssh := &mockSSH{outputs: map[string]string{
		"sysctl -n fs.suid_dumpable": "0",
	}}
	pos.RegisterOSExecutors(eng, ssh)

	p := &policy.Policy{
		Metadata: policy.PolicyMetadata{Name: "test", Framework: "cis", Scope: "host"},
		Rules: []policy.Rule{
			{ID: "1.5.1", Title: "Core dumps", Severity: "medium", Check: policy.RuleCheck{
				Type: "sysctl_value", Key: "fs.suid_dumpable", Expected: "0",
			}},
		},
		SHA256: "test",
	}
	result, err := eng.Scan(context.Background(), "srv1", "host", p)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Summary.Passed)
}
