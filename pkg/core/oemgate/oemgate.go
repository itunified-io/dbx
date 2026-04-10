// Package oemgate enforces OEM Management Pack declarations.
package oemgate

import (
	"fmt"

	"github.com/itunified-io/dbx/pkg/core/target"
)

// Result of an OEM pack gate check.
type Result struct {
	Allowed bool
	Reason  string
}

// Gate checks OEM Management Pack declarations.
type Gate struct {
	mode string
}

// New creates an OEM gate with the given enforcement mode.
func New(mode string) *Gate {
	return &Gate{mode: mode}
}

// Check evaluates whether the target declares the required OEM pack.
func (g *Gate) Check(lic *target.OracleLicense, requiredPack string) Result {
	if lic == nil {
		return Result{Allowed: true}
	}

	for _, p := range lic.OEMPacks {
		if p == requiredPack {
			return Result{Allowed: true}
		}
	}

	reason := fmt.Sprintf("OEM pack %q not declared on target", requiredPack)
	if g.mode == "strict" {
		return Result{Allowed: false, Reason: reason}
	}
	return Result{Allowed: true, Reason: reason}
}
