package otel

import (
	"context"
	"time"
)

// Span is a minimal OTEL-compatible span representation. A full OTLP
// exporter (sibling PR) maps these onto the official Span proto.
type Span struct {
	Name        string
	StartTime   time.Time
	EndTime     time.Time
	Attributes  []Attribute
	Status      Status
	Error       error
	ServiceName string
}

// Status mirrors OTEL StatusCode without depending on the SDK.
type Status int

const (
	StatusUnset Status = iota
	StatusOK
	StatusError
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusError:
		return "ERROR"
	}
	return "UNSET"
}

// Duration returns end - start, or 0 if either is unset.
func (s Span) Duration() time.Duration {
	if s.StartTime.IsZero() || s.EndTime.IsZero() {
		return 0
	}
	return s.EndTime.Sub(s.StartTime)
}

// AttributeMap returns Attributes flattened to a map for lookups.
// Later writes override earlier ones (consistent with OTEL semantics).
func (s Span) AttributeMap() map[string]string {
	m := make(map[string]string, len(s.Attributes))
	for _, a := range s.Attributes {
		m[a.Key] = a.Value
	}
	return m
}

// Exporter is the minimal interface a span exporter must satisfy.
// Implementations:
//   - StdoutExporter (testing, debugging)
//   - OTLPExporter   (sibling PR — POSTs to OTLP HTTP endpoint)
//   - NoopExporter   (audit dual-sink: when OTEL endpoint is unset, fall through silently)
type Exporter interface {
	Export(ctx context.Context, spans []Span) error
	Shutdown(ctx context.Context) error
}

// NoopExporter discards every span. Used as the default when no OTEL
// endpoint is configured — keeps callers simple (always have a valid
// exporter) without forcing infrastructure setup.
type NoopExporter struct{}

func (NoopExporter) Export(_ context.Context, _ []Span) error { return nil }
func (NoopExporter) Shutdown(_ context.Context) error         { return nil }

// SpanBuilder helps construct a Span with the dbx-standard attributes.
type SpanBuilder struct {
	span Span
}

// NewSpan starts a new SpanBuilder with the given name + service.
// Use the ForTarget / ForSkill / WithStep helpers to chain attributes,
// then End() to materialize the Span for export.
func NewSpan(name, serviceName string) *SpanBuilder {
	return &SpanBuilder{
		span: Span{
			Name:        name,
			StartTime:   time.Now(),
			ServiceName: serviceName,
			Attributes: []Attribute{
				StringAttr(AttrServiceName, serviceName),
			},
		},
	}
}

// WithStep sets the Plan-RAG step_id and skill name.
func (b *SpanBuilder) WithStep(stepID, skill string) *SpanBuilder {
	b.span.Attributes = append(b.span.Attributes,
		StringAttr(AttrStepID, stepID),
		StringAttr(AttrSkill, skill),
	)
	return b
}

// WithAuditHash sets dbx.audit_hash so the span correlates with the
// signed JSONL audit chain entry (ADR-0095).
func (b *SpanBuilder) WithAuditHash(hash string) *SpanBuilder {
	b.span.Attributes = append(b.span.Attributes, StringAttr(AttrDbxAuditHash, hash))
	return b
}

// WithLicenseTier sets dbx.license_tier (per ADR-0094).
func (b *SpanBuilder) WithLicenseTier(tier string) *SpanBuilder {
	b.span.Attributes = append(b.span.Attributes, StringAttr(AttrDbxLicenseTier, tier))
	return b
}

// WithDecision records the tool-call decision + optional deny rule
// (Items 1 + 2 of agentic-AI hardening).
func (b *SpanBuilder) WithDecision(decision, denyRule string) *SpanBuilder {
	b.span.Attributes = append(b.span.Attributes, StringAttr(AttrDecision, decision))
	if denyRule != "" {
		b.span.Attributes = append(b.span.Attributes, StringAttr(AttrDenyRule, denyRule))
	}
	return b
}

// WithAttrs appends arbitrary attributes (e.g. from Target.OTELAttrs()).
func (b *SpanBuilder) WithAttrs(attrs ...Attribute) *SpanBuilder {
	b.span.Attributes = append(b.span.Attributes, attrs...)
	return b
}

// EndOK marks the span successful and returns it.
func (b *SpanBuilder) EndOK() Span {
	b.span.EndTime = time.Now()
	b.span.Status = StatusOK
	return b.span
}

// EndError marks the span errored with the given error and returns it.
func (b *SpanBuilder) EndError(err error) Span {
	b.span.EndTime = time.Now()
	b.span.Status = StatusError
	b.span.Error = err
	return b.span
}
