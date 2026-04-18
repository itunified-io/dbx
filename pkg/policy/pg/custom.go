package pg

import (
	"github.com/itunified-io/dbx/pkg/policy"
	poracl "github.com/itunified-io/dbx/pkg/policy/oracle"
)

// RegisterCustomExecutors registers executors for custom PostgreSQL policies.
func RegisterCustomExecutors(eng *policy.Engine, db poracl.DBQuerier) {
	RegisterPGExecutors(eng, db)
}
