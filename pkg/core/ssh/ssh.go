package ssh

import "fmt"

// ExecRequest describes an SSH command to execute.
type ExecRequest struct {
	User         string
	Host         string
	KeyPath      string
	Command      string
	Args         []string
	StdinContent string // For piping scripts (e.g., RMAN .rcv files)
}

// Executor validates and builds SSH commands against an allowlist.
type Executor struct {
	allowlist Allowlist
}

// NewExecutor creates an SSH executor with the given allowlist.
func NewExecutor(al Allowlist) *Executor {
	return &Executor{allowlist: al}
}

// IsAllowed checks whether a (domain, command) pair is in the allowlist.
func (e *Executor) IsAllowed(domain, command string) bool {
	return e.allowlist.Has(domain, command)
}

// BuildArgs constructs SSH command arguments. Returns an error if the command
// is not in any allowlisted domain.
func (e *Executor) BuildArgs(user, host, keyPath, command string, cmdArgs []string) ([]string, error) {
	allowed := false
	for domain := range e.allowlist {
		if e.allowlist.Has(domain, command) {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("command %q is not in the SSH allowlist", command)
	}

	args := []string{
		"-i", keyPath,
		"-o", "StrictHostKeyChecking=accept-new",
		user + "@" + host,
		command,
	}
	args = append(args, cmdArgs...)
	return args, nil
}
