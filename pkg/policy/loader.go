package policy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFile loads a single policy YAML file.
func LoadFile(path string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("policy load: %w", err)
	}
	return parsePolicy(data, path)
}

// LoadDirectory loads all *.yaml files from a directory.
func LoadDirectory(dir string) ([]*Policy, error) {
	var policies []*Policy
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("policy load dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || (!strings.HasSuffix(e.Name(), ".yaml") && !strings.HasSuffix(e.Name(), ".yml")) {
			continue
		}
		p, err := LoadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, nil
}

func parsePolicy(data []byte, path string) (*Policy, error) {
	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("policy parse %s: %w", path, err)
	}
	p.SHA256 = fmt.Sprintf("%x", sha256.Sum256(data))
	p.Path = path
	return &p, nil
}

// PolicyRegistry holds all loaded policies, keyed by scope + framework.
type PolicyRegistry struct {
	policies map[string][]*Policy
}

// NewRegistry creates an empty registry.
func NewRegistry() *PolicyRegistry {
	return &PolicyRegistry{policies: make(map[string][]*Policy)}
}

// Register adds a policy to the registry.
func (r *PolicyRegistry) Register(p *Policy) {
	key := p.Metadata.Scope + ":" + p.Metadata.Framework
	r.policies[key] = append(r.policies[key], p)
}

// Get returns policies for a scope and optional framework filter.
func (r *PolicyRegistry) Get(scope, framework string) []*Policy {
	if framework != "" {
		return r.policies[scope+":"+framework]
	}
	var result []*Policy
	for k, v := range r.policies {
		if strings.HasPrefix(k, scope+":") {
			result = append(result, v...)
		}
	}
	return result
}

// All returns all registered policies.
func (r *PolicyRegistry) All() []*Policy {
	var result []*Policy
	for _, v := range r.policies {
		result = append(result, v...)
	}
	return result
}

// Reload clears the registry and reloads from directories.
func (r *PolicyRegistry) Reload(dirs ...string) error {
	r.policies = make(map[string][]*Policy)
	for _, dir := range dirs {
		policies, err := LoadDirectory(dir)
		if err != nil {
			return err
		}
		for _, p := range policies {
			r.Register(p)
		}
	}
	return nil
}
