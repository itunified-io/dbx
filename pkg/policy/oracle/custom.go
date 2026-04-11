package oracle

import "github.com/itunified-io/dbx/pkg/policy"

// RegisterCustomExecutors registers executors for custom Oracle policies.
func RegisterCustomExecutors(eng *policy.Engine, db DBQuerier) {
	RegisterOracleExecutors(eng, db)
}
