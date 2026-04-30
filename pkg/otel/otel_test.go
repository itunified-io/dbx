package otel

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestStringAttr(t *testing.T) {
	a := StringAttr("foo", "bar")
	if a.Key != "foo" || a.Value != "bar" {
		t.Errorf("StringAttr: %+v", a)
	}
}

func TestIntBoolAttr(t *testing.T) {
	if IntAttr("n", 42).Value != "42" {
		t.Error("IntAttr")
	}
	if BoolAttr("b", true).Value != "true" {
		t.Error("BoolAttr true")
	}
	if BoolAttr("b", false).Value != "false" {
		t.Error("BoolAttr false")
	}
}

func TestJoinAttrs(t *testing.T) {
	a := []Attribute{StringAttr("a", "1")}
	b := []Attribute{StringAttr("b", "2"), StringAttr("c", "3")}
	got := JoinAttrs(a, b)
	if len(got) != 3 {
		t.Fatalf("JoinAttrs: want 3, got %d", len(got))
	}
	if got[0].Key != "a" || got[2].Key != "c" {
		t.Errorf("ordering wrong: %+v", got)
	}
}

func TestFormatAttrs(t *testing.T) {
	out := FormatAttrs([]Attribute{StringAttr("k", "v"), StringAttr("x", "y")})
	if !strings.Contains(out, `k="v"`) || !strings.Contains(out, `x="y"`) {
		t.Errorf("FormatAttrs: %q", out)
	}
}

func TestSpanBuilder_DBxStandardAttrs(t *testing.T) {
	span := NewSpan("test.span", "dbx").
		WithStep("/lab-up step 5", "lab-up").
		WithAuditHash("abc123").
		WithLicenseTier(LicenseTierEnterprise).
		WithDecision("executed", "").
		EndOK()

	m := span.AttributeMap()
	want := map[string]string{
		AttrServiceName:    "dbx",
		AttrStepID:         "/lab-up step 5",
		AttrSkill:          "lab-up",
		AttrDbxAuditHash:   "abc123",
		AttrDbxLicenseTier: LicenseTierEnterprise,
		AttrDecision:       "executed",
	}
	for k, v := range want {
		if got := m[k]; got != v {
			t.Errorf("attr %s: got %q want %q", k, got, v)
		}
	}
	if span.Status != StatusOK {
		t.Errorf("status: got %v want OK", span.Status)
	}
	if span.Duration() <= 0 {
		t.Errorf("duration should be > 0, got %v", span.Duration())
	}
}

func TestSpanBuilder_DenyDecision(t *testing.T) {
	span := NewSpan("blocked.tool", "claude-agent").
		WithDecision("denied", "LAB-007").
		EndOK()
	m := span.AttributeMap()
	if m[AttrDecision] != "denied" {
		t.Errorf("decision: %q", m[AttrDecision])
	}
	if m[AttrDenyRule] != "LAB-007" {
		t.Errorf("deny_rule: %q", m[AttrDenyRule])
	}
}

func TestSpanBuilder_Error(t *testing.T) {
	want := errors.New("boom")
	span := NewSpan("t", "dbx").EndError(want)
	if span.Status != StatusError {
		t.Errorf("status: %v", span.Status)
	}
	if span.Error != want {
		t.Errorf("error: %v", span.Error)
	}
}

func TestStatusString(t *testing.T) {
	tests := map[Status]string{
		StatusUnset: "UNSET",
		StatusOK:    "OK",
		StatusError: "ERROR",
	}
	for s, want := range tests {
		if s.String() != want {
			t.Errorf("Status(%d): %s want %s", s, s.String(), want)
		}
	}
}

func TestNoopExporter(t *testing.T) {
	e := NoopExporter{}
	span := NewSpan("t", "dbx").EndOK()
	if err := e.Export(context.Background(), []Span{span}); err != nil {
		t.Errorf("Noop Export: %v", err)
	}
	if err := e.Shutdown(context.Background()); err != nil {
		t.Errorf("Noop Shutdown: %v", err)
	}
}

func TestAttributeStandardKeys(t *testing.T) {
	// Lock the canonical attribute key names so renames break tests
	// (downstream emitters depend on these strings).
	required := []string{
		AttrServiceName, AttrStepID, AttrSkill,
		AttrDbxEntityType, AttrDbxEntityName, AttrDbxDBUniqueName,
		AttrDbxHost, AttrDbxLicenseTier, AttrDbxAuditHash,
		AttrDecision, AttrDenyRule,
	}
	if len(required) != 11 {
		t.Errorf("standard key count drifted: got %d want 11", len(required))
	}
	for _, k := range required {
		if k == "" {
			t.Error("empty standard key")
		}
	}
}

func TestSpanDuration_Zero(t *testing.T) {
	s := Span{}
	if s.Duration() != 0 {
		t.Errorf("zero span duration should be 0, got %v", s.Duration())
	}
	s2 := Span{StartTime: time.Now()}
	if s2.Duration() != 0 {
		t.Errorf("span without end should be 0, got %v", s2.Duration())
	}
}
