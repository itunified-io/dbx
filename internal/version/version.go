// Package version holds the build version, injected at compile time via ldflags.
package version

// Version is set by the linker at build time.
var Version = "dev"

// Commit is the git short commit hash, set by the linker.
var Commit = "unknown"

// Date is the build date, set by the linker.
var Date = "unknown"

// Edition distinguishes OSS from enterprise builds.
var Edition = "oss"

// Info returns a formatted version string.
func Info() string {
	return Version + " (" + Commit + ") " + Edition + " built " + Date
}
