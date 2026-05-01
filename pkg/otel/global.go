package otel

import "sync"

// globalExporter is the package-level Exporter used by emitters that
// don't carry an Exporter explicitly through their call graph (the 8
// install primitives in pkg/provision/install are the first consumers).
//
// Defaults to NoopExporter so callers always have a valid sink without
// any wiring; production code calls SetGlobalExporter from main() to
// install an OTLPExporter (or similar).
var (
	globalExporter Exporter = NoopExporter{}
	globalMu       sync.RWMutex
)

// GlobalExporter returns the currently registered Exporter. Always
// non-nil (NoopExporter when no explicit exporter is set).
func GlobalExporter() Exporter {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalExporter
}

// SetGlobalExporter installs e as the package-level Exporter. Passing
// nil resets to NoopExporter so callers can never end up with a nil
// exporter at Export() time.
func SetGlobalExporter(e Exporter) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if e == nil {
		e = NoopExporter{}
	}
	globalExporter = e
}
