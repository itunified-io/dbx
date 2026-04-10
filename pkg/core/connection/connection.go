// Package connection provides database connection management interfaces.
package connection

import (
	"context"
	"fmt"

	"github.com/itunified-io/dbx/pkg/core/target"
)

// QueryResult holds the result of a SQL query.
type QueryResult struct {
	Columns []string
	Rows    [][]any
}

// Manager handles database connections for different engines.
type Manager interface {
	Query(ctx context.Context, tgt *target.Target, query string, args ...any) (*QueryResult, error)
	Exec(ctx context.Context, tgt *target.Target, stmt string, args ...any) (int64, error)
	Ping(ctx context.Context, tgt *target.Target) error
	Close() error
}

// NewManager creates a connection manager based on the target type.
// In P1 this returns a stub. godror (Oracle) is added in P2, pgx (PG) in P12.
func NewManager(tgt *target.Target) (Manager, error) {
	switch {
	case tgt.IsOracle():
		return &stubManager{engine: "oracle"}, nil
	case tgt.IsPostgres():
		return &stubManager{engine: "postgres"}, nil
	case tgt.IsHost():
		return &stubManager{engine: "host"}, nil
	default:
		return nil, fmt.Errorf("unsupported target type: %s", tgt.Type)
	}
}

type stubManager struct {
	engine string
}

func (s *stubManager) Query(ctx context.Context, tgt *target.Target, query string, args ...any) (*QueryResult, error) {
	return nil, fmt.Errorf("%s connection not yet implemented", s.engine)
}

func (s *stubManager) Exec(ctx context.Context, tgt *target.Target, stmt string, args ...any) (int64, error) {
	return 0, fmt.Errorf("%s connection not yet implemented", s.engine)
}

func (s *stubManager) Ping(ctx context.Context, tgt *target.Target) error {
	return fmt.Errorf("%s connection not yet implemented", s.engine)
}

func (s *stubManager) Close() error {
	return nil
}
