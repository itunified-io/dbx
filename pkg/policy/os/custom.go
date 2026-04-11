package os

import "github.com/itunified-io/dbx/pkg/policy"

// RegisterCustomExecutors registers executors for custom OS policies.
// Custom policies use the same check types as CIS/STIG.
func RegisterCustomExecutors(eng *policy.Engine, ssh SSHRunner) {
	RegisterOSExecutors(eng, ssh)
}
