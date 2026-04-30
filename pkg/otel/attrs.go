// Package otel provides OpenTelemetry span emission helpers for dbx and
// downstream tools (proxctl, linuxctl, mcp-oracle-*).
//
// Design (per itunified-io/infrastructure ADR-0103 + ADR-0103a):
//
//   - dbx Enterprise audit chain (Ed25519-signed JSONL per ADR-0095) remains
//     the canonical, compliance-grade record. OTEL spans are an observability
//     mirror — best-effort, never blocks the signed write.
//   - Each audit record's signed hash appears as `dbx.audit_hash` on the
//     corresponding span, providing operator-driven cross-correlation.
//   - Span attributes follow the `dbx.*` namespace mapped to ADR-0097
//     Cloud Control target_type values.
//
// This package keeps OTEL SDK dependency out of the core dbx module by
// defining a minimal Attribute + Span type. A sibling package
// (pkg/otel/exporter) will provide the OTLP HTTP exporter once wiring
// actually flows production traffic.
package otel

import (
	"fmt"
	"strconv"
)

// Attribute is a key-value pair attached to a Span. Mirrors the OTEL
// attribute model but avoids pulling in go.opentelemetry.io/otel/attribute
// at the foundation layer.
type Attribute struct {
	Key   string
	Value string
}

// String renders an attribute in `key=value` form for log output.
func (a Attribute) String() string {
	return a.Key + "=" + a.Value
}

// StringAttr is the canonical constructor for string-valued attributes.
func StringAttr(key, val string) Attribute {
	return Attribute{Key: key, Value: val}
}

// IntAttr formats an int as an Attribute.
func IntAttr(key string, val int) Attribute {
	return Attribute{Key: key, Value: strconv.Itoa(val)}
}

// BoolAttr formats a bool as "true" or "false".
func BoolAttr(key string, val bool) Attribute {
	if val {
		return Attribute{Key: key, Value: "true"}
	}
	return Attribute{Key: key, Value: "false"}
}

// Standard attribute keys used across dbx + downstream emitters.
// Matches the conventions documented in ADR-0103a.
const (
	// Resource-level (set per emitter, attached to all spans from that emitter)
	AttrServiceName = "service.name" // claude-agent | proxctl | linuxctl | dbx | mcp-oracle-ee-dataguard

	// Plan-RAG / step correlation (ADR-0096)
	AttrStepID = "step_id" // e.g. "/lab-up step 5"
	AttrSkill  = "skill"   // calling skill name

	// dbx Cloud Control target identifiers (ADR-0097)
	AttrDbxEntityType    = "dbx.entity_type"     // oracle_database | rac_database | oracle_pdb | …
	AttrDbxEntityName    = "dbx.entity_name"     // ext3-orcl | ORCLPRI | ORCL_DG
	AttrDbxDBUniqueName  = "dbx.db_unique_name"  // ORCLPRI
	AttrDbxHost          = "dbx.host"            // ext3adm1.itunified.io
	AttrDbxLicenseTier   = "dbx.license_tier"    // community | business | enterprise
	AttrDbxAuditHash     = "dbx.audit_hash"      // hex SHA-256 of signed JSONL record (correlation key)

	// Decision metadata (cross-link with replay traces, ADR-0101)
	AttrDecision = "decision"  // executed | skipped | denied | failed
	AttrDenyRule = "deny_rule" // policy rule ID, e.g. "LAB-007"
)

// LicenseTier values, matching ADR-0094.
const (
	LicenseTierCommunity  = "community"
	LicenseTierBusiness   = "business"
	LicenseTierEnterprise = "enterprise"
)

// JoinAttrs concatenates attribute slices for ergonomic span building.
// The result preserves ordering; callers that need stable serialization
// should sort by key separately.
func JoinAttrs(slices ...[]Attribute) []Attribute {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	out := make([]Attribute, 0, total)
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}

// FormatAttrs renders attributes as `k1=v1 k2=v2` for log + debug output.
func FormatAttrs(attrs []Attribute) string {
	if len(attrs) == 0 {
		return ""
	}
	out := ""
	for i, a := range attrs {
		if i > 0 {
			out += " "
		}
		out += fmt.Sprintf("%s=%q", a.Key, a.Value)
	}
	return out
}
