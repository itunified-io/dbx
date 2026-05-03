package target

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// safeName matches filesystem-safe target names: alphanumeric + - _ . only,
// must not start with '.', max length 128.
var safeNameRE = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)

// StoreDir returns the absolute path to the target registry directory
// (~/.dbx/targets). It does not create the directory.
func StoreDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fall back to $HOME directly; UserHomeDir only fails on truly
		// degenerate environments.
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".dbx", "targets")
}

// validateName ensures the name is safe to use as a filesystem component.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("target store: name must not be empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("target store: invalid name %q", name)
	}
	if !safeNameRE.MatchString(name) {
		return fmt.Errorf("target store: invalid name %q (allowed: alphanumeric, '-', '_', '.', max 128, must start alphanumeric or '_')", name)
	}
	return nil
}

// Save writes t to <StoreDir()>/<t.Name>.yaml with mode 0600. It creates
// the store directory if missing. Save overwrites any existing file with
// the same name (idempotent).
func Save(t *Target) error {
	if t == nil {
		return fmt.Errorf("target store: nil target")
	}
	if err := validateName(t.Name); err != nil {
		return err
	}
	dir := StoreDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("target store: mkdir %s: %w", dir, err)
	}
	data, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("target store: marshal %s: %w", t.Name, err)
	}
	path := filepath.Join(dir, t.Name+".yaml")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("target store: write %s: %w", path, err)
	}
	return nil
}

// Load reads <StoreDir()>/<name>.yaml and parses it into a Target.
// Returns a wrapped ErrTargetNotFound when the file does not exist.
func Load(name string) (*Target, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	path := filepath.Join(StoreDir(), name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("target store: %w: %s", ErrTargetNotFound, name)
		}
		return nil, fmt.Errorf("target store: read %s: %w", path, err)
	}
	t, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("target store: parse %s: %w", path, err)
	}
	return t, nil
}

// List returns all targets in StoreDir() sorted by name. Non-yaml files
// are skipped. A missing store directory returns an empty slice and nil
// error.
func List() ([]*Target, error) {
	dir := StoreDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("target store: readdir %s: %w", dir, err)
	}
	out := make([]*Target, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("target store: read %s: %w", e.Name(), err)
		}
		t, err := Parse(data)
		if err != nil {
			return nil, fmt.Errorf("target store: parse %s: %w", e.Name(), err)
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Remove deletes <StoreDir()>/<name>.yaml. Idempotent: returns nil when
// the file is already gone.
func Remove(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	path := filepath.Join(StoreDir(), name+".yaml")
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("target store: remove %s: %w", path, err)
	}
	return nil
}
