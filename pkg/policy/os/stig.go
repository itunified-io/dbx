package os

import "github.com/itunified-io/dbx/pkg/policy"

// STIGSeverityMap maps DISA CAT levels to policy severity.
var STIGSeverityMap = map[string]string{
	"CAT I":   "critical",
	"CAT II":  "high",
	"CAT III": "medium",
}

// RegisterSTIGExecutors registers STIG executors — same check types as CIS.
// The difference is in the YAML policy file (STIG IDs, CAT levels).
func RegisterSTIGExecutors(eng *policy.Engine, ssh SSHRunner) {
	RegisterOSExecutors(eng, ssh)
}
