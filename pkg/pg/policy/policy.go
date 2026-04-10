// Package policy provides markdown policy file management.
package policy

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PolicyFile represents a loaded policy file.
type PolicyFile struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	SHA256     string    `json:"sha256"`
	Severity   string    `json:"severity"`
	LastLoaded time.Time `json:"last_loaded"`
}

// PolicyStore represents the collection of loaded policies.
type PolicyStore struct {
	Dir      string       `json:"dir"`
	Policies []PolicyFile `json:"policies"`
	LoadedAt time.Time    `json:"loaded_at"`
}

// Status scans a policy directory and returns all .md policy files.
func Status(policyDir string) (*PolicyStore, error) {
	store := &PolicyStore{Dir: policyDir, LoadedAt: time.Now()}
	entries, err := os.ReadDir(policyDir)
	if err != nil {
		return nil, fmt.Errorf("policy dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		path := filepath.Join(policyDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		hash := sha256.Sum256(data)
		store.Policies = append(store.Policies, PolicyFile{
			Name:       e.Name(),
			Path:       path,
			SHA256:     fmt.Sprintf("%x", hash),
			LastLoaded: time.Now(),
		})
	}
	return store, nil
}

// Reload re-scans the policy directory.
func Reload(policyDir string) (*PolicyStore, error) {
	return Status(policyDir)
}
