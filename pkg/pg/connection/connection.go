// Package connection provides a multi-profile PostgreSQL connection registry.
package connection

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProfileConfig holds the connection parameters for a PostgreSQL profile.
type ProfileConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Database string `json:"database" yaml:"database"`
	SSLMode  string `json:"sslmode" yaml:"sslmode"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"-" yaml:"-"`
	PoolMin  int32  `json:"pool_min" yaml:"pool_min"`
	PoolMax  int32  `json:"pool_max" yaml:"pool_max"`
}

// DSN returns the PostgreSQL connection string for this config.
func (c ProfileConfig) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}
	port := c.Port
	if port == 0 {
		port = 5432
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, port, c.Database, sslMode,
	)
}

// Profile represents a named database connection profile.
type Profile struct {
	Name   string        `json:"name"`
	Config ProfileConfig `json:"config"`
	Pool   *pgxpool.Pool `json:"-"`
}

// ProfileInfo is a serializable representation of a profile (no pool, no password).
type ProfileInfo struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Database  string `json:"database"`
	SSLMode   string `json:"sslmode"`
	User      string `json:"user"`
	Connected bool   `json:"connected"`
}

// Registry manages multiple named PostgreSQL connection profiles.
type Registry struct {
	mu       sync.RWMutex
	profiles map[string]*Profile
	active   string
}

// NewRegistry creates an empty connection registry.
func NewRegistry() *Registry {
	return &Registry{
		profiles: make(map[string]*Profile),
	}
}

// Add registers a new profile. Returns an error if the name is already taken.
func (r *Registry) Add(name string, cfg ProfileConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}
	r.profiles[name] = &Profile{Name: name, Config: cfg}

	// Auto-activate the first profile added.
	if r.active == "" {
		r.active = name
	}
	return nil
}

// Connect opens a pgxpool connection for the named profile.
func (r *Registry) Connect(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.profiles[name]
	if !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	if p.Pool != nil {
		return nil // already connected
	}

	cfg, err := pgxpool.ParseConfig(p.Config.DSN())
	if err != nil {
		return fmt.Errorf("parse config for %q: %w", name, err)
	}
	if p.Config.PoolMin > 0 {
		cfg.MinConns = p.Config.PoolMin
	}
	if p.Config.PoolMax > 0 {
		cfg.MaxConns = p.Config.PoolMax
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("connect %q: %w", name, err)
	}
	p.Pool = pool
	return nil
}

// Disconnect closes the pool for the named profile.
func (r *Registry) Disconnect(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.profiles[name]
	if !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	if p.Pool != nil {
		p.Pool.Close()
		p.Pool = nil
	}
	return nil
}

// Get returns the profile with the given name, or an error if not found.
func (r *Registry) Get(name string) (*Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return p, nil
}

// Active returns the name of the currently active profile.
func (r *Registry) Active() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.active
}

// Switch sets the active profile. Returns an error if the profile does not exist.
func (r *Registry) Switch(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	r.active = name
	return nil
}

// List returns a sorted slice of ProfileInfo for all registered profiles.
func (r *Registry) List() []ProfileInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]ProfileInfo, 0, len(r.profiles))
	for _, p := range r.profiles {
		sslMode := p.Config.SSLMode
		if sslMode == "" {
			sslMode = "prefer"
		}
		port := p.Config.Port
		if port == 0 {
			port = 5432
		}
		infos = append(infos, ProfileInfo{
			Name:      p.Name,
			Host:      p.Config.Host,
			Port:      port,
			Database:  p.Config.Database,
			SSLMode:   sslMode,
			User:      p.Config.User,
			Connected: p.Pool != nil,
		})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos
}

// Remove deletes a profile, closing its pool if connected.
// Returns an error if the profile does not exist.
func (r *Registry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.profiles[name]
	if !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	if p.Pool != nil {
		p.Pool.Close()
	}
	delete(r.profiles, name)

	if r.active == name {
		r.active = ""
		// Auto-activate the first remaining profile (sorted for determinism).
		names := make([]string, 0, len(r.profiles))
		for n := range r.profiles {
			names = append(names, n)
		}
		sort.Strings(names)
		if len(names) > 0 {
			r.active = names[0]
		}
	}
	return nil
}

// ActivePool returns the pgxpool.Pool for the active profile, or an error
// if no profile is active or the active profile is not connected.
func (r *Registry) ActivePool() (*pgxpool.Pool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.active == "" {
		return nil, fmt.Errorf("no active profile")
	}
	p, ok := r.profiles[r.active]
	if !ok {
		return nil, fmt.Errorf("active profile %q not found", r.active)
	}
	if p.Pool == nil {
		return nil, fmt.Errorf("active profile %q is not connected", r.active)
	}
	return p.Pool, nil
}
