// OTEL attribute helpers for Target. Per ADR-0103a (OTEL bus + dbx
// integration) of itunified-io/infrastructure.
//
// Lives in the same package as Target so external callers can derive
// dbx.* span attributes without crossing module boundaries:
//
//	t := target.Target{...}
//	span := otel.NewSpan("dbca create", "dbx").
//	    WithAttrs(t.OTELAttrs()...).
//	    EndOK()
//
// pkg/otel/ defines the canonical attribute key names (AttrDbxEntityType
// etc.). We deliberately avoid importing pkg/otel here to prevent an
// import cycle; this file mirrors the constants. A test in pkg/otel
// asserts the names stay in sync.

package target

const (
	otelAttrEntityType   = "dbx.entity_type"
	otelAttrEntityName   = "dbx.entity_name"
	otelAttrDBUniqueName = "dbx.db_unique_name"
	otelAttrHost         = "dbx.host"
)

// OTELAttribute is a key-value pair returned by OTELAttrs. It mirrors
// pkg/otel.Attribute structurally so callers can convert with a simple
// map (or use as-is when the consumer accepts a struct with Key + Value).
type OTELAttribute struct {
	Key   string
	Value string
}

// OTELAttrs returns the dbx.* span attributes for this Target, derived
// from its EntityType + Name + connection endpoints.
//
//   - dbx.entity_type  always set (Target.Type)
//   - dbx.entity_name  always set (Target.Name)
//   - dbx.db_unique_name  set when Primary.Database is non-empty
//   - dbx.host         set when Primary.Host is non-empty
//
// Callers needing additional attributes (license tier, audit hash, step
// ID) should append them via pkg/otel.SpanBuilder methods.
func (t *Target) OTELAttrs() []OTELAttribute {
	out := []OTELAttribute{
		{Key: otelAttrEntityType, Value: string(t.Type)},
		{Key: otelAttrEntityName, Value: t.Name},
	}
	if t.Primary != nil {
		if t.Primary.Database != "" {
			out = append(out, OTELAttribute{Key: otelAttrDBUniqueName, Value: t.Primary.Database})
		}
		if t.Primary.Host != "" {
			out = append(out, OTELAttribute{Key: otelAttrHost, Value: t.Primary.Host})
		}
	} else if t.SSH != nil && t.SSH.Host != "" {
		// Fall back to SSH host when no Primary endpoint (e.g. cluster, host targets)
		out = append(out, OTELAttribute{Key: otelAttrHost, Value: t.SSH.Host})
	}
	return out
}
