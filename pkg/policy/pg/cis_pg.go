// Package pg provides SQL-based policy check executors for PostgreSQL.
package pg

import (
	"github.com/itunified-io/dbx/pkg/policy"
	poracl "github.com/itunified-io/dbx/pkg/policy/oracle"
)

// RegisterPGExecutors registers SQL-based executors for PostgreSQL policies.
// PG uses the same SQLExecutor as Oracle — the difference is in the policy YAML (PG-specific queries).
func RegisterPGExecutors(eng *policy.Engine, db poracl.DBQuerier) {
	eng.RegisterExecutor("sql", poracl.NewSQLExecutor(db))
}
