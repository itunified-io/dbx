// Package os provides OS-level policy check executors for CIS Linux, DISA STIG, and custom policies.
package os

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/itunified-io/dbx/pkg/policy"
)

// SSHRunner abstracts SSH command execution for testability.
type SSHRunner interface {
	Run(ctx context.Context, command string) (string, error)
}

// KernelModuleExecutor checks if a kernel module is loaded/disabled.
type KernelModuleExecutor struct{ ssh SSHRunner }

func NewKernelModuleExecutor(ssh SSHRunner) *KernelModuleExecutor {
	return &KernelModuleExecutor{ssh: ssh}
}

func (e *KernelModuleExecutor) Execute(ctx context.Context, check policy.RuleCheck) (policy.CheckResult, error) {
	out, err := e.ssh.Run(ctx, fmt.Sprintf("lsmod | grep -c %s", check.Module))
	if err != nil {
		return policy.CheckResult{Status: "error", Message: err.Error(), EvaluatedAt: time.Now()}, nil
	}
	count := strings.TrimSpace(out)
	expected, _ := check.Expected.(string)
	if expected == "disabled" && count != "0" {
		return policy.CheckResult{
			Status: "fail", Actual: "loaded (count=" + count + ")", Expected: "disabled",
			EvaluatedAt: time.Now(),
		}, nil
	}
	out, err = e.ssh.Run(ctx, fmt.Sprintf("modprobe -n -v %s", check.Module))
	if err == nil && !strings.Contains(out, "install /bin/true") && !strings.Contains(out, "install /bin/false") {
		return policy.CheckResult{
			Status: "fail", Actual: "no install override", Expected: "install /bin/true",
			EvaluatedAt: time.Now(),
		}, nil
	}
	return policy.CheckResult{Status: "pass", Actual: "disabled", EvaluatedAt: time.Now()}, nil
}

// SysctlExecutor checks sysctl kernel parameter values.
type SysctlExecutor struct{ ssh SSHRunner }

func NewSysctlExecutor(ssh SSHRunner) *SysctlExecutor {
	return &SysctlExecutor{ssh: ssh}
}

func (e *SysctlExecutor) Execute(ctx context.Context, check policy.RuleCheck) (policy.CheckResult, error) {
	out, err := e.ssh.Run(ctx, fmt.Sprintf("sysctl -n %s", check.Key))
	if err != nil {
		return policy.CheckResult{Status: "error", Message: err.Error(), EvaluatedAt: time.Now()}, nil
	}
	actual := strings.TrimSpace(out)
	expected, _ := check.Expected.(string)
	if actual != expected {
		return policy.CheckResult{
			Status: "fail", Actual: actual, Expected: expected,
			EvaluatedAt: time.Now(),
		}, nil
	}
	return policy.CheckResult{Status: "pass", Actual: actual, EvaluatedAt: time.Now()}, nil
}

// FileContentExecutor checks file content against a regex pattern.
type FileContentExecutor struct{ ssh SSHRunner }

func NewFileContentExecutor(ssh SSHRunner) *FileContentExecutor {
	return &FileContentExecutor{ssh: ssh}
}

func (e *FileContentExecutor) Execute(ctx context.Context, check policy.RuleCheck) (policy.CheckResult, error) {
	out, err := e.ssh.Run(ctx, fmt.Sprintf("grep -E '%s' %s", check.Pattern, check.Path))
	if err != nil {
		return policy.CheckResult{Status: "fail", Actual: "(not found)", Expected: check.Pattern, EvaluatedAt: time.Now()}, nil
	}
	line := strings.TrimSpace(out)
	expected, _ := check.Expected.(string)
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		actual := parts[len(parts)-1]
		if actual != expected {
			return policy.CheckResult{
				Status: "fail", Actual: actual, Expected: expected,
				EvaluatedAt: time.Now(),
			}, nil
		}
	}
	return policy.CheckResult{Status: "pass", Actual: line, EvaluatedAt: time.Now()}, nil
}

// FilePermissionExecutor checks file ownership and permissions.
type FilePermissionExecutor struct{ ssh SSHRunner }

func NewFilePermissionExecutor(ssh SSHRunner) *FilePermissionExecutor {
	return &FilePermissionExecutor{ssh: ssh}
}

func (e *FilePermissionExecutor) Execute(ctx context.Context, check policy.RuleCheck) (policy.CheckResult, error) {
	out, err := e.ssh.Run(ctx, fmt.Sprintf("stat -c '%%a %%U %%G' %s", check.Path))
	if err != nil {
		return policy.CheckResult{Status: "error", Message: err.Error(), EvaluatedAt: time.Now()}, nil
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) < 3 {
		return policy.CheckResult{Status: "error", Message: "unexpected stat output", Actual: out, EvaluatedAt: time.Now()}, nil
	}
	perm, owner, group := parts[0], parts[1], parts[2]
	expectedPerm := strings.TrimLeft(check.Permission, "0")
	if expectedPerm == "" {
		expectedPerm = "0"
	}
	actualPerm := strings.TrimLeft(perm, "0")
	if actualPerm == "" {
		actualPerm = "0"
	}

	if actualPerm != expectedPerm || owner != check.Owner || group != check.Group {
		return policy.CheckResult{
			Status:      "fail",
			Actual:      fmt.Sprintf("%s %s %s", perm, owner, group),
			Expected:    fmt.Sprintf("%s %s %s", check.Permission, check.Owner, check.Group),
			EvaluatedAt: time.Now(),
		}, nil
	}
	return policy.CheckResult{Status: "pass", Actual: out, EvaluatedAt: time.Now()}, nil
}

// CommandOutputExecutor runs an arbitrary command and checks the output.
type CommandOutputExecutor struct{ ssh SSHRunner }

func NewCommandOutputExecutor(ssh SSHRunner) *CommandOutputExecutor {
	return &CommandOutputExecutor{ssh: ssh}
}

func (e *CommandOutputExecutor) Execute(ctx context.Context, check policy.RuleCheck) (policy.CheckResult, error) {
	out, err := e.ssh.Run(ctx, check.Command)
	if err != nil {
		return policy.CheckResult{Status: "error", Message: err.Error(), EvaluatedAt: time.Now()}, nil
	}
	actual := strings.TrimSpace(out)
	expected, _ := check.Expected.(string)
	if actual != expected {
		return policy.CheckResult{
			Status: "fail", Actual: actual, Expected: expected,
			EvaluatedAt: time.Now(),
		}, nil
	}
	return policy.CheckResult{Status: "pass", Actual: actual, EvaluatedAt: time.Now()}, nil
}

// RegisterOSExecutors registers all OS-level check executors with the engine.
func RegisterOSExecutors(eng *policy.Engine, ssh SSHRunner) {
	eng.RegisterExecutor("kernel_module", NewKernelModuleExecutor(ssh))
	eng.RegisterExecutor("sysctl_value", NewSysctlExecutor(ssh))
	eng.RegisterExecutor("file_content", NewFileContentExecutor(ssh))
	eng.RegisterExecutor("file_permission", NewFilePermissionExecutor(ssh))
	eng.RegisterExecutor("command_output", NewCommandOutputExecutor(ssh))
}
