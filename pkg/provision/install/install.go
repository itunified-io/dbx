// Package install ships Oracle install primitives — runInstaller,
// root.sh, asmca, netca, oracleasm/afd disk labeling — invoked by
// /lab-up Phase D skills via dbxcli provision install <action>.
//
// All functions in this package require Enterprise license tier
// (license.RequireTier checked at the cobra layer, not here).
//
// # Idempotency patterns
//
// Two patterns are used across this package depending on whether the
// underlying Oracle command is itself idempotent. Tasks 4-8 add new
// primitives and MUST pick the correct pattern.
//
// ## Touchfile (single-file) — for IDEMPOTENT operations only
//
// Used by RootSh. Oracle's root.sh is documented as safe to re-run, so
// a single sentinel touchfile is written by the same shell pipeline as
// the script itself ("script && touch <file>"). Any failure mode short
// of "script ran cleanly" leaves the touchfile absent, and a re-run
// simply re-invokes the script.
//
// This pattern is INSUFFICIENT for non-idempotent commands: if the
// remote process completes the install but the local exec connection
// drops before "touch" runs, the next invocation has no way to
// distinguish "never ran" from "ran but wasn't recorded" — and
// re-invoking runInstaller / asmca / dbca on a half-installed home
// can corrupt the inventory.
//
// ## Two-phase sentinel (.partial → .installed) — for NON-IDEMPOTENT operations
//
// Required for runInstaller, asmca, dbca, and any other primitive
// where re-running on a partially-completed prior run is unsafe. The
// pattern is:
//
//  1. BEFORE invoking the install command, write <home>/.dbx/<op>.partial.
//  2. AFTER the command exits 0, atomically rename <op>.partial →
//     <op>.installed (mv, NOT a second touch — rename is atomic on a
//     single filesystem; touch is not).
//  3. Detection: <op>.installed present → DetectionStateInstalled;
//     <op>.partial present without <op>.installed → DetectionStatePartial
//     (operator must run the matching reverter before retrying);
//     neither present → DetectionStateAbsent.
//
// The .partial sentinel intentionally outlives the running process so
// that a hung or killed installer is observable on the next probe.
// Tasks 4-8 implementing runInstaller / asmca / dbca MUST follow this
// shape; do not reach for the touchfile pattern as a shortcut.
package install
