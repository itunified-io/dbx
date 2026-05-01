package install

import (
	"context"

	"github.com/itunified-io/dbx/pkg/otel"
)

// emitSpan exports a single span via the package-level OTEL exporter.
// Pattern (used in every *WithExec primitive):
//
//	defer func() { emitSpan(ctx, sb, retErr) }()
//
// Best-effort: NoopExporter (the default) makes this a no-op when no
// OTEL endpoint is configured. Errors from Export are intentionally
// swallowed — the signed JSONL audit chain (ADR-0095) is the canonical
// record; OTEL is the observability mirror per ADR-0103a.
func emitSpan(ctx context.Context, sb *otel.SpanBuilder, retErr error) {
	if sb == nil {
		return
	}
	var span otel.Span
	if retErr != nil {
		span = sb.EndError(retErr)
	} else {
		span = sb.EndOK()
	}
	_ = otel.GlobalExporter().Export(ctx, []otel.Span{span})
}
