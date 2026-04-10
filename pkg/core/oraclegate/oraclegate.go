// Package oraclegate enforces Oracle edition/option license declarations.
package oraclegate

import (
	"fmt"
	"strings"

	"github.com/itunified-io/dbx/pkg/core/target"
)

// Decision is the gate's verdict.
type Decision int

const (
	Allow     Decision = iota
	Warn
	Block
	AuditOnly
)

// Requirement declares what a tool needs from Oracle licensing.
type Requirement struct {
	Edition string
	Options []string
}

// Result is the gate's check result.
type Result struct {
	Decision Decision
	Reason   string
}

// Gate checks Oracle license declarations against tool requirements.
type Gate struct {
	mode string
}

// New creates an Oracle license gate with the given enforcement mode.
func New(mode string) *Gate {
	return &Gate{mode: mode}
}

// Check evaluates whether the target's Oracle license satisfies the requirement.
func (g *Gate) Check(lic *target.OracleLicense, req Requirement) Result {
	if lic == nil {
		return Result{Decision: Allow}
	}

	if req.Edition == "enterprise" && lic.Edition != "enterprise" {
		return g.deny(fmt.Sprintf("requires edition=%s but target declares %s", req.Edition, lic.Edition))
	}

	declared := make(map[string]bool, len(lic.Options))
	for _, o := range lic.Options {
		declared[o] = true
	}
	var missing []string
	for _, o := range req.Options {
		if !declared[o] {
			missing = append(missing, o)
		}
	}
	if len(missing) > 0 {
		return g.deny(fmt.Sprintf("requires undeclared options: %s", strings.Join(missing, ", ")))
	}

	return Result{Decision: Allow}
}

func (g *Gate) deny(reason string) Result {
	switch g.mode {
	case "warn":
		return Result{Decision: Warn, Reason: reason}
	case "audit-only":
		return Result{Decision: AuditOnly, Reason: reason}
	default:
		return Result{Decision: Block, Reason: reason}
	}
}
