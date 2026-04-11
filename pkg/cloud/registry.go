package cloud

import (
	"fmt"
	"sync"
)

// ProviderRegistry manages registered cloud providers.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[ProviderID]CloudProvider
}

// NewProviderRegistry creates an empty registry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[ProviderID]CloudProvider),
	}
}

// Register adds a provider to the registry.
func (r *ProviderRegistry) Register(id ProviderID, p CloudProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[id] = p
}

// Get returns the provider for the given ID.
func (r *ProviderRegistry) Get(id ProviderID) (CloudProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[id]
	if !ok {
		return nil, fmt.Errorf("no provider registered for %s", id)
	}
	return p, nil
}

// List returns all registered provider IDs.
func (r *ProviderRegistry) List() []ProviderID {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]ProviderID, 0, len(r.providers))
	for id := range r.providers {
		ids = append(ids, id)
	}
	return ids
}
