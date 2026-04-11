package oracle

import "github.com/itunified-io/dbx/pkg/policy"

// STIGSeverityMap maps DISA CAT levels to Oracle policy severity.
var STIGSeverityMap = map[string]string{
	"CAT I":   "critical",
	"CAT II":  "high",
	"CAT III": "medium",
}

// RegisterSTIGExecutors registers STIG executors — same SQL check type.
func RegisterSTIGExecutors(eng *policy.Engine, db DBQuerier) {
	RegisterOracleExecutors(eng, db)
}
