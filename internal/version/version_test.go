package version

import (
	"testing"
)

func TestVersionDefaults(t *testing.T) {
	// When built without ldflags, defaults are set
	if Version == "" {
		t.Error("Version must not be empty")
	}
	if Edition != "oss" {
		t.Errorf("Default edition should be oss, got %s", Edition)
	}
}

func TestInfo(t *testing.T) {
	info := Info()
	if info == "" {
		t.Error("Info() must not be empty")
	}
}
