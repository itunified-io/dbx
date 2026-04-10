package oemgate_test

import (
	"testing"

	"github.com/itunified-io/dbx/pkg/core/oemgate"
	"github.com/itunified-io/dbx/pkg/core/target"
	"github.com/stretchr/testify/assert"
)

func TestAllowWhenPackDeclared(t *testing.T) {
	gate := oemgate.New("strict")
	lic := &target.OracleLicense{
		OEMPacks: []string{"diagnostics", "lifecycle"},
	}
	result := gate.Check(lic, "diagnostics")
	assert.True(t, result.Allowed)
}

func TestBlockWhenPackMissing(t *testing.T) {
	gate := oemgate.New("strict")
	lic := &target.OracleLicense{
		OEMPacks: []string{"diagnostics"},
	}
	result := gate.Check(lic, "tuning")
	assert.False(t, result.Allowed)
	assert.Contains(t, result.Reason, "tuning")
}

func TestWarnMode(t *testing.T) {
	gate := oemgate.New("warn")
	lic := &target.OracleLicense{OEMPacks: []string{}}
	result := gate.Check(lic, "diagnostics")
	assert.True(t, result.Allowed)
	assert.NotEmpty(t, result.Reason)
}

func TestNilLicenseSkips(t *testing.T) {
	gate := oemgate.New("strict")
	result := gate.Check(nil, "diagnostics")
	assert.True(t, result.Allowed)
}
