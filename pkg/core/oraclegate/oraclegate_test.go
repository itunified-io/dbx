package oraclegate_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/core/oraclegate"
	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/stretchr/testify/assert"
)

func eeLicense() *target.OracleLicense {
	return &target.OracleLicense{
		Edition: "enterprise",
		Options: []string{"partitioning", "diagnostics_pack", "tuning_pack"},
	}
}

func se2License() *target.OracleLicense {
	return &target.OracleLicense{
		Edition: "standard2",
		Options: []string{},
	}
}

func TestAllowWhenOptionDeclared(t *testing.T) {
	gate := oraclegate.New("strict")
	result := gate.Check(eeLicense(), oraclegate.Requirement{
		Edition: "enterprise",
		Options: []string{"diagnostics_pack"},
	})
	assert.Equal(t, oraclegate.Allow, result.Decision)
}

func TestBlockWhenOptionMissing(t *testing.T) {
	gate := oraclegate.New("strict")
	result := gate.Check(eeLicense(), oraclegate.Requirement{
		Edition: "enterprise",
		Options: []string{"advanced_security"},
	})
	assert.Equal(t, oraclegate.Block, result.Decision)
	assert.Contains(t, result.Reason, "advanced_security")
}

func TestBlockSE2ForEEFeature(t *testing.T) {
	gate := oraclegate.New("strict")
	result := gate.Check(se2License(), oraclegate.Requirement{
		Edition: "enterprise",
	})
	assert.Equal(t, oraclegate.Block, result.Decision)
	assert.Contains(t, result.Reason, "enterprise")
}

func TestWarnMode(t *testing.T) {
	gate := oraclegate.New("warn")
	result := gate.Check(se2License(), oraclegate.Requirement{
		Edition: "enterprise",
	})
	assert.Equal(t, oraclegate.Warn, result.Decision)
}

func TestAuditOnlyMode(t *testing.T) {
	gate := oraclegate.New("audit-only")
	result := gate.Check(se2License(), oraclegate.Requirement{
		Edition: "enterprise",
	})
	assert.Equal(t, oraclegate.AuditOnly, result.Decision)
}

func TestNoGateForPostgres(t *testing.T) {
	gate := oraclegate.New("strict")
	result := gate.Check(nil, oraclegate.Requirement{Edition: "enterprise"})
	assert.Equal(t, oraclegate.Allow, result.Decision)
}
