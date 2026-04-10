package target

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ErrTargetNotFound is returned when a target name is not in the registry.
var ErrTargetNotFound = errors.New("target not found")

// Registry holds all loaded targets indexed by name.
type Registry struct {
	targets map[string]*Target
}

// NewRegistry loads all *.yaml files from dir into a registry.
func NewRegistry(dir string) (*Registry, error) {
	r := &Registry{targets: make(map[string]*Target)}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return r, nil
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", e.Name(), err)
		}
		tgt, err := Parse(data)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", e.Name(), err)
		}
		r.targets[tgt.Name] = tgt
	}
	return r, nil
}

// Get returns a target by name or ErrTargetNotFound.
func (r *Registry) Get(name string) (*Target, error) {
	tgt, ok := r.targets[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTargetNotFound, name)
	}
	return tgt, nil
}

// Count returns the number of registered targets.
func (r *Registry) Count() int {
	return len(r.targets)
}

// List returns all targets sorted by name.
func (r *Registry) List() []*Target {
	out := make([]*Target, 0, len(r.targets))
	for _, t := range r.targets {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ListByType returns targets matching the given entity type.
func (r *Registry) ListByType(et EntityType) []*Target {
	var out []*Target
	for _, t := range r.targets {
		if t.Type == et {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Resolve looks up a target by name and validates entity_type.
func (r *Registry) Resolve(name string, entityType string) (*Target, error) {
	tgt, err := r.Get(name)
	if err != nil {
		return nil, err
	}
	if entityType != "" && EntityType(entityType) != tgt.Type {
		return nil, fmt.Errorf("entity_type mismatch: target %q is %s, not %s", name, tgt.Type, entityType)
	}
	return tgt, nil
}
